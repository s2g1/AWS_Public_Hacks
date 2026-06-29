package correspondence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CorrespondenceType classifies the type of correspondence to generate.
type CorrespondenceType string

const (
	CorrespondenceTypeApprovalConfirmation  CorrespondenceType = "APPROVAL_CONFIRMATION"
	CorrespondenceTypeRejectionNotice       CorrespondenceType = "REJECTION_NOTICE"
	CorrespondenceTypeREAResponse           CorrespondenceType = "REA_RESPONSE"
	CorrespondenceTypeEscalationNotification CorrespondenceType = "ESCALATION_NOTIFICATION"
)

// OutputFormat specifies the format of the generated correspondence.
type OutputFormat string

const (
	OutputFormatEmailHTML          OutputFormat = "EMAIL_HTML"
	OutputFormatPDFContent         OutputFormat = "PDF_CONTENT"
	OutputFormatPortalNotification OutputFormat = "PORTAL_NOTIFICATION"
)

// CorrespondenceStatus tracks the lifecycle state of generated correspondence.
type CorrespondenceStatus string

const (
	CorrespondenceStatusDraft         CorrespondenceStatus = "DRAFT"
	CorrespondenceStatusPendingReview CorrespondenceStatus = "PENDING_REVIEW"
	CorrespondenceStatusSent          CorrespondenceStatus = "SENT"
)

// CorrespondenceRequest contains the input data needed to generate correspondence.
type CorrespondenceRequest struct {
	Type              CorrespondenceType `json:"type"`
	Recipient         string             `json:"recipient"`
	PaymentID         string             `json:"paymentId"`
	ContractNumber    string             `json:"contractNumber"`
	Amount            string             `json:"amount"`
	Reason            string             `json:"reason"`
	AdditionalContext map[string]string  `json:"additionalContext,omitempty"`
}

// GeneratedCorrespondence contains the output of the correspondence generation.
type GeneratedCorrespondence struct {
	Subject     string               `json:"subject"`
	Body        string               `json:"body"`
	Format      OutputFormat         `json:"format"`
	GeneratedAt time.Time            `json:"generatedAt"`
	Status      CorrespondenceStatus `json:"status"`
}

// BedrockClient defines the interface for invoking Bedrock models.
// This interface allows for mocking in tests.
type BedrockClient interface {
	InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error)
}

// BedrockRequest represents the request payload for Claude on Bedrock.
type BedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	Messages         []BedrockMessage `json:"messages"`
}

// BedrockMessage represents a message in the Bedrock conversation.
type BedrockMessage struct {
	Role    string                `json:"role"`
	Content []BedrockContentBlock `json:"content"`
}

// BedrockContentBlock represents a content block in a Bedrock message.
type BedrockContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// BedrockResponse represents the response from Claude on Bedrock.
type BedrockResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// CorrespondenceGenerator generates professional correspondence using Amazon Bedrock.
type CorrespondenceGenerator struct {
	bedrockClient BedrockClient
	modelID       string
}

// NewCorrespondenceGenerator creates a new CorrespondenceGenerator with the given Bedrock client.
func NewCorrespondenceGenerator(client BedrockClient, modelID string) *CorrespondenceGenerator {
	return &CorrespondenceGenerator{
		bedrockClient: client,
		modelID:       modelID,
	}
}

// GenerateCorrespondence generates professional correspondence based on the request type and format.
func (g *CorrespondenceGenerator) GenerateCorrespondence(ctx context.Context, req CorrespondenceRequest, format OutputFormat) (*GeneratedCorrespondence, error) {
	// Validate required fields.
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	// Build the prompt based on correspondence type.
	prompt := buildCorrespondencePrompt(req, format)

	// Invoke Bedrock to generate the correspondence.
	bedrockReq := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []BedrockContentBlock{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	requestBytes, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bedrock request: %w", err)
	}

	responseBytes, err := g.bedrockClient.InvokeModel(ctx, g.modelID, requestBytes)
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

	// Parse the generated correspondence from the response.
	generated, err := parseCorrespondenceResponse(response.Content[0].Text, req, format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse correspondence response: %w", err)
	}

	return generated, nil
}

// validateRequest checks that all required fields are populated in the request.
func validateRequest(req CorrespondenceRequest) error {
	if req.Type == "" {
		return fmt.Errorf("correspondence type is required")
	}
	if req.Recipient == "" {
		return fmt.Errorf("recipient is required")
	}
	if req.PaymentID == "" {
		return fmt.Errorf("payment ID is required")
	}

	// Validate correspondence type is a known value.
	switch req.Type {
	case CorrespondenceTypeApprovalConfirmation,
		CorrespondenceTypeRejectionNotice,
		CorrespondenceTypeREAResponse,
		CorrespondenceTypeEscalationNotification:
		// valid
	default:
		return fmt.Errorf("unknown correspondence type: %s", req.Type)
	}

	return nil
}

