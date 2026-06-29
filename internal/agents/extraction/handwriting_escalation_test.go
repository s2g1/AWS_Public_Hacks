package extraction

import (
	"sort"
	"testing"

	"federal-payment-processing/internal/models"
)

func TestCheckHandwritingEscalation_FieldsBelowThresholdFlagged(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.50,
			IsHandwritten: true,
		},
		"amount": {
			Value:         "1500.00",
			Confidence:    0.60,
			IsHandwritten: true,
		},
	}

	flagged, needsEscalation := CheckHandwritingEscalation(fields)

	if !needsEscalation {
		t.Fatal("expected escalation needed for handwritten fields below threshold")
	}
	if len(flagged) != 2 {
		t.Fatalf("expected 2 flagged fields, got %d", len(flagged))
	}

	sort.Strings(flagged)
	if flagged[0] != "amount" || flagged[1] != "payee" {
		t.Fatalf("expected [amount, payee], got %v", flagged)
	}
}

func TestCheckHandwritingEscalation_FieldsAboveThresholdNotFlagged(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.70,
			IsHandwritten: true,
		},
		"amount": {
			Value:         "1500.00",
			Confidence:    0.80,
			IsHandwritten: true,
		},
	}

	flagged, needsEscalation := CheckHandwritingEscalation(fields)

	if needsEscalation {
		t.Fatal("expected no escalation for handwritten fields above threshold")
	}
	if flagged != nil {
		t.Fatalf("expected nil flagged fields, got %v", flagged)
	}
}

func TestCheckHandwritingEscalation_NonHandwrittenFieldsIgnored(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.30, // low confidence but NOT handwritten
			IsHandwritten: false,
		},
		"amount": {
			Value:         "1500.00",
			Confidence:    0.40, // low confidence but NOT handwritten
			IsHandwritten: false,
		},
	}

	flagged, needsEscalation := CheckHandwritingEscalation(fields)

	if needsEscalation {
		t.Fatal("expected no escalation for non-handwritten fields")
	}
	if flagged != nil {
		t.Fatalf("expected nil flagged fields, got %v", flagged)
	}
}

func TestCheckHandwritingEscalation_MixedFields(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.50,
			IsHandwritten: true, // handwritten + below threshold → flagged
		},
		"amount": {
			Value:         "1500.00",
			Confidence:    0.90,
			IsHandwritten: false, // not handwritten → ignored
		},
		"date": {
			Value:         "2024-01-15",
			Confidence:    0.70,
			IsHandwritten: true, // handwritten + above threshold → not flagged
		},
		"invoiceNumber": {
			Value:         "INV-001",
			Confidence:    0.40,
			IsHandwritten: false, // not handwritten → ignored
		},
	}

	flagged, needsEscalation := CheckHandwritingEscalation(fields)

	if !needsEscalation {
		t.Fatal("expected escalation needed for at least one handwritten field below threshold")
	}
	if len(flagged) != 1 {
		t.Fatalf("expected 1 flagged field, got %d", len(flagged))
	}
	if flagged[0] != "payee" {
		t.Fatalf("expected 'payee' to be flagged, got %q", flagged[0])
	}
}

func TestCheckHandwritingEscalation_EmptyFields(t *testing.T) {
	flagged, needsEscalation := CheckHandwritingEscalation(map[string]models.ExtractedField{})

	if needsEscalation {
		t.Fatal("expected no escalation for empty fields")
	}
	if flagged != nil {
		t.Fatalf("expected nil flagged fields, got %v", flagged)
	}
}

func TestCheckHandwritingEscalation_NilFields(t *testing.T) {
	flagged, needsEscalation := CheckHandwritingEscalation(nil)

	if needsEscalation {
		t.Fatal("expected no escalation for nil fields")
	}
	if flagged != nil {
		t.Fatalf("expected nil flagged fields, got %v", flagged)
	}
}

func TestCheckHandwritingEscalation_ExactlyAtThresholdNotFlagged(t *testing.T) {
	fields := map[string]models.ExtractedField{
		"payee": {
			Value:         "John Doe",
			Confidence:    0.65, // exactly at threshold → NOT below
			IsHandwritten: true,
		},
	}

	flagged, needsEscalation := CheckHandwritingEscalation(fields)

	if needsEscalation {
		t.Fatal("expected no escalation for field exactly at threshold")
	}
	if flagged != nil {
		t.Fatalf("expected nil flagged fields, got %v", flagged)
	}
}

func TestEscalationReasonHandwritingReviewConstant(t *testing.T) {
	if EscalationReasonHandwritingReview != "handwriting_review" {
		t.Fatalf("expected escalation reason 'handwriting_review', got %q", EscalationReasonHandwritingReview)
	}
}
