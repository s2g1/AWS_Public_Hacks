package models

import "time"

// ApprovalLevel represents the required level of authority for payment approval.
type ApprovalLevel string

const (
	ApprovalLevelPurchaseCard             ApprovalLevel = "PURCHASE_CARD"
	ApprovalLevelSupervisor               ApprovalLevel = "SUPERVISOR"
	ApprovalLevelContractingOfficer       ApprovalLevel = "CONTRACTING_OFFICER"
	ApprovalLevelSeniorContractingOfficer ApprovalLevel = "SENIOR_CONTRACTING_OFFICER"
	ApprovalLevelAgencyHead               ApprovalLevel = "AGENCY_HEAD"
)

// ApprovalLevelOrder defines the ranking of approval levels from lowest to highest.
// Used for elevation logic.
var ApprovalLevelOrder = []ApprovalLevel{
	ApprovalLevelPurchaseCard,
	ApprovalLevelSupervisor,
	ApprovalLevelContractingOfficer,
	ApprovalLevelSeniorContractingOfficer,
	ApprovalLevelAgencyHead,
}

// Priority represents the urgency level of a payment routing decision.
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityNormal Priority = "NORMAL"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

// PriorityOrder defines the ranking of priorities from lowest to highest.
// Used for elevation logic.
var PriorityOrder = []Priority{
	PriorityLow,
	PriorityNormal,
	PriorityHigh,
	PriorityUrgent,
}

// RoutingStatus represents the outcome of the routing decision.
type RoutingStatus string

const (
	RoutingStatusRouted    RoutingStatus = "ROUTED"
	RoutingStatusEscalated RoutingStatus = "ESCALATED"
)

// RoutingDecision contains the complete output of the routing agent.
type RoutingDecision struct {
	Status        RoutingStatus `json:"status"`
	Approver      string        `json:"approver,omitempty"`
	ApprovalLevel ApprovalLevel `json:"approvalLevel"`
	Priority      Priority      `json:"priority"`
	Rationale     string        `json:"rationale"`
	RoutedAt      time.Time     `json:"routedAt"`
}
