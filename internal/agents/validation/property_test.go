package validation

import (
	"fmt"
	"testing"

	"federal-payment-processing/internal/models"

	"pgregory.net/rapid"
)

// **Validates: Requirements 3.4, 3.5, 3.6**

// genSeverity generates an arbitrary Severity value.
func genSeverity() *rapid.Generator[models.Severity] {
	return rapid.Custom(func(t *rapid.T) models.Severity {
		severities := []models.Severity{
			models.SeverityCritical,
			models.SeverityError,
			models.SeverityWarning,
		}
		idx := rapid.IntRange(0, len(severities)-1).Draw(t, "severityIndex")
		return severities[idx]
	})
}

// genValidationIssue generates an arbitrary ValidationIssue with the given severity.
func genValidationIssueWithSeverity(severity models.Severity) *rapid.Generator[models.ValidationIssue] {
	return rapid.Custom(func(t *rapid.T) models.ValidationIssue {
		field := rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "field")
		message := rapid.StringMatching(`[A-Za-z ]{5,20}`).Draw(t, "message")
		return models.ValidationIssue{
			Severity: severity,
			Field:    field,
			Message:  message,
		}
	})
}

// genIssueSlice generates a slice of ValidationIssues with arbitrary severities.
func genIssueSlice() *rapid.Generator[[]models.ValidationIssue] {
	return rapid.Custom(func(t *rapid.T) []models.ValidationIssue {
		n := rapid.IntRange(0, 10).Draw(t, "issueCount")
		issues := make([]models.ValidationIssue, n)
		for i := 0; i < n; i++ {
			sev := genSeverity().Draw(t, "severity")
			issues[i] = genValidationIssueWithSeverity(sev).Draw(t, "issue")
		}
		return issues
	})
}

// TestProperty_CriticalSeverity_AlwaysRejected verifies that if any issue has
// CRITICAL severity, determineStatus returns REJECTED regardless of other issues.
func TestProperty_CriticalSeverity_AlwaysRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a base set of issues (may or may not contain CRITICAL)
		baseIssues := genIssueSlice().Draw(t, "baseIssues")

		// Inject at least one CRITICAL issue
		criticalIssue := genValidationIssueWithSeverity(models.SeverityCritical).Draw(t, "criticalIssue")
		issues := append(baseIssues, criticalIssue)

		status := determineStatus(issues)
		if status != models.ValidationStatusRejected {
			t.Fatalf("expected REJECTED when CRITICAL issue present, got %s (issues: %+v)", status, issues)
		}
	})
}

// TestProperty_ErrorWithoutCritical_AlwaysNeedsReview verifies that if any issue
// has ERROR severity and no issues have CRITICAL severity, determineStatus returns NEEDS_REVIEW.
func TestProperty_ErrorWithoutCritical_AlwaysNeedsReview(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate only WARNING and ERROR issues (no CRITICAL)
		n := rapid.IntRange(0, 8).Draw(t, "warningCount")
		issues := make([]models.ValidationIssue, 0, n+1)
		for i := 0; i < n; i++ {
			// Only WARNING issues as filler
			issues = append(issues, genValidationIssueWithSeverity(models.SeverityWarning).Draw(t, "warningIssue"))
		}

		// Inject at least one ERROR issue
		errorIssue := genValidationIssueWithSeverity(models.SeverityError).Draw(t, "errorIssue")
		issues = append(issues, errorIssue)

		status := determineStatus(issues)
		if status != models.ValidationStatusNeedsReview {
			t.Fatalf("expected NEEDS_REVIEW when ERROR (no CRITICAL) present, got %s (issues: %+v)", status, issues)
		}
	})
}

// TestProperty_WarningsOnly_AlwaysValid verifies that if all issues are WARNING
// severity (no CRITICAL, no ERROR), determineStatus returns VALID.
func TestProperty_WarningsOnly_AlwaysValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate only WARNING issues
		n := rapid.IntRange(1, 10).Draw(t, "warningCount")
		issues := make([]models.ValidationIssue, n)
		for i := 0; i < n; i++ {
			issues[i] = genValidationIssueWithSeverity(models.SeverityWarning).Draw(t, "warningIssue")
		}

		status := determineStatus(issues)
		if status != models.ValidationStatusValid {
			t.Fatalf("expected VALID when only WARNING issues present, got %s (issues: %+v)", status, issues)
		}
	})
}

// TestProperty_NoIssues_AlwaysValid verifies that when no issues exist,
// determineStatus returns VALID.
func TestProperty_NoIssues_AlwaysValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Empty issues slice
		issues := []models.ValidationIssue{}

		status := determineStatus(issues)
		if status != models.ValidationStatusValid {
			t.Fatalf("expected VALID when no issues present, got %s", status)
		}
	})
}

// --- Integration tests through ValidatePayment ---

// genDocumentType generates a valid DocumentType that has required fields defined.
func genDocumentType() *rapid.Generator[models.DocumentType] {
	return rapid.Custom(func(t *rapid.T) models.DocumentType {
		docTypes := []models.DocumentType{
			models.DocumentTypeInvoice,
			models.DocumentTypePurchaseOrder,
			models.DocumentTypeTravelVoucher,
			models.DocumentTypeGrantPayment,
			models.DocumentTypeContractPayment,
		}
		idx := rapid.IntRange(0, len(docTypes)-1).Draw(t, "docTypeIndex")
		return docTypes[idx]
	})
}

