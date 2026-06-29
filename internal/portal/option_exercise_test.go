package portal

import (
	"strings"
	"testing"
	"time"
)

func testModIDGen() string {
	return "MOD-TEST-001"
}

func newOptionTestContract() *Contract {
	deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	return &Contract{
		ContractID:     "C-001",
		ContractNumber: "W911NF-25-C-0001",
		ContractType:   ContractTypeFFP,
		TotalCeiling:   1000000.00,
		TotalObligated: 500000.00,
		TotalExpended:  200000.00,
		Status:         ContractStatusActive,
		CLINs: []ContractLineItem{
			{
				CLINID:                 "CLIN-001",
				CLINNumber:             "0001",
				Description:            "Base period labor",
				CLINType:               CLINTypeFFP,
				CLINStatus:             CLINStatusActive,
				Ceiling:                300000.00,
				Obligated:              300000.00,
				Expended:               200000.00,
			},
			{
				CLINID:                 "CLIN-002",
				CLINNumber:             "0002",
				Description:            "Option Year 1 labor",
				CLINType:               CLINTypeOption,
				CLINStatus:             CLINStatusActive,
				Ceiling:                200000.00,
				Obligated:              200000.00,
				Expended:               0,
				OptionExerciseDeadline: &deadline,
			},
		},
		Modifications: []Modification{},
	}
}

func TestExerciseOption_ValidExercise(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	err := ExerciseOptionWithIDGen(contract, "CLIN-002", now, testModIDGen)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify CLIN status updated to EXERCISED
	if contract.CLINs[1].CLINStatus != CLINStatusExercised {
		t.Errorf("expected CLIN status EXERCISED, got: %s", contract.CLINs[1].CLINStatus)
	}

	// Verify contract TotalObligated increased
	expectedObligated := 700000.00
	if contract.TotalObligated != expectedObligated {
		t.Errorf("expected TotalObligated %.2f, got: %.2f", expectedObligated, contract.TotalObligated)
	}

	// Verify modification was created
	if len(contract.Modifications) != 1 {
		t.Fatalf("expected 1 modification, got: %d", len(contract.Modifications))
	}
	mod := contract.Modifications[0]
	if mod.ModificationID != "MOD-TEST-001" {
		t.Errorf("expected modification ID MOD-TEST-001, got: %s", mod.ModificationID)
	}
	if mod.Amount != 200000.00 {
		t.Errorf("expected modification amount 200000.00, got: %.2f", mod.Amount)
	}
}

func TestExerciseOption_CLINNotFound(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	err := ExerciseOptionWithIDGen(contract, "CLIN-NONEXISTENT", now, testModIDGen)
	if err == nil {
		t.Fatal("expected error for non-existent CLIN, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected error about CLIN not existing, got: %v", err)
	}
}

func TestExerciseOption_WrongType(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	// CLIN-001 is FFP, not OPTION
	err := ExerciseOptionWithIDGen(contract, "CLIN-001", now, testModIDGen)
	if err == nil {
		t.Fatal("expected error for non-option CLIN, got nil")
	}
	if !strings.Contains(err.Error(), "not an option type") {
		t.Errorf("expected error about wrong type, got: %v", err)
	}
}

func TestExerciseOption_WrongStatus(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	// Set the option CLIN to a non-ACTIVE status
	contract.CLINs[1].CLINStatus = CLINStatusExercised

	err := ExerciseOptionWithIDGen(contract, "CLIN-002", now, testModIDGen)
	if err == nil {
		t.Fatal("expected error for wrong status, got nil")
	}
	if !strings.Contains(err.Error(), "not in ACTIVE status") {
		t.Errorf("expected error about status, got: %v", err)
	}
}

func TestExerciseOption_ExpiredDeadline(t *testing.T) {
	contract := newOptionTestContract()
	// Set now to after the deadline (2025-12-31)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	err := ExerciseOptionWithIDGen(contract, "CLIN-002", now, testModIDGen)
	if err == nil {
		t.Fatal("expected error for expired deadline, got nil")
	}
	if !strings.Contains(err.Error(), "deadline has expired") {
		t.Errorf("expected error about expired deadline, got: %v", err)
	}
}

func TestExerciseOption_WouldExceedCeiling(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	// Set obligation so high that exercising the option would exceed ceiling
	contract.TotalObligated = 900000.00 // CLIN-002 obligated is 200000, total would be 1,100,000 > ceiling of 1,000,000

	err := ExerciseOptionWithIDGen(contract, "CLIN-002", now, testModIDGen)
	if err == nil {
		t.Fatal("expected error for exceeding ceiling, got nil")
	}
	if !strings.Contains(err.Error(), "exceed contract ceiling") {
		t.Errorf("expected error about exceeding ceiling, got: %v", err)
	}
}

func TestExerciseOption_NilDeadline(t *testing.T) {
	contract := newOptionTestContract()
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	// Remove the deadline — should still succeed regardless of time
	contract.CLINs[1].OptionExerciseDeadline = nil

	err := ExerciseOptionWithIDGen(contract, "CLIN-002", now, testModIDGen)
	if err != nil {
		t.Fatalf("expected no error when deadline is nil, got: %v", err)
	}
}
