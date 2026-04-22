package transactions

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emekachisom/kolowise/internal/mlclient"
	"github.com/emekachisom/kolowise/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
	ML *mlclient.Client
}

func NewHandler(db *pgxpool.Pool, ml *mlclient.Client) *Handler {
	return &Handler{
		DB: db,
		ML: ml,
	}
}

func (h *Handler) CreateManual(c *gin.Context) {
	var req ManualTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req.Direction = strings.ToLower(strings.TrimSpace(req.Direction))
	if req.Direction != "credit" && req.Direction != "debit" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be credit or debit"})
		return
	}

	amountKobo, err := utils.AmountStringToKobo(req.Amount)
	if err != nil || amountKobo <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	txnDate := time.Now().UTC()
	if strings.TrimSpace(req.TxnDate) != "" {
		parsed, err := parseDate(req.TxnDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid txn_date"})
			return
		}
		txnDate = parsed.UTC()
	}

	var exists bool
	err = h.DB.QueryRow(
		c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2)`,
		req.AccountID,
		userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not found"})
		return
	}

	resolvedCategory := h.resolveCategory(
		c.Request.Context(),
		req.Category,
		req.Narration,
		req.MerchantName,
		req.Direction,
		amountKobo,
	)

	dbTx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer dbTx.Rollback(c.Request.Context())

	query := `
		INSERT INTO transactions (
			user_id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'manual')
		RETURNING id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source, created_at
	`

	var res TransactionResponse
	err = dbTx.QueryRow(
		c.Request.Context(),
		query,
		userID,
		req.AccountID,
		amountKobo,
		req.Direction,
		strings.TrimSpace(req.Narration),
		strings.TrimSpace(req.MerchantName),
		resolvedCategory,
		txnDate,
	).Scan(
		&res.ID,
		&res.AccountID,
		&res.AmountKobo,
		&res.Direction,
		&res.Narration,
		&res.MerchantName,
		&res.Category,
		&res.TxnDate,
		&res.Source,
		&res.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	delta := amountKobo
	if req.Direction == "debit" {
		delta = -amountKobo
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`UPDATE accounts SET current_balance_kobo = current_balance_kobo + $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`,
		delta,
		req.AccountID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account balance"})
		return
	}

	if err := dbTx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	res.Amount = utils.KoboToAmountString(res.AmountKobo)
	c.JSON(http.StatusCreated, gin.H{"transaction": res})
}

func (h *Handler) UploadCSV(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accountID := strings.TrimSpace(c.PostForm("account_id"))
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}

	var exists bool
	err := h.DB.QueryRow(
		c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2)`,
		accountID,
		userID,
	).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not found"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	parsedRows, err := ParseCSV(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbTx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer dbTx.Rollback(c.Request.Context())

	var totalDelta int64
	for _, row := range parsedRows {
		resolvedCategory := h.resolveCategory(
			c.Request.Context(),
			row.Category,
			row.Narration,
			row.MerchantName,
			row.Direction,
			row.AmountKobo,
		)

		_, err := dbTx.Exec(
			c.Request.Context(),
			`
			INSERT INTO transactions (
				user_id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'csv')
			`,
			userID,
			accountID,
			row.AmountKobo,
			row.Direction,
			row.Narration,
			row.MerchantName,
			resolvedCategory,
			row.TxnDate.UTC(),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert csv transaction"})
			return
		}

		if row.Direction == "credit" {
			totalDelta += row.AmountKobo
		} else {
			totalDelta -= row.AmountKobo
		}
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`UPDATE accounts SET current_balance_kobo = current_balance_kobo + $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`,
		totalDelta,
		accountID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account balance"})
		return
	}

	if err := dbTx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit csv import"})
		return
	}

	c.JSON(http.StatusCreated, CSVImportResponse{
		InsertedCount: len(parsedRows),
		Source:        "csv",
	})
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accountID := strings.TrimSpace(c.Query("account_id"))

	limit := 20
	offset := 0

	if val := strings.TrimSpace(c.Query("limit")); val != "" {
		parsed, err := strconv.Atoi(val)
		if err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if val := strings.TrimSpace(c.Query("offset")); val != "" {
		parsed, err := strconv.Atoi(val)
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var rows pgxRows
	var err error

	if accountID != "" {
		rows, err = h.DB.Query(
			c.Request.Context(),
			`
			SELECT id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source, created_at
			FROM transactions
			WHERE user_id = $1 AND account_id = $2
			ORDER BY txn_date DESC, created_at DESC
			LIMIT $3 OFFSET $4
			`,
			userID,
			accountID,
			limit,
			offset,
		)
	} else {
		rows, err = h.DB.Query(
			c.Request.Context(),
			`
			SELECT id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source, created_at
			FROM transactions
			WHERE user_id = $1
			ORDER BY txn_date DESC, created_at DESC
			LIMIT $2 OFFSET $3
			`,
			userID,
			limit,
			offset,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}
	defer rows.Close()

	results := make([]TransactionResponse, 0)
	for rows.Next() {
		var t TransactionResponse
		if err := rows.Scan(
			&t.ID,
			&t.AccountID,
			&t.AmountKobo,
			&t.Direction,
			&t.Narration,
			&t.MerchantName,
			&t.Category,
			&t.TxnDate,
			&t.Source,
			&t.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan transaction"})
			return
		}

		t.Amount = utils.KoboToAmountString(t.AmountKobo)
		results = append(results, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": results,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *Handler) resolveCategory(
	ctx context.Context,
	currentCategory string,
	narration string,
	merchantName string,
	direction string,
	amountKobo int64,
) string {
	currentCategory = strings.TrimSpace(strings.ToLower(currentCategory))
	if currentCategory != "" {
		return currentCategory
	}

	if h.ML == nil {
		return "uncategorized"
	}

	amountFloat := float64(amountKobo) / 100.0

	resp, err := h.ML.PredictCategory(ctx, mlclient.CategoryPredictionRequest{
		Narration:    narration,
		MerchantName: merchantName,
		Amount:       amountFloat,
		Direction:    direction,
	})
	if err != nil || resp == nil || strings.TrimSpace(resp.Category) == "" {
		return "uncategorized"
	}

	return strings.ToLower(strings.TrimSpace(resp.Category))
}

type pgxRows interface {
	Close()
	Next() bool
	Scan(dest ...interface{}) error
}
