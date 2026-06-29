package compliance

import (
	"context"
	"encoding/json"
	"fmt"

	"federal-payment-processing/internal/models"
)

// BedrockClient is an interface for invoking Amazon Bedrock models,
// enabling mocking in tests.
type BedrockClient interface {
	// InvokeModel sends a payload to the specified Bedrock model and returns the response.
	InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error)
}

// FARChecker evaluates payments against Federal Acquisition Regulation (FAR) rules
// using Amazon Bedrock Claude Sonnet for rule interpretation.
type FARChecker struct {
	bedrockClient BedrockClient
}

// NewFARChecker creates a new FARChecker with the given Bedrock client.
func NewFARChecker(client BedrockClient) *FARChecker {
	return &FARChecker{bedrockClient: client}
}

// BedrockModelID is the model identifier for Claude Sonnet used in FAR evaluation.
const BedrockModelID = "anthropic.claude-sonnet-20241022-v2:0"

// bedrockRequest represents the request payload sent to Bedrock Claude.
type bedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	Messages         []bedrockMessage `json:"messages"`
}

// bedrockMessage represents a single message in the Bedrock conversation.
type bedrockMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// bedrockResponse represents the response from Bedrock Claude.
type bedrockResponse struct {
	Content []bedrockContentBlock `json:"content"`
}

// bedrockContentBlock represents a content block in the Bedrock response.
type bedrockContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// farEvaluationResponse represents the expected JSON response from the FAR evaluation.
type farEvaluationResponse struct {
	Flags []farFlag `json:"flags"`
}

// farFlag represents a single FAR compliance flag from the model response.
type farFlag struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// CheckFARCompliance evaluates a payment against FAR rules using Bedrock Claude Sonnet.
// It builds a prompt with payment details, invokes the model, and parses the response
// into compliance flags.
func (f *FARChecker) CheckFARCompliance(ctx context.Context, extraction *models.ExtractionResult) ([]models.ComplianceFlag, error) {
	if f.bedrockClient == nil {
		return nil, fmt.Errorf("bedrock client is not configured")
	}

	// Extract payment details from the extraction result.
	payee := getFieldValue(extraction, "payee")
	amount := getFieldValue(extraction, "amount")
	category := determineCategory(extraction)

	// Build the FAR check prompt.
	prompt := buildFARCheckPrompt(payee, amount, category)

	// Build Bedrock request payload.
	reqPayload := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1024,
		Messages: []bedrockMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bedrock request: %w", err)
	}

	// Invoke Bedrock model.
	responseBytes, err := f.bedrockClient.InvokeModel(ctx, BedrockModelID, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke bedrock model: %w", err)
	}

	// Parse response.
	flags, err := parseFARResponse(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FAR evaluation response: %w", err)
	}

	return flags, nil
}

// buildFARCheckPrompt constructs the prompt for FAR compliance evaluation.
func buildFARCheckPrompt(payee, amount, category string) string {
	return fmt.Sprintf(`You are a Federal Acquisition Regulation (FAR) compliance evaluator. Evaluate the following federal payment for FAR rule compliance.

Payment Details:
- Payee: %s
- Amount: %s
- Spend Category: %s

Evaluate this payment against applicable FAR rules for the given spend category and amount. Consider:
1. FAR Part 13 (Simplified Acquisition Procedures) for amounts under the simplified acquisition threshold
2. FAR Part 15 (Contracting by Negotiation) for larger procurements
3. FAR Part 19 (Small Business Programs) set-aside requirements
4. FAR Part 31 (Contract Cost Principles) for cost allowability
5. FAR Part 32 (Contract Financing) for payment timing and methods

Respond ONLY with a JSON object in the following format (no additional text):
{
  "flags": [
    {
      "rule": "FAR_RULE_IDENTIFIER",
      "severity": "BLOCKING" or "REQUIRES_REVIEW",
      "message": "Description of the compliance concern"
    }
  ]
}

If no FAR violations are found, respond with:
{"flags": []}`, payee, amount, category)
}

// parseFARResponse parses the Bedrock response into compliance flags.
func parseFARResponse(responseBytes []byte) ([]models.ComplianceFlag, error) {
	// Parse the Bedrock response envelope.
	var response bedrockResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bedrock response: %w", err)
	}

	// Extract text content from the response.
	var responseText string
	for _, block := range response.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	if responseText == "" {
		// No text content in response; treat as no flags.
		return []models.ComplianceFlag{}, nil
	}

	// Parse the JSON from the response text.
	var farResult farEvaluationResponse
	if err := json.Unmarshal([]byte(responseText), &farResult); err != nil {
		return nil, fmt.Errorf("failed to parse FAR evaluation JSON: %w", err)
	}

	// Convert farFlags to models.ComplianceFlag with appropriate severity mapping.
	var flags []models.ComplianceFlag
	for _, f := range farResult.Flags {
		severity := mapFARSeverity(f.Severity)
		flags = append(flags, models.ComplianceFlag{
			Rule:     f.Rule,
			Severity: severity,
			Message:  f.Message,
		})
	}

	return flags, nil
}

// mapFARSeverity maps a string severity from the model response to a FlagSeverity value.
func mapFARSeverity(severity string) models.FlagSeverity {
	switch severity {
	case "BLOCKING":
		return models.FlagSeverityBlocking
	case "REQUIRES_REVIEW":
		return models.FlagSeverityRequiresReview
	default:
		// Default to REQUIRES_REVIEW for unknown severities to avoid blocking payments
		// without explicit blocking determination.
		return models.FlagSeverityRequiresReview
	}
}

// getFieldValue extracts the normalized (or raw) value from an extraction field.
func getFieldValue(extraction *models.ExtractionResult, fieldName string) string {
	field, ok := extraction.Fields[fieldName]
	if !ok {
		return ""
	}
	if field.Normalized != "" {
		return field.Normalized
	}
	return field.Value
}

// determineCategory infers the spend category from the document type and fields.
func determineCategory(extraction *models.ExtractionResult) string {
	switch extraction.DocumentType {
	case models.DocumentTypeInvoice:
		return "SUPPLIES_AND_SERVICES"
	case models.DocumentTypePurchaseOrder:
		return "PURCHASE_ORDER"
	case models.DocumentTypeTravelVoucher:
		return "TRAVEL"
	case models.DocumentTypeGrantPayment:
		return "GRANT"
	case models.DocumentTypeContractPayment:
		return "CONTRACT"
	default:
		return "GENERAL"
	}
}
