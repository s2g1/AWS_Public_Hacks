package compliance

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockBedrockClient implements BedrockClient for testing.
type mockBedrockClient struct {
	response []byte
	err      error
	// capturedModelID records the model ID used in the invocation.
	capturedModelID string
	// capturedPayload records the payload sent to InvokeModel.
	capturedPayload []byte
}

func (m *mockBedrockClient) InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
	m.capturedModelID = modelID
	m.capturedPayload = payload
	return m.response, m.err
}

// buildMockBedrockResponse creates a properly formatted Bedrock response envelope
// containing the given FAR evaluation JSON.
func buildMockBedrockResponse(farJSON string) []byte {
	resp := bedrockResponse{
		Content: []bedrockContentBlock{
			{
				Type: "text",
				Text: farJSON,
			},
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestCheckFARCompliance_NoViolations(t *testing.T) {
	farJSON := `{"flags": []}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Acme Corp", Normalized: "Acme Corp", Confidence: 0.95},
			"amount": {Value: "5000.00", Normalized: "5000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(flags))
	}
}

func TestCheckFARCompliance_BlockingViolation(t *testing.T) {
	farJSON := `{"flags": [{"rule": "FAR_31_205", "severity": "BLOCKING", "message": "Unallowable cost: entertainment expenses"}]}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Vendor Inc", Normalized: "Vendor Inc", Confidence: 0.95},
			"amount": {Value: "15000.00", Normalized: "15000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Rule != "FAR_31_205" {
		t.Errorf("expected rule FAR_31_205, got %s", flags[0].Rule)
	}
	if flags[0].Severity != models.FlagSeverityBlocking {
		t.Errorf("expected BLOCKING severity, got %s", flags[0].Severity)
	}
	if flags[0].Message != "Unallowable cost: entertainment expenses" {
		t.Errorf("unexpected message: %s", flags[0].Message)
	}
}

func TestCheckFARCompliance_RequiresReviewViolation(t *testing.T) {
	farJSON := `{"flags": [{"rule": "FAR_19_502", "severity": "REQUIRES_REVIEW", "message": "Small business set-aside may apply for this amount"}]}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypePurchaseOrder,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Big Corp", Normalized: "Big Corp", Confidence: 0.90},
			"amount": {Value: "200000.00", Normalized: "200000.00", Confidence: 0.90},
		},
		OverallConfidence: 0.90,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Rule != "FAR_19_502" {
		t.Errorf("expected rule FAR_19_502, got %s", flags[0].Rule)
	}
	if flags[0].Severity != models.FlagSeverityRequiresReview {
		t.Errorf("expected REQUIRES_REVIEW severity, got %s", flags[0].Severity)
	}
}

func TestCheckFARCompliance_MultipleViolations(t *testing.T) {
	farJSON := `{"flags": [
		{"rule": "FAR_13_003", "severity": "REQUIRES_REVIEW", "message": "Exceeds micro-purchase threshold without competition"},
		{"rule": "FAR_32_905", "severity": "REQUIRES_REVIEW", "message": "Payment timing may not comply with Prompt Payment Act"}
	]}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Quick Services LLC", Normalized: "Quick Services LLC", Confidence: 0.88},
			"amount": {Value: "8500.00", Normalized: "8500.00", Confidence: 0.92},
		},
		OverallConfidence: 0.88,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	if flags[0].Rule != "FAR_13_003" {
		t.Errorf("expected first rule FAR_13_003, got %s", flags[0].Rule)
	}
	if flags[1].Rule != "FAR_32_905" {
		t.Errorf("expected second rule FAR_32_905, got %s", flags[1].Rule)
	}
}

func TestCheckFARCompliance_BedrockError(t *testing.T) {
	client := &mockBedrockClient{
		err: errors.New("bedrock service unavailable"),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err == nil {
		t.Fatal("expected error when bedrock fails")
	}
	if flags != nil {
		t.Errorf("expected nil flags on error, got %v", flags)
	}
}

func TestCheckFARCompliance_NilClient(t *testing.T) {
	checker := NewFARChecker(nil)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	_, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err == nil {
		t.Fatal("expected error with nil client")
	}
}

func TestCheckFARCompliance_MalformedJSON(t *testing.T) {
	// Response with invalid JSON in the text content.
	client := &mockBedrockClient{
		response: buildMockBedrockResponse("this is not valid json"),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	_, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err == nil {
		t.Fatal("expected error when response contains malformed JSON")
	}
}

func TestCheckFARCompliance_EmptyResponseContent(t *testing.T) {
	// Response with no text content blocks.
	resp := bedrockResponse{
		Content: []bedrockContentBlock{},
	}
	respBytes, _ := json.Marshal(resp)
	client := &mockBedrockClient{
		response: respBytes,
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 0 {
		t.Errorf("expected 0 flags for empty response, got %d", len(flags))
	}
}

func TestCheckFARCompliance_UsesCorrectModelID(t *testing.T) {
	farJSON := `{"flags": []}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	_, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.capturedModelID != BedrockModelID {
		t.Errorf("expected model ID %s, got %s", BedrockModelID, client.capturedModelID)
	}
}

func TestCheckFARCompliance_PromptContainsPaymentDetails(t *testing.T) {
	farJSON := `{"flags": []}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Acme Corp", Normalized: "Acme Corp", Confidence: 0.95},
			"amount": {Value: "75000.00", Normalized: "75000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	_, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the payload contains the payment details.
	var req bedrockRequest
	if err := json.Unmarshal(client.capturedPayload, &req); err != nil {
		t.Fatalf("failed to unmarshal captured payload: %v", err)
	}
	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message in request, got %d", len(req.Messages))
	}

	promptContent := req.Messages[0].Content
	if promptContent == "" {
		t.Fatal("expected non-empty prompt content")
	}
	// Verify key payment details are in the prompt.
	if !containsSubstring(promptContent, "Acme Corp") {
		t.Error("expected prompt to contain payee name 'Acme Corp'")
	}
	if !containsSubstring(promptContent, "75000.00") {
		t.Error("expected prompt to contain amount '75000.00'")
	}
	if !containsSubstring(promptContent, "SUPPLIES_AND_SERVICES") {
		t.Error("expected prompt to contain category 'SUPPLIES_AND_SERVICES'")
	}
}

func TestCheckFARCompliance_UnknownSeverityDefaultsToRequiresReview(t *testing.T) {
	farJSON := `{"flags": [{"rule": "FAR_UNKNOWN", "severity": "INFO", "message": "Informational note"}]}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Test Corp", Normalized: "Test Corp", Confidence: 0.95},
			"amount": {Value: "1000.00", Normalized: "1000.00", Confidence: 0.95},
		},
		OverallConfidence: 0.95,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Severity != models.FlagSeverityRequiresReview {
		t.Errorf("expected REQUIRES_REVIEW for unknown severity, got %s", flags[0].Severity)
	}
}

func TestCheckFARCompliance_MissingFields(t *testing.T) {
	farJSON := `{"flags": []}`
	client := &mockBedrockClient{
		response: buildMockBedrockResponse(farJSON),
	}
	checker := NewFARChecker(client)

	// Extraction with no payee or amount fields.
	extraction := &models.ExtractionResult{
		DocumentType:      models.DocumentTypeUnknown,
		Fields:            map[string]models.ExtractedField{},
		OverallConfidence: 0.50,
	}

	flags, err := checker.CheckFARCompliance(context.Background(), extraction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 0 {
		t.Errorf("expected 0 flags for empty extraction, got %d", len(flags))
	}
}

func TestDetermineCategory(t *testing.T) {
	tests := []struct {
		docType  models.DocumentType
		expected string
	}{
		{models.DocumentTypeInvoice, "SUPPLIES_AND_SERVICES"},
		{models.DocumentTypePurchaseOrder, "PURCHASE_ORDER"},
		{models.DocumentTypeTravelVoucher, "TRAVEL"},
		{models.DocumentTypeGrantPayment, "GRANT"},
		{models.DocumentTypeContractPayment, "CONTRACT"},
		{models.DocumentTypeUnknown, "GENERAL"},
	}

	for _, tc := range tests {
		extraction := &models.ExtractionResult{DocumentType: tc.docType}
		result := determineCategory(extraction)
		if result != tc.expected {
			t.Errorf("determineCategory(%s) = %s, want %s", tc.docType, result, tc.expected)
		}
	}
}

func TestMapFARSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected models.FlagSeverity
	}{
		{"BLOCKING", models.FlagSeverityBlocking},
		{"REQUIRES_REVIEW", models.FlagSeverityRequiresReview},
		{"UNKNOWN", models.FlagSeverityRequiresReview},
		{"", models.FlagSeverityRequiresReview},
	}

	for _, tc := range tests {
		result := mapFARSeverity(tc.input)
		if result != tc.expected {
			t.Errorf("mapFARSeverity(%q) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

// containsSubstring checks if s contains the given substring.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
