package portal

import (
	"fmt"
	"math"
)

// epsilon is the tolerance for floating point comparisons in financial calculations.
const epsilon = 0.0001

// CalculateContractTotals sums the CLIN-level obligated and expended amounts
// to produce contract-level totals.
func CalculateContractTotals(contract *Contract) (totalObligated, totalExpended float64) {
	for _, clin := range contract.CLINs {
		totalObligated += clin.Obligated
		totalExpended += clin.Expended
	}
	return totalObligated, totalExpended
}

// ValidateCLINSummation checks that the sum of CLIN obligated amounts equals
// the contract's TotalObligated, and the sum of CLIN expended amounts equals
// the contract's TotalExpended. Returns a descriptive error if there is a mismatch.
func ValidateCLINSummation(contract *Contract) error {
	sumObligated, sumExpended := CalculateContractTotals(contract)

	obligatedMismatch := math.Abs(sumObligated-contract.TotalObligated) > epsilon
	expendedMismatch := math.Abs(sumExpended-contract.TotalExpended) > epsilon

	if obligatedMismatch && expendedMismatch {
		return fmt.Errorf(
			"CLIN summation mismatch: obligated sum %.4f != contract total obligated %.4f; expended sum %.4f != contract total expended %.4f",
			sumObligated, contract.TotalObligated, sumExpended, contract.TotalExpended,
		)
	}
	if obligatedMismatch {
		return fmt.Errorf(
			"CLIN summation mismatch: obligated sum %.4f != contract total obligated %.4f",
			sumObligated, contract.TotalObligated,
		)
	}
	if expendedMismatch {
		return fmt.Errorf(
			"CLIN summation mismatch: expended sum %.4f != contract total expended %.4f",
			sumExpended, contract.TotalExpended,
		)
	}

	return nil
}

// RecalculateContractTotals recalculates and sets the contract's TotalObligated
// and TotalExpended from the CLIN data. This is useful after modifying CLIN values
// to keep the contract totals in sync.
func RecalculateContractTotals(contract *Contract) {
	contract.TotalObligated, contract.TotalExpended = CalculateContractTotals(contract)
}

// GetContractFinancials returns the financial figures for a contract as viewed by a user.
// Per Requirement 16.3, all users (regardless of role) see identical financial figures
// for obligations, expenditures, and ceilings.
func GetContractFinancials(contract *Contract, user *PortalUser) (ceiling, obligated, expended float64) {
	return contract.TotalCeiling, contract.TotalObligated, contract.TotalExpended
}
