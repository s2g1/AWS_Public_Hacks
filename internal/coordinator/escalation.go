package coordinator

import (
	"federal-payment-processing/internal/models"
	"fmt"
)

// ExtractionThreshold is the minimum overall confidence required to proceed
// with automated processing. Payments below this threshold are escalated
// to human review.
const ExtractionThreshold = 0.75

// CheckExtractionConfidence evaluates whether the extraction result meets
// the minimum confidence threshold. If the overall confidence is below
// ExtractionThreshold, the payment is transitioned to ESCALATED status
// and further automated processing is halted.
//
// Returns:
//   - (true, nil) if the payment was escalated due to low confidence
//   - (false, nil) if confidence is sufficient and processing may continue
//   - (false, error) if the status transition fails
func CheckExtractionConfidence(record *models.PaymentRecord, extractionResult *models.ExtractionResult) (bool, error) {
	if record == nil {
		return false, fmt.Errorf("payment record must not be nil")
	}
	if extractionResult == nil {
		return false, fmt.Errorf("extraction result must not be nil")
	}

	if extractionResult.OverallConfidence >= ExtractionThreshold {
		return false, nil
	}

	reason := fmt.Sprintf(
		"Low extraction confidence: %.2f (threshold: %.2f)",
		extractionResult.OverallConfidence,
		ExtractionThreshold,
	)

	err := models.TransitionPayment(record, models.PaymentStatusEscalated, "coordinator", reason)
	if err != nil {
		return false, fmt.Errorf("failed to escalate payment: %w", err)
	}

	return true, nil
}
