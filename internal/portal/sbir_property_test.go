package portal

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 21.1, 21.4**
// Property 24: SBIR CLIN Status Gate
// 1. ACTIVE or EXERCISED status always allows SBIR invoice validation to proceed (no status error)
// 2. COMPLETED, EXPIRED, or NOT_EXERCISED status always produces a status error

func TestProperty24_ActiveOrExercised_NoStatusError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Draw a valid (allowing) CLIN status: ACTIVE or EXERCISED
		statusIdx := rapid.IntRange(0, 1).Draw(t, "statusIdx")
		allowedStatuses := []CLINStatus{CLINStatusActive, CLINStatusExercised}
		clinStatus := allowedStatuses[statusIdx]

		// Generate a reasonable obligation that is large enough for the invoice
		obligated := rapid.Float64Range(10000, 10_000_000).Draw(t, "obligated")

		// Generate expended that is less than obligated (so expenditure check passes)
		expended := rapid.Float64Range(0, obligated*0.8).Draw(t, "expended")

		// Generate an invoice amount that fits within remaining obligation
		remaining := obligated - expended
		invoiceAmount := rapid.Float64Range(0.01, remaining).Draw(t, "invoiceAmount")

		contract := &Contract{
			ContractID: "CTR-SBIR-PBT",
			CLINs: []ContractLineItem{
				{
					CLINID:     "CLIN-PBT-001",
					CLINStatus: clinStatus,
					Obligated:  obligated,
					Expended:   expended,
				},
			},
		}

		// Use FFP with milestone accepted so no contract-type error
		err := ValidateSBIRInvoice(contract, "CLIN-PBT-001", invoiceAmount, ContractTypeFFP, true)
		if err != nil {
			t.Fatalf("expected no error for CLIN status %s (ACTIVE or EXERCISED), got: %v",
				clinStatus, err)
		}
	})
}

func TestProperty24_CompletedExpiredOrNotExercised_StatusError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Draw an invalid (blocking) CLIN status: COMPLETED, EXPIRED, or NOT_EXERCISED
		statusIdx := rapid.IntRange(0, 2).Draw(t, "statusIdx")
		blockedStatuses := []CLINStatus{CLINStatusCompleted, CLINStatusExpired, CLINStatusNotExercised}
		clinStatus := blockedStatuses[statusIdx]

		// Generate arbitrary financial values (they shouldn't matter since status gate fires first)
		obligated := rapid.Float64Range(1000, 10_000_000).Draw(t, "obligated")
		expended := rapid.Float64Range(0, obligated).Draw(t, "expended")
		invoiceAmount := rapid.Float64Range(0.01, 100_000).Draw(t, "invoiceAmount")

		contract := &Contract{
			ContractID: "CTR-SBIR-PBT",
			CLINs: []ContractLineItem{
				{
					CLINID:     "CLIN-PBT-001",
					CLINStatus: clinStatus,
					Obligated:  obligated,
					Expended:   expended,
				},
			},
		}

		// Use FFP with milestone accepted — only the status gate should fire
		err := ValidateSBIRInvoice(contract, "CLIN-PBT-001", invoiceAmount, ContractTypeFFP, true)
		if err == nil {
			t.Fatalf("expected status error for CLIN status %s, got nil", clinStatus)
		}
		// Verify the error mentions the CLIN status
		if !strings.Contains(err.Error(), string(clinStatus)) {
			t.Fatalf("expected error to mention status %s, got: %v", clinStatus, err)
		}
		// Verify the error mentions ACTIVE or EXERCISED as the required statuses
		if !strings.Contains(err.Error(), "ACTIVE") || !strings.Contains(err.Error(), "EXERCISED") {
			t.Fatalf("expected error to mention required statuses ACTIVE/EXERCISED, got: %v", err)
		}
	})
}

// **Validates: Requirements 21.2, 22.3**
// Property 25: SBIR Expenditure Ceiling Enforcement
// 1. When invoice amount + expended <= obligation, validation passes (no expenditure error)
// 2. When invoice amount + expended > obligation, validation fails with a "held" error

func TestProperty25_ExpenditureWithinObligation_Passes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an obligation amount
		obligated := rapid.Float64Range(1, 10_000_000).Draw(t, "obligated")

		// Generate expended that is less than obligated (leave room for invoice)
		expended := rapid.Float64Range(0, obligated).Draw(t, "expended")

		// Generate an invoice amount such that expended + invoice <= obligated
		remaining := obligated - expended
		invoiceAmount := rapid.Float64Range(0, remaining).Draw(t, "invoiceAmount")

		contract := &Contract{
			ContractID: "CTR-SBIR-PROP",
			CLINs: []ContractLineItem{
				{
					CLINID:     "CLIN-PROP-001",
					CLINStatus: CLINStatusActive,
					Obligated:  obligated,
					Expended:   expended,
				},
			},
		}

		// Use FFP with milestone accepted so only the expenditure check is tested
		err := ValidateSBIRInvoice(contract, "CLIN-PROP-001", invoiceAmount, ContractTypeFFP, true)
		if err != nil {
			t.Fatalf("expected no error when invoiceAmount (%.2f) + expended (%.2f) = %.2f <= obligated (%.2f), got: %v",
				invoiceAmount, expended, invoiceAmount+expended, obligated, err)
		}
	})
}

func TestProperty25_ExpenditureExceedsObligation_Held(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an obligation amount
		obligated := rapid.Float64Range(1, 10_000_000).Draw(t, "obligated")

		// Generate expended that is within obligation
		expended := rapid.Float64Range(0, obligated).Draw(t, "expended")

		// Generate an invoice amount that causes expended + invoice > obligated
		remaining := obligated - expended
		excess := rapid.Float64Range(0.01, 5_000_000).Draw(t, "excess")
		invoiceAmount := remaining + excess

		contract := &Contract{
			ContractID: "CTR-SBIR-PROP",
			CLINs: []ContractLineItem{
				{
					CLINID:     "CLIN-PROP-001",
					CLINStatus: CLINStatusActive,
					Obligated:  obligated,
					Expended:   expended,
				},
			},
		}

		// Use FFP with milestone accepted so contract-type checks pass
		err := ValidateSBIRInvoice(contract, "CLIN-PROP-001", invoiceAmount, ContractTypeFFP, true)
		if err == nil {
			t.Fatalf("expected error when invoiceAmount (%.2f) + expended (%.2f) = %.2f > obligated (%.2f), got nil",
				invoiceAmount, expended, invoiceAmount+expended, obligated)
		}
		if !strings.Contains(err.Error(), "held") {
			t.Fatalf("expected error to contain 'held', got: %v", err)
		}
	})
}
