package correspondence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockBedrockClient is a test double for the BedrockClient interface.
type mockBedrockClient struct {
	// responseFunc allows tests to customize the response based on the request.
	responseFunc func(ctx context.Context, modelID string, payload []byte) ([]byte, error)
}

func (m *mockBedrockClient) InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
	if m.responseFunc != nil {
		return m.responseFunc(ctx, modelID, payload)
	}
	return nil, fmt.Errorf("no response configured")
}

// newMockClientWithResponse creates a mock that returns a predefined correspondence response.
func newMockClientWithResponse(subject, body string) *mockBedrockClient {
	return &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			respJSON := fmt.Sprintf(`{"subject": %q, "body": %q}`, subject, body)
			bedrockResp := BedrockResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: respJSON},
				},
			}
			return json.Marshal(bedrockResp)
		},
	}
}

// newMockClientThatCaptures creates a mock that captures the request and returns a valid response.
func newMockClientThatCaptures(captured *[]byte) *mockBedrockClient {
	return &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			*captured = append([]byte{}, payload...)
			respJSON := `{"subject": "Test Subject", "body": "Test body content"}`
			bedrockResp := BedrockResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: respJSON},
				},
			}
			return json.Marshal(bedrockResp)
		},
	}
}

func TestGenerateCorrespondence_ApprovalConfirmation(t *testing.T) {
	subject := "Payment Approval Confirmation - PAY-001"
	body := "Dear Acme Corp, your payment of $50,000.00 has been approved and will be disbursed shortly."

	mock := newMockClientWithResponse(subject, body)
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:           CorrespondenceTypeApprovalConfirmation,
		Recipient:      "Acme Corp",
		PaymentID:      "PAY-001",
		ContractNumber: "W91234-21-C-0001",
		Amount:         "$50,000.00",
	}

	result, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Subject != subject {
		t.Errorf("expected subject %q, got %q", subject, result.Subject)
	}
	if result.Body != body {
		t.Errorf("expected body %q, got %q", body, result.Body)
	}
	if result.Format != OutputFormatEmailHTML {
		t.Errorf("expected format %q, got %q", OutputFormatEmailHTML, result.Format)
	}
	if result.Status != CorrespondenceStatusDraft {
		t.Errorf("expected status DRAFT, got %q", result.Status)
	}
	if result.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
}

func TestGenerateCorrespondence_RejectionNotice(t *testing.T) {
	subject := "Payment Rejection Notice - PAY-002"
	body := "Dear Tech Solutions Inc, your payment request has been rejected. Reason: Missing invoice documentation. Next steps: Please resubmit with complete documentation."

	mock := newMockClientWithResponse(subject, body)
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:           CorrespondenceTypeRejectionNotice,
		Recipient:      "Tech Solutions Inc",
		PaymentID:      "PAY-002",
		ContractNumber: "FA8750-22-C-0002",
		Amount:         "$125,000.00",
		Reason:         "Missing invoice documentation",
	}

	result, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatPDFContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Subject != subject {
		t.Errorf("expected subject %q, got %q", subject, result.Subject)
	}
	if result.Format != OutputFormatPDFContent {
		t.Errorf("expected format %q, got %q", OutputFormatPDFContent, result.Format)
	}
	if result.Status != CorrespondenceStatusDraft {
		t.Errorf("expected status DRAFT, got %q", result.Status)
	}
}

func TestGenerateCorrespondence_REAResponse(t *testing.T) {
	subject := "Response to Request for Equitable Adjustment - REA-003"
	body := "Dear Defense Systems LLC, this letter is in response to your Request for Equitable Adjustment submitted under contract N00024-20-C-5312. The government has reviewed your request and approved the adjustment."

	mock := newMockClientWithResponse(subject, body)
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:           CorrespondenceTypeREAResponse,
		Recipient:      "Defense Systems LLC",
		PaymentID:      "REA-003",
		ContractNumber: "N00024-20-C-5312",
		Amount:         "$250,000.00",
		Reason:         "Approved - scope change justified",
	}

	result, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatPDFContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Subject != subject {
		t.Errorf("expected subject %q, got %q", subject, result.Subject)
	}
	if result.Format != OutputFormatPDFContent {
		t.Errorf("expected format %q, got %q", OutputFormatPDFContent, result.Format)
	}
}

