package portal

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 22.1, 22.2**
// Property 19: Obligation Cannot Exceed Ceiling
// 1. For any contract where TotalObligated > TotalCeiling, ValidateObligationIntegrity always returns an error
// 2. For any contract where TotalObligated <= TotalCeiling, ValidateObligationIntegrity always returns nil

func TestProperty19_ObligationExceedsCeiling_ReturnsError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a ceiling value
		totalCeiling := rapid.Float64Range(0, 100_000_000).Draw(t, "totalCeiling")

		// Generate an obligated amount that strictly exceeds the ceiling
		// Add a positive offset to ensure TotalObligated > TotalCeiling
		excess := rapid.Float64Range(0.01, 50_000_000).Draw(t, "excess")
		totalObligated := totalCeiling + excess

		contract := &Contract{
			ContractID:     "CTR-TEST",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
		}

		err := ValidateObligationIntegrity(contract)
		if err == nil {
			t.Fatalf("expected error when TotalObligated (%.2f) > TotalCeiling (%.2f), got nil",
				totalObligated, totalCeiling)
		}
	})
}

func TestProperty19_ObligationWithinCeiling_ReturnsNil(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a ceiling value
		totalCeiling := rapid.Float64Range(0, 100_000_000).Draw(t, "totalCeiling")

		// Generate an obligated amount that is at or below the ceiling
		totalObligated := rapid.Float64Range(0, totalCeiling).Draw(t, "totalObligated")

		contract := &Contract{
			ContractID:     "CTR-TEST",
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
		}

		err := ValidateObligationIntegrity(contract)
		if err != nil {
			t.Fatalf("expected nil when TotalObligated (%.2f) <= TotalCeiling (%.2f), got: %v",
				totalObligated, totalCeiling, err)
		}
	})
}
