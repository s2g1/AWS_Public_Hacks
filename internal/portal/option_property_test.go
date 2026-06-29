package portal

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// **Validates: Requirements 19.1, 19.2, 19.3**
// Property 22: Option Exercise Constraints
// Tests that ExerciseOption enforces CLIN type, status, deadline, and ceiling constraints.

// Property 22.1: Non-OPTION type CLINs always produce an error.
func TestProperty22_NonOptionTypeCLINAlwaysErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a non-OPTION CLIN type
		nonOptionTypes := []CLINType{CLINTypeFFP, CLINTypeCPFF, CLINTypeCPIF, CLINTypeTAndM}
		typeIdx := rapid.IntRange(0, len(nonOptionTypes)-1).Draw(t, "typeIdx")
		clinType := nonOptionTypes[typeIdx]

		deadline := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
		clinObligated := rapid.Float64Range(1, 500_000).Draw(t, "clinObligated")
		totalCeiling := rapid.Float64Range(clinObligated*2, 10_000_000).Draw(t, "totalCeiling")
		totalObligated := rapid.Float64Range(0, totalCeiling-clinObligated).Draw(t, "totalObligated")

		contract := &Contract{
			ContractID:     "C-PROP",
			ContractNumber: "W911NF-25-C-0001",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
			CLINs: []ContractLineItem{
				{
					CLINID:                 "CLIN-TARGET",
					CLINNumber:             "0001",
					Description:            "Test CLIN",
					CLINType:               clinType,
					CLINStatus:             CLINStatusActive,
					Ceiling:                clinObligated,
					Obligated:              clinObligated,
					OptionExerciseDeadline: &deadline,
				},
			},
		}

		now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
		err := ExerciseOptionWithIDGen(contract, "CLIN-TARGET", now, func() string { return "MOD-TEST" })

		if err == nil {
			t.Fatalf("expected error for non-OPTION CLIN type %s, got nil", clinType)
		}
	})
}

// Property 22.2: Non-ACTIVE status CLINs always produce an error.
func TestProperty22_NonActiveStatusCLINAlwaysErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a non-ACTIVE CLIN status
		nonActiveStatuses := []CLINStatus{CLINStatusExercised, CLINStatusCompleted, CLINStatusExpired, CLINStatusNotExercised}
		statusIdx := rapid.IntRange(0, len(nonActiveStatuses)-1).Draw(t, "statusIdx")
		clinStatus := nonActiveStatuses[statusIdx]

		deadline := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
		clinObligated := rapid.Float64Range(1, 500_000).Draw(t, "clinObligated")
		totalCeiling := rapid.Float64Range(clinObligated*2, 10_000_000).Draw(t, "totalCeiling")
		totalObligated := rapid.Float64Range(0, totalCeiling-clinObligated).Draw(t, "totalObligated")

		contract := &Contract{
			ContractID:     "C-PROP",
			ContractNumber: "W911NF-25-C-0001",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
			CLINs: []ContractLineItem{
				{
					CLINID:                 "CLIN-TARGET",
					CLINNumber:             "0001",
					Description:            "Test Option CLIN",
					CLINType:               CLINTypeOption,
					CLINStatus:             clinStatus,
					Ceiling:                clinObligated,
					Obligated:              clinObligated,
					OptionExerciseDeadline: &deadline,
				},
			},
		}

		now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
		err := ExerciseOptionWithIDGen(contract, "CLIN-TARGET", now, func() string { return "MOD-TEST" })

		if err == nil {
			t.Fatalf("expected error for non-ACTIVE CLIN status %s, got nil", clinStatus)
		}
	})
}

