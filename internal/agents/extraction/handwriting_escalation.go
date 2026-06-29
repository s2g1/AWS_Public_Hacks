package extraction

import (
	"federal-payment-processing/internal/models"
)

// EscalationReasonHandwritingReview is the escalation reason used when
// handwritten fields have confidence below the HandwritingThreshold.
const EscalationReasonHandwritingReview = "handwriting_review"

// CheckHandwritingEscalation scans all extracted fields and identifies those
// that are handwritten with confidence below the HandwritingThreshold (0.65).
// It returns the list of field names needing human verification and a boolean
// indicating whether escalation is needed (i.e., at least one field was flagged).
func CheckHandwritingEscalation(fields map[string]models.ExtractedField) ([]string, bool) {
	if len(fields) == 0 {
		return nil, false
	}

	var flagged []string
	for name, field := range fields {
		if field.IsHandwritten && field.Confidence < HandwritingThreshold {
			flagged = append(flagged, name)
		}
	}

	if len(flagged) == 0 {
		return nil, false
	}

	return flagged, true
}
