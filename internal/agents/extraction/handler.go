package extraction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"federal-payment-processing/internal/models"
)

// BedrockClient defines the interface for invoking Bedrock models.
// This interface allows for mocking in tests.
type BedrockClient interface {
	InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error)
}

// S3Client defines the interface for retrieving objects from S3.
type S3Client interface {
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
}

// ExtractionEvent is the input event for the extraction Lambda handler.
type ExtractionEvent struct {
	DocumentPath string `json:"documentPath"` // S3 path in format "s3://bucket/key"
	PaymentID    string `json:"paymentId"`
}

// Handler implements the extraction agent logic.
type Handler struct {
	BedrockClient BedrockClient
	S3Client      S3Client
	ModelID       string
}

// BedrockRequest represents the request payload for Claude on Bedrock.
type BedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	Messages         []BedrockMessage `json:"messages"`
}

// BedrockMessage represents a message in the Bedrock conversation.
type BedrockMessage struct {
	Role    string                 `json:"role"`
	Content []BedrockContentBlock `json:"content"`
}

// BedrockContentBlock represents a content block in a Bedrock message.
type BedrockContentBlock struct {
	Type      string             `json:"type"`
	Text      string             `json:"text,omitempty"`
	Source    *BedrockImageSource `json:"source,omitempty"`
}

// BedrockImageSource represents an image source for multimodal input.
type BedrockImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// BedrockResponse represents the response from Claude on Bedrock.
type BedrockResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// ClassificationResponse represents the parsed classification output.
type ClassificationResponse struct {
	DocumentType string `json:"documentType"`
}

// FieldExtractionResponse represents the parsed field extraction output.
type FieldExtractionResponse struct {
	Fields []struct {
		Name       string  `json:"name"`
		Value      string  `json:"value"`
		Confidence float64 `json:"confidence"`
		Normalized string  `json:"normalized,omitempty"`
	} `json:"fields"`
}

// Handle processes a document extraction event and returns an ExtractionResult.
func (h *Handler) Handle(ctx context.Context, event ExtractionEvent) (*models.ExtractionResult, error) {
	startTime := time.Now()

	// Step 1: Parse S3 path and retrieve document
	bucket, key, err := parseS3Path(event.DocumentPath)
	if err != nil {
		return nil, fmt.Errorf("invalid document path: %w", err)
	}

	documentBytes, err := h.S3Client.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve document from S3: %w", err)
	}

	// Step 2: Classify document type using Bedrock
	documentType, err := h.classifyDocument(ctx, documentBytes)
	if err != nil {
		return nil, fmt.Errorf("document classification failed: %w", err)
	}

	// Step 3: Extract fields based on document type
	fields, err := h.extractFields(ctx, documentBytes, documentType)
	if err != nil {
		return nil, fmt.Errorf("field extraction failed: %w", err)
	}

	// Step 4: Check required fields and calculate overall confidence
	requiredFields := models.RequiredFieldsByDocType[documentType]
	overallConfidence := calculateOverallConfidence(fields, requiredFields)

	processingTime := time.Since(startTime).Milliseconds()

	return &models.ExtractionResult{
		DocumentType:      documentType,
		Fields:            fields,
		OverallConfidence: overallConfidence,
		RawText:           "",
		ProcessingTimeMs:  processingTime,
	}, nil
}

// classifyDocument invokes Bedrock to classify the document type.
func (h *Handler) classifyDocument(ctx context.Context, documentBytes []byte) (models.DocumentType, error) {
	prompt := buildClassificationPrompt()

	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        256,
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContentBlock{
					{
						Type: "image",
						Source: &BedrockImageSource{
							Type:      "base64",
							MediaType: detectMediaType(documentBytes),
							Data:      encodeBase64(documentBytes),
						},
					},
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return models.DocumentTypeUnknown, fmt.Errorf("failed to marshal classification request: %w", err)
	}

	responseBytes, err := h.BedrockClient.InvokeModel(ctx, h.ModelID, requestBytes)
	if err != nil {
		return models.DocumentTypeUnknown, fmt.Errorf("bedrock invocation failed: %w", err)
	}

	var response BedrockResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return models.DocumentTypeUnknown, fmt.Errorf("failed to parse bedrock response: %w", err)
	}

	if len(response.Content) == 0 {
		return models.DocumentTypeUnknown, fmt.Errorf("empty response from bedrock")
	}

	return parseDocumentType(response.Content[0].Text), nil
}

