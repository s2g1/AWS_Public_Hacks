package models

import (
	"fmt"
	"time"
)

// ValidateTransition checks whether transitioning from the current status to the
// next status is allowed according to the payment state machine definition.
// Returns nil if the transition is valid, or a descriptive error if not.
func ValidateTransition(current, next PaymentStatus) error {
	allowed, exists := ValidTransitions[current]
	if !exists {
		return fmt.Errorf("unknown current status %q: no transitions defined", current)
	}

	for _, s := range allowed {
		if s == next {
			return nil
		}
	}

	if len(allowed) == 0 {
		return fmt.Errorf("invalid transition: status %q is terminal and cannot transition to %q", current, next)
	}

	return fmt.Errorf("invalid transition from %q to %q: allowed targets are %v", current, next, allowed)
}

// TransitionPayment validates the transition, updates the payment record status,
// appends an audit entry, and updates the record timestamp.
// Returns nil on success or a descriptive error if the transition is invalid.
func TransitionPayment(record *PaymentRecord, newStatus PaymentStatus, actor string, reason string) error {
	if record == nil {
		return fmt.Errorf("cannot transition nil payment record")
	}

	if err := ValidateTransition(record.Status, newStatus); err != nil {
		return err
	}

	previousStatus := record.Status
	record.Status = newStatus

	entry := AuditEntry{
		Timestamp:      time.Now(),
		Actor:          actor,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Reason:         reason,
	}
	record.AuditTrail = append(record.AuditTrail, entry)

	record.UpdatedAt = time.Now()

	return nil
}
