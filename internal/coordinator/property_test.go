package coordinator

import (
	"federal-payment-processing/internal/models"
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 2.1, 2.2**

// genConfidenceBelow generates a random confidence value in [0.0, 0.75).
func genConfidenceBelow() *rapid.Generator[float64] {
	return rapid.Custom(func(t *rapid.T) float64 {
		// Generate integer in [0, 7499] and divide by 10000 to get [0.0, 0.7499]
		n := rapid.IntRange(0, 7499).Draw(t, "confidenceBelow")
		return float64(n) / 10000.0
	})
}

// genConfidenceAtOrAbove generates a random confidence value in [0.75, 1.0].
func genConfidenceAtOrAbove() *rapid.Generator[float64] {
	return rapid.Custom(func(t *rapid.T) float64 {
		// Generate integer in [7500, 10000] and divide by 10000 to get [0.75, 1.0]
		n := rapid.IntRange(7500, 10000).Draw(t, "confidenceAtOrAbove")
		return float64(n) / 10000.0
	})
}

// TestProperty_LowConfidenceEscalates verifies that for any overall confidence < 0.75,
// CheckExtractionConfidence escalates the payment (returns true).
func TestProperty_LowConfidenceEscalates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		confidence := genConfidenceBelow().Draw(t, "confidence")

		record := &models.PaymentRecord{
			PaymentID: "PAY-PROP-001",
			Status:    models.PaymentStatusExtracted,
		}
		result := &models.ExtractionResult{
			OverallConfidence: confidence,
		}

		escalated, err := CheckExtractionConfidence(record, result)
		if err != nil {
			t.Fatalf("unexpected error for confidence %.4f: %v", confidence, err)
		}
		if !escalated {
			t.Fatalf("expected escalation for confidence %.4f (below threshold 0.75), but was not escalated", confidence)
		}
	})
}

// TestProperty_SufficientConfidenceDoesNotEscalate verifies that for any overall
// confidence >= 0.75, CheckExtractionConfidence does NOT escalate (returns false).
func TestProperty_SufficientConfidenceDoesNotEscalate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		confidence := genConfidenceAtOrAbove().Draw(t, "confidence")

		record := &models.PaymentRecord{
			PaymentID: "PAY-PROP-002",
			Status:    models.PaymentStatusExtracted,
		}
		result := &models.ExtractionResult{
			OverallConfidence: confidence,
		}

		escalated, err := CheckExtractionConfidence(record, result)
		if err != nil {
			t.Fatalf("unexpected error for confidence %.4f: %v", confidence, err)
		}
		if escalated {
			t.Fatalf("expected no escalation for confidence %.4f (at or above threshold 0.75), but was escalated", confidence)
		}
	})
}

// TestProperty_EscalatedStatusIsSet verifies that when a payment is escalated due to
// low confidence, the record's status is set to ESCALATED.
func TestProperty_EscalatedStatusIsSet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		confidence := genConfidenceBelow().Draw(t, "confidence")

		record := &models.PaymentRecord{
			PaymentID: "PAY-PROP-003",
			Status:    models.PaymentStatusExtracted,
		}
		result := &models.ExtractionResult{
			OverallConfidence: confidence,
		}

		escalated, err := CheckExtractionConfidence(record, result)
		if err != nil {
			t.Fatalf("unexpected error for confidence %.4f: %v", confidence, err)
		}
		if !escalated {
			t.Fatalf("expected escalation for confidence %.4f", confidence)
		}
		if record.Status != models.PaymentStatusEscalated {
			t.Fatalf("expected status ESCALATED after escalation with confidence %.4f, got %s", confidence, record.Status)
		}
	})
}

// TestProperty_NotEscalatedStatusUnchanged verifies that when confidence is sufficient
// (>= 0.75), the record's status remains EXTRACTED (unchanged).
func TestProperty_NotEscalatedStatusUnchanged(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		confidence := genConfidenceAtOrAbove().Draw(t, "confidence")

		record := &models.PaymentRecord{
			PaymentID: "PAY-PROP-004",
			Status:    models.PaymentStatusExtracted,
		}
		result := &models.ExtractionResult{
			OverallConfidence: confidence,
		}

		escalated, err := CheckExtractionConfidence(record, result)
		if err != nil {
			t.Fatalf("unexpected error for confidence %.4f: %v", confidence, err)
		}
		if escalated {
			t.Fatalf("expected no escalation for confidence %.4f", confidence)
		}
		if record.Status != models.PaymentStatusExtracted {
			t.Fatalf("expected status to remain EXTRACTED for confidence %.4f, got %s", confidence, record.Status)
		}
	})
}

// TestProperty_ThresholdBoundaryNotEscalated verifies that exactly 0.75 is treated
// as passing (not escalated). This is a focused boundary test using property testing
// to confirm the >= comparison behavior.
func TestProperty_ThresholdBoundaryNotEscalated(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Always use exactly 0.75 to confirm boundary behavior
		record := &models.PaymentRecord{
			PaymentID: "PAY-PROP-005",
			Status:    models.PaymentStatusExtracted,
		}
		result := &models.ExtractionResult{
			OverallConfidence: ExtractionThreshold, // exactly 0.75
		}

		escalated, err := CheckExtractionConfidence(record, result)
		if err != nil {
			t.Fatalf("unexpected error at threshold boundary: %v", err)
		}
		if escalated {
			t.Fatal("expected no escalation at exact threshold boundary (0.75), but payment was escalated")
		}
		if record.Status != models.PaymentStatusExtracted {
			t.Fatalf("expected status to remain EXTRACTED at threshold boundary, got %s", record.Status)
		}
	})
}
