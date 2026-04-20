package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CategoryPredictionRequest struct {
	Narration    string  `json:"narration"`
	MerchantName string  `json:"merchant_name"`
	Amount       float64 `json:"amount"`
	Direction    string  `json:"direction"`
}

type CategoryPredictionResponse struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	ModelName  string  `json:"model_name"`
}

type SafeToSaveRequest struct {
	AvailableBalance  float64 `json:"available_balance"`
	AvgMonthlyIncome  float64 `json:"avg_monthly_income"`
	AvgMonthlyExpense float64 `json:"avg_monthly_expense"`
	EmergencyBuffer   float64 `json:"emergency_buffer"`
	MonthlySurplus    float64 `json:"monthly_surplus"`
	ActiveGoalNeed    float64 `json:"active_goal_need"`
}

type SafeToSaveResponse struct {
	RecommendedAmount float64 `json:"recommended_amount"`
	ModelName         string  `json:"model_name"`
}

func (c *Client) PredictCategory(ctx context.Context, req CategoryPredictionRequest) (*CategoryPredictionResponse, error) {
	url := c.BaseURL + "/predict/category"
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ml service category prediction failed with status %d", resp.StatusCode)
	}

	var out CategoryPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) PredictSafeToSave(ctx context.Context, req SafeToSaveRequest) (*SafeToSaveResponse, error) {
	url := c.BaseURL + "/predict/safe-to-save"
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ml service safe-to-save prediction failed with status %d", resp.StatusCode)
	}

	var out SafeToSaveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
