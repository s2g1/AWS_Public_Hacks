package extraction

import (
	"testing"

	"federal-payment-processing/internal/models"
)

func TestShouldEscalateForHandwriting_NoFields(t *testing.T) {
	escalate, fields := ShouldEscalateForHandwriting(nil)
	if escalate {
		t.Error("expected no escalation for nil fields")
	}
	if fields != nil {
		t.Error("expected nil field list for nil input")
	}
}

func TestShouldEscalateForHandwriting_NoHandwrittenFields(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "Acme Corp",
			Confidence:    0.95,
			IsHandwritten: false,
		},
		"amount": {
			Value:         "$1000.00",
			Confidence:    0.90,
			IsHandwritten: false,
		},
	}

	escalate, names := ShouldEscalateForHandwriting(fields)
	if escalate {
		t.Error("expected no escalation when no handwritten fields")
	}
	if names != nil {
		t.Error("expected nil names when no escalation needed")
	}
}

func TestShouldEscalateForHandwriting_HandwrittenAboveThreshold(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.72, // Above HandwritingThreshold (0.65)
			IsHandwritten: true,
		},
		"amount": {
			Value:         "$500.00",
			Confidence:    0.90,
			IsHandwritten: false,
		},
	}

	escalate, names := ShouldEscalateForHandwriting(fields)
	if escalate {
		t.Error("expected no escalation when handwritten field is above threshold")
	}
	if names != nil {
		t.Error("expected nil names when no escalation needed")
	}
}

func TestShouldEscalateForHandwriting_HandwrittenBelowThreshold(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "J?hn D?e",
			Confidence:    0.55, // Below HandwritingThreshold (0.65)
			IsHandwritten: true,
		},
		"amount": {
			Value:         "$500.00",
			Confidence:    0.90,
			IsHandwritten: false,
		},
	}

	escalate, names := ShouldEscalateForHandwriting(fields)
	if !escalate {
		t.Error("expected escalation when handwritten field is below threshold")
	}
	if len(names) != 1 || names[0] != "payee" {
		t.Errorf("expected [payee], got %v", names)
	}
}

func TestShouldEscalateForHandwriting_MultipleHandwrittenBelowThreshold(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "unclear",
			Confidence:    0.40,
			IsHandwritten: true,
		},
		"amount": {
			Value:         "?500",
			Confidence:    0.50,
			IsHandwritten: true,
		},
		"date": {
			Value:         "2024-01-15",
			Confidence:    0.95,
			IsHandwritten: false,
		},
	}

	escalate, names := ShouldEscalateForHandwriting(fields)
	if !escalate {
		t.Error("expected escalation for multiple low-confidence handwritten fields")
	}
	if len(names) != 2 {
		t.Errorf("expected 2 flagged fields, got %d: %v", len(names), names)
	}

	// Check both fields are in the list (order not guaranteed with maps)
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["payee"] {
		t.Error("expected 'payee' in flagged fields")
	}
	if !nameSet["amount"] {
		t.Error("expected 'amount' in flagged fields")
	}
}

func TestShouldEscalateForHandwriting_ExactThreshold(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "Jane Smith",
			Confidence:    HandwritingThreshold, // Exactly at threshold (0.65)
			IsHandwritten: true,
		},
	}

	escalate, _ := ShouldEscalateForHandwriting(fields)
	if escalate {
		t.Error("expected no escalation at exact threshold boundary (< not <=)")
	}
}
