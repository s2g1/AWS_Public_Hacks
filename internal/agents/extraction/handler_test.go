package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockBedrockClient implements BedrockClient for testing.
type mockBedrockClient struct {
	responses [][]byte
	errors    []error
	callIdx   int
}

func (m *mockBedrockClient) InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
	idx := m.callIdx
	m.callIdx++
	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return nil, fmt.Errorf("unexpected call index %d", idx)
}

// mockS3Client implements S3Client for testing.
type mockS3Client struct {
	data []byte
	err  error
}

func (m *mockS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	return m.data, m.err
}

func makeBedrockResponse(text string) []byte {
	resp := BedrockResponse{
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{Type: "text", Text: text},
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestHandler_Handle_InvoiceExtraction(t *testing.T) {
	classificationResp := makeBedrockResponse(`{"documentType": "INVOICE"}`)
	extractionResp := makeBedrockResponse(`{"fields": [
		{"name": "payee", "value": "Acme Corp", "confidence": 0.95, "normalized": "Acme Corp"},
		{"name": "amount", "value": "$1,500.00", "confidence": 0.92, "normalized": "1500.00"},
		{"name": "invoiceNumber", "value": "INV-2024-001", "confidence": 0.98, "normalized": "INV-2024-001"},
		{"name": "date", "value": "03/15/2024", "confidence": 0.90, "normalized": "2024-03-15"}
	]}`)

	handler := &Handler{
		BedrockClient: &mockBedrockClient{
			responses: [][]byte{classificationResp, extractionResp},
		},
		S3Client: &mockS3Client{
			data: []byte("%PDF-1.4 fake document content"),
		},
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
	}

	result, err := handler.Handle(context.Background(), ExtractionEvent{
		DocumentPath: "s3://my-bucket/documents/invoice.pdf",
		PaymentID:    "pay-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DocumentType != models.DocumentTypeInvoice {
		t.Errorf("expected document type INVOICE, got %s", result.DocumentType)
	}

	if len(result.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(result.Fields))
	}

	// Overall confidence should be minimum of all fields = 0.90
	if result.OverallConfidence != 0.90 {
		t.Errorf("expected overall confidence 0.90, got %f", result.OverallConfidence)
	}

	// Check specific fields
	if result.Fields["payee"].Value != "Acme Corp" {
		t.Errorf("expected payee 'Acme Corp', got '%s'", result.Fields["payee"].Value)
	}
	if result.Fields["amount"].Confidence != 0.92 {
		t.Errorf("expected amount confidence 0.92, got %f", result.Fields["amount"].Confidence)
	}
}

func TestHandler_Handle_MissingRequiredFields(t *testing.T) {
	classificationResp := makeBedrockResponse(`{"documentType": "INVOICE"}`)
	// Only extract payee and amount, missing invoiceNumber and date
	extractionResp := makeBedrockResponse(`{"fields": [
		{"name": "payee", "value": "Acme Corp", "confidence": 0.95, "normalized": "Acme Corp"},
		{"name": "amount", "value": "$1,500.00", "confidence": 0.92, "normalized": "1500.00"}
	]}`)

	handler := &Handler{
		BedrockClient: &mockBedrockClient{
			responses: [][]byte{classificationResp, extractionResp},
		},
		S3Client: &mockS3Client{
			data: []byte("%PDF-1.4 fake document content"),
		},
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
	}

	result, err := handler.Handle(context.Background(), ExtractionEvent{
		DocumentPath: "s3://my-bucket/documents/invoice.pdf",
		PaymentID:    "pay-456",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Missing required fields should be added with confidence 0.0
	invoiceNum, exists := result.Fields["invoiceNumber"]
	if !exists {
		t.Fatal("expected invoiceNumber field to be present")
	}
	if invoiceNum.Confidence != 0.0 {
		t.Errorf("expected invoiceNumber confidence 0.0, got %f", invoiceNum.Confidence)
	}
	if invoiceNum.Value != "" {
		t.Errorf("expected invoiceNumber value empty, got '%s'", invoiceNum.Value)
	}

	dateField, exists := result.Fields["date"]
	if !exists {
		t.Fatal("expected date field to be present")
	}
	if dateField.Confidence != 0.0 {
		t.Errorf("expected date confidence 0.0, got %f", dateField.Confidence)
	}

	// Overall confidence should be 0.0 due to missing fields
	if result.OverallConfidence != 0.0 {
		t.Errorf("expected overall confidence 0.0, got %f", result.OverallConfidence)
	}
}

func TestHandler_Handle_PurchaseOrder(t *testing.T) {
	classificationResp := makeBedrockResponse(`{"documentType": "PURCHASE_ORDER"}`)
	extractionResp := makeBedrockResponse(`{"fields": [
		{"name": "vendor", "value": "Tech Solutions Inc", "confidence": 0.88, "normalized": "Tech Solutions Inc"},
		{"name": "items", "value": "Laptop computers x10", "confidence": 0.85, "normalized": "Laptop computers x10"},
		{"name": "totalAmount", "value": "$25,000.00", "confidence": 0.91, "normalized": "25000.00"},
		{"name": "poNumber", "value": "PO-2024-789", "confidence": 0.97, "normalized": "PO-2024-789"}
	]}`)

	handler := &Handler{
		BedrockClient: &mockBedrockClient{
			responses: [][]byte{classificationResp, extractionResp},
		},
		S3Client: &mockS3Client{
			data: []byte("%PDF-1.4 fake document content"),
		},
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
	}

	result, err := handler.Handle(context.Background(), ExtractionEvent{
		DocumentPath: "s3://bucket/po.pdf",
		PaymentID:    "pay-789",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DocumentType != models.DocumentTypePurchaseOrder {
		t.Errorf("expected PURCHASE_ORDER, got %s", result.DocumentType)
	}

	// Min confidence is 0.85 (items field)
	if result.OverallConfidence != 0.85 {
		t.Errorf("expected overall confidence 0.85, got %f", result.OverallConfidence)
	}
}

func TestHandler_Handle_InvalidS3Path(t *testing.T) {
	handler := &Handler{
		BedrockClient: &mockBedrockClient{},
		S3Client:      &mockS3Client{},
		ModelID:       "test-model",
	}

	_, err := handler.Handle(context.Background(), ExtractionEvent{
		DocumentPath: "invalid-path",
		PaymentID:    "pay-123",
	})

	if err == nil {
		t.Fatal("expected error for invalid S3 path")
	}
}

func TestHandler_Handle_S3Error(t *testing.T) {
	handler := &Handler{
		BedrockClient: &mockBedrockClient{},
		S3Client: &mockS3Client{
			err: fmt.Errorf("access denied"),
		},
		ModelID: "test-model",
	}

	_, err := handler.Handle(context.Background(), ExtractionEvent{
		DocumentPath: "s3://bucket/key.pdf",
		PaymentID:    "pay-123",
	})

	if err == nil {
		t.Fatal("expected error for S3 failure")
	}
}

func TestParseS3Path(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		bucket    string
		key       string
		expectErr bool
	}{
		{"valid path", "s3://my-bucket/path/to/file.pdf", "my-bucket", "path/to/file.pdf", false},
		{"simple path", "s3://bucket/key", "bucket", "key", false},
		{"no prefix", "bucket/key", "", "", true},
		{"empty key", "s3://bucket/", "", "", true},
		{"no key", "s3://bucket", "", "", true},
		{"empty bucket", "s3:///key", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bucket, key, err := parseS3Path(tc.path)
			if tc.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bucket != tc.bucket {
				t.Errorf("expected bucket %s, got %s", tc.bucket, bucket)
			}
			if key != tc.key {
				t.Errorf("expected key %s, got %s", tc.key, key)
			}
		})
	}
}

func TestParseDocumentType(t *testing.T) {
	tests := []struct {
		input    string
		expected models.DocumentType
	}{
		{`{"documentType": "INVOICE"}`, models.DocumentTypeInvoice},
		{`{"documentType": "PURCHASE_ORDER"}`, models.DocumentTypePurchaseOrder},
		{`{"documentType": "TRAVEL_VOUCHER"}`, models.DocumentTypeTravelVoucher},
		{`{"documentType": "GRANT_PAYMENT"}`, models.DocumentTypeGrantPayment},
		{`{"documentType": "CONTRACT_PAYMENT"}`, models.DocumentTypeContractPayment},
		{`{"documentType": "UNKNOWN"}`, models.DocumentTypeUnknown},
		{"INVOICE", models.DocumentTypeInvoice},
		{"purchase_order", models.DocumentTypePurchaseOrder},
		{"something else", models.DocumentTypeUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseDocumentType(tc.input)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestCalculateOverallConfidence(t *testing.T) {
	tests := []struct {
		name           string
		fields         map[string]models.ExtractedField
		requiredFields []string
		expected       float64
	}{
		{
			name: "all fields present - min is used",
			fields: map[string]models.ExtractedField{
				"payee":         {Confidence: 0.95},
				"amount":        {Confidence: 0.80},
				"invoiceNumber": {Confidence: 0.92},
				"date":          {Confidence: 0.88},
			},
			requiredFields: []string{"payee", "amount", "invoiceNumber", "date"},
			expected:       0.80,
		},
		{
			name: "missing required field sets confidence to 0",
			fields: map[string]models.ExtractedField{
				"payee":  {Confidence: 0.95},
				"amount": {Confidence: 0.80},
			},
			requiredFields: []string{"payee", "amount", "invoiceNumber", "date"},
			expected:       0.0,
		},
		{
			name:           "no fields at all",
			fields:         map[string]models.ExtractedField{},
			requiredFields: []string{"payee"},
			expected:       0.0,
		},
		{
			name: "no required fields for UNKNOWN type",
			fields: map[string]models.ExtractedField{
				"someField": {Confidence: 0.75},
			},
			requiredFields: []string{},
			expected:       0.75,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateOverallConfidence(tc.fields, tc.requiredFields)
			if result != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestDetectMediaType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{"PDF", []byte{0x25, 0x50, 0x44, 0x46, 0x2D}, "application/pdf"},
		{"PNG", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D}, "image/png"},
		{"JPEG", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"TIFF LE", []byte{0x49, 0x49, 0x2A, 0x00}, "image/tiff"},
		{"TIFF BE", []byte{0x4D, 0x4D, 0x00, 0x2A}, "image/tiff"},
		{"unknown defaults to PDF", []byte{0x00, 0x01, 0x02, 0x03}, "application/pdf"},
		{"short data defaults to PDF", []byte{0x00}, "application/pdf"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectMediaType(tc.data)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestParseExtractionResponse(t *testing.T) {
	t.Run("valid JSON response", func(t *testing.T) {
		response := `{"fields": [
			{"name": "payee", "value": "Test Corp", "confidence": 0.95, "normalized": "Test Corp"},
			{"name": "amount", "value": "$100", "confidence": 0.88, "normalized": "100.00"}
		]}`

		fields, err := parseExtractionResponse(response, models.DocumentTypeInvoice)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(fields))
		}

		if fields["payee"].Value != "Test Corp" {
			t.Errorf("expected payee 'Test Corp', got '%s'", fields["payee"].Value)
		}
		if fields["amount"].Confidence != 0.88 {
			t.Errorf("expected amount confidence 0.88, got %f", fields["amount"].Confidence)
		}
	})

	t.Run("JSON with markdown code block", func(t *testing.T) {
		response := "```json\n{\"fields\": [{\"name\": \"payee\", \"value\": \"Corp\", \"confidence\": 0.9}]}\n```"

		fields, err := parseExtractionResponse(response, models.DocumentTypeInvoice)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(fields) != 1 {
			t.Errorf("expected 1 field, got %d", len(fields))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := parseExtractionResponse("not json at all", models.DocumentTypeInvoice)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("normalized defaults to value", func(t *testing.T) {
		response := `{"fields": [{"name": "payee", "value": "Test", "confidence": 0.9}]}`

		fields, err := parseExtractionResponse(response, models.DocumentTypeInvoice)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fields["payee"].Normalized != "Test" {
			t.Errorf("expected normalized to default to value 'Test', got '%s'", fields["payee"].Normalized)
		}
	})
}

func TestHandler_Handle_AllDocumentTypes(t *testing.T) {
	docTypes := []struct {
		docType        models.DocumentType
		requiredFields []string
	}{
		{models.DocumentTypeInvoice, []string{"payee", "amount", "invoiceNumber", "date"}},
		{models.DocumentTypePurchaseOrder, []string{"vendor", "items", "totalAmount", "poNumber"}},
		{models.DocumentTypeTravelVoucher, []string{"traveler", "dates", "expenses", "totalClaim"}},
		{models.DocumentTypeGrantPayment, []string{"payee", "amount", "grantNumber", "date"}},
		{models.DocumentTypeContractPayment, []string{"payee", "amount", "contractNumber", "date"}},
	}

	for _, tc := range docTypes {
		t.Run(string(tc.docType), func(t *testing.T) {
			// Build extraction response with all required fields
			var fieldsJSON []string
			for _, f := range tc.requiredFields {
				fieldsJSON = append(fieldsJSON, fmt.Sprintf(
					`{"name": "%s", "value": "test-value", "confidence": 0.90, "normalized": "test-value"}`, f))
			}
			extractionJSON := fmt.Sprintf(`{"fields": [%s]}`, joinStrings(fieldsJSON, ","))

			classificationResp := makeBedrockResponse(fmt.Sprintf(`{"documentType": "%s"}`, tc.docType))
			extractionResp := makeBedrockResponse(extractionJSON)

			handler := &Handler{
				BedrockClient: &mockBedrockClient{
					responses: [][]byte{classificationResp, extractionResp},
				},
				S3Client: &mockS3Client{
					data: []byte("%PDF-1.4 fake"),
				},
				ModelID: "test-model",
			}

			result, err := handler.Handle(context.Background(), ExtractionEvent{
				DocumentPath: "s3://bucket/doc.pdf",
				PaymentID:    "pay-test",
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.DocumentType != tc.docType {
				t.Errorf("expected %s, got %s", tc.docType, result.DocumentType)
			}

			for _, f := range tc.requiredFields {
				if _, exists := result.Fields[f]; !exists {
					t.Errorf("expected field %s to be present", f)
				}
			}

			if result.OverallConfidence != 0.90 {
				t.Errorf("expected overall confidence 0.90, got %f", result.OverallConfidence)
			}
		})
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
