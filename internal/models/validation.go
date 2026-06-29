package models

import "time"

// Severity indicates the severity level of a validation issue.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityError    Severity = "ERROR"
	SeverityWarning  Severity = "WARNING"
)

// ValidationIssue represents a single problem found during payment validation.
type ValidationIssue struct {
	Severity Severity `json:"severity"`
	Field    string   `json:"field"`
	Message  string   `json:"message"`
}

// ValidationStatus represents the outcome of validation.
type ValidationStatus string

const (
	ValidationStatusValid      ValidationStatus = "VALID"
	ValidationStatusNeedsReview ValidationStatus = "NEEDS_REVIEW"
	ValidationStatusRejected   ValidationStatus = "REJECTED"
)

// ValidationResult contains the complete output of the validation agent.
type ValidationResult struct {
	Status      ValidationStatus  `json:"status"`
	Issues      []ValidationIssue `json:"issues"`
	ValidatedAt time.Time         `json:"validatedAt"`
}
