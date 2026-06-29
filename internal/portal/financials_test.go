package portal

import (
	"math"
	"testing"
)

func TestCalculateContractTotals_EmptyCLINs(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{},
	}

	totalObligated, totalExpended := CalculateContractTotals(contract)

	if totalObligated != 0.0 {
		t.Errorf("expected totalObligated=0.0, got %f", totalObligated)
	}
	if totalExpended != 0.0 {
		t.Errorf("expected totalExpended=0.0, got %f", totalExpended)
	}
}

func TestCalculateContractTotals_SingleCLIN(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{
			{Obligated: 50000.00, Expended: 30000.00},
		},
	}

	totalObligated, totalExpended := CalculateContractTotals(contract)

	if totalObligated != 50000.00 {
		t.Errorf("expected totalObligated=50000.00, got %f", totalObligated)
	}
	if totalExpended != 30000.00 {
		t.Errorf("expected totalExpended=30000.00, got %f", totalExpended)
	}
}

func TestCalculateContractTotals_MultipleCLINs(t *testing.T) {
	contract := &Contract{
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
			{Obligated: 50000.00, Expended: 25000.00},
		},
	}

	totalObligated, totalExpended := CalculateContractTotals(contract)

	if math.Abs(totalObligated-350000.00) > epsilon {
		t.Errorf("expected totalObligated=350000.00, got %f", totalObligated)
	}
	if math.Abs(totalExpended-235000.00) > epsilon {
		t.Errorf("expected totalExpended=235000.00, got %f", totalExpended)
	}
}

func TestValidateCLINSummation_ValidMatch(t *testing.T) {
	contract := &Contract{
		TotalObligated: 350000.00,
		TotalExpended:  235000.00,
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
			{Obligated: 50000.00, Expended: 25000.00},
		},
	}

	err := ValidateCLINSummation(contract)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCLINSummation_ObligatedMismatch(t *testing.T) {
	contract := &Contract{
		TotalObligated: 400000.00, // wrong - should be 350000
		TotalExpended:  235000.00,
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
			{Obligated: 50000.00, Expended: 25000.00},
		},
	}

	err := ValidateCLINSummation(contract)
	if err == nil {
		t.Error("expected error for obligated mismatch, got nil")
	}
}

func TestValidateCLINSummation_ExpendedMismatch(t *testing.T) {
	contract := &Contract{
		TotalObligated: 350000.00,
		TotalExpended:  300000.00, // wrong - should be 235000
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
			{Obligated: 50000.00, Expended: 25000.00},
		},
	}

	err := ValidateCLINSummation(contract)
	if err == nil {
		t.Error("expected error for expended mismatch, got nil")
	}
}

func TestValidateCLINSummation_BothMismatch(t *testing.T) {
	contract := &Contract{
		TotalObligated: 999999.00,
		TotalExpended:  888888.00,
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
		},
	}

	err := ValidateCLINSummation(contract)
	if err == nil {
		t.Error("expected error for both mismatch, got nil")
	}
}

func TestValidateCLINSummation_EmptyCLINsZeroTotals(t *testing.T) {
	contract := &Contract{
		TotalObligated: 0.0,
		TotalExpended:  0.0,
		CLINs:          []ContractLineItem{},
	}

	err := ValidateCLINSummation(contract)
	if err != nil {
		t.Errorf("expected no error for empty CLINs with zero totals, got: %v", err)
	}
}

func TestValidateCLINSummation_EmptyCLINsNonZeroTotals(t *testing.T) {
	contract := &Contract{
		TotalObligated: 100000.00,
		TotalExpended:  50000.00,
		CLINs:          []ContractLineItem{},
	}

	err := ValidateCLINSummation(contract)
	if err == nil {
		t.Error("expected error for empty CLINs with non-zero totals, got nil")
	}
}

func TestValidateCLINSummation_FloatingPointTolerance(t *testing.T) {
	// Values that would produce floating point rounding issues without epsilon tolerance
	contract := &Contract{
		TotalObligated: 0.3,
		TotalExpended:  0.3,
		CLINs: []ContractLineItem{
			{Obligated: 0.1, Expended: 0.1},
			{Obligated: 0.1, Expended: 0.1},
			{Obligated: 0.1, Expended: 0.1},
		},
	}

	err := ValidateCLINSummation(contract)
	if err != nil {
		t.Errorf("expected no error with floating point tolerance, got: %v", err)
	}
}

func TestValidateCLINSummation_SingleCLINMatching(t *testing.T) {
	contract := &Contract{
		TotalObligated: 75000.50,
		TotalExpended:  42000.25,
		CLINs: []ContractLineItem{
			{Obligated: 75000.50, Expended: 42000.25},
		},
	}

	err := ValidateCLINSummation(contract)
	if err != nil {
		t.Errorf("expected no error for single CLIN matching, got: %v", err)
	}
}

func TestRecalculateContractTotals_UpdatesFromCLINs(t *testing.T) {
	contract := &Contract{
		TotalObligated: 0.0, // intentionally wrong
		TotalExpended:  0.0, // intentionally wrong
		CLINs: []ContractLineItem{
			{Obligated: 100000.00, Expended: 60000.00},
			{Obligated: 200000.00, Expended: 150000.00},
		},
	}

	RecalculateContractTotals(contract)

	if math.Abs(contract.TotalObligated-300000.00) > epsilon {
		t.Errorf("expected TotalObligated=300000.00, got %f", contract.TotalObligated)
	}
	if math.Abs(contract.TotalExpended-210000.00) > epsilon {
		t.Errorf("expected TotalExpended=210000.00, got %f", contract.TotalExpended)
	}
}

func TestRecalculateContractTotals_EmptyCLINs(t *testing.T) {
	contract := &Contract{
		TotalObligated: 100000.00, // should be reset to 0
		TotalExpended:  50000.00,  // should be reset to 0
		CLINs:          []ContractLineItem{},
	}

	RecalculateContractTotals(contract)

	if contract.TotalObligated != 0.0 {
		t.Errorf("expected TotalObligated=0.0, got %f", contract.TotalObligated)
	}
	if contract.TotalExpended != 0.0 {
		t.Errorf("expected TotalExpended=0.0, got %f", contract.TotalExpended)
	}
}

func TestRecalculateContractTotals_ValidationPassesAfterRecalculate(t *testing.T) {
	contract := &Contract{
		TotalObligated: 999.00, // wrong values
		TotalExpended:  888.00,
		CLINs: []ContractLineItem{
			{Obligated: 50000.00, Expended: 30000.00},
			{Obligated: 75000.00, Expended: 50000.00},
		},
	}

	// Should fail before recalculation
	err := ValidateCLINSummation(contract)
	if err == nil {
		t.Error("expected error before recalculation")
	}

	// Recalculate and validate
	RecalculateContractTotals(contract)
	err = ValidateCLINSummation(contract)
	if err != nil {
		t.Errorf("expected no error after recalculation, got: %v", err)
	}
}
