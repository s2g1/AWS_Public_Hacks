package models

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// **Validates: Requirements 14.1, 14.2, 14.4**

// validTransitionChains returns all statuses that can be reached via forward transitions
// from RECEIVED (excluding terminal/error paths for generating longer chains).
var happyPathChain = []PaymentStatus{
	PaymentStatusReceived,
	PaymentStatusExtracting,
	PaymentStatusExtracted,
	PaymentStatusValidating,
	PaymentStatusValidated,
	PaymentStatusCheckingCompliance,
	PaymentStatusCompliant,
	PaymentStatusRouting,
	PaymentStatusRouted,
	PaymentStatusApproving,
	PaymentStatusApproved,
	PaymentStatusDisbursing,
	PaymentStatusDisbursed,
}

// generateValidTransitionSequence generates a random sequence of valid transitions
// starting from RECEIVED using the ValidTransitions map.
func generateValidTransitionSequence(t *rapid.T) []PaymentStatus {
	// Start from RECEIVED and pick random valid next states
	current := PaymentStatusReceived
	sequence := []PaymentStatus{current}

	// Generate between 1 and 12 transitions (max depth of the happy path)
	numTransitions := rapid.IntRange(1, 12).Draw(t, "numTransitions")

	for i := 0; i < numTransitions; i++ {
		nextStates := ValidTransitions[current]
		if len(nextStates) == 0 {
			// Terminal state reached
			break
		}
		// Pick a random next state from valid transitions
		idx := rapid.IntRange(0, len(nextStates)-1).Draw(t, "nextStateIdx")
		next := nextStates[idx]
		sequence = append(sequence, next)
		current = next

		// If we hit a terminal state, stop
		if current.IsTerminal() {
			break
		}
	}

	return sequence
}

// TestProperty_AuditTrailCountMatchesTransitions verifies that after N successful
// transitions, the audit trail has exactly N entries.
func TestProperty_AuditTrailCountMatchesTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sequence := generateValidTransitionSequence(t)

		record := &PaymentRecord{
			PaymentID:  "PAY-PROP-001",
			Status:     PaymentStatusReceived,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AuditTrail: []AuditEntry{},
		}

		successfulTransitions := 0
		for i := 1; i < len(sequence); i++ {
			actor := rapid.StringMatching(`[a-z]+-agent`).Draw(t, "actor")
			reason := rapid.StringMatching(`[a-z ]+`).Draw(t, "reason")
			err := TransitionPayment(record, sequence[i], actor, reason)
			if err == nil {
				successfulTransitions++
			}
		}

		// Property: audit trail has exactly N entries for N successful transitions
		if len(record.AuditTrail) != successfulTransitions {
			t.Fatalf("expected %d audit entries for %d successful transitions, got %d",
				successfulTransitions, successfulTransitions, len(record.AuditTrail))
		}
	})
}

// TestProperty_AuditEntryRecordsCorrectStatuses verifies that each audit entry
// records the correct previous status and new status matching the transition.
func TestProperty_AuditEntryRecordsCorrectStatuses(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sequence := generateValidTransitionSequence(t)

		record := &PaymentRecord{
			PaymentID:  "PAY-PROP-002",
			Status:     PaymentStatusReceived,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AuditTrail: []AuditEntry{},
		}

		// Track expected previous/new statuses
		type expectedTransition struct {
			prev PaymentStatus
			next PaymentStatus
		}
		var expected []expectedTransition

		for i := 1; i < len(sequence); i++ {
			prevStatus := record.Status
			actor := rapid.StringMatching(`[a-z]+-agent`).Draw(t, "actor")
			reason := rapid.StringMatching(`[a-z ]+`).Draw(t, "reason")
			err := TransitionPayment(record, sequence[i], actor, reason)
			if err == nil {
				expected = append(expected, expectedTransition{prev: prevStatus, next: sequence[i]})
			}
		}

		// Property: each audit entry has correct previous and new status
		if len(record.AuditTrail) != len(expected) {
			t.Fatalf("audit trail length %d does not match expected transitions %d",
				len(record.AuditTrail), len(expected))
		}
		for i, entry := range record.AuditTrail {
			if entry.PreviousStatus != expected[i].prev {
				t.Fatalf("entry %d: expected previous status %q, got %q",
					i, expected[i].prev, entry.PreviousStatus)
			}
			if entry.NewStatus != expected[i].next {
				t.Fatalf("entry %d: expected new status %q, got %q",
					i, expected[i].next, entry.NewStatus)
			}
		}
	})
}

