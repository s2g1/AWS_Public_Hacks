package validation

import (
	"errors"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockPaymentStore implements PaymentStore for testing.
type mockPaymentStore struct {
	matches []DuplicateMatch
	err     error
	// Capture query parameters for assertion
	calledPayee       string
	calledAmount      string
	calledDate        string
	calledLookbackDays int
}

func (m *mockPaymentStore) QueryDuplicates(payee string, amount string, date string, lookbackDays int) ([]DuplicateMatch, error) {
	m.calledPayee = payee
	m.calledAmount = amount
	m.calledDate = date
	m.calledLookbackDays = lookbackDays
	return m.matches, m.err
}

func TestValidatePaymentWithDuplicateCheck_NoDuplicates_ReturnsValid(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{matches: nil, err: nil}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %d: %+v", len(result.Issues), result.Issues)
	}
}

func TestValidatePaymentWithDuplicateCheck_DuplicateFound_AddsWarning(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{
		matches: []DuplicateMatch{
			{PaymentID: "PAY-2024-001", Date: "2024-06-10"},
		},
		err: nil,
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// WARNING alone should not affect status — still VALID
	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID (warning doesn't reject), got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "duplicate" &&
			issue.Message == "Potential duplicate: PAY-2024-001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WARNING issue for duplicate payment PAY-2024-001")
	}
}

func TestValidatePaymentWithDuplicateCheck_MultipleDuplicates_AddsMultipleWarnings(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{
		matches: []DuplicateMatch{
			{PaymentID: "PAY-2024-001", Date: "2024-06-10"},
			{PaymentID: "PAY-2024-002", Date: "2024-06-12"},
		},
		err: nil,
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}

	duplicateWarnings := 0
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "duplicate" {
			duplicateWarnings++
		}
	}
	if duplicateWarnings != 2 {
		t.Errorf("expected 2 duplicate warnings, got %d", duplicateWarnings)
	}
}

func TestValidatePaymentWithDuplicateCheck_StoreError_SkipsDuplicateCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{
		matches: nil,
		err:     errors.New("DynamoDB connection failed"),
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// Should still return valid result without duplicate issues
	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID when store errors, got %s", result.Status)
	}
	for _, issue := range result.Issues {
		if issue.Field == "duplicate" {
			t.Error("expected no duplicate issues when store returns error")
		}
	}
}

func TestValidatePaymentWithDuplicateCheck_MissingPayeeField_SkipsDuplicateCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "payee")
	store := &mockPaymentStore{
		matches: []DuplicateMatch{{PaymentID: "PAY-001", Date: "2024-06-10"}},
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// Should still flag missing payee as CRITICAL (from base validation)
	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected REJECTED for missing payee, got %s", result.Status)
	}
	// But no duplicate warning since payee field is missing
	for _, issue := range result.Issues {
		if issue.Field == "duplicate" {
			t.Error("expected no duplicate issues when payee field is missing")
		}
	}
}

func TestValidatePaymentWithDuplicateCheck_MissingAmountField_SkipsDuplicateCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "amount")
	store := &mockPaymentStore{
		matches: []DuplicateMatch{{PaymentID: "PAY-001", Date: "2024-06-10"}},
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// No duplicate check performed
	for _, issue := range result.Issues {
		if issue.Field == "duplicate" {
			t.Error("expected no duplicate issues when amount field is missing")
		}
	}
}

func TestValidatePaymentWithDuplicateCheck_MissingDateField_SkipsDuplicateCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "date")
	store := &mockPaymentStore{
		matches: []DuplicateMatch{{PaymentID: "PAY-001", Date: "2024-06-10"}},
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// No duplicate check performed
	for _, issue := range result.Issues {
		if issue.Field == "duplicate" {
			t.Error("expected no duplicate issues when date field is missing")
		}
	}
}

func TestValidatePaymentWithDuplicateCheck_QueriesWithCorrectParams(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{matches: nil, err: nil}

	ValidatePaymentWithDuplicateCheck(extraction, store)

	if store.calledPayee != "Acme Corp" {
		t.Errorf("expected payee 'Acme Corp', got %q", store.calledPayee)
	}
	if store.calledAmount != "1500.00" {
		t.Errorf("expected amount '1500.00', got %q", store.calledAmount)
	}
	if store.calledDate != "2024-06-15" {
		t.Errorf("expected date '2024-06-15', got %q", store.calledDate)
	}
	if store.calledLookbackDays != 30 {
		t.Errorf("expected lookbackDays 30, got %d", store.calledLookbackDays)
	}
}

func TestValidatePaymentWithDuplicateCheck_UsesValueWhenNormalizedEmpty(t *testing.T) {
	extraction := validInvoiceExtraction()
	// Set normalized to empty to test fallback to Value
	payee := extraction.Fields["payee"]
	payee.Normalized = ""
	payee.Value = "Fallback Payee"
	extraction.Fields["payee"] = payee

	amount := extraction.Fields["amount"]
	amount.Normalized = ""
	amount.Value = "999.99"
	extraction.Fields["amount"] = amount

	date := extraction.Fields["date"]
	date.Normalized = ""
	date.Value = "2024-07-01"
	extraction.Fields["date"] = date

	store := &mockPaymentStore{matches: nil, err: nil}

	ValidatePaymentWithDuplicateCheck(extraction, store)

	if store.calledPayee != "Fallback Payee" {
		t.Errorf("expected payee 'Fallback Payee', got %q", store.calledPayee)
	}
	if store.calledAmount != "999.99" {
		t.Errorf("expected amount '999.99', got %q", store.calledAmount)
	}
	if store.calledDate != "2024-07-01" {
		t.Errorf("expected date '2024-07-01', got %q", store.calledDate)
	}
}

func TestValidatePaymentWithDuplicateCheck_DuplicateWithExistingValidationErrors(t *testing.T) {
	extraction := validInvoiceExtraction()
	// Make amount invalid to produce an ERROR
	f := extraction.Fields["amount"]
	f.Normalized = "not-valid"
	extraction.Fields["amount"] = f

	store := &mockPaymentStore{
		matches: []DuplicateMatch{
			{PaymentID: "PAY-DUP-001", Date: "2024-06-10"},
		},
	}

	result := ValidatePaymentWithDuplicateCheck(extraction, store)

	// Status should be NEEDS_REVIEW from the amount error (not from the duplicate warning)
	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	// Should have both the amount error AND the duplicate warning
	hasAmountError := false
	hasDuplicateWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityError && issue.Field == "amount" {
			hasAmountError = true
		}
		if issue.Severity == models.SeverityWarning && issue.Field == "duplicate" {
			hasDuplicateWarning = true
		}
	}
	if !hasAmountError {
		t.Error("expected ERROR for invalid amount")
	}
	if !hasDuplicateWarning {
		t.Error("expected WARNING for duplicate")
	}
}

func TestValidatePaymentWithDuplicateCheck_ExistingValidatePaymentStillWorks(t *testing.T) {
	// Verify that the original ValidatePayment function still works unchanged
	extraction := validInvoiceExtraction()
	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected original ValidatePayment to return VALID, got %s", result.Status)
	}
}
