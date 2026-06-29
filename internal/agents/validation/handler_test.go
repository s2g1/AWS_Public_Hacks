package validation

import (
	"testing"
	"time"

	"federal-payment-processing/internal/models"
)

// Helper to build a valid ExtractionResult for an INVOICE with all required fields.
func validInvoiceExtraction() models.ExtractionResult {
	return models.ExtractionResult{
		DocumentType:      models.DocumentTypeInvoice,
		OverallConfidence: 0.95,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      "Acme Corp",
				Confidence: 0.95,
				Normalized: "Acme Corp",
			},
			"amount": {
				Value:      "$1,500.00",
				Confidence: 0.92,
				Normalized: "1500.00",
			},
			"invoiceNumber": {
				Value:      "INV-2024-001",
				Confidence: 0.98,
				Normalized: "INV-2024-001",
			},
			"date": {
				Value:      "2024-06-15",
				Confidence: 0.90,
				Normalized: "2024-06-15",
			},
		},
	}
}

func TestValidatePayment_AllFieldsValid_ReturnsValid(t *testing.T) {
	extraction := validInvoiceExtraction()

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %d: %+v", len(result.Issues), result.Issues)
	}
}

func TestValidatePayment_MissingRequiredField_ReturnsCriticalAndRejected(t *testing.T) {
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "payee")

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected status REJECTED, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityCritical && issue.Field == "payee" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CRITICAL issue for missing 'payee' field")
	}
}

func TestValidatePayment_LowConfidenceField_ReturnsCriticalAndRejected(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Confidence = 0.50 // below FIELD_CONFIDENCE_THRESHOLD (0.80)
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected status REJECTED, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityCritical && issue.Field == "amount" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CRITICAL issue for low-confidence 'amount' field")
	}
}

func TestValidatePayment_ConfidenceExactlyAtThreshold_NoIssue(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Confidence = 0.80 // exactly at FIELD_CONFIDENCE_THRESHOLD
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID, got %s", result.Status)
	}
}

func TestValidatePayment_InvalidCurrencyFormat_ReturnsError(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Normalized = "not-a-number"
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityError && issue.Field == "amount" && issue.Message == "Invalid currency format" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ERROR issue for invalid currency format")
	}
}

func TestValidatePayment_NegativeAmount_ReturnsError(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Normalized = "-500.00"
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityError && issue.Field == "amount" && issue.Message == "Amount must be positive" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ERROR issue for negative amount")
	}
}

func TestValidatePayment_ZeroAmount_ReturnsError(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Normalized = "0.00"
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityError && issue.Field == "amount" && issue.Message == "Amount must be positive" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ERROR issue for zero amount")
	}
}

func TestValidatePayment_InvalidDateFormat_ReturnsError(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["date"]
	f.Normalized = "not-a-date"
	extraction.Fields["date"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusNeedsReview {
		t.Errorf("expected status NEEDS_REVIEW, got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityError && issue.Field == "date" && issue.Message == "Invalid date format" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ERROR issue for invalid date format")
	}
}

func TestValidatePayment_FutureDate_ReturnsWarning(t *testing.T) {
	extraction := validInvoiceExtraction()
	futureDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	f := extraction.Fields["date"]
	f.Normalized = futureDate
	extraction.Fields["date"] = f

	result := ValidatePayment(extraction)

	// WARNING alone should still be VALID
	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID (warnings don't affect status), got %s", result.Status)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityWarning && issue.Field == "date" && issue.Message == "Future date detected" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WARNING issue for future date")
	}
}

func TestValidatePayment_CriticalOverridesError(t *testing.T) {
	// When both CRITICAL and ERROR issues exist, status should be REJECTED
	extraction := validInvoiceExtraction()
	delete(extraction.Fields, "payee") // CRITICAL: missing field
	f := extraction.Fields["amount"]
	f.Normalized = "not-a-number" // ERROR: invalid format
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected status REJECTED (CRITICAL overrides ERROR), got %s", result.Status)
	}
}

func TestValidatePayment_UnknownDocType_NoCompletenessCheck(t *testing.T) {
	extraction := models.ExtractionResult{
		DocumentType:      models.DocumentTypeUnknown,
		OverallConfidence: 0.90,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      "$100.00",
				Confidence: 0.90,
				Normalized: "100.00",
			},
		},
	}

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusValid {
		t.Errorf("expected status VALID for unknown doc type, got %s", result.Status)
	}
}

func TestValidatePayment_AmountWithCurrencySymbol_ParsesCorrectly(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Normalized = "$2,500.75"
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	// Should parse fine and not produce any amount-related issues
	for _, issue := range result.Issues {
		if issue.Field == "amount" {
			t.Errorf("expected no amount issues for valid currency string, got: %+v", issue)
		}
	}
}