// TestProperty_AuditEntryHasNonEmptyActorAndNonZeroTimestamp verifies that every
// audit entry has a non-empty actor and a non-zero timestamp.
func TestProperty_AuditEntryHasNonEmptyActorAndNonZeroTimestamp(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sequence := generateValidTransitionSequence(t)

		record := &PaymentRecord{
			PaymentID:  "PAY-PROP-003",
			Status:     PaymentStatusReceived,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AuditTrail: []AuditEntry{},
		}

		for i := 1; i < len(sequence); i++ {
			actor := rapid.StringMatching(`[a-z]+-agent`).Draw(t, "actor")
			reason := rapid.StringMatching(`[a-z ]+`).Draw(t, "reason")
			TransitionPayment(record, sequence[i], actor, reason)
		}

		// Property: every audit entry has non-empty actor and non-zero timestamp
		for i, entry := range record.AuditTrail {
			if entry.Actor == "" {
				t.Fatalf("entry %d: actor is empty", i)
			}
			if entry.Timestamp.IsZero() {
				t.Fatalf("entry %d: timestamp is zero", i)
			}
		}
	})
}

// TestProperty_AuditTrailChronologicalOrder verifies that audit trail entries
// are in chronological order (each timestamp >= previous).
func TestProperty_AuditTrailChronologicalOrder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sequence := generateValidTransitionSequence(t)

		record := &PaymentRecord{
			PaymentID:  "PAY-PROP-004",
			Status:     PaymentStatusReceived,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AuditTrail: []AuditEntry{},
		}

		for i := 1; i < len(sequence); i++ {
			actor := rapid.StringMatching(`[a-z]+-agent`).Draw(t, "actor")
			reason := rapid.StringMatching(`[a-z ]+`).Draw(t, "reason")
			TransitionPayment(record, sequence[i], actor, reason)
		}

		// Property: timestamps are in chronological order
		for i := 1; i < len(record.AuditTrail); i++ {
			if record.AuditTrail[i].Timestamp.Before(record.AuditTrail[i-1].Timestamp) {
				t.Fatalf("entry %d timestamp %v is before entry %d timestamp %v",
					i, record.AuditTrail[i].Timestamp, i-1, record.AuditTrail[i-1].Timestamp)
			}
		}
	})
}

// TestProperty_AuditTrailCountGteTransitions verifies that the count of audit
// trail entries is always >= the number of status transitions (Requirement 14.4).
func TestProperty_AuditTrailCountGteTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sequence := generateValidTransitionSequence(t)

		record := &PaymentRecord{
			PaymentID:  "PAY-PROP-005",
			Status:     PaymentStatusReceived,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			AuditTrail: []AuditEntry{},
		}

		transitionCount := 0
		for i := 1; i < len(sequence); i++ {
			actor := rapid.StringMatching(`[a-z]+-agent`).Draw(t, "actor")
			reason := rapid.StringMatching(`[a-z ]+`).Draw(t, "reason")
			err := TransitionPayment(record, sequence[i], actor, reason)
			if err == nil {
				transitionCount++
			}
		}

		// Property: audit trail entries >= number of status transitions (Req 14.4)
		if len(record.AuditTrail) < transitionCount {
			t.Fatalf("audit trail count %d is less than transition count %d",
				len(record.AuditTrail), transitionCount)
		}
	})
}