func TestGenerateCorrespondence_EscalationNotification(t *testing.T) {
	subject := "URGENT: Payment Escalation Requires Review - PAY-004"
	body := "Attention: A payment of $1,500,000.00 requires your immediate review. The payment has been escalated due to: Amount exceeds senior contracting officer threshold."

	mock := newMockClientWithResponse(subject, body)
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:           CorrespondenceTypeEscalationNotification,
		Recipient:      "John Smith, Agency Head",
		PaymentID:      "PAY-004",
		ContractNumber: "GS-35F-0001T",
		Amount:         "$1,500,000.00",
		Reason:         "Amount exceeds senior contracting officer threshold",
	}

	result, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatPortalNotification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Subject != subject {
		t.Errorf("expected subject %q, got %q", subject, result.Subject)
	}
	if result.Format != OutputFormatPortalNotification {
		t.Errorf("expected format %q, got %q", OutputFormatPortalNotification, result.Format)
	}
}

func TestGenerateCorrespondence_DifferentOutputFormats(t *testing.T) {
	formats := []OutputFormat{
		OutputFormatEmailHTML,
		OutputFormatPDFContent,
		OutputFormatPortalNotification,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var captured []byte
			mock := newMockClientThatCaptures(&captured)
			gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

			req := CorrespondenceRequest{
				Type:           CorrespondenceTypeApprovalConfirmation,
				Recipient:      "Test Corp",
				PaymentID:      "PAY-FMT-001",
				ContractNumber: "C-001",
				Amount:         "$10,000.00",
			}

			result, err := gen.GenerateCorrespondence(context.Background(), req, format)
			if err != nil {
				t.Fatalf("unexpected error for format %s: %v", format, err)
			}

			if result.Format != format {
				t.Errorf("expected format %q, got %q", format, result.Format)
			}

			// Verify the prompt includes format-specific instructions.
			var bedrockReq BedrockRequest
			if err := json.Unmarshal(captured, &bedrockReq); err != nil {
				t.Fatalf("failed to unmarshal captured request: %v", err)
			}

			promptText := bedrockReq.Messages[0].Content[0].Text
			switch format {
			case OutputFormatEmailHTML:
				if !strings.Contains(promptText, "HTML") {
					t.Error("email HTML format prompt should mention HTML")
				}
			case OutputFormatPDFContent:
				if !strings.Contains(promptText, "PDF") {
					t.Error("PDF format prompt should mention PDF")
				}
			case OutputFormatPortalNotification:
				if !strings.Contains(promptText, "portal notification") {
					t.Error("portal notification format prompt should mention portal notification")
				}
			}
		})
	}
}

func TestGenerateCorrespondence_MissingRequiredFields(t *testing.T) {
	mock := newMockClientWithResponse("subject", "body")
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	tests := []struct {
		name    string
		req     CorrespondenceRequest
		wantErr string
	}{
		{
			name: "missing type",
			req: CorrespondenceRequest{
				Recipient: "Test",
				PaymentID: "PAY-001",
			},
			wantErr: "correspondence type is required",
		},
		{
			name: "missing recipient",
			req: CorrespondenceRequest{
				Type:      CorrespondenceTypeApprovalConfirmation,
				PaymentID: "PAY-001",
			},
			wantErr: "recipient is required",
		},
		{
			name: "missing payment ID",
			req: CorrespondenceRequest{
				Type:      CorrespondenceTypeApprovalConfirmation,
				Recipient: "Test",
			},
			wantErr: "payment ID is required",
		},
		{
			name: "invalid type",
			req: CorrespondenceRequest{
				Type:      "INVALID_TYPE",
				Recipient: "Test",
				PaymentID: "PAY-001",
			},
			wantErr: "unknown correspondence type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.GenerateCorrespondence(context.Background(), tt.req, OutputFormatEmailHTML)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestGenerateCorrespondence_BedrockError(t *testing.T) {
	mock := &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			return nil, fmt.Errorf("service unavailable")
		},
	}
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:      CorrespondenceTypeApprovalConfirmation,
		Recipient: "Test Corp",
		PaymentID: "PAY-001",
	}

	_, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "bedrock invocation failed") {
		t.Errorf("expected bedrock error, got: %v", err)
	}
}

