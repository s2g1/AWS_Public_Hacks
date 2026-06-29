package validation

import (
	"federal-payment-processing/internal/models"
)

// PayeeRegistry defines the interface for looking up payees in the registered
// vendor/payee registry (backed by DynamoDB in production).
type PayeeRegistry interface {
	// Lookup checks if the given payee name exists in the registry.
	// Returns true if found, false if not found.
	Lookup(payeeName string) (bool, error)
}

// ValidatePaymentFull performs comprehensive validation combining:
//   - All standard field validation from ValidatePayment (completeness, format, dates)
//   - Duplicate detection via PaymentStore (from task 4.3)
//   - Payee registry cross-reference
//
// If the payee is not found in the registry, a WARNING severity issue is added
// with field "payee" and message "Payee not in registry". Since it's a WARNING,
// it does not affect the overall validation status (status stays VALID if no
// CRITICAL or ERROR issues exist).
func ValidatePaymentFull(extraction models.ExtractionResult, store PaymentStore, registry PayeeRegistry) models.ValidationResult {
	// Run validation with duplicate check (which includes base validation)
	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// Perform payee registry cross-reference
	payeeField, hasPayee := extraction.Fields["payee"]
	if !hasPayee {
		return result
	}

	payeeName := payeeField.Normalized
	if payeeName == "" {
		payeeName = payeeField.Value
	}

	if payeeName == "" {
		return result
	}

	found, err := registry.Lookup(payeeName)
	if err != nil {
		// If registry lookup fails, skip the check (don't block validation)
		return result
	}

	if !found {
		result.Issues = append(result.Issues, models.ValidationIssue{
			Severity: models.SeverityWarning,
			Field:    "payee",
			Message:  "Payee not in registry",
		})
	}

	return result
}
