package chatbot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", or "system"
	Content string `json:"content"` // The message text
}

// ChatRequest represents an incoming chat request from a user.
type ChatRequest struct {
	Message        string `json:"message"`
	SessionID      string `json:"sessionId"`
	CurrentPage    string `json:"currentPage"`
	UserID         string `json:"userId"`
	UserRole       string `json:"userRole"`
	OrganizationID string `json:"organizationId"`
}

// ChatResponse represents the chatbot's response to a user message.
type ChatResponse struct {
	Response string   `json:"response"`
	Sources  []string `json:"sources,omitempty"`
}

// BedrockChatClient defines the interface for invoking Bedrock models for chat.
type BedrockChatClient interface {
	InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error)
}

// BedrockChatRequest represents the request payload for Claude on Bedrock.
type BedrockChatRequest struct {
	AnthropicVersion string              `json:"anthropic_version"`
	MaxTokens        int                 `json:"max_tokens"`
	System           string              `json:"system,omitempty"`
	Messages         []BedrockChatMsg    `json:"messages"`
}

// BedrockChatMsg represents a message in the Bedrock conversation.
type BedrockChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// BedrockChatResponse represents the response from Claude on Bedrock.
type BedrockChatResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// ChatHandler manages chatbot interactions with Bedrock Claude.
type ChatHandler struct {
	bedrockClient BedrockChatClient
	modelID       string
	knowledgeBase string
	mu            sync.RWMutex
	sessions      map[string][]ChatMessage
}

// NewChatHandler creates a new ChatHandler instance.
func NewChatHandler(client BedrockChatClient, modelID string, knowledgeBase string) *ChatHandler {
	return &ChatHandler{
		bedrockClient: client,
		modelID:       modelID,
		knowledgeBase: knowledgeBase,
		sessions:      make(map[string][]ChatMessage),
	}
}

