package portal

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 18.1**
// Property 20: REA Validation Rules
// 1. Any negative or zero requested amount always produces an error
// 2. An empty AffectedCLINs list always produces an error
// 3. Any CLIN ID not present on the contract always produces an error
// 4. When all conditions are met (positive amount, non-empty CLINs, all CLINs exist on contract), submission succeeds

func TestProperty20_NegativeOrZeroAmountAlwaysProducesError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a non-positive amount (zero or negative)
		amount := rapid.Float64Range(-1_000_000, 0).Draw(t, "amount")

		contract := &Contract{
			ContractID:     "CONTRACT-PROP",
			ContractNumber: "PROP-001",
			CLINs: []ContractLineItem{
				{CLINID: "CLIN-A", CLINNumber: "0001", CLINStatus: CLINStatusActive, Ceiling: 100000, Obligated: 50000},
			},
		}

		req := REASubmissionRequest{
			RequestedAmount: amount,
			AffectedCLINs:  []string{"CLIN-A"},
			Justification:  "test justification",
			SubmittedBy:    "user@test.com",
		}

		result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)

		// Property: non-positive amount always produces an error
		if err == nil {
			t.Fatalf("expected error for amount=%f, but got nil error with result: %+v", amount, result)
		}
		if result != nil {
			t.Fatalf("expected nil result for amount=%f, but got: %+v", amount, result)
		}
	})
}

func TestProperty20_EmptyAffectedCLINsAlwaysProducesError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a positive amount (valid)
		amount := rapid.Float64Range(0.01, 1_000_000).Draw(t, "amount")

		contract := &Contract{
			ContractID:     "CONTRACT-PROP",
			ContractNumber: "PROP-001",
			CLINs: []ContractLineItem{
				{CLINID: "CLIN-A", CLINNumber: "0001", CLINStatus: CLINStatusActive, Ceiling: 100000, Obligated: 50000},
			},
		}

		req := REASubmissionRequest{
			RequestedAmount: amount,
			AffectedCLINs:  []string{},
			Justification:  "test justification",
			SubmittedBy:    "user@test.com",
		}

		result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)

		// Property: empty CLINs list always produces an error
		if err == nil {
			t.Fatalf("expected error for empty AffectedCLINs with amount=%f, but got nil error", amount)
		}
		if result != nil {
			t.Fatalf("expected nil result for empty AffectedCLINs, but got: %+v", result)
		}
	})
}

func TestProperty20_NonExistentCLINAlwaysProducesError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate valid contract with known CLINs
		numCLINs := rapid.IntRange(1, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make(map[string]bool)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z]{3}[0-9]{2}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    100000,
				Obligated:  50000,
			}
			clinIDs[id] = true
		}

		contract := &Contract{
			ContractID:     "CONTRACT-PROP",
			ContractNumber: "PROP-001",
			CLINs:          clins,
		}

		// Generate a CLIN ID that does NOT exist on the contract
		fakeCLIN := rapid.StringMatching(`FAKE-[A-Z]{4}[0-9]{3}`).Draw(t, "fakeCLIN")
		// Ensure the fake CLIN is not in the contract (extremely unlikely but defensive)
		for clinIDs[fakeCLIN] {
			fakeCLIN = fakeCLIN + "X"
		}

		amount := rapid.Float64Range(0.01, 1_000_000).Draw(t, "amount")

		req := REASubmissionRequest{
			RequestedAmount: amount,
			AffectedCLINs:  []string{fakeCLIN},
			Justification:  "test justification",
			SubmittedBy:    "user@test.com",
		}

		result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)

		// Property: referencing a CLIN not on the contract always produces an error
		if err == nil {
			t.Fatalf("expected error for non-existent CLIN %q, but got nil error", fakeCLIN)
		}
		if result != nil {
			t.Fatalf("expected nil result for non-existent CLIN, but got: %+v", result)
		}
	})
}

func TestProperty20_ValidSubmissionSucceeds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contract with 1-5 CLINs
		numCLINs := rapid.IntRange(1, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make([]string, numCLINs)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z]{2}[0-9]{2}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    100000,
				Obligated:  50000,
			}
			clinIDs[i] = id
		}

		contract := &Contract{
			ContractID:     "CONTRACT-PROP",
			ContractNumber: "PROP-001",
			CLINs:          clins,
		}

		// Select a non-empty subset of the contract's CLINs
		subsetSize := rapid.IntRange(1, numCLINs).Draw(t, "subsetSize")
		selectedCLINs := make([]string, subsetSize)
		for i := 0; i < subsetSize; i++ {
			idx := rapid.IntRange(0, numCLINs-1).Draw(t, "clinIdx")
			selectedCLINs[i] = clinIDs[idx]
		}

		// Generate a positive amount
		amount := rapid.Float64Range(0.01, 1_000_000).Draw(t, "amount")

		req := REASubmissionRequest{
			RequestedAmount: amount,
			AffectedCLINs:  selectedCLINs,
			Justification:  rapid.StringMatching(`[a-z ]{5,50}`).Draw(t, "justification"),
			SubmittedBy:    rapid.StringMatching(`[a-z]+@[a-z]+\\.com`).Draw(t, "submittedBy"),
		}

		result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)

		// Property: when all conditions are met, submission succeeds
		if err != nil {
			t.Fatalf("expected successful submission with amount=%f, CLINs=%v, got error: %v",
				amount, selectedCLINs, err)
		}
		if result == nil {
			t.Fatal("expected non-nil result for valid submission")
		}
		if result.REA == nil {
			t.Fatal("expected REA to be created")
		}
		if result.REA.Status != REAStatusSubmitted {
			t.Fatalf("expected Status=SUBMITTED, got %s", result.REA.Status)
		}
		if result.REA.RequestedAmount != amount {
			t.Fatalf("expected RequestedAmount=%f, got %f", amount, result.REA.RequestedAmount)
		}
		if len(result.AuditEntries) == 0 {
			t.Fatal("expected at least one audit entry")
		}
		if len(result.Notifications) == 0 {
			t.Fatal("expected at least one notification")
		}
	})
}
