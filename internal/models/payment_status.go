package models

// PaymentStatus represents the lifecycle state of a payment record.
type PaymentStatus string

const (
	PaymentStatusReceived           PaymentStatus = "RECEIVED"
	PaymentStatusExtracting         PaymentStatus = "EXTRACTING"
	PaymentStatusExtracted          PaymentStatus = "EXTRACTED"
	PaymentStatusValidating         PaymentStatus = "VALIDATING"
	PaymentStatusValidated          PaymentStatus = "VALIDATED"
	PaymentStatusCheckingCompliance PaymentStatus = "CHECKING_COMPLIANCE"
	PaymentStatusCompliant          PaymentStatus = "COMPLIANT"
	PaymentStatusRouting            PaymentStatus = "ROUTING"
	PaymentStatusRouted             PaymentStatus = "ROUTED"
	PaymentStatusApproving          PaymentStatus = "APPROVING"
	PaymentStatusApproved           PaymentStatus = "APPROVED"
	PaymentStatusDisbursing         PaymentStatus = "DISBURSING"
	PaymentStatusDisbursed          PaymentStatus = "DISBURSED"
	PaymentStatusRejected           PaymentStatus = "REJECTED"
	PaymentStatusEscalated          PaymentStatus = "ESCALATED"
	PaymentStatusFailed             PaymentStatus = "FAILED"
)

// ValidTransitions defines the allowed state transitions for the payment state machine.
// The key is the current status and the value is the set of statuses reachable from it.
var ValidTransitions = map[PaymentStatus][]PaymentStatus{
	PaymentStatusReceived:           {PaymentStatusExtracting, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusExtracting:         {PaymentStatusExtracted, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusExtracted:          {PaymentStatusValidating, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusValidating:         {PaymentStatusValidated, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusValidated:          {PaymentStatusCheckingCompliance, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusCheckingCompliance: {PaymentStatusCompliant, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusCompliant:          {PaymentStatusRouting, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusRouting:            {PaymentStatusRouted, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusRouted:             {PaymentStatusApproving, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusApproving:          {PaymentStatusApproved, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusApproved:           {PaymentStatusDisbursing, PaymentStatusRejected, PaymentStatusEscalated, PaymentStatusFailed},
	PaymentStatusDisbursing:         {PaymentStatusDisbursed, PaymentStatusFailed},
	PaymentStatusDisbursed:          {},
	PaymentStatusRejected:           {},
	PaymentStatusEscalated:          {PaymentStatusExtracting, PaymentStatusValidating, PaymentStatusCheckingCompliance, PaymentStatusRouting, PaymentStatusApproving},
	PaymentStatusFailed:             {},
}

// IsTerminal returns true if the status is a terminal state (no further transitions allowed,
// except ESCALATED which can resume).
func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentStatusDisbursed || s == PaymentStatusRejected || s == PaymentStatusFailed
}
