package chatbot

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// mockBedrockChatClient is a test double for the BedrockChatClient interface.
type mockBedrockChatClient struct {
	response    string
	err         error
	lastPayload []byte
}

func (m *mockBedrockChatClient) InvokeModel(_ context.Context, _ string, payload []byte) ([]byte, error) {
	m.lastPayload = payload
	if m.err != nil {
		return nil, m.err
	}
	resp := BedrockChatResponse{
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{Type: "text", Text: m.response},
		},
	}
	return json.Marshal(resp)
}

func TestHandleChat_BasicMessageResponseFlow(t *testing.T) {
	mock := &mockBedrockChatClient{response: "Hello! How can I help you?"}
	handler := NewChatHandler(mock, "test-model", "Test knowledge base content")

	req := ChatRequest{
		Message:        "Hi there",
		SessionID:      "session-1",
		CurrentPage:    "/dashboard",
		UserID:         "user-1",
		UserRole:       "CONTRACTING_OFFICER",
		OrganizationID: "org-1",
	}

	resp, err := handler.HandleChat(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Response != "Hello! How can I help you?" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
}

func TestHandleChat_SessionContextMaintained(t *testing.T) {
	mock := &mockBedrockChatClient{response: "First response"}
	handler := NewChatHandler(mock, "test-model", "Knowledge")

	sessionID := "session-persist"

	// First message
	req1 := ChatRequest{
		Message:        "What is a CLIN?",
		SessionID:      sessionID,
		CurrentPage:    "/contracts",
		UserID:         "user-1",
		UserRole:       "CONTRACTOR",
		OrganizationID: "org-abc",
	}
	_, err := handler.HandleChat(context.Background(), req1)
	if err != nil {
		t.Fatalf("first message failed: %v", err)
	}

	// Verify session has messages
	session := handler.GetSession(sessionID)
	if len(session) != 2 {
		t.Fatalf("expected 2 messages in session after first exchange, got %d", len(session))
	}
	if session[0].Role != "user" || session[0].Content != "What is a CLIN?" {
		t.Errorf("unexpected first message: %+v", session[0])
	}
	if session[1].Role != "assistant" || session[1].Content != "First response" {
		t.Errorf("unexpected second message: %+v", session[1])
	}

	// Second message - session should include prior context
	mock.response = "Second response"
	req2 := ChatRequest{
		Message:        "Tell me more",
		SessionID:      sessionID,
		CurrentPage:    "/contracts",
		UserID:         "user-1",
		UserRole:       "CONTRACTOR",
		OrganizationID: "org-abc",
	}
	_, err = handler.HandleChat(context.Background(), req2)
	if err != nil {
		t.Fatalf("second message failed: %v", err)
	}

	// Verify the Bedrock payload included previous messages
	var bedrockReq BedrockChatRequest
	if err := json.Unmarshal(mock.lastPayload, &bedrockReq); err != nil {
		t.Fatalf("failed to unmarshal bedrock request: %v", err)
	}

	// Should have 3 messages: user1, assistant1, user2
	if len(bedrockReq.Messages) != 3 {
		t.Fatalf("expected 3 messages in bedrock request, got %d", len(bedrockReq.Messages))
	}
	if bedrockReq.Messages[0].Content != "What is a CLIN?" {
		t.Errorf("expected first message content to be preserved, got: %s", bedrockReq.Messages[0].Content)
	}
}

func TestHandleChat_SystemPromptIncludesRBACConstraints(t *testing.T) {
	mock := &mockBedrockChatClient{response: "Response with RBAC"}
	handler := NewChatHandler(mock, "test-model", "Knowledge base")

	req := ChatRequest{
		Message:        "Show me contracts",
		SessionID:      "session-rbac",
		CurrentPage:    "/contracts",
		UserID:         "user-1",
		UserRole:       "CONTRACTOR",
		OrganizationID: "org-xyz",
	}

	_, err := handler.HandleChat(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the system prompt sent to Bedrock
	var bedrockReq BedrockChatRequest
	if err := json.Unmarshal(mock.lastPayload, &bedrockReq); err != nil {
		t.Fatalf("failed to unmarshal bedrock request: %v", err)
	}

	if bedrockReq.System == "" {
		t.Fatal("expected system prompt to be non-empty")
	}

	// Verify RBAC section exists
	if !strings.Contains(bedrockReq.System, "Data Access Constraints") {
		t.Error("system prompt should contain RBAC constraints section")
	}
	if !strings.Contains(bedrockReq.System, "org-xyz") {
		t.Error("system prompt should reference the user's organization ID")
	}
}

func TestHandleChat_COGetsAllContractsAccessible(t *testing.T) {
	mock := &mockBedrockChatClient{response: "CO response"}
	handler := NewChatHandler(mock, "test-model", "Knowledge")

	req := ChatRequest{
		Message:        "List all contracts",
		SessionID:      "session-co",
		CurrentPage:    "/contracts",
		UserID:         "co-user",
		UserRole:       "CONTRACTING_OFFICER",
		OrganizationID: "org-gov",
	}

	_, err := handler.HandleChat(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bedrockReq BedrockChatRequest
	if err := json.Unmarshal(mock.lastPayload, &bedrockReq); err != nil {
		t.Fatalf("failed to unmarshal bedrock request: %v", err)
	}

	// CO should get "all contracts are accessible" in the RBAC section
	if !strings.Contains(bedrockReq.System, "all contracts are accessible") {
		t.Error("CO system prompt should contain 'all contracts are accessible'")
	}
}

func TestHandleChat_ContractorGetsOrgScopedConstraint(t *testing.T) {
	mock := &mockBedrockChatClient{response: "Contractor response"}
	handler := NewChatHandler(mock, "test-model", "Knowledge")

	req := ChatRequest{
		Message:        "Show my invoices",
		SessionID:      "session-contractor",
		CurrentPage:    "/payments",
		UserID:         "contractor-user",
		UserRole:       "CONTRACTOR",
		OrganizationID: "org-acme",
	}

	_, err := handler.HandleChat(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bedrockReq BedrockChatRequest
	if err := json.Unmarshal(mock.lastPayload, &bedrockReq); err != nil {
		t.Fatalf("failed to unmarshal bedrock request: %v", err)
	}

	// Contractor should get org-scoped constraint
	if !strings.Contains(bedrockReq.System, "org-acme") {
		t.Error("Contractor system prompt should reference their org ID")
	}
	if !strings.Contains(bedrockReq.System, "organizationId matches") {
		t.Error("Contractor system prompt should contain org-scoped constraint language")
	}
	// Should NOT have "all contracts are accessible"
	if strings.Contains(bedrockReq.System, "all contracts are accessible") {
		t.Error("Contractor system prompt should NOT contain 'all contracts are accessible'")
	}
}
