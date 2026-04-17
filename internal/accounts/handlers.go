package accounts

import (
	"net/http"
	"strings"

	"github.com/emekachisom/kolowise/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req.AccountName = strings.TrimSpace(req.AccountName)
	req.InstitutionName = strings.TrimSpace(req.InstitutionName)
	req.AccountType = strings.TrimSpace(strings.ToLower(req.AccountType))

	currency := strings.TrimSpace(strings.ToUpper(req.Currency))
	if currency == "" {
		currency = "NGN"
	}

	var openingBalanceKobo int64
	var err error
	if strings.TrimSpace(req.OpeningBalance) != "" {
		openingBalanceKobo, err = utils.AmountStringToKobo(req.OpeningBalance)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid opening_balance"})
			return
		}
	}

	query := `
		INSERT INTO accounts (
			user_id, account_name, institution_name, account_type, current_balance_kobo, currency
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, account_name, institution_name, account_type, current_balance_kobo, currency, created_at
	`

	var account AccountResponse
	err = h.DB.QueryRow(
		c.Request.Context(),
		query,
		userID,
		req.AccountName,
		req.InstitutionName,
		req.AccountType,
		openingBalanceKobo,
		currency,
	).Scan(
		&account.ID,
		&account.AccountName,
		&account.InstitutionName,
		&account.AccountType,
		&account.CurrentBalanceKobo,
		&account.Currency,
		&account.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	account.CurrentBalance = utils.KoboToAmountString(account.CurrentBalanceKobo)
	c.JSON(http.StatusCreated, gin.H{"account": account})
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rows, err := h.DB.Query(
		c.Request.Context(),
		`
		SELECT id, account_name, institution_name, account_type, current_balance_kobo, currency, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list accounts"})
		return
	}
	defer rows.Close()

	accounts := make([]AccountResponse, 0)
	for rows.Next() {
		var a AccountResponse
		if err := rows.Scan(
			&a.ID,
			&a.AccountName,
			&a.InstitutionName,
			&a.AccountType,
			&a.CurrentBalanceKobo,
			&a.Currency,
			&a.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan account"})
			return
		}

		a.CurrentBalance = utils.KoboToAmountString(a.CurrentBalanceKobo)
		accounts = append(accounts, a)
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}
