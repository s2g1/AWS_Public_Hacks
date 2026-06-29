package coordinator

import (
	"federal-payment-processing/internal/models"
	"strings"
	"testing"
)

func TestCheckExtractionConfidence_BelowThreshold_Escalates(t *testing.T) {
	record := &models.PaymentRecord{
		PaymentID: "PAY-001",
		Status:    models.PaymentStatusExtracted,
	}
	result := &models.ExtractionResult{
		OverallConfidence: 0.60,
	}

	escalated, err := CheckExtractionConfidence(record, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !escalated {
		t.Fatal("expected payment to be escalated when confidence is below threshold")
	}
	if record.Status != models.PaymentStatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", record.Status)
	}

	// Verify audit trail was recorded
	if len(record.AuditTrail) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(record.AuditTrail))
	}
	entry := record.AuditTrail[0]
	if entry.Actor != "coordinator" {
		t.Errorf("expected actor 'coordinator', got %q", entry.Actor)
	}
	if entry.PreviousStatus != models.PaymentStatusExtracted {
		t.Errorf("expected previous status EXTRACTED, got %s", entry.PreviousStatus)
	}
	if entry.NewStatus != models.PaymentStatusEscalated {
		t.Errorf("expected new status ESCALATED, got %s", entry.NewStatus)
	}
	if !strings.Contains(entry.Reason, "Low extraction confidence") {
		t.Errorf("expected reason to contain 'Low extraction confidence', got %q", entry.Reason)
	}
	if !strings.Contains(entry.Reason, "0.60") {
		t.Errorf("expected reason to contain confidence value '0.60', got %q", entry.Reason)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp in audit entry")
	}
}

func TestCheckExtractionConfidence_AtThreshold_Passes(t *testing.T) {
	record := &models.PaymentRecord{
		PaymentID: "PAY-002",
		Status:    models.PaymentStatusExtracted,
	}
	result := &models.ExtractionResult{
		OverallConfidence: 0.75, // exactly at threshold
	}

	escalated, err := CheckExtractionConfidence(record, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if escalated {
		t.Fatal("expected payment NOT to be escalated when confidence equals threshold")
	}
	if record.Status != models.PaymentStatusExtracted {
		t.Errorf("expected status to remain EXTRACTED, got %s", record.Status)
	}
	if len(record.AuditTrail) != 0 {
		t.Errorf("expected no audit entries, got %d", len(record.AuditTrail))
	}
}

func TestCheckExtractionConfidence_AboveThreshold_Passes(t *testing.T) {
	record := &models.PaymentRecord{
		PaymentID: "PAY-003",
		Status:    models.PaymentStatusExtracted,
	}
	result := &models.ExtractionResult{
		OverallConfidence: 0.92,
	}

	escalated, err := CheckExtractionConfidence(record, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if escalated {
		t.Fatal("expected payment NOT to be escalated when confidence is above threshold")
	}
	if record.Status != models.PaymentStatusExtracted {
		t.Errorf("expected status to remain EXTRACTED, got %s", record.Status)
	}
	if len(record.AuditTrail) != 0 {
		t.Errorf("expected no audit entries, got %d", len(record.AuditTrail))
	}
}

func TestCheckExtractionConfidence_TransitionError(t *testing.T) {
	// Use a terminal status that cannot transition to ESCALATED
	record := &models.PaymentRecord{
		PaymentID: "PAY-004",
		Status:    models.PaymentStatusDisbursed, // terminal state
	}
	result := &models.ExtractionResult{
		OverallConfidence: 0.50,
	}

	escalated, err := CheckExtractionConfidence(record, result)
	if err == nil {
		t.Fatal("expected an error when transition is invalid")
	}
	if escalated {
		t.Fatal("expected escalated to be false when transition fails")
	}
	if !strings.Contains(err.Error(), "failed to escalate payment") {
		t.Errorf("expected error to contain 'failed to escalate payment', got %q", err.Error())
	}
	// Status should remain unchanged
	if record.Status != models.PaymentStatusDisbursed {
		t.Errorf("expected status to remain DISBURSED, got %s", record.Status)
	}
}

func TestCheckExtractionConfidence_NilRecord(t *testing.T) {
	result := &models.ExtractionResult{
		OverallConfidence: 0.50,
	}

	escalated, err := CheckExtractionConfidence(nil, result)
	if err == nil {
		t.Fatal("expected error for nil payment record")
	}
	if escalated {
		t.Fatal("expected escalated to be false for nil record")
	}
}

func TestCheckExtractionConfidence_NilExtractionResult(t *testing.T) {
	record := &models.PaymentRecord{
		PaymentID: "PAY-005",
		Status:    models.PaymentStatusExtracted,
	}

	escalated, err := CheckExtractionConfidence(record, nil)
	if err == nil {
		t.Fatal("expected error for nil extraction result")
	}
	if escalated {
		t.Fatal("expected escalated to be false for nil extraction result")
	}
}

func TestCheckExtractionConfidence_ZeroConfidence_Escalates(t *testing.T) {
	record := &models.PaymentRecord{
		PaymentID: "PAY-006",
		Status:    models.PaymentStatusExtracted,
	}
	result := &models.ExtractionResult{
		OverallConfidence: 0.0,
	}

	escalated, err := CheckExtractionConfidence(record, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !escalated {
		t.Fatal("expected payment to be escalated when confidence is 0.0")
	}
	if record.Status != models.PaymentStatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", record.Status)
	}
}