// extractFields invokes Bedrock to extract fields based on document type.
func (h *Handler) extractFields(ctx context.Context, documentBytes []byte, docType models.DocumentType) (map[string]models.ExtractedField, error) {
	prompt := buildExtractionPrompt(docType)

	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContentBlock{
					{
						Type: "image",
						Source: &BedrockImageSource{
							Type:      "base64",
							MediaType: detectMediaType(documentBytes),
							Data:      encodeBase64(documentBytes),
						},
					},
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extraction request: %w", err)
	}

	responseBytes, err := h.BedrockClient.InvokeModel(ctx, h.ModelID, requestBytes)
	if err != nil {
		return nil, fmt.Errorf("bedrock invocation failed: %w", err)
	}

	var response BedrockResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse bedrock response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from bedrock")
	}

	return parseExtractionResponse(response.Content[0].Text, docType)
}

// calculateOverallConfidence computes the overall confidence as the minimum of
// all field confidences. Missing required fields are added with confidence 0.0.
func calculateOverallConfidence(fields map[string]models.ExtractedField, requiredFields []string) float64 {
	// Add missing required fields with confidence 0.0
	for _, required := range requiredFields {
		if _, exists := fields[required]; !exists {
			fields[required] = models.ExtractedField{
				Value:      "",
				Confidence: 0.0,
				Normalized: "",
			}
		}
	}

	// If no fields at all, return 0.0
	if len(fields) == 0 {
		return 0.0
	}

	// Overall confidence is the minimum of all field confidences
	minConfidence := 1.0
	for _, field := range fields {
		if field.Confidence < minConfidence {
			minConfidence = field.Confidence
		}
	}

	return minConfidence
}

// parseS3Path extracts bucket and key from an S3 path (s3://bucket/key).
func parseS3Path(path string) (string, string, error) {
	if !strings.HasPrefix(path, "s3://") {
		return "", "", fmt.Errorf("path must start with s3://: %s", path)
	}
	trimmed := strings.TrimPrefix(path, "s3://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid S3 path format, expected s3://bucket/key: %s", path)
	}
	return parts[0], parts[1], nil
}

// buildClassificationPrompt creates the prompt for document classification.
func buildClassificationPrompt() string {
	return `Analyze this document image and classify it into exactly one of the following types:
- INVOICE
- PURCHASE_ORDER
- TRAVEL_VOUCHER
- GRANT_PAYMENT
- CONTRACT_PAYMENT
- UNKNOWN

Respond with ONLY a JSON object in this format:
{"documentType": "TYPE_HERE"}

Do not include any other text or explanation.`
}