func TestValidatePayment_AmountFieldUsesValueWhenNormalizedEmpty(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["amount"]
	f.Normalized = ""
	f.Value = "2000.00"
	extraction.Fields["amount"] = f

	result := ValidatePayment(extraction)

	for _, issue := range result.Issues {
		if issue.Field == "amount" {
			t.Errorf("expected no amount issues when Value fallback is valid, got: %+v", issue)
		}
	}
}

func TestValidatePayment_DateFieldUsesValueWhenNormalizedEmpty(t *testing.T) {
	extraction := validInvoiceExtraction()
	f := extraction.Fields["date"]
	f.Normalized = ""
	f.Value = "2024-06-15"
	extraction.Fields["date"] = f

	result := ValidatePayment(extraction)

	for _, issue := range result.Issues {
		if issue.Field == "date" {
			t.Errorf("expected no date issues when Value fallback is valid, got: %+v", issue)
		}
	}
}

func TestParseCurrency_ValidFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100.00", 100.00},
		{"$1,500.00", 1500.00},
		{"$100", 100.00},
		{"1000", 1000.00},
		{"$1,000,000.50", 1000000.50},
		{"  $500.00  ", 500.00},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseCurrency(tc.input)
			if err != nil {
				t.Errorf("unexpected error parsing %q: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("parseCurrency(%q) = %f, want %f", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseCurrency_InvalidFormats(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"$abc",
		"not-a-number",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseCurrency(input)
			if err == nil {
				t.Errorf("expected error parsing %q, got nil", input)
			}
		})
	}
}

func TestParseDate_ValidFormats(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"2024-06-15"},
		{"06/15/2024"},
		{"6/5/2024"},
		{"06-15-2024"},
		{"Jun 15, 2024"},
		{"June 15, 2024"},
		{"2024/06/15"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := parseDate(tc.input)
			if err != nil {
				t.Errorf("unexpected error parsing %q: %v", tc.input, err)
			}
		})
	}
}

func TestParseDate_InvalidFormats(t *testing.T) {
	tests := []string{
		"",
		"not-a-date",
		"2024-13-01", // invalid month
		"32/01/2024", // invalid day
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDate(input)
			if err == nil {
				t.Errorf("expected error parsing %q, got nil", input)
			}
		})
	}
}

func TestDetermineStatus_NoIssues_ReturnsValid(t *testing.T) {
	status := determineStatus(nil)
	if status != models.ValidationStatusValid {
		t.Errorf("expected VALID, got %s", status)
	}
}

func TestDetermineStatus_WarningsOnly_ReturnsValid(t *testing.T) {
	issues := []models.ValidationIssue{
		{Severity: models.SeverityWarning, Field: "date", Message: "Future date"},
	}
	status := determineStatus(issues)
	if status != models.ValidationStatusValid {
		t.Errorf("expected VALID, got %s", status)
	}
}

func TestDetermineStatus_ErrorOnly_ReturnsNeedsReview(t *testing.T) {
	issues := []models.ValidationIssue{
		{Severity: models.SeverityError, Field: "amount", Message: "Invalid format"},
	}
	status := determineStatus(issues)
	if status != models.ValidationStatusNeedsReview {
		t.Errorf("expected NEEDS_REVIEW, got %s", status)
	}
}

func TestDetermineStatus_CriticalPresent_ReturnsRejected(t *testing.T) {
	issues := []models.ValidationIssue{
		{Severity: models.SeverityCritical, Field: "payee", Message: "Missing"},
		{Severity: models.SeverityError, Field: "amount", Message: "Invalid"},
	}
	status := determineStatus(issues)
	if status != models.ValidationStatusRejected {
		t.Errorf("expected REJECTED, got %s", status)
	}
}

func TestValidatePayment_PurchaseOrderMissingFields(t *testing.T) {
	extraction := models.ExtractionResult{
		DocumentType:      models.DocumentTypePurchaseOrder,
		OverallConfidence: 0.90,
		Fields: map[string]models.ExtractedField{
			"vendor": {Value: "Vendor Inc", Confidence: 0.90, Normalized: "Vendor Inc"},
			// Missing: items, totalAmount, poNumber
		},
	}

	result := ValidatePayment(extraction)

	if result.Status != models.ValidationStatusRejected {
		t.Errorf("expected REJECTED for missing PO fields, got %s", result.Status)
	}

	// Should have 3 CRITICAL issues for missing fields
	critCount := 0
	for _, issue := range result.Issues {
		if issue.Severity == models.SeverityCritical {
			critCount++
		}
	}
	if critCount != 3 {
		t.Errorf("expected 3 CRITICAL issues, got %d", critCount)
	}
}
