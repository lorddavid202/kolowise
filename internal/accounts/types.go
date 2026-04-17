package accounts

import "time"

type CreateAccountRequest struct {
	AccountName     string `json:"account_name" binding:"required"`
	InstitutionName string `json:"institution_name"`
	AccountType     string `json:"account_type" binding:"required"`
	OpeningBalance  string `json:"opening_balance"`
	Currency        string `json:"currency"`
}

type AccountResponse struct {
	ID                 string    `json:"id"`
	AccountName        string    `json:"account_name"`
	InstitutionName    string    `json:"institution_name"`
	AccountType        string    `json:"account_type"`
	CurrentBalanceKobo int64     `json:"current_balance_kobo"`
	CurrentBalance     string    `json:"current_balance"`
	Currency           string    `json:"currency"`
	CreatedAt          time.Time `json:"created_at"`
}
