package extraction

import (
	"federal-payment-processing/internal/models"
)

// ShouldEscalateForHandwriting checks whether any handwritten fields have confidence
// below HandwritingThreshold (0.65). Returns true and the list of problematic field names
// if escalation is needed.
func ShouldEscalateForHandwriting(fields map[string]models.ExtractedField) (bool, []string) {
	if len(fields) == 0 {
		return false, nil
	}

	var lowConfidenceFields []string

	for name, field := range fields {
		if field.IsHandwritten && field.Confidence < HandwritingThreshold {
			lowConfidenceFields = append(lowConfidenceFields, name)
		}
	}

	if len(lowConfidenceFields) > 0 {
		return true, lowConfidenceFields
	}

	return false, nil
}
