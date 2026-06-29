package routing

import (
	"fmt"
	"time"

	"federal-payment-processing/internal/models"
)

// Approver represents a person with authority to approve payments at a given level.
type Approver struct {
	Name              string
	IsOnLeave         bool
	DelegationExpired bool
	Delegate          *Approver
}

// ApproverRegistry provides access to approvers by approval level.
type ApproverRegistry interface {
	GetApprover(level models.ApprovalLevel) (*Approver, error)
}

// DetermineRouteWithDelegation extends routing by checking approver availability
// and delegating or escalating as needed per Requirements 11.1, 11.2.
func (h *RoutingHandler) DetermineRouteWithDelegation(
	extraction *models.ExtractionResult,
	compliance *models.ComplianceResult,
	registry ApproverRegistry,
) (*models.RoutingDecision, error) {
	amount, err := parseAmount(extraction)
	if err != nil {
		return nil, fmt.Errorf("routing: failed to parse amount: %w", err)
	}

	approvalLevel, priority := determineApprovalLevelAndPriority(amount)

	// Look up the primary approver for the determined level.
	approver, err := registry.GetApprover(approvalLevel)
	if err != nil {
		return nil, fmt.Errorf("routing: failed to look up approver: %w", err)
	}

	assignedApprover := approver

	// Check if the primary approver is unavailable (on leave or delegation expired).
	if approver.IsOnLeave || approver.DelegationExpired {
		if approver.Delegate != nil {
			// Route to the designated delegate.
			assignedApprover = approver.Delegate
		} else {
			// No delegate available: escalate with URGENT priority.
			return &models.RoutingDecision{
				Status:        models.RoutingStatusEscalated,
				ApprovalLevel: approvalLevel,
				Priority:      models.PriorityUrgent,
				Rationale: fmt.Sprintf(
					"No available approver at %s level: primary approver %s is unavailable and no delegate is designated. Escalated with URGENT priority, admin notified.",
					approvalLevel, approver.Name,
				),
				RoutedAt: time.Now(),
			}, nil
		}
	}

	rationale := buildRationale(amount, approvalLevel, priority)

	return &models.RoutingDecision{
		Status:        models.RoutingStatusRouted,
		Approver:      assignedApprover.Name,
		ApprovalLevel: approvalLevel,
		Priority:      priority,
		Rationale:     rationale,
		RoutedAt:      time.Now(),
	}, nil
}