// genHighConfidenceField generates an ExtractedField with confidence >= 0.80.
func genHighConfidenceField() *rapid.Generator[models.ExtractedField] {
	return rapid.Custom(func(t *rapid.T) models.ExtractedField {
		conf := rapid.Float64Range(0.80, 1.0).Draw(t, "confidence")
		return models.ExtractedField{
			Value:      "valid-value",
			Confidence: conf,
			Normalized: "valid-value",
		}
	})
}

// genValidAmountField generates an ExtractedField that has a valid positive amount.
func genValidAmountField() *rapid.Generator[models.ExtractedField] {
	return rapid.Custom(func(t *rapid.T) models.ExtractedField {
		amount := rapid.Float64Range(0.01, 1000000.0).Draw(t, "amount")
		conf := rapid.Float64Range(0.80, 1.0).Draw(t, "confidence")
		return models.ExtractedField{
			Value:      fmt.Sprintf("%.2f", amount),
			Confidence: conf,
			Normalized: fmt.Sprintf("%.2f", amount),
		}
	})
}

// genValidDateField generates an ExtractedField with a valid past date.
func genValidDateField() *rapid.Generator[models.ExtractedField] {
	return rapid.Custom(func(t *rapid.T) models.ExtractedField {
		year := rapid.IntRange(2020, 2024).Draw(t, "year")
		month := rapid.IntRange(1, 12).Draw(t, "month")
		day := rapid.IntRange(1, 28).Draw(t, "day")
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		conf := rapid.Float64Range(0.80, 1.0).Draw(t, "confidence")
		return models.ExtractedField{
			Value:      dateStr,
			Confidence: conf,
			Normalized: dateStr,
		}
	})
}

// TestProperty_ValidatePayment_AllRequiredFieldsHighConfidence_StatusValid verifies that
// when all required fields are present with high confidence and valid formats,
// ValidatePayment returns VALID status.
func TestProperty_ValidatePayment_AllRequiredFieldsHighConfidence_StatusValid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		docType := genDocumentType().Draw(t, "docType")
		requiredFields := models.RequiredFieldsByDocType[docType]

		fields := make(map[string]models.ExtractedField)
		for _, fieldName := range requiredFields {
			if fieldName == "amount" || fieldName == "totalAmount" || fieldName == "totalClaim" {
				fields[fieldName] = genValidAmountField().Draw(t, fieldName)
			} else if fieldName == "date" || fieldName == "dates" {
				fields[fieldName] = genValidDateField().Draw(t, fieldName)
			} else {
				fields[fieldName] = genHighConfidenceField().Draw(t, fieldName)
			}
		}

		extraction := models.ExtractionResult{
			DocumentType:      docType,
			Fields:            fields,
			OverallConfidence: 0.95,
		}

		result := ValidatePayment(extraction)
		if result.Status != models.ValidationStatusValid {
			t.Fatalf("expected VALID when all required fields present with high confidence, got %s (docType: %s, issues: %+v)",
				result.Status, docType, result.Issues)
		}
	})
}

// TestProperty_ValidatePayment_MissingRequiredField_StatusRejected verifies that
// when at least one required field is missing, ValidatePayment returns REJECTED.
func TestProperty_ValidatePayment_MissingRequiredField_StatusRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		docType := genDocumentType().Draw(t, "docType")
		requiredFields := models.RequiredFieldsByDocType[docType]
		if len(requiredFields) == 0 {
			t.Skip("no required fields for this doc type")
		}

		// Pick a field to omit
		omitIdx := rapid.IntRange(0, len(requiredFields)-1).Draw(t, "omitIndex")

		fields := make(map[string]models.ExtractedField)
		for i, fieldName := range requiredFields {
			if i == omitIdx {
				continue // skip this field
			}
			if fieldName == "amount" || fieldName == "totalAmount" || fieldName == "totalClaim" {
				fields[fieldName] = genValidAmountField().Draw(t, fieldName)
			} else if fieldName == "date" || fieldName == "dates" {
				fields[fieldName] = genValidDateField().Draw(t, fieldName)
			} else {
				fields[fieldName] = genHighConfidenceField().Draw(t, fieldName)
			}
		}

		extraction := models.ExtractionResult{
			DocumentType:      docType,
			Fields:            fields,
			OverallConfidence: 0.95,
		}

		result := ValidatePayment(extraction)
		if result.Status != models.ValidationStatusRejected {
			t.Fatalf("expected REJECTED when required field %q missing, got %s (docType: %s, issues: %+v)",
				requiredFields[omitIdx], result.Status, docType, result.Issues)
		}
	})
}

// TestProperty_ValidatePayment_InvalidAmount_StatusNeedsReview verifies that
// when amount has invalid format (and all required fields are present), status is NEEDS_REVIEW or REJECTED.
func TestProperty_ValidatePayment_InvalidAmount_StatusNeedsReview(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Use UNKNOWN doc type so no required-field CRITICAL issues are generated
		fields := make(map[string]models.ExtractedField)
		fields["amount"] = models.ExtractedField{
			Value:      "not-a-number",
			Confidence: 0.90,
			Normalized: "not-a-number",
		}

		extraction := models.ExtractionResult{
			DocumentType:      models.DocumentTypeUnknown,
			Fields:            fields,
			OverallConfidence: 0.95,
		}

		result := ValidatePayment(extraction)
		if result.Status != models.ValidationStatusNeedsReview {
			t.Fatalf("expected NEEDS_REVIEW when amount is invalid format, got %s (issues: %+v)",
				result.Status, result.Issues)
		}
	})
}
