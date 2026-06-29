package models

import "time"

// AuditEntry records a single event in the payment audit trail.
type AuditEntry struct {
	Timestamp     time.Time     `json:"timestamp"`
	Actor         string        `json:"actor"`
	PreviousStatus PaymentStatus `json:"previousStatus"`
	NewStatus     PaymentStatus `json:"newStatus"`
	Reason        string        `json:"reason"`
}
