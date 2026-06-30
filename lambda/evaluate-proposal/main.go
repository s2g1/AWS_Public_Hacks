package main

import (
	"context"
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
	ProposalText    string  `json:"proposalText"`
	DocumentBase64  string  `json:"documentBase64"`
	DocumentName    string  `json:"documentName"`
	SolicitationSOW string  `json:"solicitationSOW"`
	PriceProposal   float64 `json:"priceProposal"`
	CompanyName     string  `json:"companyName"`
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

// --- Anthropic Claude API format (used by Claude Sonnet 4.6) ---

type ClaudeRequest struct {
	AnthropicVersion string         `json:"anthropic_version"`
	MaxTokens        int            `json:"max_tokens"`
	Messages         []ClaudeMessage `json:"messages"`
}

type ClaudeMessage struct {
	Role    string               `json:"role"`
	Content []ClaudeContentBlock `json:"content"`
}

type ClaudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.LambdaFunctionURLResponse{StatusCode: 200, Headers: headers}, nil
	}

	// Parse request body
	var req EvaluationRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body: "+err.Error(), headers), nil
	}

	// Build the prompt
	prompt := buildEvaluationPrompt(req)

	// Invoke Bedrock with Claude Sonnet 4.6 (global cross-region inference)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return errorResponse(500, "Failed to load AWS config: "+err.Error(), headers), nil
	}

	client := bedrockruntime.NewFromConfig(cfg)
	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "global.anthropic.claude-sonnet-4-6"
	}

	claudeReq := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Messages: []ClaudeMessage{
			{
				Role: "user",
				Content: []ClaudeContentBlock{
					{Type: "text", Text: prompt},
				},
			},
		},
	}

	reqBytes, err := json.Marshal(claudeReq)
	if err != nil {
		return errorResponse(500, "Failed to marshal request: "+err.Error(), headers), nil
	}

	result, err := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &modelID,
		ContentType: strPtr("application/json"),
		Body:        reqBytes,
	})
	if err != nil {
		return errorResponse(500, "Bedrock invocation failed: "+err.Error(), headers), nil
	}

	// Parse Claude response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(result.Body, &claudeResp); err != nil {
		return errorResponse(500, "Failed to parse Bedrock response: "+err.Error(), headers), nil
	}

	if len(claudeResp.Content) == 0 {
		return errorResponse(500, "Empty response from Bedrock", headers), nil
	}

	responseText := claudeResp.Content[0].Text

	// Parse the evaluation JSON
	evaluation, err := parseEvaluation(responseText, req.PriceProposal)
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
		sb.WriteString(fmt.Sprintf("Statement of Work / Requirements:\n%s\n\n", req.SolicitationSOW))
	}

	if req.ProposalText != "" {
		sb.WriteString(fmt.Sprintf("Proposal Technical Approach:\n%s\n\n", req.ProposalText))
	}

	if req.CompanyName != "" {
		sb.WriteString(fmt.Sprintf("Proposing Company: %s\n\n", req.CompanyName))
	}

	totalPrice := req.PriceProposal
	if totalPrice <= 0 {
		totalPrice = 2500000
	}

	sb.WriteString(fmt.Sprintf("Total Proposed Price: $%.2f\n\n", totalPrice))

	sb.WriteString(fmt.Sprintf(`Generate a realistic evaluation. The CLIN ceiling values MUST sum to exactly $%.2f.

Respond with ONLY this JSON structure (no markdown, no code fences, no explanation):
{"summary":"2-3 sentence evaluation summary referencing the proposal content","clinBreakdown":[{"clinNumber":"0001","description":"Research & Development","type":"CPFF","ceiling":<number>},{"clinNumber":"0002","description":"System Integration & Testing","type":"CPFF","ceiling":<number>},{"clinNumber":"0003","description":"Program Management","type":"FFP","ceiling":<number>}],"boeAllocation":"R&D: X%% ($amount) | Integration: Y%% ($amount) | PM: Z%% ($amount) | Total: $total","score":<integer 75-95>,"recommendation":"APPROVE"}

Rules:
- CLINs MUST sum to exactly the total proposed price
- Score should be between 75-95
- Score 85+ = APPROVE, 70-84 = REVIEW
- Reference the proposal content in your summary
- Output ONLY the JSON object, nothing else`, totalPrice))

	return sb.String()
}

func parseEvaluation(response string, priceProposal float64) (*EvaluationResponse, error) {
	response = strings.TrimSpace(response)

	// Strip markdown code blocks if present
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Extract JSON
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