// buildCorrespondencePrompt constructs the prompt for Bedrock based on correspondence type.
func buildCorrespondencePrompt(req CorrespondenceRequest, format OutputFormat) string {
	var typeInstruction string

	switch req.Type {
	case CorrespondenceTypeApprovalConfirmation:
		typeInstruction = fmt.Sprintf(
			`Generate a professional payment approval confirmation letter.
The letter should inform the recipient that their payment has been approved.

Key details:
- Recipient: %s
- Payment ID: %s
- Contract Number: %s
- Amount: %s

The tone should be formal and positive. Include:
- Confirmation that the payment of %s has been approved
- Reference to payment ID and contract number
- Expected disbursement timeline
- Contact information for questions`,
			req.Recipient, req.PaymentID, req.ContractNumber, req.Amount, req.Amount)

	case CorrespondenceTypeRejectionNotice:
		typeInstruction = fmt.Sprintf(
			`Generate a professional payment rejection notice.
The letter should inform the recipient that their payment has been rejected with specific reasons and next steps.

Key details:
- Recipient: %s
- Payment ID: %s
- Contract Number: %s
- Amount: %s
- Rejection Reason: %s

The tone should be formal and respectful. Include:
- Clear statement that the payment has been rejected
- Specific reason(s) for rejection
- Next steps the recipient can take (e.g., resubmission, appeal process)
- Contact information for questions or disputes
- Timeline for any appeal process`,
			req.Recipient, req.PaymentID, req.ContractNumber, req.Amount, req.Reason)

	case CorrespondenceTypeREAResponse:
		typeInstruction = fmt.Sprintf(
			`Generate formal government correspondence responding to a Request for Equitable Adjustment (REA).
The letter should follow federal government correspondence style and conventions.

Key details:
- Recipient (Contractor): %s
- Payment/REA ID: %s
- Contract Number: %s
- Amount: %s
- Response/Decision: %s

The tone should be formal government correspondence style. Include:
- Reference to the original REA submission
- The government's decision and rationale
- Any contract modifications resulting from the decision
- Next steps and timeline
- Official signature block placeholder`,
			req.Recipient, req.PaymentID, req.ContractNumber, req.Amount, req.Reason)

	case CorrespondenceTypeEscalationNotification:
		typeInstruction = fmt.Sprintf(
			`Generate an urgent escalation notification requiring immediate action.
The letter should alert the reviewer that a payment requires their attention.

Key details:
- Recipient (Reviewer): %s
- Payment ID: %s
- Contract Number: %s
- Amount: %s
- Escalation Reason: %s

The tone should be urgent and action-oriented. Include:
- Clear statement that immediate action is needed
- The specific reason for escalation
- What action the reviewer needs to take
- Deadline or urgency level
- Link/reference to the payment for review`,
			req.Recipient, req.PaymentID, req.ContractNumber, req.Amount, req.Reason)
	}

	// Add format instructions.
	var formatInstruction string
	switch format {
	case OutputFormatEmailHTML:
		formatInstruction = `Format the output as an HTML email body. Use appropriate HTML tags for structure (paragraphs, headings, etc.). Do not include <html>, <head>, or <body> tags - just the inner content.`
	case OutputFormatPDFContent:
		formatInstruction = `Format the output as plain text suitable for PDF generation. Use proper spacing, formal letter structure with date, address block, salutation, body paragraphs, and closing.`
	case OutputFormatPortalNotification:
		formatInstruction = `Format the output as a concise portal notification. Keep it brief (2-3 sentences) with the essential information and a call to action.`
	}

	// Add additional context if provided.
	additionalCtx := ""
	if len(req.AdditionalContext) > 0 {
		additionalCtx = "\n\nAdditional context:\n"
		for key, value := range req.AdditionalContext {
			additionalCtx += fmt.Sprintf("- %s: %s\n", key, value)
		}
	}

	return fmt.Sprintf(`%s
%s
%s

Respond with ONLY a JSON object in this format:
{"subject": "Email subject line", "body": "The full correspondence body"}

Do not include any other text or explanation outside the JSON.`, typeInstruction, formatInstruction, additionalCtx)
}

// correspondenceResponseJSON represents the parsed response from Bedrock.
type correspondenceResponseJSON struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// parseCorrespondenceResponse parses the Bedrock response into a GeneratedCorrespondence.
func parseCorrespondenceResponse(response string, req CorrespondenceRequest, format OutputFormat) (*GeneratedCorrespondence, error) {
	// Try to extract JSON from the response (handle markdown code blocks).
	cleaned := response
	if idx := findJSONStart(cleaned); idx >= 0 {
		endIdx := findJSONEnd(cleaned, idx)
		if endIdx > idx {
			cleaned = cleaned[idx : endIdx+1]
		}
	}

	var parsed correspondenceResponseJSON
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse correspondence response JSON: %w", err)
	}

	if parsed.Subject == "" || parsed.Body == "" {
		return nil, fmt.Errorf("correspondence response missing subject or body")
	}

	return &GeneratedCorrespondence{
		Subject:     parsed.Subject,
		Body:        parsed.Body,
		Format:      format,
		GeneratedAt: time.Now(),
		Status:      CorrespondenceStatusDraft,
	}, nil
}

// findJSONStart finds the index of the first '{' in the string.
func findJSONStart(s string) int {
	for i, c := range s {
		if c == '{' {
			return i
		}
	}
	return -1
}

// findJSONEnd finds the index of the last '}' in the string.
func findJSONEnd(s string, startIdx int) int {
	lastIdx := -1
	for i := startIdx; i < len(s); i++ {
		if s[i] == '}' {
			lastIdx = i
		}
	}
	return lastIdx
}
