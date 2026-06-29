package extraction

import (
	"context"

	"federal-payment-processing/internal/models"
)

// HandwritingThreshold is the confidence threshold below which handwritten fields
// are flagged for human verification.
const HandwritingThreshold = 0.65

// HandwritingConfidenceCap is the maximum confidence score assigned to handwritten
// fields, even when the handwriting is clearly legible.
const HandwritingConfidenceCap = 0.80

// TextractField represents a single field detected by Amazon Textract.
type TextractField struct {
	Name          string  `json:"name"`
	Value         string  `json:"value"`
	Confidence    float64 `json:"confidence"`
	IsHandwritten bool    `json:"isHandwritten"`
}

// TextractResult represents the result of a Textract document analysis.
type TextractResult struct {
	Fields []TextractField `json:"fields"`
}

// TextractClient defines the interface for invoking Amazon Textract.
// This interface allows for mocking in tests.
type TextractClient interface {
	DetectDocumentText(ctx context.Context, documentBytes []byte) (*TextractResult, error)
}

// EnhanceWithHandwriting uses Textract as a fallback for fields with low confidence
// from Bedrock OCR. It detects handwritten content and adjusts confidence scores
// accordingly. Handwritten fields are capped at HandwritingConfidenceCap (0.80)
// even when clearly legible.
func EnhanceWithHandwriting(
	ctx context.Context,
	fields map[string]models.ExtractedField,
	documentBytes []byte,
	textractClient TextractClient,
) map[string]models.ExtractedField {
	if textractClient == nil || len(fields) == 0 {
		return fields
	}

	// Identify fields with low confidence that could benefit from Textract
	hasLowConfidence := false
	for _, field := range fields {
		if field.Confidence < HandwritingConfidenceCap {
			hasLowConfidence = true
			break
		}
	}

	if !hasLowConfidence {
		return fields
	}

	// Invoke Textract to detect handwriting
	textractResult, err := textractClient.DetectDocumentText(ctx, documentBytes)
	if err != nil || textractResult == nil {
		// If Textract fails, return the original fields unchanged
		return fields
	}

	// Build a lookup from Textract results
	textractLookup := make(map[string]TextractField)
	for _, tf := range textractResult.Fields {
		textractLookup[tf.Name] = tf
	}

	// Enhance fields with Textract information
	enhanced := make(map[string]models.ExtractedField, len(fields))
	for name, field := range fields {
		tf, found := textractLookup[name]
		if !found {
			// No Textract data for this field, keep original
			enhanced[name] = field
			continue
		}

		if tf.IsHandwritten {
			// Field is handwritten: cap confidence at HandwritingConfidenceCap
			confidence := tf.Confidence
			if confidence > HandwritingConfidenceCap {
				confidence = HandwritingConfidenceCap
			}

			enhanced[name] = models.ExtractedField{
				Value:         field.Value,
				Confidence:    confidence,
				BoundingBox:   field.BoundingBox,
				Normalized:    field.Normalized,
				IsHandwritten: true,
			}
		} else {
			// Field is printed text: use the better confidence between Bedrock and Textract
			if field.Confidence < tf.Confidence {
				enhanced[name] = models.ExtractedField{
					Value:         field.Value,
					Confidence:    tf.Confidence,
					BoundingBox:   field.BoundingBox,
					Normalized:    field.Normalized,
					IsHandwritten: false,
				}
			} else {
				enhanced[name] = field
			}
		}
	}

	return enhanced
}