// buildExtractionPrompt creates the prompt for field extraction based on document type.
func buildExtractionPrompt(docType models.DocumentType) string {
	var fieldsDescription string
	switch docType {
	case models.DocumentTypeInvoice:
		fieldsDescription = `Extract the following fields from this invoice:
- payee: The name of the entity being paid
- amount: The total payment amount (in currency format)
- invoiceNumber: The invoice identifier/number
- date: The invoice date`
	case models.DocumentTypePurchaseOrder:
		fieldsDescription = `Extract the following fields from this purchase order:
- vendor: The vendor/supplier name
- items: Description of items or services
- totalAmount: The total amount
- poNumber: The purchase order number`
	case models.DocumentTypeTravelVoucher:
		fieldsDescription = `Extract the following fields from this travel voucher:
- traveler: The traveler's name
- dates: The travel dates
- expenses: Description of expenses
- totalClaim: The total claim amount`
	case models.DocumentTypeGrantPayment:
		fieldsDescription = `Extract the following fields from this grant payment document:
- payee: The grant recipient name
- amount: The payment amount
- grantNumber: The grant identifier
- date: The payment date`
	case models.DocumentTypeContractPayment:
		fieldsDescription = `Extract the following fields from this contract payment document:
- payee: The contractor/payee name
- amount: The payment amount
- contractNumber: The contract identifier
- date: The payment date`
	default:
		fieldsDescription = `Extract any identifiable fields from this document including:
- payee/vendor name
- amounts
- dates
- reference numbers`
	}

	return fmt.Sprintf(`%s

For each field, provide a confidence score between 0.0 and 1.0 indicating how certain you are of the extraction.

Respond with ONLY a JSON object in this format:
{"fields": [{"name": "fieldName", "value": "extractedValue", "confidence": 0.95, "normalized": "normalizedValue"}]}

Rules:
- confidence should be 0.95-1.0 for clearly readable fields
- confidence should be 0.7-0.94 for partially readable fields
- confidence should be 0.3-0.69 for poorly readable fields
- confidence should be below 0.3 for guessed fields
- normalized should contain a cleaned version of the value (e.g., currency without symbols for amounts, ISO date format for dates)
- Do not include any other text or explanation.`, fieldsDescription)
}

// parseDocumentType converts a classification response string to a DocumentType.
func parseDocumentType(response string) models.DocumentType {
	// Try to parse as JSON first
	response = strings.TrimSpace(response)
	var classification ClassificationResponse
	if err := json.Unmarshal([]byte(response), &classification); err == nil {
		response = classification.DocumentType
	}

	// Normalize the response
	normalized := strings.ToUpper(strings.TrimSpace(response))

	switch {
	case strings.Contains(normalized, "INVOICE"):
		return models.DocumentTypeInvoice
	case strings.Contains(normalized, "PURCHASE_ORDER"):
		return models.DocumentTypePurchaseOrder
	case strings.Contains(normalized, "TRAVEL_VOUCHER"):
		return models.DocumentTypeTravelVoucher
	case strings.Contains(normalized, "GRANT_PAYMENT"):
		return models.DocumentTypeGrantPayment
	case strings.Contains(normalized, "CONTRACT_PAYMENT"):
		return models.DocumentTypeContractPayment
	default:
		return models.DocumentTypeUnknown
	}
}

// parseExtractionResponse parses the Bedrock extraction response into ExtractedField map.
func parseExtractionResponse(response string, docType models.DocumentType) (map[string]models.ExtractedField, error) {
	response = strings.TrimSpace(response)

	// Try to extract JSON from the response (handle markdown code blocks)
	if idx := strings.Index(response, "{"); idx >= 0 {
		endIdx := strings.LastIndex(response, "}")
		if endIdx > idx {
			response = response[idx : endIdx+1]
		}
	}

	var extracted FieldExtractionResponse
	if err := json.Unmarshal([]byte(response), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse extraction response JSON: %w", err)
	}

	fields := make(map[string]models.ExtractedField)
	for _, f := range extracted.Fields {
		normalized := f.Normalized
		if normalized == "" {
			normalized = f.Value
		}
		fields[f.Name] = models.ExtractedField{
			Value:      f.Value,
			Confidence: f.Confidence,
			Normalized: normalized,
		}
	}

	return fields, nil
}

// detectMediaType determines the MIME type of the document bytes.
func detectMediaType(data []byte) string {
	if len(data) < 4 {
		return "application/pdf"
	}

	// Check magic bytes
	switch {
	case data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46:
		return "application/pdf"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "image/png"
	case data[0] == 0xFF && data[1] == 0xD8:
		return "image/jpeg"
	case (data[0] == 0x49 && data[1] == 0x49) || (data[0] == 0x4D && data[1] == 0x4D):
		return "image/tiff"
	default:
		return "application/pdf"
	}
}

// encodeBase64 encodes bytes to base64 string.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
