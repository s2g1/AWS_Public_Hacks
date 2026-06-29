package portal

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 16.3**
// Property 16: Financial Data Consistency Across Roles
// For any contract, when viewed by different role types (CO, COR, PCO, PM, Contractor),
// the financial figures (TotalCeiling, TotalObligated, TotalExpended) are always identical.

func TestProperty16_FinancialDataConsistencyAcrossRoles(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random contract with financial data
		totalCeiling := rapid.Float64Range(0, 100_000_000).Draw(t, "totalCeiling")
		totalObligated := rapid.Float64Range(0, totalCeiling).Draw(t, "totalObligated")
		totalExpended := rapid.Float64Range(0, totalObligated).Draw(t, "totalExpended")

		contract := &Contract{
			ContractID:     rapid.StringMatching(`[A-Z0-9]{8}`).Draw(t, "contractId"),
			ContractNumber: rapid.StringMatching(`[A-Z]{2}-[0-9]{4}`).Draw(t, "contractNumber"),
			TotalCeiling:   totalCeiling,
			TotalObligated: totalObligated,
			TotalExpended:  totalExpended,
		}

		// All possible roles in the system
		allRoles := []PortalRole{
			PortalRoleContractingOfficer,
			PortalRoleCOR,
			PortalRoleProcuringContractingOfficer,
			PortalRoleProgramManager,
			PortalRoleContractor,
		}

		// Pick a random role index to use as the baseline
		baseIdx := rapid.IntRange(0, len(allRoles)-1).Draw(t, "baseRoleIdx")
		baseUser := &PortalUser{
			UserID: "user-base",
			Name:   "Base User",
			Role:   allRoles[baseIdx],
		}

		baseCeiling, baseObligated, baseExpended := GetContractFinancials(contract, baseUser)

		// Verify against the raw contract fields
		if baseCeiling != contract.TotalCeiling {
			t.Fatalf("base ceiling %f != contract.TotalCeiling %f", baseCeiling, contract.TotalCeiling)
		}
		if baseObligated != contract.TotalObligated {
			t.Fatalf("base obligated %f != contract.TotalObligated %f", baseObligated, contract.TotalObligated)
		}
		if baseExpended != contract.TotalExpended {
			t.Fatalf("base expended %f != contract.TotalExpended %f", baseExpended, contract.TotalExpended)
		}

		// Every role must see the same financial figures
		for _, role := range allRoles {
			user := &PortalUser{
				UserID: "user-" + string(role),
				Name:   "User " + string(role),
				Role:   role,
			}

			ceiling, obligated, expended := GetContractFinancials(contract, user)

			if ceiling != baseCeiling {
				t.Fatalf("role %s ceiling %f != base ceiling %f", role, ceiling, baseCeiling)
			}
			if obligated != baseObligated {
				t.Fatalf("role %s obligated %f != base obligated %f", role, obligated, baseObligated)
			}
			if expended != baseExpended {
				t.Fatalf("role %s expended %f != base expended %f", role, expended, baseExpended)
			}
		}
	})
}
