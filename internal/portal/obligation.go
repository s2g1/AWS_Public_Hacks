package portal

import "fmt"

// CLINExpenditureViolation represents a CLIN whose expenditures exceed its obligations.
type CLINExpenditureViolation struct {
	CLINID   string  `json:"clinId"`
	Expended float64 `json:"expended"`
	Obligated float64 `json:"obligated"`
	Excess   float64 `json:"excess"`
}

// ValidateObligationIntegrity enforces that total contract obligations never
// exceed the total contract ceiling. Returns an error if TotalObligated > TotalCeiling.
// Validates: Requirement 22.1
func ValidateObligationIntegrity(contract *Contract) error {
	if contract.TotalObligated > contract.TotalCeiling {
		return fmt.Errorf(
			"obligation integrity violation: total obligated %.2f exceeds total ceiling %.2f",
			contract.TotalObligated, contract.TotalCeiling,
		)
	}
	return nil
}

// ValidateCLINExpenditureLimits checks each CLIN's expenditures against its
// obligations. Returns a list of violations where expenditures exceed obligations.
// Validates: Requirement 22.2
func ValidateCLINExpenditureLimits(contract *Contract) []CLINExpenditureViolation {
	var violations []CLINExpenditureViolation
	for _, clin := range contract.CLINs {
		if clin.Expended > clin.Obligated {
			violations = append(violations, CLINExpenditureViolation{
				CLINID:    clin.CLINID,
				Expended:  clin.Expended,
				Obligated: clin.Obligated,
				Excess:    clin.Expended - clin.Obligated,
			})
		}
	}
	return violations
}

// CheckExpenditureAllowed checks if adding amount to a CLIN's expended total
// would exceed its obligation. Returns an error with notification details if it
// would exceed, nil if the expenditure is allowed.
// Validates: Requirement 22.3
func CheckExpenditureAllowed(contract *Contract, clinID string, amount float64) error {
	for _, clin := range contract.CLINs {
		if clin.CLINID == clinID {
			newExpended := clin.Expended + amount
			if newExpended > clin.Obligated {
				return fmt.Errorf(
					"expenditure blocked: adding %.2f to CLIN %s would result in %.2f expended, exceeding obligation of %.2f; payment held, contracting officer notified",
					amount, clinID, newExpended, clin.Obligated,
				)
			}
			return nil
		}
	}
	return fmt.Errorf("CLIN %q not found on contract %s", clinID, contract.ContractID)
}
