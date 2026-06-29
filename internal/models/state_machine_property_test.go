package models

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 13.1, 13.3**

// allStatuses returns all defined PaymentStatus values from the ValidTransitions map.
func allStatuses() []PaymentStatus {
	statuses := make([]PaymentStatus, 0, len(ValidTransitions))
	for s := range ValidTransitions {
		statuses = append(statuses, s)
	}
	return statuses
}

// genPaymentStatus generates an arbitrary valid PaymentStatus from the state machine.
func genPaymentStatus() *rapid.Generator[PaymentStatus] {
	return rapid.Custom(func(t *rapid.T) PaymentStatus {
		statuses := allStatuses()
		idx := rapid.IntRange(0, len(statuses)-1).Draw(t, "statusIndex")
		return statuses[idx]
	})
}

// TestProperty_ValidTransitionsAccepted verifies that for any valid status and any
// allowed next status from ValidTransitions, ValidateTransition returns nil.
func TestProperty_ValidTransitionsAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		current := genPaymentStatus().Draw(t, "current")
		allowed := ValidTransitions[current]
		if len(allowed) == 0 {
			t.Skip("no transitions from terminal state")
		}
		nextIdx := rapid.IntRange(0, len(allowed)-1).Draw(t, "nextIndex")
		next := allowed[nextIdx]

		err := ValidateTransition(current, next)
		if err != nil {
			t.Fatalf("expected valid transition from %q to %q to succeed, got error: %v", current, next, err)
		}
	})
}

// TestProperty_InvalidTransitionsRejected verifies that for any valid status and any
// status NOT in its allowed transitions, ValidateTransition returns an error.
func TestProperty_InvalidTransitionsRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		current := genPaymentStatus().Draw(t, "current")
		target := genPaymentStatus().Draw(t, "target")

		// Check if target is in the allowed set for current
		allowed := ValidTransitions[current]
		isAllowed := false
		for _, s := range allowed {
			if s == target {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			t.Skip("target is in allowed transitions, not testing invalid case")
		}

		err := ValidateTransition(current, target)
		if err == nil {
			t.Fatalf("expected transition from %q to %q to be rejected (not in allowed set), but got nil error", current, target)
		}
	})
}

// TestProperty_TerminalStatesHaveNoTransitions verifies that terminal states
// (DISBURSED, REJECTED, FAILED) have no valid transitions.
// ESCALATED is not terminal and can resume processing.
func TestProperty_TerminalStatesHaveNoTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		terminalStatuses := []PaymentStatus{
			PaymentStatusDisbursed,
			PaymentStatusRejected,
			PaymentStatusFailed,
		}
		idx := rapid.IntRange(0, len(terminalStatuses)-1).Draw(t, "terminalIndex")
		terminal := terminalStatuses[idx]

		// Terminal states must have empty transition sets
		transitions := ValidTransitions[terminal]
		if len(transitions) != 0 {
			t.Fatalf("terminal state %q should have no transitions, but has %v", terminal, transitions)
		}

		// Any transition attempt from a terminal state should fail
		target := genPaymentStatus().Draw(t, "target")
		err := ValidateTransition(terminal, target)
		if err == nil {
			t.Fatalf("expected transition from terminal state %q to %q to be rejected, but got nil error", terminal, target)
		}
	})
}

// TestProperty_StateGraphWellFormed verifies the state machine graph is well-formed:
// every status referenced in transition targets is itself a key in ValidTransitions.
func TestProperty_StateGraphWellFormed(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		current := genPaymentStatus().Draw(t, "current")
		allowed := ValidTransitions[current]
		if len(allowed) == 0 {
			t.Skip("no transitions to verify")
		}
		nextIdx := rapid.IntRange(0, len(allowed)-1).Draw(t, "nextIndex")
		target := allowed[nextIdx]

		// Every target in the transitions must be a defined state (key in ValidTransitions)
		if _, exists := ValidTransitions[target]; !exists {
			t.Fatalf("transition target %q from state %q is not a defined state in ValidTransitions", target, current)
		}
	})
}

// TestProperty_NoSelfTransitions verifies that no state can transition to itself.
func TestProperty_NoSelfTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		current := genPaymentStatus().Draw(t, "current")
		allowed := ValidTransitions[current]

		for _, target := range allowed {
			if target == current {
				t.Fatalf("state %q has a self-transition, which is not allowed", current)
			}
		}
	})
}
