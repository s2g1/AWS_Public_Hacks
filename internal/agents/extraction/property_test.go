package extraction

import (
	"testing"

	"federal-payment-processing/internal/models"

	"pgregory.net/rapid"
)

// **Validates: Requirements 1.3, 1.7**

// genConfidence generates a random confidence value between 0.0 and 1.0 inclusive.
func genConfidence() *rapid.Generator[float64] {
	return rapid.Custom(func(t *rapid.T) float64 {
		// Generate integer in [0, 1000] and divide by 1000 to get [0.0, 1.0]
		v := rapid.IntRange(0, 1000).Draw(t, "confidenceInt")
		return float64(v) / 1000.0
	})
}

// genFieldName generates a random field name string.
func genFieldName() *rapid.Generator[string] {
	fieldNames := []string{
		"payee", "amount", "invoiceNumber", "date",
		"vendor", "items", "totalAmount", "poNumber",
		"traveler", "dates", "expenses", "totalClaim",
		"grantNumber", "contractNumber", "memo", "reference",
	}
	return rapid.Custom(func(t *rapid.T) string {
		idx := rapid.IntRange(0, len(fieldNames)-1).Draw(t, "fieldNameIdx")
		return fieldNames[idx]
	})
}

// genFieldsMap generates a random map of extracted fields with random confidences.
func genFieldsMap(minSize, maxSize int) *rapid.Generator[map[string]models.ExtractedField] {
	return rapid.Custom(func(t *rapid.T) map[string]models.ExtractedField {
		size := rapid.IntRange(minSize, maxSize).Draw(t, "mapSize")
		fields := make(map[string]models.ExtractedField)
		for i := 0; i < size; i++ {
			name := genFieldName().Draw(t, "fieldName")
			conf := genConfidence().Draw(t, "confidence")
			fields[name] = models.ExtractedField{
				Value:      "test-value",
				Confidence: conf,
				Normalized: "test-value",
			}
		}
		return fields
	})
}

// TestProperty_OverallConfidenceIsMinimumFieldConfidence verifies that for any set
// of extracted fields with random confidences, the overall confidence equals the
// minimum field confidence across all fields (including any added missing required fields).
func TestProperty_OverallConfidenceIsMinimumFieldConfidence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fields := genFieldsMap(1, 10).Draw(t, "fields")
		requiredFields := []string{} // No required fields missing

		// Make a copy to avoid in-place modification affecting our checks
		fieldsCopy := make(map[string]models.ExtractedField)
		for k, v := range fields {
			fieldsCopy[k] = v
		}

		result := calculateOverallConfidence(fieldsCopy, requiredFields)

		// Find the expected minimum
		expectedMin := 1.0
		for _, field := range fieldsCopy {
			if field.Confidence < expectedMin {
				expectedMin = field.Confidence
			}
		}

		if result != expectedMin {
			t.Fatalf("overall confidence %f does not equal minimum field confidence %f for fields %v",
				result, expectedMin, fieldsCopy)
		}
	})
}

// TestProperty_MissingRequiredFieldSetsConfidenceToZero verifies that when any
// required field is missing from the fields map, the overall confidence is 0.0.
func TestProperty_MissingRequiredFieldSetsConfidenceToZero(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate fields with at least one field present
		fields := genFieldsMap(1, 5).Draw(t, "fields")

		// Generate a required field name that is NOT in the fields map
		missingField := rapid.Custom(func(t *rapid.T) string {
			candidates := []string{
				"requiredA", "requiredB", "requiredC", "requiredD", "requiredE",
			}
			idx := rapid.IntRange(0, len(candidates)-1).Draw(t, "missingFieldIdx")
			return candidates[idx]
		}).Draw(t, "missingField")

		// Ensure the missing field is not already in fields
		delete(fields, missingField)

		requiredFields := []string{missingField}

		result := calculateOverallConfidence(fields, requiredFields)

		if result != 0.0 {
			t.Fatalf("expected overall confidence 0.0 when required field %q is missing, got %f",
				missingField, result)
		}
	})
}

// TestProperty_SingleFieldConfidenceEqualsOverall verifies that for a single
// extracted field, the overall confidence equals that field's confidence.
func TestProperty_SingleFieldConfidenceEqualsOverall(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		conf := genConfidence().Draw(t, "confidence")
		fieldName := genFieldName().Draw(t, "fieldName")

		fields := map[string]models.ExtractedField{
			fieldName: {
				Value:      "test-value",
				Confidence: conf,
				Normalized: "test-value",
			},
		}

		result := calculateOverallConfidence(fields, []string{})

		if result != conf {
			t.Fatalf("for single field %q with confidence %f, expected overall %f but got %f",
				fieldName, conf, conf, result)
		}
	})
}

// TestProperty_OverallConfidenceBoundedZeroToOne verifies that the overall
// confidence is always between 0.0 and 1.0 inclusive, regardless of inputs.
func TestProperty_OverallConfidenceBoundedZeroToOne(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fields := genFieldsMap(0, 10).Draw(t, "fields")

		// Generate some required fields (some may overlap with existing fields)
		numRequired := rapid.IntRange(0, 5).Draw(t, "numRequired")
		requiredFields := make([]string, numRequired)
		for i := 0; i < numRequired; i++ {
			requiredFields[i] = genFieldName().Draw(t, "requiredField")
		}

		result := calculateOverallConfidence(fields, requiredFields)

		if result < 0.0 || result > 1.0 {
			t.Fatalf("overall confidence %f is outside [0.0, 1.0] range for fields=%v required=%v",
				result, fields, requiredFields)
		}
	})
}
