package portal

import "testing"

func TestValidateObligationIntegrity_UnderCeiling(t *testing.T) {
	contract := &Contract{
		TotalCeiling:   1000000.00,
		TotalObligated: 750000.00,
	}

	err := ValidateObligationIntegrity(contract)
	if err != nil {
		t.Errorf("expected no error when obligation is under ceiling, got: %v", err)
	}
}

func TestValidateObligationIntegrity_EqualToCeiling(t *testing.T) {
	contract := &Contract{
		TotalCeiling:   500000.00,
		TotalObligated: 500000.00,
	}

	err := ValidateObligationIntegrity(contract)
	if err != nil {
		t.Errorf("expected no error when obligation equals ceiling, got: %v", err)
	}
}

func TestValidateObligationIntegrity_ExceedsCeiling(t *testing.T) {
	contract := &Contract{
		TotalCeiling:   1000000.00,
		TotalObligated: 1000001.00,
	}

	err := ValidateObligationIntegrity(contract)
	if err == nil {
		t.Error("expected error when obligation exceeds ceiling, got nil")
	}
}

func TestValidateObligationIntegrity_ZeroValues(t *testing.T) {
	contract := &Contract{
		TotalCeiling:   0.0,
		TotalObligated: 0.0,
	}

	err := ValidateObligationIntegrity(contract)
	if err != nil {
		t.Errorf("expected no error with zero values, got: %v", err)
	}
}

func TestValidateCLINExpenditureLimits_WithinLimits(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 50000.00},
			{CLINID: "CLIN-002", Obligated: 200000.00, Expended: 150000.00},
			{CLINID: "CLIN-003", Obligated: 75000.00, Expended: 75000.00},
		},
	}

	violations := ValidateCLINExpenditureLimits(contract)
	if len(violations) != 0 {
		t.Errorf("expected no violations when all CLINs are within limits, got %d", len(violations))
	}
}

func TestValidateCLINExpenditureLimits_ExceedingLimit(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 50000.00},
			{CLINID: "CLIN-002", Obligated: 200000.00, Expended: 250000.00}, // exceeds
			{CLINID: "CLIN-003", Obligated: 75000.00, Expended: 80000.00},   // exceeds
		},
	}

	violations := ValidateCLINExpenditureLimits(contract)
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	// Check first violation (CLIN-002)
	if violations[0].CLINID != "CLIN-002" {
		t.Errorf("expected first violation CLINID=CLIN-002, got %s", violations[0].CLINID)
	}
	if violations[0].Excess != 50000.00 {
		t.Errorf("expected first violation excess=50000.00, got %f", violations[0].Excess)
	}

	// Check second violation (CLIN-003)
	if violations[1].CLINID != "CLIN-003" {
		t.Errorf("expected second violation CLINID=CLIN-003, got %s", violations[1].CLINID)
	}
	if violations[1].Excess != 5000.00 {
		t.Errorf("expected second violation excess=5000.00, got %f", violations[1].Excess)
	}
}

func TestValidateCLINExpenditureLimits_EmptyCLINs(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{},
	}

	violations := ValidateCLINExpenditureLimits(contract)
	if len(violations) != 0 {
		t.Errorf("expected no violations for empty CLINs, got %d", len(violations))
	}
}

func TestCheckExpenditureAllowed_WithinLimit(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-001",
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 50000.00},
		},
	}

	err := CheckExpenditureAllowed(contract, "CLIN-001", 25000.00)
	if err != nil {
		t.Errorf("expected expenditure to be allowed, got: %v", err)
	}
}

func TestCheckExpenditureAllowed_ExactlyAtLimit(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-001",
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 50000.00},
		},
	}

	err := CheckExpenditureAllowed(contract, "CLIN-001", 50000.00)
	if err != nil {
		t.Errorf("expected expenditure at exact limit to be allowed, got: %v", err)
	}
}

func TestCheckExpenditureAllowed_ExceedsLimit(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-001",
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 90000.00},
		},
	}

	err := CheckExpenditureAllowed(contract, "CLIN-001", 15000.00)
	if err == nil {
		t.Error("expected error when expenditure would exceed obligation, got nil")
	}
}

func TestCheckExpenditureAllowed_CLINNotFound(t *testing.T) {
	contract := &Contract{
		ContractID: "CTR-001",
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", Obligated: 100000.00, Expended: 50000.00},
		},
	}

	err := CheckExpenditureAllowed(contract, "CLIN-999", 10000.00)
	if err == nil {
		t.Error("expected error for non-existent CLIN, got nil")
	}
}
