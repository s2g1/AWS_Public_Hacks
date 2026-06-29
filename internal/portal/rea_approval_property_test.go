package portal

import (
	"math"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// **Validates: Requirements 18.3, 18.4**
// Property 21: REA Approval Adjusts Ceilings
// 1. After APPROVED response, total ceiling increases by the requested amount
// 2. After PARTIALLY_APPROVED response, total ceiling increases by the approved amount (< requested)
// 3. After DENIED response, total ceiling remains unchanged
// 4. After APPROVED, affected CLINs each get an equal portion of the approved amount added to their ceilings

func TestProperty21_ApprovedIncreasesTotalCeilingByRequestedAmount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contract with 1-5 CLINs
		numCLINs := rapid.IntRange(1, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make([]string, numCLINs)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z][0-9]{3}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    rapid.Float64Range(10000, 500000).Draw(t, "clinCeiling"),
				Obligated:  rapid.Float64Range(5000, 250000).Draw(t, "clinObligated"),
			}
			clinIDs[i] = id
		}

		initialTotalCeiling := rapid.Float64Range(100000, 2000000).Draw(t, "totalCeiling")
		contract := &Contract{
			ContractID:     "CONTRACT-PROP21",
			ContractNumber: "PROP21-001",
			TotalCeiling:   initialTotalCeiling,
			TotalObligated: rapid.Float64Range(50000, 1000000).Draw(t, "totalObligated"),
			CLINs:          clins,
		}

		// Generate a positive requested amount
		requestedAmount := rapid.Float64Range(1.0, 500000).Draw(t, "requestedAmount")

		// Select a non-empty subset of CLINs as affected
		subsetSize := rapid.IntRange(1, numCLINs).Draw(t, "subsetSize")
		affectedCLINs := make([]string, subsetSize)
		for i := 0; i < subsetSize; i++ {
			affectedCLINs[i] = clinIDs[rapid.IntRange(0, numCLINs-1).Draw(t, "clinIdx")]
		}

		rea := &REA{
			REAID:           "REA-PROP21",
			ContractID:      "CONTRACT-PROP21",
			RequestedAmount: requestedAmount,
			AffectedCLINs:   affectedCLINs,
			Status:          REAStatusSubmitted,
			Justification:   "test",
			SubmittedBy:     "contractor@test.com",
			SubmittedAt:     time.Now(),
		}

		response := REAResponse{
			ResponseType: REAStatusApproved,
			Rationale:    "Approved",
			RespondedBy:  "co@gov.com",
			RespondedAt:  time.Now(),
		}

		_, err := RespondToREA(contract, rea, response)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Property: total ceiling increases by exactly the requested amount
		expectedCeiling := initialTotalCeiling + requestedAmount
		if math.Abs(contract.TotalCeiling-expectedCeiling) > 0.01 {
			t.Fatalf("expected total ceiling %f, got %f (initial=%f, requested=%f)",
				expectedCeiling, contract.TotalCeiling, initialTotalCeiling, requestedAmount)
		}
	})
}

