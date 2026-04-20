package goals

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

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
	var req CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	targetAmountKobo, err := utils.AmountStringToKobo(req.TargetAmount)
	if err != nil || targetAmountKobo <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_amount"})
		return
	}

	targetDate, err := parseGoalDate(req.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_date"})
		return
	}

	priority := req.Priority
	if priority <= 0 {
		priority = 1
	}

	query := `
		INSERT INTO savings_goals (
			user_id, title, target_amount_kobo, current_amount_kobo, target_date, priority, status
		)
		VALUES ($1, $2, $3, 0, $4, $5, 'active')
		RETURNING id, title, target_amount_kobo, current_amount_kobo, target_date, priority, status, created_at
	`

	var goal GoalResponse
	err = h.DB.QueryRow(
		c.Request.Context(),
		query,
		userID,
		req.Title,
		targetAmountKobo,
		targetDate,
		priority,
	).Scan(
		&goal.ID,
		&goal.Title,
		&goal.TargetAmountKobo,
		&goal.CurrentAmountKobo,
		&goal.TargetDate,
		&goal.Priority,
		&goal.Status,
		&goal.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal"})
		return
	}

	enrichGoal(&goal)
	c.JSON(http.StatusCreated, gin.H{"goal": goal})
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
		SELECT id, title, target_amount_kobo, current_amount_kobo, target_date, priority, status, created_at
		FROM savings_goals
		WHERE user_id = $1
		ORDER BY
			CASE status WHEN 'active' THEN 1 WHEN 'completed' THEN 2 ELSE 3 END,
			priority ASC,
			created_at DESC
		`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list goals"})
		return
	}
	defer rows.Close()

	goals := make([]GoalResponse, 0)
	for rows.Next() {
		var g GoalResponse
		if err := rows.Scan(
			&g.ID,
			&g.Title,
			&g.TargetAmountKobo,
			&g.CurrentAmountKobo,
			&g.TargetDate,
			&g.Priority,
			&g.Status,
			&g.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan goal"})
			return
		}

		enrichGoal(&g)
		goals = append(goals, g)
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (h *Handler) Contribute(c *gin.Context) {
	var req ContributeGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	goalID := c.Param("id")
	if strings.TrimSpace(goalID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal id is required"})
		return
	}

	amountKobo, err := utils.AmountStringToKobo(req.Amount)
	if err != nil || amountKobo <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	dbTx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer dbTx.Rollback(c.Request.Context())

	var title string
	var currentAmountKobo int64
	var targetAmountKobo int64
	var status string

	err = dbTx.QueryRow(
		c.Request.Context(),
		`
		SELECT title, current_amount_kobo, target_amount_kobo, status
		FROM savings_goals
		WHERE id = $1 AND user_id = $2
		LIMIT 1
		`,
		goalID,
		userID,
	).Scan(&title, &currentAmountKobo, &targetAmountKobo, &status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal not found"})
		return
	}

	if status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal is not active"})
		return
	}

	var currentBalanceKobo int64
	err = dbTx.QueryRow(
		c.Request.Context(),
		`
		SELECT current_balance_kobo
		FROM accounts
		WHERE id = $1 AND user_id = $2
		LIMIT 1
		`,
		req.AccountID,
		userID,
	).Scan(&currentBalanceKobo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not found"})
		return
	}

	if currentBalanceKobo < amountKobo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient account balance"})
		return
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`
		INSERT INTO goal_contributions (goal_id, user_id, account_id, amount_kobo, note)
		VALUES ($1, $2, $3, $4, $5)
		`,
		goalID,
		userID,
		req.AccountID,
		amountKobo,
		strings.TrimSpace(req.Note),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record contribution"})
		return
	}

	newCurrentAmountKobo := currentAmountKobo + amountKobo
	newStatus := "active"
	if newCurrentAmountKobo >= targetAmountKobo {
		newStatus = "completed"
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`
		UPDATE savings_goals
		SET current_amount_kobo = $1, status = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		`,
		newCurrentAmountKobo,
		newStatus,
		goalID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update goal"})
		return
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`
		INSERT INTO transactions (
			user_id, account_id, amount_kobo, direction, narration, merchant_name, category, txn_date, source
		)
		VALUES ($1, $2, $3, 'debit', $4, '', 'goal_savings', NOW(), 'goal_contribution')
		`,
		userID,
		req.AccountID,
		amountKobo,
		fmt.Sprintf("Goal contribution: %s", title),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction record"})
		return
	}

	_, err = dbTx.Exec(
		c.Request.Context(),
		`
		UPDATE accounts
		SET current_balance_kobo = current_balance_kobo - $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		`,
		amountKobo,
		req.AccountID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account balance"})
		return
	}

	if err := dbTx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit contribution"})
		return
	}

	var goal GoalResponse
	err = h.DB.QueryRow(
		c.Request.Context(),
		`
		SELECT id, title, target_amount_kobo, current_amount_kobo, target_date, priority, status, created_at
		FROM savings_goals
		WHERE id = $1 AND user_id = $2
		LIMIT 1
		`,
		goalID,
		userID,
	).Scan(
		&goal.ID,
		&goal.Title,
		&goal.TargetAmountKobo,
		&goal.CurrentAmountKobo,
		&goal.TargetDate,
		&goal.Priority,
		&goal.Status,
		&goal.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated goal"})
		return
	}

	enrichGoal(&goal)
	c.JSON(http.StatusOK, gin.H{"goal": goal})
}

func enrichGoal(g *GoalResponse) {
	g.TargetAmount = utils.KoboToAmountString(g.TargetAmountKobo)
	g.CurrentAmount = utils.KoboToAmountString(g.CurrentAmountKobo)

	remaining := g.TargetAmountKobo - g.CurrentAmountKobo
	if remaining < 0 {
		remaining = 0
	}
	g.RemainingKobo = remaining
	g.RemainingAmount = utils.KoboToAmountString(remaining)

	if g.TargetAmountKobo > 0 {
		progress := (float64(g.CurrentAmountKobo) / float64(g.TargetAmountKobo)) * 100
		g.ProgressPercent = math.Min(progress, 100)
	}
}

func parseGoalDate(input string) (time.Time, error) {
	input = strings.TrimSpace(input)

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	for _, f := range formats {
		t, err := time.Parse(f, input)
		if err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid target date")
}
