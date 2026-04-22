package recommendations

type SafeToSaveResponse struct {
	Engine                string `json:"engine"`
	ModelName             string `json:"model_name"`
	RuleVersion           string `json:"rule_version"`
	RecommendedAmountKobo int64  `json:"recommended_amount_kobo"`
	RecommendedAmount     string `json:"recommended_amount"`
	AvailableBalanceKobo  int64  `json:"available_balance_kobo"`
	AvailableBalance      string `json:"available_balance"`
	AvgMonthlyIncomeKobo  int64  `json:"avg_monthly_income_kobo"`
	AvgMonthlyIncome      string `json:"avg_monthly_income"`
	AvgMonthlyExpenseKobo int64  `json:"avg_monthly_expense_kobo"`
	AvgMonthlyExpense     string `json:"avg_monthly_expense"`
	EmergencyBufferKobo   int64  `json:"emergency_buffer_kobo"`
	EmergencyBuffer       string `json:"emergency_buffer"`
	MonthlySurplusKobo    int64  `json:"monthly_surplus_kobo"`
	MonthlySurplus        string `json:"monthly_surplus"`
	ActiveGoalNeedKobo    int64  `json:"active_goal_need_kobo"`
	ActiveGoalNeed        string `json:"active_goal_need"`
	Reason                string `json:"reason"`
}
