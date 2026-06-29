package models

import "time"

// PaymentRecord tracks the complete lifecycle of a payment from ingestion through disbursement.
type PaymentRecord struct {
	PaymentID          string              `json:"paymentId"`
	Status             PaymentStatus       `json:"status"`
	DocumentPath       string              `json:"documentPath"`
	ExtractedData      *ExtractionResult   `json:"extractedData,omitempty"`
	ValidationResult   *ValidationResult   `json:"validationResult,omitempty"`
	ComplianceResult   *ComplianceResult   `json:"complianceResult,omitempty"`
	RoutingDecision    *RoutingDecision    `json:"routingDecision,omitempty"`
	DisbursementResult *DisbursementResult `json:"disbursementResult,omitempty"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
	AuditTrail         []AuditEntry        `json:"auditTrail"`
}
