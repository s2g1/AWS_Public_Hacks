package models

import "time"

// Decision represents the outcome decision of an agent's processing step.
type Decision string

const (
	DecisionApprove   Decision = "APPROVE"
	DecisionReject    Decision = "REJECT"
	DecisionEscalate  Decision = "ESCALATE"
	DecisionNeedsInfo Decision = "NEEDS_INFO"
)

// TraceContext provides distributed tracing context for an agent message.
type TraceContext struct {
	WorkflowID      string   `json:"workflowId"`
	StepNumber      int      `json:"stepNumber"`
	ParentMessageID string   `json:"parentMessageId"`
	AgentChain      []string `json:"agentChain"`
}

// AgentMessage is the structured envelope for inter-agent communication.
type AgentMessage struct {
	MessageID    string       `json:"messageId"`
	SourceAgent  string       `json:"sourceAgent"`
	TargetAgent  string       `json:"targetAgent"`
	PaymentID    string       `json:"paymentId"`
	MessageType  string       `json:"messageType"`
	Payload      interface{}  `json:"payload"`
	Confidence   float64      `json:"confidence"`
	Decision     Decision     `json:"decision"`
	Timestamp    time.Time    `json:"timestamp"`
	TraceContext TraceContext `json:"traceContext"`
}
