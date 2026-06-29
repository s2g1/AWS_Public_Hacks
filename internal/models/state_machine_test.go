package models

import (
	"testing"
	"time"
)

func TestValidateTransition_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current PaymentStatus
		next    PaymentStatus
	}{
		{"received to extracting", PaymentStatusReceived, PaymentStatusExtracting},
		{"extracting to extracted", PaymentStatusExtracting, PaymentStatusExtracted},
		{"extracted to validating", PaymentStatusExtracted, PaymentStatusValidating},
		{"validating to validated", PaymentStatusValidating, PaymentStatusValidated},
		{"validated to checking compliance", PaymentStatusValidated, PaymentStatusCheckingCompliance},
		{"checking compliance to compliant", PaymentStatusCheckingCompliance, PaymentStatusCompliant},
		{"compliant to routing", PaymentStatusCompliant, PaymentStatusRouting},
		{"routing to routed", PaymentStatusRouting, PaymentStatusRouted},
		{"routed to approving", PaymentStatusRouted, PaymentStatusApproving},
		{"approving to approved", PaymentStatusApproving, PaymentStatusApproved},
		{"approved to disbursing", PaymentStatusApproved, PaymentStatusDisbursing},
		{"disbursing to disbursed", PaymentStatusDisbursing, PaymentStatusDisbursed},
		{"received to rejected", PaymentStatusReceived, PaymentStatusRejected},
		{"received to escalated", PaymentStatusReceived, PaymentStatusEscalated},
		{"received to failed", PaymentStatusReceived, PaymentStatusFailed},
		{"escalated to extracting", PaymentStatusEscalated, PaymentStatusExtracting},
		{"escalated to validating", PaymentStatusEscalated, PaymentStatusValidating},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.current, tt.next)
			if err != nil {
				t.Errorf("expected valid transition from %q to %q, got error: %v", tt.current, tt.next, err)
			}
		})
	}
}

func TestValidateTransition_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current PaymentStatus
		next    PaymentStatus
	}{
		{"received to approved", PaymentStatusReceived, PaymentStatusApproved},
		{"received to disbursed", PaymentStatusReceived, PaymentStatusDisbursed},
		{"disbursed to received", PaymentStatusDisbursed, PaymentStatusReceived},
		{"rejected to received", PaymentStatusRejected, PaymentStatusReceived},
		{"failed to received", PaymentStatusFailed, PaymentStatusReceived},
		{"extracting to compliant", PaymentStatusExtracting, PaymentStatusCompliant},
		{"disbursing to rejected", PaymentStatusDisbursing, PaymentStatusRejected},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.current, tt.next)
			if err == nil {
				t.Errorf("expected invalid transition from %q to %q to return error, got nil", tt.current, tt.next)
			}
		})
	}
}

func TestValidateTransition_UnknownStatus(t *testing.T) {
	err := ValidateTransition(PaymentStatus("UNKNOWN"), PaymentStatusReceived)
	if err == nil {
		t.Error("expected error for unknown status, got nil")
	}
}

func TestTransitionPayment_Success(t *testing.T) {
	record := &PaymentRecord{
		PaymentID:  "PAY-001",
		Status:     PaymentStatusReceived,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AuditTrail: []AuditEntry{},
	}

	err := TransitionPayment(record, PaymentStatusExtracting, "extraction-agent", "starting extraction")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if record.Status != PaymentStatusExtracting {
		t.Errorf("expected status %q, got %q", PaymentStatusExtracting, record.Status)
	}

	if len(record.AuditTrail) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(record.AuditTrail))
	}

	entry := record.AuditTrail[0]
	if entry.Actor != "extraction-agent" {
		t.Errorf("expected actor %q, got %q", "extraction-agent", entry.Actor)
	}
	if entry.PreviousStatus != PaymentStatusReceived {
		t.Errorf("expected previous status %q, got %q", PaymentStatusReceived, entry.PreviousStatus)
	}
	if entry.NewStatus != PaymentStatusExtracting {
		t.Errorf("expected new status %q, got %q", PaymentStatusExtracting, entry.NewStatus)
	}
	if entry.Reason != "starting extraction" {
		t.Errorf("expected reason %q, got %q", "starting extraction", entry.Reason)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestTransitionPayment_InvalidTransition(t *testing.T) {
	record := &PaymentRecord{
		PaymentID:  "PAY-002",
		Status:     PaymentStatusReceived,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AuditTrail: []AuditEntry{},
	}

	err := TransitionPayment(record, PaymentStatusDisbursed, "system", "skip to end")
	if err == nil {
		t.Fatal("expected error for invalid transition, got nil")
	}

	// Status should remain unchanged
	if record.Status != PaymentStatusReceived {
		t.Errorf("expected status to remain %q, got %q", PaymentStatusReceived, record.Status)
	}

	// No audit entry should be added
	if len(record.AuditTrail) != 0 {
		t.Errorf("expected 0 audit entries, got %d", len(record.AuditTrail))
	}
}

func TestTransitionPayment_NilRecord(t *testing.T) {
	err := TransitionPayment(nil, PaymentStatusExtracting, "agent", "test")
	if err == nil {
		t.Fatal("expected error for nil record, got nil")
	}
}

func TestTransitionPayment_MultipleTransitions(t *testing.T) {
	record := &PaymentRecord{
		PaymentID:  "PAY-003",
		Status:     PaymentStatusReceived,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AuditTrail: []AuditEntry{},
	}

	transitions := []struct {
		status PaymentStatus
		actor  string
		reason string
	}{
		{PaymentStatusExtracting, "extraction-agent", "starting extraction"},
		{PaymentStatusExtracted, "extraction-agent", "extraction complete"},
		{PaymentStatusValidating, "validation-agent", "starting validation"},
		{PaymentStatusValidated, "validation-agent", "validation passed"},
	}

	for _, tr := range transitions {
		err := TransitionPayment(record, tr.status, tr.actor, tr.reason)
		if err != nil {
			t.Fatalf("unexpected error transitioning to %q: %v", tr.status, err)
		}
	}

	if record.Status != PaymentStatusValidated {
		t.Errorf("expected final status %q, got %q", PaymentStatusValidated, record.Status)
	}

	if len(record.AuditTrail) != 4 {
		t.Errorf("expected 4 audit entries, got %d", len(record.AuditTrail))
	}
}

func TestTransitionPayment_TerminalStateRejectsTransition(t *testing.T) {
	record := &PaymentRecord{
		PaymentID:  "PAY-004",
		Status:     PaymentStatusDisbursed,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AuditTrail: []AuditEntry{},
	}

	err := TransitionPayment(record, PaymentStatusReceived, "system", "reset attempt")
	if err == nil {
		t.Fatal("expected error transitioning from terminal state, got nil")
	}
}
