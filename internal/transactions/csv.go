package transactions

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emekachisom/kolowise/pkg/utils"
)

type ParsedCSVTransaction struct {
	AmountKobo   int64
	Direction    string
	Narration    string
	MerchantName string
	Category     string
	TxnDate      time.Time
}

func ParseCSV(file io.Reader) ([]ParsedCSVTransaction, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("csv must contain header and at least one row")
	}

	headerMap := make(map[string]int)
	for i, h := range rows[0] {
		headerMap[normalizeHeader(h)] = i
	}

	requiredDateIdx, ok := findHeader(headerMap, "txn_date", "date")
	if !ok {
		return nil, fmt.Errorf("csv must include txn_date or date column")
	}

	amountIdx, hasAmount := findHeader(headerMap, "amount")
	directionIdx, hasDirection := findHeader(headerMap, "direction")
	debitIdx, hasDebit := findHeader(headerMap, "debit")
	creditIdx, hasCredit := findHeader(headerMap, "credit")

	if !(hasAmount && hasDirection) && !(hasDebit || hasCredit) {
		return nil, fmt.Errorf("csv must have either amount+direction columns or debit/credit columns")
	}

	narrationIdx, _ := findHeader(headerMap, "narration", "description")
	merchantIdx, _ := findHeader(headerMap, "merchant_name", "merchant")
	categoryIdx, _ := findHeader(headerMap, "category")

	results := make([]ParsedCSVTransaction, 0, len(rows)-1)

	for i, row := range rows[1:] {
		lineNo := i + 2

		txnDate, err := parseDate(row[requiredDateIdx])
		if err != nil {
			return nil, fmt.Errorf("row %d invalid date: %w", lineNo, err)
		}

		var amountKobo int64
		var direction string

		if hasAmount && hasDirection {
			amountKobo, err = utils.AmountStringToKobo(row[amountIdx])
			if err != nil {
				return nil, fmt.Errorf("row %d invalid amount: %w", lineNo, err)
			}
			direction = strings.ToLower(strings.TrimSpace(row[directionIdx]))
			if direction != "credit" && direction != "debit" {
				return nil, fmt.Errorf("row %d direction must be credit or debit", lineNo)
			}
		} else {
			debitVal := ""
			creditVal := ""

			if hasDebit && debitIdx < len(row) {
				debitVal = strings.TrimSpace(row[debitIdx])
			}
			if hasCredit && creditIdx < len(row) {
				creditVal = strings.TrimSpace(row[creditIdx])
			}

			switch {
			case debitVal != "":
				amountKobo, err = utils.AmountStringToKobo(debitVal)
				if err != nil {
					return nil, fmt.Errorf("row %d invalid debit amount: %w", lineNo, err)
				}
				direction = "debit"
			case creditVal != "":
				amountKobo, err = utils.AmountStringToKobo(creditVal)
				if err != nil {
					return nil, fmt.Errorf("row %d invalid credit amount: %w", lineNo, err)
				}
				direction = "credit"
			default:
				return nil, fmt.Errorf("row %d must have either debit or credit value", lineNo)
			}
		}

		get := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		results = append(results, ParsedCSVTransaction{
			AmountKobo:   amountKobo,
			Direction:    direction,
			Narration:    get(narrationIdx),
			MerchantName: get(merchantIdx),
			Category:     get(categoryIdx),
			TxnDate:      txnDate,
		})
	}

	return results, nil
}

func normalizeHeader(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func findHeader(headerMap map[string]int, keys ...string) (int, bool) {
	for _, k := range keys {
		if idx, ok := headerMap[k]; ok {
			return idx, true
		}
	}
	return -1, false
}

func parseDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)

	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02/01/2006",
		"02/01/2006 15:04",
		"01/02/2006",
		"01/02/2006 15:04",
	}

	for _, f := range formats {
		t, err := time.Parse(f, val)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %s", val)
}
