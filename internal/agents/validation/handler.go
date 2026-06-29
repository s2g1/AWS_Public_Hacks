package validation

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"federal-payment-processing/internal/models"
)

// FIELD_CONFIDENCE_THRESHOLD is the minimum confidence score a required field
// must have to pass the completeness check.
const FIELD_CONFIDENCE_THRESHOLD = 0.80

// ValidatePayment performs validation on an ExtractionResult, checking completeness,
// format, and date validity. It returns a ValidationResult with status and issues.
func ValidatePayment(extraction models.ExtractionResult) models.ValidationResult {
	var issues []models.ValidationIssue

	// Step 1: Completeness check
	issues = append(issues, checkCompleteness(extraction)...)

	// Step 2: Format validation (amount)
	issues = append(issues, checkAmountFormat(extraction)...)

	// Step 3: Date validation
	issues = append(issues, checkDateValidity(extraction)...)

	// Step 4: Determine overall status
	status := determineStatus(issues)

	return models.ValidationResult{
		Status:      status,
		Issues:      issues,
		ValidatedAt: time.Now(),
	}
}

// checkCompleteness verifies all required fields for the document type are present
// and have confidence >= FIELD_CONFIDENCE_THRESHOLD.
func checkCompleteness(extraction models.ExtractionResult) []models.ValidationIssue {
	var issues []models.ValidationIssue

	requiredFields, ok := models.RequiredFieldsByDocType[extraction.DocumentType]
	if !ok {
		return issues
	}

	for _, fieldName := range requiredFields {
		field, exists := extraction.Fields[fieldName]
		if !exists || field.Confidence < FIELD_CONFIDENCE_THRESHOLD {
			issues = append(issues, models.ValidationIssue{
				Severity: models.SeverityCritical,
				Field:    fieldName,
				Message:  "Required field missing or low confidence",
			})
		}
	}

	return issues
}

// checkAmountFormat validates that the "amount" field, if present, is a valid
// currency format (parseable as a float) and is positive.
func checkAmountFormat(extraction models.ExtractionResult) []models.ValidationIssue {
	var issues []models.ValidationIssue

	field, exists := extraction.Fields["amount"]
	if !exists {
		return issues
	}

	value := field.Normalized
	if value == "" {
		value = field.Value
	}

	amount, err := parseCurrency(value)
	if err != nil {
		issues = append(issues, models.ValidationIssue{
			Severity: models.SeverityError,
			Field:    "amount",
			Message:  "Invalid currency format",
		})
		return issues
	}

	if amount <= 0 {
		issues = append(issues, models.ValidationIssue{
			Severity: models.SeverityError,
			Field:    "amount",
			Message:  "Amount must be positive",
		})
	}

	return issues
}

// checkDateValidity validates that the "date" field, if present, is in a valid
// date format. Future dates generate a WARNING.
func checkDateValidity(extraction models.ExtractionResult) []models.ValidationIssue {
	var issues []models.ValidationIssue

	field, exists := extraction.Fields["date"]
	if !exists {
		return issues
	}

	value := field.Normalized
	if value == "" {
		value = field.Value
	}

	parsedDate, err := parseDate(value)
	if err != nil {
		issues = append(issues, models.ValidationIssue{
			Severity: models.SeverityError,
			Field:    "date",
			Message:  "Invalid date format",
		})
		return issues
	}

	if parsedDate.After(time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)) {
		issues = append(issues, models.ValidationIssue{
			Severity: models.SeverityWarning,
			Field:    "date",
			Message:  "Future date detected",
		})
	}

	return issues
}

// determineStatus determines the overall validation status based on issue severities.
// CRITICAL → REJECTED, ERROR (no CRITICAL) → NEEDS_REVIEW, else → VALID.
func determineStatus(issues []models.ValidationIssue) models.ValidationStatus {
	hasCritical := false
	hasError := false

	for _, issue := range issues {
		switch issue.Severity {
		case models.SeverityCritical:
			hasCritical = true
		case models.SeverityError:
			hasError = true
		}
	}

	if hasCritical {
		return models.ValidationStatusRejected
	}
	if hasError {
		return models.ValidationStatusNeedsReview
	}
	return models.ValidationStatusValid
}

// parseCurrency parses a currency string into a float64.
// It handles optional currency symbols ($), commas as thousands separators, and whitespace.
func parseCurrency(s string) (float64, error) {
	// Remove common currency symbols, commas, and whitespace
	cleaned := strings.TrimSpace(s)
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	if cleaned == "" {
		return 0, fmt.Errorf("empty currency value")
	}

	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid currency format: %s", s)
	}

	return amount, nil
}

// parseDate attempts to parse a date string in common formats.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01-02-2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"2006/01/02",
	}

	trimmed := strings.TrimSpace(s)
	for _, format := range formats {
		if t, err := time.Parse(format, trimmed); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", s)
}