// HandleChat processes a chat request and returns a response.
func (h *ChatHandler) HandleChat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Message == "" {
		return nil, fmt.Errorf("message cannot be empty")
	}
	if req.SessionID == "" {
		return nil, fmt.Errorf("sessionId cannot be empty")
	}

	// Load or create session
	h.mu.Lock()
	session, exists := h.sessions[req.SessionID]
	if !exists {
		session = []ChatMessage{}
	}
	h.mu.Unlock()

	// Build system prompt
	systemPrompt := h.buildSystemPrompt(req)

	// Append user message to session
	userMsg := ChatMessage{
		Role:    "user",
		Content: req.Message,
	}
	session = append(session, userMsg)

	// Build Bedrock messages from session (only user/assistant messages)
	bedrockMessages := make([]BedrockChatMsg, 0, len(session))
	for _, msg := range session {
		if msg.Role == "user" || msg.Role == "assistant" {
			bedrockMessages = append(bedrockMessages, BedrockChatMsg{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// Invoke Bedrock Claude
	bedrockReq := BedrockChatRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		System:           systemPrompt,
		Messages:         bedrockMessages,
	}

	reqBytes, err := json.Marshal(bedrockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bedrock request: %w", err)
	}

	respBytes, err := h.bedrockClient.InvokeModel(ctx, h.modelID, reqBytes)
	if err != nil {
		return nil, fmt.Errorf("bedrock invocation failed: %w", err)
	}

	// Parse response
	var bedrockResp BedrockChatResponse
	if err := json.Unmarshal(respBytes, &bedrockResp); err != nil {
		return nil, fmt.Errorf("failed to parse bedrock response: %w", err)
	}

	if len(bedrockResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from bedrock")
	}

	responseText := bedrockResp.Content[0].Text

	// Append assistant message to session
	assistantMsg := ChatMessage{
		Role:    "assistant",
		Content: responseText,
	}
	session = append(session, assistantMsg)

	// Save session
	h.mu.Lock()
	h.sessions[req.SessionID] = session
	h.mu.Unlock()

	// Extract sources from response if any
	sources := extractSources(responseText)

	return &ChatResponse{
		Response: responseText,
		Sources:  sources,
	}, nil
}

// buildSystemPrompt constructs the system prompt with knowledge base, page context, and RBAC constraints.
func (h *ChatHandler) buildSystemPrompt(req ChatRequest) string {
	var sb strings.Builder

	// Base knowledge base
	sb.WriteString("You are a helpful assistant for the Federal Payment Processing Platform. ")
	sb.WriteString("You help users understand the application, navigate features, and answer questions about their data.\n\n")

	// App documentation context
	if h.knowledgeBase != "" {
		sb.WriteString("## Application Documentation\n")
		sb.WriteString(h.knowledgeBase)
		sb.WriteString("\n\n")
	}

	// Current page context
	if req.CurrentPage != "" {
		sb.WriteString("## Current Page Context\n")
		sb.WriteString(fmt.Sprintf("The user is currently viewing the '%s' page. ", req.CurrentPage))
		sb.WriteString("Provide contextual help relevant to this page when appropriate.\n\n")
	}

	// User role context
	if req.UserRole != "" {
		sb.WriteString("## User Role\n")
		sb.WriteString(fmt.Sprintf("The user has the role: %s. ", req.UserRole))
		sb.WriteString(buildRoleDescription(req.UserRole))
		sb.WriteString("\n\n")
	}

	// RBAC constraints
	sb.WriteString("## Data Access Constraints (RBAC)\n")
	if req.UserRole == "CONTRACTING_OFFICER" {
		sb.WriteString("For CO role, all contracts are accessible. ")
		sb.WriteString(fmt.Sprintf("Only provide data for contracts in the user's assigned portfolio. User organizationId: %s.\n", req.OrganizationID))
	} else {
		sb.WriteString(fmt.Sprintf("Only provide data for contracts where organizationId matches %s. ", req.OrganizationID))
		sb.WriteString("Do NOT expose data from contracts the user doesn't have access to. ")
		sb.WriteString("If asked about data outside their access, inform the user that they do not have permission to view that information.\n")
	}

	sb.WriteString("\n## Response Guidelines\n")
	sb.WriteString("- Be concise and helpful\n")
	sb.WriteString("- Reference specific page features when helping with navigation\n")
	sb.WriteString("- When providing data, cite the source (e.g., contract ID, payment ID)\n")
	sb.WriteString("- Never fabricate data - if you don't have the information, say so\n")

	return sb.String()
}

// buildRoleDescription returns a description of what each role can do.
func buildRoleDescription(role string) string {
	switch role {
	case "CONTRACTING_OFFICER":
		return "This role can view all contracts in their portfolio, respond to REAs, exercise options, and manage obligations."
	case "COR":
		return "This role can view contracts associated with their organization."
	case "PROCURING_CONTRACTING_OFFICER":
		return "This role can view contracts in their organization, submit REAs, update EAC, and submit invoices."
	case "PROGRAM_MANAGER":
		return "This role can view contracts in their organization, submit REAs, and update EAC."
	case "CONTRACTOR":
		return "This role can view contracts associated with their organization, submit REAs, update EAC, and submit invoices."
	default:
		return "Role permissions are limited to viewing accessible data."
	}
}

// extractSources attempts to extract cited sources from the response text.
func extractSources(response string) []string {
	var sources []string
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Source:") || strings.HasPrefix(trimmed, "- Source:") {
			source := strings.TrimPrefix(trimmed, "- ")
			source = strings.TrimPrefix(source, "Source:")
			source = strings.TrimSpace(source)
			if source != "" {
				sources = append(sources, source)
			}
		}
	}
	return sources
}

// GetSession returns the conversation history for a given session ID.
func (h *ChatHandler) GetSession(sessionID string) []ChatMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, exists := h.sessions[sessionID]
	if !exists {
		return nil
	}
	// Return a copy
	result := make([]ChatMessage, len(session))
	copy(result, session)
	return result
}