func TestGenerateCorrespondence_EmptyBedrockResponse(t *testing.T) {
	mock := &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			resp := BedrockResponse{Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{}}
			return json.Marshal(resp)
		},
	}
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:      CorrespondenceTypeApprovalConfirmation,
		Recipient: "Test Corp",
		PaymentID: "PAY-001",
	}

	_, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "empty response from bedrock") {
		t.Errorf("expected empty response error, got: %v", err)
	}
}

func TestGenerateCorrespondence_AdditionalContext(t *testing.T) {
	var captured []byte
	mock := newMockClientThatCaptures(&captured)
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:           CorrespondenceTypeApprovalConfirmation,
		Recipient:      "Test Corp",
		PaymentID:      "PAY-CTX-001",
		ContractNumber: "C-001",
		Amount:         "$10,000.00",
		AdditionalContext: map[string]string{
			"approvedBy": "Jane Doe",
			"department": "Defense Logistics",
		},
	}

	_, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the prompt includes additional context.
	var bedrockReq BedrockRequest
	if err := json.Unmarshal(captured, &bedrockReq); err != nil {
		t.Fatalf("failed to unmarshal captured request: %v", err)
	}

	promptText := bedrockReq.Messages[0].Content[0].Text
	if !strings.Contains(promptText, "approvedBy") {
		t.Error("prompt should include additional context key 'approvedBy'")
	}
	if !strings.Contains(promptText, "Jane Doe") {
		t.Error("prompt should include additional context value 'Jane Doe'")
	}
}

func TestGenerateCorrespondence_PromptContainsRecipient(t *testing.T) {
	types := []CorrespondenceType{
		CorrespondenceTypeApprovalConfirmation,
		CorrespondenceTypeRejectionNotice,
		CorrespondenceTypeREAResponse,
		CorrespondenceTypeEscalationNotification,
	}

	for _, corrType := range types {
		t.Run(string(corrType), func(t *testing.T) {
			var captured []byte
			mock := newMockClientThatCaptures(&captured)
			gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

			req := CorrespondenceRequest{
				Type:           corrType,
				Recipient:      "Unique Recipient Name",
				PaymentID:      "PAY-TYPE-001",
				ContractNumber: "C-001",
				Amount:         "$75,000.00",
				Reason:         "Test reason for correspondence",
			}

			_, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var bedrockReq BedrockRequest
			if err := json.Unmarshal(captured, &bedrockReq); err != nil {
				t.Fatalf("failed to unmarshal captured request: %v", err)
			}

			promptText := bedrockReq.Messages[0].Content[0].Text
			if !strings.Contains(promptText, "Unique Recipient Name") {
				t.Errorf("prompt for %s should contain recipient name", corrType)
			}
			if !strings.Contains(promptText, "PAY-TYPE-001") {
				t.Errorf("prompt for %s should contain payment ID", corrType)
			}
		})
	}
}

func TestGenerateCorrespondence_InvalidJSON(t *testing.T) {
	mock := &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			bedrockResp := BedrockResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "This is not valid JSON at all"},
				},
			}
			return json.Marshal(bedrockResp)
		},
	}
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:      CorrespondenceTypeApprovalConfirmation,
		Recipient: "Test Corp",
		PaymentID: "PAY-001",
	}

	_, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "failed to parse correspondence response") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestGenerateCorrespondence_ResponseWithCodeBlock(t *testing.T) {
	// Test that the parser handles responses wrapped in markdown code blocks.
	mock := &mockBedrockClient{
		responseFunc: func(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
			// Simulate Bedrock returning JSON inside a code block.
			text := "```json\n{\"subject\": \"Approval Notice\", \"body\": \"Your payment has been approved.\"}\n```"
			bedrockResp := BedrockResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: text},
				},
			}
			return json.Marshal(bedrockResp)
		},
	}
	gen := NewCorrespondenceGenerator(mock, "anthropic.claude-3-sonnet-20240229-v1:0")

	req := CorrespondenceRequest{
		Type:      CorrespondenceTypeApprovalConfirmation,
		Recipient: "Test Corp",
		PaymentID: "PAY-001",
	}

	result, err := gen.GenerateCorrespondence(context.Background(), req, OutputFormatEmailHTML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Subject != "Approval Notice" {
		t.Errorf("expected subject 'Approval Notice', got %q", result.Subject)
	}
	if result.Body != "Your payment has been approved." {
		t.Errorf("expected body 'Your payment has been approved.', got %q", result.Body)
	}
}