// Property 22.3: Exercising when now > deadline always produces an error.
func TestProperty22_ExpiredDeadlineAlwaysErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a deadline in the past relative to "now"
		deadlineYear := rapid.IntRange(2020, 2024).Draw(t, "deadlineYear")
		deadlineMonth := rapid.IntRange(1, 12).Draw(t, "deadlineMonth")
		deadlineDay := rapid.IntRange(1, 28).Draw(t, "deadlineDay")
		deadline := time.Date(deadlineYear, time.Month(deadlineMonth), deadlineDay, 0, 0, 0, 0, time.UTC)

		// "now" is always after the deadline
		daysAfter := rapid.IntRange(1, 365*5).Draw(t, "daysAfter")
		now := deadline.AddDate(0, 0, daysAfter)

		clinObligated := rapid.Float64Range(1, 500_000).Draw(t, "clinObligated")
		totalCeiling := rapid.Float64Range(clinObligated*2, 10_000_000).Draw(t, "totalCeiling")
		totalObligated := rapid.Float64Range(0, totalCeiling-clinObligated).Draw(t, "totalObligated")

		contract := &Contract{
			ContractID:     "C-PROP",
			ContractNumber: "W911NF-25-C-0001",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
			CLINs: []ContractLineItem{
				{
					CLINID:                 "CLIN-TARGET",
					CLINNumber:             "0001",
					Description:            "Test Option CLIN",
					CLINType:               CLINTypeOption,
					CLINStatus:             CLINStatusActive,
					Ceiling:                clinObligated,
					Obligated:              clinObligated,
					OptionExerciseDeadline: &deadline,
				},
			},
		}

		err := ExerciseOptionWithIDGen(contract, "CLIN-TARGET", now, func() string { return "MOD-TEST" })

		if err == nil {
			t.Fatalf("expected error for expired deadline (deadline=%v, now=%v), got nil", deadline, now)
		}
	})
}

// Property 22.4: When conditions are valid (OPTION type, ACTIVE, before deadline, within ceiling),
// exercise succeeds and CLIN status becomes EXERCISED.
func TestProperty22_ValidConditionsSucceedAndCLINBecomesExercised(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate valid conditions
		clinObligated := rapid.Float64Range(1, 500_000).Draw(t, "clinObligated")
		// totalCeiling must be large enough to accommodate current obligations + CLIN obligation
		totalObligated := rapid.Float64Range(0, 5_000_000).Draw(t, "totalObligated")
		totalCeiling := rapid.Float64Range(totalObligated+clinObligated, totalObligated+clinObligated+5_000_000).Draw(t, "totalCeiling")

		// Generate a deadline in the future relative to "now"
		now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
		daysUntilDeadline := rapid.IntRange(1, 365*5).Draw(t, "daysUntilDeadline")
		deadline := now.AddDate(0, 0, daysUntilDeadline)

		contract := &Contract{
			ContractID:     "C-PROP",
			ContractNumber: "W911NF-25-C-0001",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
			CLINs: []ContractLineItem{
				{
					CLINID:                 "CLIN-TARGET",
					CLINNumber:             "0001",
					Description:            "Test Option CLIN",
					CLINType:               CLINTypeOption,
					CLINStatus:             CLINStatusActive,
					Ceiling:                clinObligated,
					Obligated:              clinObligated,
					OptionExerciseDeadline: &deadline,
				},
			},
			Modifications: []Modification{},
		}

		originalObligated := contract.TotalObligated

		err := ExerciseOptionWithIDGen(contract, "CLIN-TARGET", now, func() string { return "MOD-TEST" })

		if err != nil {
			t.Fatalf("expected no error for valid option exercise, got: %v", err)
		}

		// Verify CLIN status is now EXERCISED
		if contract.CLINs[0].CLINStatus != CLINStatusExercised {
			t.Fatalf("expected CLIN status EXERCISED, got: %s", contract.CLINs[0].CLINStatus)
		}

		// Verify contract TotalObligated increased by CLIN's Obligated amount
		expectedObligated := originalObligated + clinObligated
		if contract.TotalObligated != expectedObligated {
			t.Fatalf("expected TotalObligated %.2f, got: %.2f", expectedObligated, contract.TotalObligated)
		}

		// Verify a modification was created
		if len(contract.Modifications) != 1 {
			t.Fatalf("expected 1 modification, got: %d", len(contract.Modifications))
		}
		if contract.Modifications[0].Amount != clinObligated {
			t.Fatalf("expected modification amount %.2f, got: %.2f", clinObligated, contract.Modifications[0].Amount)
		}
	})
}