func TestProperty21_PartiallyApprovedIncreasesTotalCeilingByApprovedAmount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contract with 1-5 CLINs
		numCLINs := rapid.IntRange(1, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make([]string, numCLINs)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z][0-9]{3}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    rapid.Float64Range(10000, 500000).Draw(t, "clinCeiling"),
				Obligated:  rapid.Float64Range(5000, 250000).Draw(t, "clinObligated"),
			}
			clinIDs[i] = id
		}

		initialTotalCeiling := rapid.Float64Range(100000, 2000000).Draw(t, "totalCeiling")
		contract := &Contract{
			ContractID:     "CONTRACT-PROP21",
			ContractNumber: "PROP21-001",
			TotalCeiling:   initialTotalCeiling,
			TotalObligated: rapid.Float64Range(50000, 1000000).Draw(t, "totalObligated"),
			CLINs:          clins,
		}

		// Generate requested amount and a smaller approved amount
		requestedAmount := rapid.Float64Range(100, 500000).Draw(t, "requestedAmount")
		approvedAmount := rapid.Float64Range(0.01, requestedAmount-0.01).Draw(t, "approvedAmount")

		// Select a non-empty subset of CLINs as affected
		subsetSize := rapid.IntRange(1, numCLINs).Draw(t, "subsetSize")
		affectedCLINs := make([]string, subsetSize)
		for i := 0; i < subsetSize; i++ {
			affectedCLINs[i] = clinIDs[rapid.IntRange(0, numCLINs-1).Draw(t, "clinIdx")]
		}

		rea := &REA{
			REAID:           "REA-PROP21",
			ContractID:      "CONTRACT-PROP21",
			RequestedAmount: requestedAmount,
			AffectedCLINs:   affectedCLINs,
			Status:          REAStatusSubmitted,
			Justification:   "test",
			SubmittedBy:     "contractor@test.com",
			SubmittedAt:     time.Now(),
		}

		response := REAResponse{
			ResponseType:   REAStatusPartiallyApproved,
			ApprovedAmount: approvedAmount,
			Rationale:      "Partially approved",
			RespondedBy:    "co@gov.com",
			RespondedAt:    time.Now(),
		}

		_, err := RespondToREA(contract, rea, response)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Property: total ceiling increases by the approved amount (not the requested amount)
		expectedCeiling := initialTotalCeiling + approvedAmount
		if math.Abs(contract.TotalCeiling-expectedCeiling) > 0.01 {
			t.Fatalf("expected total ceiling %f, got %f (initial=%f, approved=%f, requested=%f)",
				expectedCeiling, contract.TotalCeiling, initialTotalCeiling, approvedAmount, requestedAmount)
		}
	})
}

func TestProperty21_DeniedLeavesTotalCeilingUnchanged(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contract with 1-5 CLINs
		numCLINs := rapid.IntRange(1, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make([]string, numCLINs)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z][0-9]{3}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    rapid.Float64Range(10000, 500000).Draw(t, "clinCeiling"),
				Obligated:  rapid.Float64Range(5000, 250000).Draw(t, "clinObligated"),
			}
			clinIDs[i] = id
		}

		initialTotalCeiling := rapid.Float64Range(100000, 2000000).Draw(t, "totalCeiling")
		contract := &Contract{
			ContractID:     "CONTRACT-PROP21",
			ContractNumber: "PROP21-001",
			TotalCeiling:   initialTotalCeiling,
			TotalObligated: rapid.Float64Range(50000, 1000000).Draw(t, "totalObligated"),
			CLINs:          clins,
		}

		// Store original CLIN ceilings
		originalCeilings := make([]float64, numCLINs)
		for i := range clins {
			originalCeilings[i] = clins[i].Ceiling
		}

		requestedAmount := rapid.Float64Range(1.0, 500000).Draw(t, "requestedAmount")

		// Select a non-empty subset of CLINs as affected
		subsetSize := rapid.IntRange(1, numCLINs).Draw(t, "subsetSize")
		affectedCLINs := make([]string, subsetSize)
		for i := 0; i < subsetSize; i++ {
			affectedCLINs[i] = clinIDs[rapid.IntRange(0, numCLINs-1).Draw(t, "clinIdx")]
		}

		rea := &REA{
			REAID:           "REA-PROP21",
			ContractID:      "CONTRACT-PROP21",
			RequestedAmount: requestedAmount,
			AffectedCLINs:   affectedCLINs,
			Status:          REAStatusSubmitted,
			Justification:   "test",
			SubmittedBy:     "contractor@test.com",
			SubmittedAt:     time.Now(),
		}

		response := REAResponse{
			ResponseType: REAStatusDenied,
			Rationale:    "Denied",
			RespondedBy:  "co@gov.com",
			RespondedAt:  time.Now(),
		}

		_, err := RespondToREA(contract, rea, response)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Property: total ceiling remains unchanged after denial
		if math.Abs(contract.TotalCeiling-initialTotalCeiling) > 0.01 {
			t.Fatalf("expected total ceiling unchanged at %f, got %f",
				initialTotalCeiling, contract.TotalCeiling)
		}

		// Property: all CLIN ceilings remain unchanged after denial
		for i := range contract.CLINs {
			if math.Abs(contract.CLINs[i].Ceiling-originalCeilings[i]) > 0.01 {
				t.Fatalf("expected CLIN %s ceiling unchanged at %f, got %f",
					contract.CLINs[i].CLINID, originalCeilings[i], contract.CLINs[i].Ceiling)
			}
		}
	})
}

