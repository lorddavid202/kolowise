package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func AmountStringToKobo(input string) (int64, error) {
	cleaned := strings.TrimSpace(input)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "₦", "")
	cleaned = strings.ReplaceAll(strings.ToUpper(cleaned), "NGN", "")
	cleaned = strings.TrimSpace(cleaned)

	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %w", err)
	}

	return int64(math.Round(val * 100)), nil
}

func KoboToAmountString(kobo int64) string {
	return fmt.Sprintf("%.2f", float64(kobo)/100)
}
