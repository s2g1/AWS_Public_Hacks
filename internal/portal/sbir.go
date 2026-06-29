package portal

import "fmt"

// SBIRInvoiceParams contains the parameters for validating an SBIR invoice.
type SBIRInvoiceParams struct {
	Contract          *Contract
	CLINID            string
	InvoiceAmount     float64
	ContractType      ContractType
	MilestoneAccepted bool
	CostAllowable    bool
}

// ValidateSBIRInvoice validates an SBIR invoice against CLIN obligations and
// contract type rules before approving payment.
//
// It performs the following checks:
// 1. Verifies the referenced CLIN is in ACTIVE or EXERCISED status
// 2. Holds payment if invoice amount would cause CLIN expenditure to exceed obligation
// 3. For CPFF/CPIF contracts: verifies cost allowability (returns error if not marked allowable)
// 4. For FFP contracts: verifies milestone has been accepted (returns error if not)
//
// Validates: Requirements 21.1, 21.2, 21.3, 21.4
func ValidateSBIRInvoice(contract *Contract, clinID string, invoiceAmount float64, contractType ContractType, milestoneAccepted bool) error {
	return ValidateSBIRInvoiceWithAllowability(SBIRInvoiceParams{
		Contract:          contract,
		CLINID:            clinID,
		InvoiceAmount:     invoiceAmount,
		ContractType:      contractType,
		MilestoneAccepted: milestoneAccepted,
		CostAllowable:    true, // default to allowable for backward compatibility
	})
}

// ValidateSBIRInvoiceWithAllowability validates an SBIR invoice with explicit
// cost allowability control for cost-type contracts.
//
// Validates: Requirements 21.1, 21.2, 21.3, 21.4
func ValidateSBIRInvoiceWithAllowability(params SBIRInvoiceParams) error {
	// Step 1: Find the referenced CLIN and verify status
	clin, err := findCLIN(params.Contract, params.CLINID)
	if err != nil {
		return err
	}

	if clin.CLINStatus != CLINStatusActive && clin.CLINStatus != CLINStatusExercised {
		return fmt.Errorf(
			"SBIR invoice rejected: CLIN %s is in %s status; must be ACTIVE or EXERCISED",
			params.CLINID, clin.CLINStatus,
		)
	}

	// Step 2: Check if invoice amount would cause expenditure to exceed obligation
	if err := CheckExpenditureAllowed(params.Contract, params.CLINID, params.InvoiceAmount); err != nil {
		return fmt.Errorf(
			"SBIR invoice held: %w",
			err,
		)
	}

	// Step 3: Contract type-specific validation
	switch params.ContractType {
	case ContractTypeCPFF, ContractTypeCPIF:
		// For cost-type contracts, verify cost allowability
		if !params.CostAllowable {
			return fmt.Errorf(
				"SBIR invoice rejected: cost not allowable for %s contract on CLIN %s",
				params.ContractType, params.CLINID,
			)
		}
	case ContractTypeFFP:
		// For firm-fixed-price contracts, verify milestone acceptance
		if !params.MilestoneAccepted {
			return fmt.Errorf(
				"SBIR invoice rejected: milestone not accepted for FFP contract on CLIN %s",
				params.CLINID,
			)
		}
	}

	return nil
}

// findCLIN looks up a CLIN by ID on the contract.
func findCLIN(contract *Contract, clinID string) (*ContractLineItem, error) {
	for i := range contract.CLINs {
		if contract.CLINs[i].CLINID == clinID {
			return &contract.CLINs[i], nil
		}
	}
	return nil, fmt.Errorf("CLIN %q not found on contract %s", clinID, contract.ContractID)
}
