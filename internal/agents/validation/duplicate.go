package validation

import (
	"fmt"

	"federal-payment-processing/internal/models"
)

// DuplicateMatch represents a previously processed payment that matches
// the current payment's payee, amount, and date within the lookback window.
type DuplicateMatch struct {
	PaymentID string `json:"paymentId"`
	Date      string `json:"date"`
}

// PaymentStore defines the interface for querying existing payments
// to detect potential duplicates.
type PaymentStore interface {
	// QueryDuplicates searches for payments matching the given payee, amount, and date
	// within the specified lookback window (in days).
	QueryDuplicates(payee string, amount string, date string, lookbackDays int) ([]DuplicateMatch, error)
}

// ValidatePaymentWithDuplicateCheck performs all standard validation checks
// (via ValidatePayment) AND queries the PaymentStore for potential duplicate payments.
// If duplicates are found, a WARNING severity issue is added referencing the matching payment ID.
// Since duplicates produce only WARNINGs, they do not affect the overall validation status
// (won't auto-reject) — they route to human review via the WARNING signal.
func ValidatePaymentWithDuplicateCheck(extraction models.ExtractionResult, store PaymentStore) models.ValidationResult {
	// Run existing validation first
	result := ValidatePayment(extraction)

	// Perform duplicate check if payee, amount, and date fields are present
	payeeField, hasPayee := extraction.Fields["payee"]
	amountField, hasAmount := extraction.Fields["amount"]
	dateField, hasDate := extraction.Fields["date"]

	if !hasPayee || !hasAmount || !hasDate {
		return result
	}

	payee := payeeField.Normalized
	if payee == "" {
		payee = payeeField.Value
	}

	amount := amountField.Normalized
	if amount == "" {
		amount = amountField.Value
	}

	date := dateField.Normalized
	if date == "" {
		date = dateField.Value
	}

	// Query for duplicates within 30-day lookback window
	matches, err := store.QueryDuplicates(payee, amount, date, 30)
	if err != nil {
		// If the store query fails, we don't block validation — just skip duplicate check
		return result
	}

	// Add a WARNING for each duplicate match found
	for _, match := range matches {
		result.Issues = append(result.Issues, models.ValidationIssue{
			Severity: models.SeverityWarning,
			Field:    "duplicate",
			Message:  fmt.Sprintf("Potential duplicate: %s", match.PaymentID),
		})
	}

	return result
}
