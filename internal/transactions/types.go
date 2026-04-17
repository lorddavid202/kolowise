package transactions

import "time"

type ManualTransactionRequest struct {
	AccountID    string `json:"account_id" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	Direction    string `json:"direction" binding:"required"`
	Narration    string `json:"narration"`
	MerchantName string `json:"merchant_name"`
	Category     string `json:"category"`
	TxnDate      string `json:"txn_date"`
}

type TransactionResponse struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	AmountKobo   int64     `json:"amount_kobo"`
	Amount       string    `json:"amount"`
	Direction    string    `json:"direction"`
	Narration    string    `json:"narration"`
	MerchantName string    `json:"merchant_name"`
	Category     string    `json:"category"`
	TxnDate      time.Time `json:"txn_date"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}

type CSVImportResponse struct {
	InsertedCount int    `json:"inserted_count"`
	Source        string `json:"source"`
}