func TestProperty21_ApprovedDistributesEquallyToAffectedCLINs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a contract with 2-5 CLINs to ensure distinct affected CLINs
		numCLINs := rapid.IntRange(2, 5).Draw(t, "numCLINs")
		clins := make([]ContractLineItem, numCLINs)
		clinIDs := make([]string, numCLINs)
		for i := 0; i < numCLINs; i++ {
			id := rapid.StringMatching(`CLIN-[A-Z][0-9]{3}`).Draw(t, "clinID")
			clins[i] = ContractLineItem{
				CLINID:     id,
				CLINNumber: "000" + string(rune('1'+i)),
				CLINStatus: CLINStatusActive,
				Ceiling:    rapid.Float64Range(10000, 500000).Draw(t, "clinCeiling"),
				Obligated:  rapid.Float64Range(5000, 250000).Draw(t, "clinObligated"),
			}
			clinIDs[i] = id
		}

		contract := &Contract{
			ContractID:     "CONTRACT-PROP21",
			ContractNumber: "PROP21-001",
			TotalCeiling:   rapid.Float64Range(100000, 2000000).Draw(t, "totalCeiling"),
			TotalObligated: rapid.Float64Range(50000, 1000000).Draw(t, "totalObligated"),
			CLINs:          clins,
		}

		// Store original ceilings
		originalCeilings := make(map[string]float64)
		for _, clin := range contract.CLINs {
			originalCeilings[clin.CLINID] = clin.Ceiling
		}

		requestedAmount := rapid.Float64Range(1.0, 500000).Draw(t, "requestedAmount")

		// Select a non-empty subset of DISTINCT CLIN IDs as affected
		subsetSize := rapid.IntRange(1, numCLINs).Draw(t, "subsetSize")
		// Use a set to ensure distinct CLINs
		affectedSet := make(map[string]bool)
		affectedCLINs := []string{}
		for len(affectedCLINs) < subsetSize {
			idx := rapid.IntRange(0, numCLINs-1).Draw(t, "clinIdx")
			id := clinIDs[idx]
			if !affectedSet[id] {
				affectedSet[id] = true
				affectedCLINs = append(affectedCLINs, id)
			}
		}

		rea := &REA{
			REAID:           "REA-PROP21",
			ContractID:      "CONTRACT-PROP21",
			RequestedAmount: requestedAmount,
			AffectedCLINs:   affectedCLINs,
			Status:          REAStatusSubmitted,
			Justification:   "test",
			SubmittedBy:     "contractor@test.com",
			SubmittedAt:     time.Now(),
		}

		response := REAResponse{
			ResponseType: REAStatusApproved,
			Rationale:    "Approved",
			RespondedBy:  "co@gov.com",
			RespondedAt:  time.Now(),
		}

		_, err := RespondToREA(contract, rea, response)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Property: each affected CLIN gets an equal portion of the approved amount
		perCLINAmount := requestedAmount / float64(len(affectedCLINs))

		for _, clin := range contract.CLINs {
			expectedCeiling := originalCeilings[clin.CLINID]
			if affectedSet[clin.CLINID] {
				expectedCeiling += perCLINAmount
			}
			if math.Abs(clin.Ceiling-expectedCeiling) > 0.01 {
				t.Fatalf("CLIN %s: expected ceiling %f, got %f (original=%f, perCLIN=%f, affected=%v)",
					clin.CLINID, expectedCeiling, clin.Ceiling, originalCeilings[clin.CLINID], perCLINAmount, affectedSet[clin.CLINID])
			}
		}
	})
}
