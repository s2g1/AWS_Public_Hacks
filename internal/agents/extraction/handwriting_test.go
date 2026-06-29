package extraction

import (
	"context"
	"fmt"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockTextractClient implements TextractClient for testing.
type mockTextractClient struct {
	result *TextractResult
	err    error
}

func (m *mockTextractClient) DetectDocumentText(ctx context.Context, documentBytes []byte) (*TextractResult, error) {
	return m.result, m.err
}

func TestEnhanceWithHandwriting_HandwrittenFieldGetsLowerConfidenceCap(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "John Smith",
			Confidence: 0.55, // Low confidence from Bedrock OCR
			Normalized: "John Smith",
		},
	}

	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{
					Name:          "payee",
					Value:         "John Smith",
					Confidence:    0.92, // High confidence from Textract
					IsHandwritten: true,
				},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// Handwritten field should be capped at 0.80
	if enhanced["payee"].Confidence > HandwritingConfidenceCap {
		t.Errorf("expected handwritten field confidence capped at %f, got %f",
			HandwritingConfidenceCap, enhanced["payee"].Confidence)
	}

	// Should be marked as handwritten
	if !enhanced["payee"].IsHandwritten {
		t.Error("expected field to be marked as handwritten")
	}
}

func TestEnhanceWithHandwriting_PrintedTextFieldsUnchanged(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"invoiceNumber": {
			Value:      "INV-2024-001",
			Confidence: 0.60, // Low confidence from Bedrock
			Normalized: "INV-2024-001",
		},
	}

	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{
					Name:          "invoiceNumber",
					Value:         "INV-2024-001",
					Confidence:    0.95, // Textract gives high confidence
					IsHandwritten: false,
				},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// Printed text field should use Textract's higher confidence (not capped)
	if enhanced["invoiceNumber"].Confidence != 0.95 {
		t.Errorf("expected printed field confidence 0.95, got %f", enhanced["invoiceNumber"].Confidence)
	}

	// Should NOT be marked as handwritten
	if enhanced["invoiceNumber"].IsHandwritten {
		t.Error("expected printed field to not be marked as handwritten")
	}
}

func TestEnhanceWithHandwriting_MixedDocumentCorrectlyIdentifiesBoth(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Jane Doe",
			Confidence: 0.50, // Handwritten field - low confidence from Bedrock
			Normalized: "Jane Doe",
		},
		"invoiceNumber": {
			Value:      "INV-123",
			Confidence: 0.70, // Printed field - moderate confidence from Bedrock
			Normalized: "INV-123",
		},
		"amount": {
			Value:      "$500.00",
			Confidence: 0.60, // Handwritten amount
			Normalized: "500.00",
		},
	}

	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{
					Name:          "payee",
					Value:         "Jane Doe",
					Confidence:    0.85,
					IsHandwritten: true,
				},
				{
					Name:          "invoiceNumber",
					Value:         "INV-123",
					Confidence:    0.97,
					IsHandwritten: false,
				},
				{
					Name:          "amount",
					Value:         "$500.00",
					Confidence:    0.72,
					IsHandwritten: true,
				},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// payee: handwritten, capped at 0.80
	if !enhanced["payee"].IsHandwritten {
		t.Error("expected payee to be marked as handwritten")
	}
	if enhanced["payee"].Confidence > HandwritingConfidenceCap {
		t.Errorf("expected payee confidence capped at %f, got %f",
			HandwritingConfidenceCap, enhanced["payee"].Confidence)
	}

	// invoiceNumber: printed, should use Textract's higher confidence
	if enhanced["invoiceNumber"].IsHandwritten {
		t.Error("expected invoiceNumber to not be marked as handwritten")
	}
	if enhanced["invoiceNumber"].Confidence != 0.97 {
		t.Errorf("expected invoiceNumber confidence 0.97, got %f", enhanced["invoiceNumber"].Confidence)
	}

	// amount: handwritten, confidence should be the Textract value (0.72) since it's below cap
	if !enhanced["amount"].IsHandwritten {
		t.Error("expected amount to be marked as handwritten")
	}
	if enhanced["amount"].Confidence != 0.72 {
		t.Errorf("expected amount confidence 0.72, got %f", enhanced["amount"].Confidence)
	}
}

