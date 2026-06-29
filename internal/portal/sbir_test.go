package portal

import (
	"strings"
	"testing"
)

func TestValidateSBIRInvoice_ValidActiveCLIN(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 10000.00, ContractTypeFFP, true)
	if err != nil {
		t.Errorf("expected no error for valid SBIR invoice with ACTIVE CLIN, got: %v", err)
	}
}

func TestValidateSBIRInvoice_ValidExercisedCLIN(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-002",
				CLINStatus: CLINStatusExercised,
				Obligated:  200000.00,
				Expended:   80000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-002", 25000.00, ContractTypeCPFF, true)
	if err != nil {
		t.Errorf("expected no error for valid SBIR invoice with EXERCISED CLIN, got: %v", err)
	}
}

func TestValidateSBIRInvoice_InvalidCLINStatus_Completed(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusCompleted,
				Obligated:  100000.00,
				Expended:   100000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 5000.00, ContractTypeFFP, true)
	if err == nil {
		t.Fatal("expected error for COMPLETED CLIN status, got nil")
	}
	if !strings.Contains(err.Error(), "COMPLETED") {
		t.Errorf("expected error to mention COMPLETED status, got: %v", err)
	}
}

func TestValidateSBIRInvoice_InvalidCLINStatus_Expired(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusExpired,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 5000.00, ContractTypeFFP, true)
	if err == nil {
		t.Fatal("expected error for EXPIRED CLIN status, got nil")
	}
	if !strings.Contains(err.Error(), "EXPIRED") {
		t.Errorf("expected error to mention EXPIRED status, got: %v", err)
	}
}

func TestValidateSBIRInvoice_InvalidCLINStatus_NotExercised(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusNotExercised,
				Obligated:  100000.00,
				Expended:   0.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 5000.00, ContractTypeFFP, true)
	if err == nil {
		t.Fatal("expected error for NOT_EXERCISED CLIN status, got nil")
	}
	if !strings.Contains(err.Error(), "NOT_EXERCISED") {
		t.Errorf("expected error to mention NOT_EXERCISED status, got: %v", err)
	}
}

func TestValidateSBIRInvoice_ExpenditureExceedsObligation(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   95000.00,
			},
		},
	}

	// Trying to invoice 10000 when only 5000 of obligation remains
	err := ValidateSBIRInvoice(contract, "CLIN-001", 10000.00, ContractTypeFFP, true)
	if err == nil {
		t.Fatal("expected error when invoice would exceed CLIN obligation, got nil")
	}
	if !strings.Contains(err.Error(), "held") {
		t.Errorf("expected error to indicate payment held, got: %v", err)
	}
}

func TestValidateSBIRInvoice_ExpenditureExactlyAtObligation(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   90000.00,
			},
		},
	}

	// Exactly at the limit should be allowed
	err := ValidateSBIRInvoice(contract, "CLIN-001", 10000.00, ContractTypeFFP, true)
	if err != nil {
		t.Errorf("expected no error when invoice exactly meets obligation, got: %v", err)
	}
}

func TestValidateSBIRInvoice_FFPWithoutMilestone(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 10000.00, ContractTypeFFP, false)
	if err == nil {
		t.Fatal("expected error for FFP contract without milestone acceptance, got nil")
	}
	if !strings.Contains(err.Error(), "milestone not accepted") {
		t.Errorf("expected error to mention milestone not accepted, got: %v", err)
	}
}

func TestValidateSBIRInvoice_FFPWithMilestone(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-001", 10000.00, ContractTypeFFP, true)
	if err != nil {
		t.Errorf("expected no error for FFP with accepted milestone, got: %v", err)
	}
}

func TestValidateSBIRInvoice_CPFFCostNotAllowable(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoiceWithAllowability(SBIRInvoiceParams{
		Contract:       contract,
		CLINID:         "CLIN-001",
		InvoiceAmount:  10000.00,
		ContractType:   ContractTypeCPFF,
		CostAllowable: false,
	})
	if err == nil {
		t.Fatal("expected error for CPFF with non-allowable cost, got nil")
	}
	if !strings.Contains(err.Error(), "cost not allowable") {
		t.Errorf("expected error to mention cost not allowable, got: %v", err)
	}
}

func TestValidateSBIRInvoice_CPIFCostNotAllowable(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoiceWithAllowability(SBIRInvoiceParams{
		Contract:       contract,
		CLINID:         "CLIN-001",
		InvoiceAmount:  10000.00,
		ContractType:   ContractTypeCPIF,
		CostAllowable: false,
	})
	if err == nil {
		t.Fatal("expected error for CPIF with non-allowable cost, got nil")
	}
	if !strings.Contains(err.Error(), "cost not allowable") {
		t.Errorf("expected error to mention cost not allowable, got: %v", err)
	}
}

func TestValidateSBIRInvoice_CPFFCostAllowable(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoiceWithAllowability(SBIRInvoiceParams{
		Contract:       contract,
		CLINID:         "CLIN-001",
		InvoiceAmount:  10000.00,
		ContractType:   ContractTypeCPFF,
		CostAllowable: true,
	})
	if err != nil {
		t.Errorf("expected no error for CPFF with allowable cost, got: %v", err)
	}
}

func TestValidateSBIRInvoice_CLINNotFound(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-SBIR-001",
		CLINs: []ContractLineItem{
			{
				CLINID:     "CLIN-001",
				CLINStatus: CLINStatusActive,
				Obligated:  100000.00,
				Expended:   50000.00,
			},
		},
	}

	err := ValidateSBIRInvoice(contract, "CLIN-999", 10000.00, ContractTypeFFP, true)
	if err == nil {
		t.Fatal("expected error for non-existent CLIN, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention CLIN not found, got: %v", err)
	}
}
