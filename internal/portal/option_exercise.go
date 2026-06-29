package portal

import (
	"errors"
	"fmt"
	"time"
)

// OptionExerciseResult contains the outcome and side effects of exercising an option.
type OptionExerciseResult struct {
	Modification  *Modification  `json:"modification"`
	AuditEntries  []AuditEntry   `json:"auditEntries"`
	Notifications []Notification `json:"notifications"`
}

// ExerciseOption exercises an option CLIN on a contract. It verifies:
//   - The CLIN exists on the contract
//   - The CLIN is of type OPTION
//   - The CLIN is in ACTIVE status
//   - The exercise deadline has not passed
//   - The new total obligation does not exceed the contract ceiling
//
// On success, it updates the CLIN status to EXERCISED, increases the contract's
// TotalObligated by the CLIN's Obligated amount, and creates a Modification record.
func ExerciseOption(contract *Contract, clinID string, now time.Time) error {
	return ExerciseOptionWithIDGen(contract, clinID, now, defaultModIDGenerator)
}

// ExerciseOptionWithIDGen is like ExerciseOption but accepts a custom ID generator for testing.
func ExerciseOptionWithIDGen(contract *Contract, clinID string, now time.Time, idGen IDGenerator) error {
	// Find the CLIN on the contract
	clinIdx := -1
	for i, clin := range contract.CLINs {
		if clin.CLINID == clinID {
			clinIdx = i
			break
		}
	}
	if clinIdx == -1 {
		return fmt.Errorf("CLIN %q does not exist on contract %s", clinID, contract.ContractID)
	}

	clin := &contract.CLINs[clinIdx]

	// Verify CLIN is an OPTION type
	if clin.CLINType != CLINTypeOption {
		return fmt.Errorf("CLIN %q is not an option type (type: %s)", clinID, clin.CLINType)
	}

	// Verify CLIN is in ACTIVE status
	if clin.CLINStatus != CLINStatusActive {
		return fmt.Errorf("CLIN %q is not in ACTIVE status (status: %s)", clinID, clin.CLINStatus)
	}

	// Verify exercise deadline has not passed
	if clin.OptionExerciseDeadline != nil && now.After(*clin.OptionExerciseDeadline) {
		return errors.New("option exercise deadline has expired")
	}

	// Verify new total obligation does not exceed contract ceiling
	newTotalObligated := contract.TotalObligated + clin.Obligated
	if newTotalObligated > contract.TotalCeiling {
		return fmt.Errorf(
			"exercising option would exceed contract ceiling: new obligation %.2f exceeds ceiling %.2f",
			newTotalObligated, contract.TotalCeiling,
		)
	}

	// All validations passed — apply changes

	// Update CLIN status to EXERCISED
	clin.CLINStatus = CLINStatusExercised

	// Increase contract TotalObligated by the CLIN's Obligated amount
	contract.TotalObligated += clin.Obligated

	// Create a Modification record on the contract
	mod := Modification{
		ModificationID: idGen(),
		Description:    fmt.Sprintf("Option exercised: CLIN %s (%s)", clin.CLINNumber, clin.Description),
		Amount:         clin.Obligated,
		CreatedAt:      now,
		CreatedBy:      "system",
	}
	contract.Modifications = append(contract.Modifications, mod)

	return nil
}

// defaultModIDGenerator returns a simple timestamp-based modification ID.
func defaultModIDGenerator() string {
	return fmt.Sprintf("MOD-%d", time.Now().UnixNano())
}