func TestEnhanceWithHandwriting_FieldBelowHandwritingThresholdFlagged(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Illegible Name",
			Confidence: 0.40, // Very low confidence from Bedrock
			Normalized: "Illegible Name",
		},
	}

	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{
					Name:          "payee",
					Value:         "Illegible Name",
					Confidence:    0.50, // Textract also has low confidence - below threshold
					IsHandwritten: true,
				},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// Field should be marked as handwritten
	if !enhanced["payee"].IsHandwritten {
		t.Error("expected field to be marked as handwritten")
	}

	// Confidence should be below HandwritingThreshold (0.65) - flagged for verification
	if enhanced["payee"].Confidence >= HandwritingThreshold {
		t.Errorf("expected handwritten field confidence below threshold %f, got %f",
			HandwritingThreshold, enhanced["payee"].Confidence)
	}

	// Confidence should be 0.50 (from Textract, below cap so not capped)
	if enhanced["payee"].Confidence != 0.50 {
		t.Errorf("expected confidence 0.50, got %f", enhanced["payee"].Confidence)
	}
}

func TestEnhanceWithHandwriting_NilTextractClient(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Test",
			Confidence: 0.55,
			Normalized: "Test",
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), nil)

	// Should return original fields unchanged
	if enhanced["payee"].Confidence != 0.55 {
		t.Errorf("expected original confidence 0.55, got %f", enhanced["payee"].Confidence)
	}
	if enhanced["payee"].IsHandwritten {
		t.Error("expected field not to be marked as handwritten with nil client")
	}
}

func TestEnhanceWithHandwriting_TextractError(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Test",
			Confidence: 0.55,
			Normalized: "Test",
		},
	}

	textractClient := &mockTextractClient{
		result: nil,
		err:    fmt.Errorf("textract service unavailable"),
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// Should return original fields unchanged on error
	if enhanced["payee"].Confidence != 0.55 {
		t.Errorf("expected original confidence 0.55, got %f", enhanced["payee"].Confidence)
	}
	if enhanced["payee"].IsHandwritten {
		t.Error("expected field not to be marked as handwritten on error")
	}
}

func TestEnhanceWithHandwriting_AllFieldsHighConfidence_SkipsTextract(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Test Corp",
			Confidence: 0.95,
			Normalized: "Test Corp",
		},
		"amount": {
			Value:      "$100.00",
			Confidence: 0.92,
			Normalized: "100.00",
		},
	}

	// This client would error if called, proving we skip Textract
	textractClient := &mockTextractClient{
		result: nil,
		err:    fmt.Errorf("should not be called"),
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// Fields should be unchanged since all are above the cap
	if enhanced["payee"].Confidence != 0.95 {
		t.Errorf("expected payee confidence 0.95, got %f", enhanced["payee"].Confidence)
	}
	if enhanced["amount"].Confidence != 0.92 {
		t.Errorf("expected amount confidence 0.92, got %f", enhanced["amount"].Confidence)
	}
}

func TestEnhanceWithHandwriting_EmptyFields(t *testing.T) {
	fields := map[string]models.ExtractedField{}

	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{Name: "payee", Value: "Test", Confidence: 0.9, IsHandwritten: true},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	if len(enhanced) != 0 {
		t.Errorf("expected empty fields, got %d", len(enhanced))
	}
}

func TestEnhanceWithHandwriting_FieldNotInTextractResults(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:      "Test Corp",
			Confidence: 0.60,
			Normalized: "Test Corp",
		},
		"date": {
			Value:      "2024-01-01",
			Confidence: 0.55,
			Normalized: "2024-01-01",
		},
	}

	// Textract only returns data for "payee", not "date"
	textractClient := &mockTextractClient{
		result: &TextractResult{
			Fields: []TextractField{
				{
					Name:          "payee",
					Value:         "Test Corp",
					Confidence:    0.88,
					IsHandwritten: true,
				},
			},
		},
	}

	enhanced := EnhanceWithHandwriting(context.Background(), fields, []byte("doc"), textractClient)

	// payee should be enhanced (handwritten, capped at 0.80)
	if !enhanced["payee"].IsHandwritten {
		t.Error("expected payee to be marked as handwritten")
	}
	if enhanced["payee"].Confidence > HandwritingConfidenceCap {
		t.Errorf("expected payee capped at %f, got %f", HandwritingConfidenceCap, enhanced["payee"].Confidence)
	}

	// date should remain unchanged since not in Textract results
	if enhanced["date"].Confidence != 0.55 {
		t.Errorf("expected date confidence 0.55 unchanged, got %f", enhanced["date"].Confidence)
	}
	if enhanced["date"].IsHandwritten {
		t.Error("expected date not marked as handwritten")
	}
}

func TestHandwritingThresholdConstant(t *testing.T) {
	if HandwritingThreshold != 0.65 {
		t.Errorf("expected HandwritingThreshold 0.65, got %f", HandwritingThreshold)
	}
}

func TestHandwritingConfidenceCapConstant(t *testing.T) {
	if HandwritingConfidenceCap != 0.80 {
		t.Errorf("expected HandwritingConfidenceCap 0.80, got %f", HandwritingConfidenceCap)
	}
}
