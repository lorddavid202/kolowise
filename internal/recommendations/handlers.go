package recommendations

import (
	"fmt"
	"math"
	"net/http"

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

	ruleRecommendedKobo := minInt64(balanceAfterBufferKobo/2, monthlySurplusKobo*40/100)

	if activeGoalNeedKobo > 0 && ruleRecommendedKobo > activeGoalNeedKobo {
		ruleRecommendedKobo = activeGoalNeedKobo
	}
	if ruleRecommendedKobo < 0 {
		ruleRecommendedKobo = 0
	}

	engine := "rule_based"
	modelName := "rule_based_v1"
	recommendedAmountKobo := ruleRecommendedKobo

	reason := fmt.Sprintf(
		"Fallback rule engine was used. Based on your current balance of NGN %s, average monthly income of NGN %s, average monthly expenses of NGN %s, and an emergency buffer of NGN %s, the system recommends saving NGN %s now.",
		utils.KoboToAmountString(availableBalanceKobo),
		utils.KoboToAmountString(avgMonthlyIncomeKobo),
		utils.KoboToAmountString(avgMonthlyExpenseKobo),
		utils.KoboToAmountString(emergencyBufferKobo),
		utils.KoboToAmountString(recommendedAmountKobo),
	)

	if h.ML != nil {
		mlResp, mlErr := h.ML.PredictSafeToSave(c.Request.Context(), mlclient.SafeToSaveRequest{
			AvailableBalance:  float64(availableBalanceKobo) / 100.0,
			AvgMonthlyIncome:  float64(avgMonthlyIncomeKobo) / 100.0,
			AvgMonthlyExpense: float64(avgMonthlyExpenseKobo) / 100.0,
			EmergencyBuffer:   float64(emergencyBufferKobo) / 100.0,
			MonthlySurplus:    float64(monthlySurplusKobo) / 100.0,
			ActiveGoalNeed:    float64(activeGoalNeedKobo) / 100.0,
		})

		if mlErr == nil && mlResp != nil {
			mlRecommendedKobo := int64(math.Round(mlResp.RecommendedAmount * 100))

			if mlRecommendedKobo < 0 {
				mlRecommendedKobo = 0
			}

			safeCapKobo := ruleRecommendedKobo
			if safeCapKobo < 0 {
				safeCapKobo = 0
			}

			if mlRecommendedKobo > safeCapKobo {
				mlRecommendedKobo = safeCapKobo
			}

			recommendedAmountKobo = mlRecommendedKobo
			engine = "ml"
			modelName = mlResp.ModelName

			reason = fmt.Sprintf(
				"ML model %s was used first. The recommendation was safety-capped by your rule-based guardrails, resulting in NGN %s.",
				modelName,
				utils.KoboToAmountString(recommendedAmountKobo),
			)
		}
	}

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
		modelName,
		reason,
	)

	resp := SafeToSaveResponse{
		Engine:                engine,
		ModelName:             modelName,
		RuleVersion:           "rule_guardrail_v1",
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
