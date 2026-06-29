package validation

import (
	"errors"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockPayeeRegistry implements PayeeRegistry for testing.
type mockPayeeRegistry struct {
	found bool
	err   error
	// Capture lookup parameter for assertion
	calledPayeeName string
}

func (m *mockPayeeRegistry) Lookup(payeeName string) (bool, error) {
	m.calledPayeeName = payeeName
	return m.found, m.err
}

// noOpPaymentStore is a PaymentStore that returns no duplicates.
type noOpPaymentStore struct{}

func (s *noOpPaymentStore) QueryDuplicates(payee string, amount string, date string, lookbackDays int) ([]DuplicateMatch, error) {
	return nil, nil
}

func TestValidatePaymentFull_PayeeFoundInRegistry_NoWarning(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: true, err: nil}

	result := ValidatePaymentFull(extraction, store, registry)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}

	for _, issue := range result.Issues {
		if issue.Field == "payee" && issue.Message == "Payee not in registry" {
			t.Error("expected no payee registry warning when payee is found")
		}
	}
}

func TestValidatePaymentFull_PayeeNotInRegistry_AddsWarning(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: false, err: nil}

	result := ValidatePaymentFull(extraction, store, registry)

	// WARNING alone should not affect status — still VALID
	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID (warning doesn't affect status), got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "payee" && issue.Message == "Payee not in registry" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WARNING issue for payee not in registry")
	}
}

func TestValidatePaymentFull_RegistryError_SkipsCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: false, err: errors.New("registry unavailable")}

	result := ValidatePaymentFull(extraction, store, registry)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID when registry errors, got %s", result.Status)
	}

	for _, issue := range result.Issues {
		if issue.Field == "payee" && issue.Message == "Payee not in registry" {
			t.Error("expected no payee registry warning when registry returns error")
		}
	}
}

func TestValidatePaymentFull_MissingPayeeField_SkipsRegistryCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "payee")
	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: false, err: nil}

	result := ValidatePaymentFull(extraction, store, registry)

	// Should have CRITICAL for missing payee (from base validation), but no registry warning
	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected REJECTED for missing payee, got %s", result.Status)
	}

	for _, issue := range result.Issues {
		if issue.Message == "Payee not in registry" {
			t.Error("expected no registry warning when payee field is missing")
		}
	}
}

func TestValidatePaymentFull_LooksUpCorrectPayeeName(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: true, err: nil}

	ValidatePaymentFull(extraction, store, registry)

	if registry.calledPayeeName != "Acme Corp" {
		t.Errorf("expected lookup for 'Acme Corp', got %q", registry.calledPayeeName)
	}
}

func TestValidatePaymentFull_UsesValueWhenNormalizedEmpty(t *testing.T) {
	extraction := validInvoiceExtraction()
	payee := extraction.Fields["payee"]
	payee.Normalized = ""
	payee.Value = "Fallback Corp"
	extraction.Fields["payee"] = payee

	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: true, err: nil}

	ValidatePaymentFull(extraction, store, registry)

	if registry.calledPayeeName != "Fallback Corp" {
		t.Errorf("expected lookup for 'Fallback Corp', got %q", registry.calledPayeeName)
	}
}

func TestValidatePaymentFull_CombinesDuplicateAndRegistryWarnings(t *testing.T) {
	extraction := validInvoiceExtraction()
	store := &mockPaymentStore{
		matches: []DuplicateMatch{{PaymentID: "PAY-001", Date: "2024-06-10"}},
		err:     nil,
	}
	registry := &mockPayeeRegistry{found: false, err: nil}

	result := ValidatePaymentFull(extraction, store, registry)

	// Should be VALID since both are just warnings
	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}

	hasDuplicateWarning := false
	hasRegistryWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "duplicate" {
			hasDuplicateWarning = true
		}
		if issue.Severity == models.SeverityWarning && issue.Field == "payee" && issue.Message == "Payee not in registry" {
			hasRegistryWarning = true
		}
	}
	if !hasDuplicateWarning {
		t.Error("expected duplicate warning")
	}
	if !hasRegistryWarning {
		t.Error("expected payee registry warning")
	}
}

func TestValidatePaymentFull_WithValidationErrors_StatusStillDetermined(t *testing.T) {
	extraction := validInvoiceExtraction()
	// Make amount invalid to produce an ERROR
	f := extraction.Fields["amount"]
	f.Normalized = "invalid"
	extraction.Fields["amount"] = f

	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: false, err: nil}

	result := ValidatePaymentFull(extraction, store, registry)

	// Status should be NEEDS_REVIEW from the amount error
	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	// Should still have registry warning in addition to format error
	hasRegistryWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "payee" && issue.Message == "Payee not in registry" {
			hasRegistryWarning = true
		}
	}
	if !hasRegistryWarning {
		t.Error("expected payee registry warning even with format errors")
	}
}

func TestValidatePaymentFull_EmptyPayeeValue_SkipsRegistryCheck(t *testing.T) {
	extraction := validInvoiceExtraction()
	payee := extraction.Fields["payee"]
	payee.Normalized = ""
	payee.Value = ""
	extraction.Fields["payee"] = payee

	store := &noOpPaymentStore{}
	registry := &mockPayeeRegistry{found: false, err: nil}

	ValidatePaymentFull(extraction, store, registry)

	// Registry should not have been called
	if registry.calledPayeeName != "" {
		t.Errorf("expected no registry lookup for empty payee, got %q", registry.calledPayeeName)
	}
}

func TestValidatePaymentFull_ExistingFunctionsStillWork(t *testing.T) {
	// Verify that simpler functions remain unchanged
	extraction := validInvoiceExtraction()

	// ValidatePayment still works
	result1 := ValidatePayment(extraction)
	if result1.Status != models.ValidationStatusValid {
		t.Errorf("expected ValidatePayment to return VALID, got %s", result1.Status)
	}

	// ValidatePaymentWithDuplicateCheck still works
	store := &noOpPaymentStore{}
	result2 := ValidatePaymentWithDuplicateCheck(extraction, store)
	if result2.Status != models.ValidationStatusValid {
		t.Errorf("expected ValidatePaymentWithDuplicateCheck to return VALID, got %s", result2.Status)
	}
}
