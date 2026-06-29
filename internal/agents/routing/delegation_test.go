package routing

import (
	"fmt"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockApproverRegistry is a test double for ApproverRegistry.
type mockApproverRegistry struct {
	approvers map[models.ApprovalLevel]*Approver
}

func (m *mockApproverRegistry) GetApprover(level models.ApprovalLevel) (*Approver, error) {
	approver, ok := m.approvers[level]
	if !ok {
		return nil, fmt.Errorf("no approver configured for level %s", level)
	}
	return approver, nil
}

// TestDelegation_ApproverAvailable verifies that when the primary approver is
// available, the payment routes normally to that approver.
func TestDelegation_ApproverAvailable(t *testing.T) {
	registry := &mockApproverRegistry{
		approvers: map[models.ApprovalLevel]*Approver{
			models.ApprovalLevelSupervisor: {
				Name:              "Jane Smith",
				IsOnLeave:         false,
				DelegationExpired: false,
				Delegate:          nil,
			},
		},
	}

	handler := NewRoutingHandler()
	extraction := extractionWithAmount("5000.00")

	decision, err := handler.DetermineRouteWithDelegation(extraction, compliantResult(), registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Status != models.RoutingStatusRouted {
		t.Errorf("expected ROUTED status, got %s", decision.Status)
	}
	if decision.Approver != "Jane Smith" {
		t.Errorf("expected approver 'Jane Smith', got %q", decision.Approver)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR level, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %s", decision.Priority)
	}
}

// TestDelegation_ApproverOnLeave_DelegateAvailable verifies that when the
// primary approver is on leave but a delegate exists, the payment routes to
// the delegate.
func TestDelegation_ApproverOnLeave_DelegateAvailable(t *testing.T) {
	delegate := &Approver{
		Name:              "Bob Johnson",
		IsOnLeave:         false,
		DelegationExpired: false,
		Delegate:          nil,
	}

	registry := &mockApproverRegistry{
		approvers: map[models.ApprovalLevel]*Approver{
			models.ApprovalLevelSupervisor: {
				Name:              "Jane Smith",
				IsOnLeave:         true,
				DelegationExpired: false,
				Delegate:          delegate,
			},
		},
	}

	handler := NewRoutingHandler()
	extraction := extractionWithAmount("5000.00")

	decision, err := handler.DetermineRouteWithDelegation(extraction, compliantResult(), registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Status != models.RoutingStatusRouted {
		t.Errorf("expected ROUTED status, got %s", decision.Status)
	}
	if decision.Approver != "Bob Johnson" {
		t.Errorf("expected delegate 'Bob Johnson', got %q", decision.Approver)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR level, got %s", decision.ApprovalLevel)
	}
}

// TestDelegation_DelegationExpired_NoDelegate verifies that when the primary
// approver has an expired delegation and no delegate is available, the payment
// is escalated with URGENT priority.
func TestDelegation_DelegationExpired_NoDelegate(t *testing.T) {
	registry := &mockApproverRegistry{
		approvers: map[models.ApprovalLevel]*Approver{
			models.ApprovalLevelContractingOfficer: {
				Name:              "Alice Brown",
				IsOnLeave:         false,
				DelegationExpired: true,
				Delegate:          nil,
			},
		},
	}

	handler := NewRoutingHandler()
	extraction := extractionWithAmount("50000.00")

	decision, err := handler.DetermineRouteWithDelegation(extraction, compliantResult(), registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Status != models.RoutingStatusEscalated {
		t.Errorf("expected ESCALATED status, got %s", decision.Status)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority, got %s", decision.Priority)
	}
	if decision.Approver != "" {
		t.Errorf("expected empty approver for escalated decision, got %q", decision.Approver)
	}
}

// TestDelegation_ApproverOnLeave_NoDelegate verifies that when the primary
// approver is on leave and no delegate is available, the payment is escalated
// with URGENT priority.
func TestDelegation_ApproverOnLeave_NoDelegate(t *testing.T) {
	registry := &mockApproverRegistry{
		approvers: map[models.ApprovalLevel]*Approver{
			models.ApprovalLevelSupervisor: {
				Name:              "Jane Smith",
				IsOnLeave:         true,
				DelegationExpired: false,
				Delegate:          nil,
			},
		},
	}

	handler := NewRoutingHandler()
	extraction := extractionWithAmount("10000.00")

	decision, err := handler.DetermineRouteWithDelegation(extraction, compliantResult(), registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Status != models.RoutingStatusEscalated {
		t.Errorf("expected ESCALATED status, got %s", decision.Status)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority, got %s", decision.Priority)
	}
	if decision.Approver != "" {
		t.Errorf("expected empty approver for escalated decision, got %q", decision.Approver)
	}
	if decision.Rationale == "" {
		t.Error("expected non-empty rationale explaining escalation reason")
	}
}
