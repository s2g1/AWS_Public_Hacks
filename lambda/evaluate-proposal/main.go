package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// EvaluationRequest is the input payload from the frontend
type EvaluationRequest struct {
	ProposalText    string `json:"proposalText"`
	DocumentBase64  string `json:"documentBase64"`
	DocumentName    string `json:"documentName"`
	SolicitationSOW string `json:"solicitationSOW"`
	PriceProposal   float64 `json:"priceProposal"`
	CompanyName     string `json:"companyName"`
}

// CLINItem represents a contract line item
type CLINItem struct {
	CLINNumber  string  `json:"clinNumber"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Ceiling     float64 `json:"ceiling"`
	Obligated   float64 `json:"obligated"`
	Expended    float64 `json:"expended"`
}

// EvaluationResponse is the output returned to the frontend
type EvaluationResponse struct {
	Summary        string     `json:"summary"`
	CLINBreakdown  []CLINItem `json:"clinBreakdown"`
	BOEAllocation  string     `json:"boeAllocation"`
	Score          int        `json:"score"`
	Recommendation string     `json:"recommendation"`
}

// BedrockRequest for Claude
type BedrockRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []Message `json:"messages"`
}

type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type BedrockResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// CORS headers
	headers := map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "POST, OPTIONS",
	}

	// Handle preflight
	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.LambdaFunctionURLResponse{StatusCode: 200, Headers: headers}, nil
	}

	// Parse request body
	var req EvaluationRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body: "+err.Error(), headers), nil
	}

	// Build the prompt for Bedrock
	prompt := buildEvaluationPrompt(req)

	// Invoke Bedrock
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return errorResponse(500, "Failed to load AWS config: "+err.Error(), headers), nil
	}

	client := bedrockruntime.NewFromConfig(cfg)
	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "anthropic.claude-3-5-sonnet-20241022-v2:0"
	}

	// Build content blocks
	contentBlocks := []ContentBlock{}

	// If document is provided, include it
	if req.DocumentBase64 != "" {
		mediaType := detectMediaType(req.DocumentBase64)
		contentBlocks = append(contentBlocks, ContentBlock{
			Type: "image",
			Source: &ImageSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      req.DocumentBase64,
			},
		})
	}

	contentBlocks = append(contentBlocks, ContentBlock{
		Type: "text",
		Text: prompt,
	})

	bedrockReq := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Messages: []Message{
			{
				Role:    "user",
				Content: contentBlocks,
			},
		},
	}

	reqBytes, err := json.Marshal(bedrockReq)
	if err != nil {
		return errorResponse(500, "Failed to marshal Bedrock request: "+err.Error(), headers), nil
	}

	result, err := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &modelID,
		ContentType: strPtr("application/json"),
		Body:        reqBytes,
	})
	if err != nil {
		return errorResponse(500, "Bedrock invocation failed: "+err.Error(), headers), nil
	}

	// Parse Bedrock response
	var bedrockResp BedrockResponse
	if err := json.Unmarshal(result.Body, &bedrockResp); err != nil {
		return errorResponse(500, "Failed to parse Bedrock response: "+err.Error(), headers), nil
	}

	if len(bedrockResp.Content) == 0 {
		return errorResponse(500, "Empty response from Bedrock", headers), nil
	}

	// Parse the evaluation from Claude's response
	evaluation, err := parseEvaluation(bedrockResp.Content[0].Text, req.PriceProposal)
	if err != nil {
		return errorResponse(500, "Failed to parse evaluation: "+err.Error(), headers), nil
	}

	respBytes, _ := json.Marshal(evaluation)
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(respBytes),
	}, nil
}

func buildEvaluationPrompt(req EvaluationRequest) string {
	var sb strings.Builder
	sb.WriteString(`You are a federal acquisition specialist evaluating a contractor proposal against a Statement of Work (SOW).

Analyze the proposal and provide a structured evaluation. You MUST respond with ONLY valid JSON, no other text.

`)

	if req.SolicitationSOW != "" {
		sb.WriteString(fmt.Sprintf("**Statement of Work / Requirements:**\n%s\n\n", req.SolicitationSOW))
	}

	if req.ProposalText != "" {
		sb.WriteString(fmt.Sprintf("**Proposal Technical Approach:**\n%s\n\n", req.ProposalText))
	}

	if req.CompanyName != "" {
		sb.WriteString(fmt.Sprintf("**Proposing Company:** %s\n\n", req.CompanyName))
	}

	totalPrice := req.PriceProposal
	if totalPrice <= 0 {
		totalPrice = 2500000 // Default estimate
	}

	sb.WriteString(fmt.Sprintf("**Total Proposed Price:** $%.2f\n\n", totalPrice))

	sb.WriteString(fmt.Sprintf(`Generate a realistic evaluation. The CLIN ceiling values MUST sum to exactly $%.2f.

Respond with ONLY this JSON structure:
{
  "summary": "2-3 sentence evaluation summary mentioning strengths and any concerns",
  "clinBreakdown": [
    {"clinNumber": "0001", "description": "Research & Development", "type": "CPFF", "ceiling": <number>},
    {"clinNumber": "0002", "description": "System Integration & Testing", "type": "CPFF", "ceiling": <number>},
    {"clinNumber": "0003", "description": "Program Management", "type": "FFP", "ceiling": <number>}
  ],
  "boeAllocation": "R&D: X%% ($amount) | Integration: Y%% ($amount) | PM: Z%% ($amount) | Total: $total",
  "score": <integer 70-98>,
  "recommendation": "APPROVE" or "REVIEW" or "REJECT"
}

Rules:
- CLINs MUST sum to exactly the total proposed price
- Score 85+ = APPROVE, 70-84 = REVIEW, below 70 = REJECT
- If proposal text is strong and aligned with SOW, score higher
- Include specific references to the proposal content in summary
- Do NOT include any text outside the JSON object`, totalPrice))

	return sb.String()
}

func parseEvaluation(response string, priceProposal float64) (*EvaluationResponse, error) {
	response = strings.TrimSpace(response)

	// Extract JSON from potential markdown code blocks
	if idx := strings.Index(response, "{"); idx >= 0 {
		endIdx := strings.LastIndex(response, "}")
		if endIdx > idx {
			response = response[idx : endIdx+1]
		}
	}

	var eval EvaluationResponse
	if err := json.Unmarshal([]byte(response), &eval); err != nil {
		return nil, fmt.Errorf("failed to parse evaluation JSON: %w (raw: %s)", err, response[:min(200, len(response))])
	}

	// Ensure CLINs have obligated/expended set to 0
	for i := range eval.CLINBreakdown {
		eval.CLINBreakdown[i].Obligated = 0
		eval.CLINBreakdown[i].Expended = 0
	}

	return &eval, nil
}

func detectMediaType(base64Data string) string {
	// Decode first few bytes to detect type
	if len(base64Data) < 8 {
		return "application/pdf"
	}
	decoded, err := base64.StdEncoding.DecodeString(base64Data[:min(16, len(base64Data))])
	if err != nil || len(decoded) < 4 {
		return "application/pdf"
	}
	switch {
	case decoded[0] == 0x25 && decoded[1] == 0x50: // %P
		return "application/pdf"
	case decoded[0] == 0x89 && decoded[1] == 0x50: // PNG
		return "image/png"
	case decoded[0] == 0xFF && decoded[1] == 0xD8: // JPEG
		return "image/jpeg"
	default:
		return "application/pdf"
	}
}

func errorResponse(code int, msg string, headers map[string]string) events.LambdaFunctionURLResponse {
	body, _ := json.Marshal(map[string]string{"error": msg})
	return events.LambdaFunctionURLResponse{
		StatusCode: code,
		Headers:    headers,
		Body:       string(body),
	}
}

func strPtr(s string) *string { return &s }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	lambda.Start(handler)
}
