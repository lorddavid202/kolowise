package recommendations

import (
	"fmt"
	"net/http"

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

func (h *Handler) SafeToSave(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var availableBalanceKobo int64
	err := h.DB.QueryRow(
		c.Request.Context(),
		`
		SELECT COALESCE(SUM(current_balance_kobo), 0)
		FROM accounts
		WHERE user_id = $1
		`,
		userID,
	).Scan(&availableBalanceKobo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute available balance"})
		return
	}

	var totalCredits90d int64
	var totalDebits90d int64
	err = h.DB.QueryRow(
		c.Request.Context(),
		`
		SELECT
			COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount_kobo ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN direction = 'debit' THEN amount_kobo ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1
		  AND txn_date >= NOW() - INTERVAL '90 days'
		`,
		userID,
	).Scan(&totalCredits90d, &totalDebits90d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute transaction history"})
		return
	}

	var activeGoalNeedKobo int64
	err = h.DB.QueryRow(
		c.Request.Context(),
		`
		SELECT COALESCE(SUM(GREATEST(target_amount_kobo - current_amount_kobo, 0)), 0)
		FROM savings_goals
		WHERE user_id = $1
		  AND status = 'active'
		`,
		userID,
	).Scan(&activeGoalNeedKobo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute goal need"})
		return
	}

	avgMonthlyIncomeKobo := totalCredits90d / 3
	avgMonthlyExpenseKobo := totalDebits90d / 3

	emergencyBufferKobo := avgMonthlyExpenseKobo * 20 / 100
	minimumBufferKobo := int64(1000000) // ₦10,000
	if emergencyBufferKobo < minimumBufferKobo {
		emergencyBufferKobo = minimumBufferKobo
	}

	monthlySurplusKobo := avgMonthlyIncomeKobo - avgMonthlyExpenseKobo
	if monthlySurplusKobo < 0 {
		monthlySurplusKobo = 0
	}

	balanceAfterBufferKobo := availableBalanceKobo - emergencyBufferKobo
	if balanceAfterBufferKobo < 0 {
		balanceAfterBufferKobo = 0
	}

	recommendationByBalance := balanceAfterBufferKobo / 2
	recommendationBySurplus := monthlySurplusKobo * 40 / 100

	recommendedAmountKobo := minInt64(recommendationByBalance, recommendationBySurplus)

	if activeGoalNeedKobo > 0 && recommendedAmountKobo > activeGoalNeedKobo {
		recommendedAmountKobo = activeGoalNeedKobo
	}

	if recommendedAmountKobo < 0 {
		recommendedAmountKobo = 0
	}

	ruleVersion := "rule_based_v1"

	reason := fmt.Sprintf(
		"Based on your current balance of NGN %s, average monthly income of NGN %s, average monthly expenses of NGN %s, and an emergency buffer of NGN %s, the system recommends saving NGN %s now.",
		utils.KoboToAmountString(availableBalanceKobo),
		utils.KoboToAmountString(avgMonthlyIncomeKobo),
		utils.KoboToAmountString(avgMonthlyExpenseKobo),
		utils.KoboToAmountString(emergencyBufferKobo),
		utils.KoboToAmountString(recommendedAmountKobo),
	)

	_, _ = h.DB.Exec(
		c.Request.Context(),
		`
		INSERT INTO savings_recommendations (
			user_id,
			recommended_amount_kobo,
			available_balance_kobo,
			avg_monthly_income_kobo,
			avg_monthly_expense_kobo,
			emergency_buffer_kobo,
			monthly_surplus_kobo,
			active_goal_need_kobo,
			rule_version,
			reason
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`,
		userID,
		recommendedAmountKobo,
		availableBalanceKobo,
		avgMonthlyIncomeKobo,
		avgMonthlyExpenseKobo,
		emergencyBufferKobo,
		monthlySurplusKobo,
		activeGoalNeedKobo,
		ruleVersion,
		reason,
	)

	resp := SafeToSaveResponse{
		RuleVersion:           ruleVersion,
		RecommendedAmountKobo: recommendedAmountKobo,
		RecommendedAmount:     utils.KoboToAmountString(recommendedAmountKobo),
		AvailableBalanceKobo:  availableBalanceKobo,
		AvailableBalance:      utils.KoboToAmountString(availableBalanceKobo),
		AvgMonthlyIncomeKobo:  avgMonthlyIncomeKobo,
		AvgMonthlyIncome:      utils.KoboToAmountString(avgMonthlyIncomeKobo),
		AvgMonthlyExpenseKobo: avgMonthlyExpenseKobo,
		AvgMonthlyExpense:     utils.KoboToAmountString(avgMonthlyExpenseKobo),
		EmergencyBufferKobo:   emergencyBufferKobo,
		EmergencyBuffer:       utils.KoboToAmountString(emergencyBufferKobo),
		MonthlySurplusKobo:    monthlySurplusKobo,
		MonthlySurplus:        utils.KoboToAmountString(monthlySurplusKobo),
		ActiveGoalNeedKobo:    activeGoalNeedKobo,
		ActiveGoalNeed:        utils.KoboToAmountString(activeGoalNeedKobo),
		Reason:                reason,
	}

	c.JSON(http.StatusOK, gin.H{"safe_to_save": resp})
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
