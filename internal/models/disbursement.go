package models

import "time"

// DisbursementStatus represents the outcome of a disbursement operation.
type DisbursementStatus string

const (
	DisbursementStatusSuccess DisbursementStatus = "SUCCESS"
	DisbursementStatusFailed  DisbursementStatus = "FAILED"
)

// PaymentConfirmation contains the details of a successful fund transfer.
type PaymentConfirmation struct {
	TransactionID string    `json:"transactionId"`
	Amount        float64   `json:"amount"`
	Payee         string    `json:"payee"`
	DisbursedAt   time.Time `json:"disbursedAt"`
	Reference     string    `json:"reference"`
}

// DisbursementResult contains the complete output of the disbursement agent.
type DisbursementResult struct {
	Status       DisbursementStatus   `json:"status"`
	Confirmation *PaymentConfirmation `json:"confirmation,omitempty"`
	Reason       string               `json:"reason,omitempty"`
	Retryable    bool                 `json:"retryable"`
}
