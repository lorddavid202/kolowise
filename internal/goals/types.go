package goals

import "time"

type CreateGoalRequest struct {
	Title        string `json:"title" binding:"required"`
	TargetAmount string `json:"target_amount" binding:"required"`
	TargetDate   string `json:"target_date" binding:"required"`
	Priority     int    `json:"priority"`
}

type ContributeGoalRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Note      string `json:"note"`
}

type GoalResponse struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	TargetAmountKobo  int64     `json:"target_amount_kobo"`
	TargetAmount      string    `json:"target_amount"`
	CurrentAmountKobo int64     `json:"current_amount_kobo"`
	CurrentAmount     string    `json:"current_amount"`
	RemainingKobo     int64     `json:"remaining_kobo"`
	RemainingAmount   string    `json:"remaining_amount"`
	ProgressPercent   float64   `json:"progress_percent"`
	TargetDate        time.Time `json:"target_date"`
	Priority          int       `json:"priority"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type ContributionResponse struct {
	ID         string    `json:"id"`
	GoalID     string    `json:"goal_id"`
	AccountID  string    `json:"account_id"`
	AmountKobo int64     `json:"amount_kobo"`
	Amount     string    `json:"amount"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}
