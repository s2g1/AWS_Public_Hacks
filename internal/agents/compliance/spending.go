package compliance

import (
	"time"

	"federal-payment-processing/internal/models"
)

// SpendingThresholds defines the spending limits for a spend category.
type SpendingThresholds struct {
	// SingleTransactionMax is the maximum amount allowed for a single transaction.
	SingleTransactionMax float64
	// AnnualMax is the maximum cumulative spend allowed per payee in a fiscal year.
	AnnualMax float64
}

// SpendingStore is an interface for retrieving cumulative spend data,
// enabling mocking in tests.
type SpendingStore interface {
	// GetCumulativeSpend returns the total amount spent for the given payee
	// in the specified fiscal year.
	GetCumulativeSpend(payee string, fiscalYear int) (float64, error)
}

// currentFiscalYear returns the current US federal fiscal year.
// The federal fiscal year starts on October 1 of the previous calendar year.
func currentFiscalYear() int {
	now := time.Now()
	if now.Month() >= time.October {
		return now.Year() + 1
	}
	return now.Year()
}

// checkSpendingThresholds evaluates a payment amount against spending thresholds
// and returns any compliance flags that should be added.
func checkSpendingThresholds(payeeName string, amount float64, thresholds SpendingThresholds, store SpendingStore) []models.ComplianceFlag {
	var flags []models.ComplianceFlag

	// Check single transaction maximum
	if amount > thresholds.SingleTransactionMax {
		flags = append(flags, models.ComplianceFlag{
			Rule:     "THRESHOLD_EXCEEDED",
			Severity: models.FlagSeverityRequiresReview,
			Message:  "Amount exceeds single transaction threshold",
		})
	}

	// Check cumulative annual spend
	if store != nil {
		fiscalYear := currentFiscalYear()
		cumulativeSpend, err := store.GetCumulativeSpend(payeeName, fiscalYear)
		if err == nil {
			if cumulativeSpend+amount > thresholds.AnnualMax {
				flags = append(flags, models.ComplianceFlag{
					Rule:     "ANNUAL_LIMIT",
					Severity: models.FlagSeverityRequiresReview,
					Message:  "Cumulative spend would exceed annual limit",
				})
			}
		}
	}

	return flags
}
