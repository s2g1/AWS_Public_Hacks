package portal

import (
	"strings"
	"testing"
)

func newREATestContract() *Contract {
	return &Contract{
		ContractID:     "CONTRACT-001",
		ContractNumber: "W15QKN-22-C-0001",
		ContractType:   ContractTypeFFP,
		TotalCeiling:   1000000.00,
		TotalObligated: 500000.00,
		TotalExpended:  250000.00,
		CLINs: []ContractLineItem{
			{CLINID: "CLIN-001", CLINNumber: "0001", Description: "Base period services", CLINType: CLINTypeFFP, CLINStatus: CLINStatusActive, Ceiling: 500000, Obligated: 300000, Expended: 150000},
			{CLINID: "CLIN-002", CLINNumber: "0002", Description: "Option period 1", CLINType: CLINTypeOption, CLINStatus: CLINStatusActive, Ceiling: 300000, Obligated: 150000, Expended: 75000},
			{CLINID: "CLIN-003", CLINNumber: "0003", Description: "Travel", CLINType: CLINTypeTAndM, CLINStatus: CLINStatusActive, Ceiling: 200000, Obligated: 50000, Expended: 25000},
		},
	}
}

func testIDGenerator() string {
	return "REA-TEST-001"
}

func TestSubmitREA_ValidSubmission(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 50000.00,
		AffectedCLINs:   []string{"CLIN-001", "CLIN-002"},
		Justification:   "Scope increase due to additional requirements",
		SubmittedBy:     "contractor@example.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify REA record
	if result.REA == nil {
		t.Fatal("expected REA to be created")
	}
	if result.REA.REAID != "REA-TEST-001" {
		t.Errorf("expected REAID=REA-TEST-001, got %s", result.REA.REAID)
	}
	if result.REA.ContractID != "CONTRACT-001" {
		t.Errorf("expected ContractID=CONTRACT-001, got %s", result.REA.ContractID)
	}
	if result.REA.RequestedAmount != 50000.00 {
		t.Errorf("expected RequestedAmount=50000.00, got %f", result.REA.RequestedAmount)
	}
	if result.REA.ApprovedAmount != 0 {
		t.Errorf("expected ApprovedAmount=0, got %f", result.REA.ApprovedAmount)
	}
	if result.REA.Status != REAStatusSubmitted {
		t.Errorf("expected Status=SUBMITTED, got %s", result.REA.Status)
	}
	if result.REA.Justification != "Scope increase due to additional requirements" {
		t.Errorf("unexpected Justification: %s", result.REA.Justification)
	}
	if result.REA.SubmittedBy != "contractor@example.com" {
		t.Errorf("expected SubmittedBy=contractor@example.com, got %s", result.REA.SubmittedBy)
	}
	if result.REA.SubmittedAt.IsZero() {
		t.Error("expected SubmittedAt to be set")
	}
	if len(result.REA.AffectedCLINs) != 2 {
		t.Errorf("expected 2 affected CLINs, got %d", len(result.REA.AffectedCLINs))
	}

	// Verify audit trail
	if len(result.AuditEntries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(result.AuditEntries))
	}
	audit := result.AuditEntries[0]
	if audit.Action != "REA_SUBMITTED" {
		t.Errorf("expected Action=REA_SUBMITTED, got %s", audit.Action)
	}
	if audit.Actor != "contractor@example.com" {
		t.Errorf("expected Actor=contractor@example.com, got %s", audit.Actor)
	}

	// Verify notification
	if len(result.Notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(result.Notifications))
	}
	notif := result.Notifications[0]
	if notif.Recipient != "contracting_officer" {
		t.Errorf("expected Recipient=contracting_officer, got %s", notif.Recipient)
	}
	if !strings.Contains(notif.Subject, "W15QKN-22-C-0001") {
		t.Errorf("expected notification subject to contain contract number, got: %s", notif.Subject)
	}
}

func TestSubmitREA_ValidSubmission_SingleCLIN(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 10000.00,
		AffectedCLINs:   []string{"CLIN-003"},
		Justification:   "Travel cost increase",
		SubmittedBy:     "pm@contractor.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.REA.Status != REAStatusSubmitted {
		t.Errorf("expected Status=SUBMITTED, got %s", result.REA.Status)
	}
	if len(result.REA.AffectedCLINs) != 1 {
		t.Errorf("expected 1 affected CLIN, got %d", len(result.REA.AffectedCLINs))
	}
}

func TestSubmitREA_ZeroAmount(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 0,
		AffectedCLINs:   []string{"CLIN-001"},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("expected error to mention 'positive', got: %v", err)
	}
}

func TestSubmitREA_NegativeAmount(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: -5000.00,
		AffectedCLINs:   []string{"CLIN-001"},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("expected error to mention 'positive', got: %v", err)
	}
}

func TestSubmitREA_EmptyCLINs(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 25000.00,
		AffectedCLINs:   []string{},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for empty CLINs, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("expected error to mention 'at least one', got: %v", err)
	}
}

func TestSubmitREA_NilCLINs(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 25000.00,
		AffectedCLINs:   nil,
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for nil CLINs, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("expected error to mention 'at least one', got: %v", err)
	}
}

func TestSubmitREA_NonExistentCLIN(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 25000.00,
		AffectedCLINs:   []string{"CLIN-001", "CLIN-999"},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for non-existent CLIN, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "CLIN-999") {
		t.Errorf("expected error to mention 'CLIN-999', got: %v", err)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected error to mention 'does not exist', got: %v", err)
	}
}

func TestSubmitREA_AllNonExistentCLINs(t *testing.T) {
	contract := newREATestContract()
	req := REASubmissionRequest{
		RequestedAmount: 25000.00,
		AffectedCLINs:   []string{"CLIN-FAKE"},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error for non-existent CLIN, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
	if !strings.Contains(err.Error(), "CLIN-FAKE") {
		t.Errorf("expected error to mention 'CLIN-FAKE', got: %v", err)
	}
}

func TestSubmitREA_ContractWithNoCLINs(t *testing.T) {
	contract := &Contract{
		ContractID:     "CONTRACT-EMPTY",
		ContractNumber: "EMPTY-001",
		CLINs:          []ContractLineItem{},
	}
	req := REASubmissionRequest{
		RequestedAmount: 10000.00,
		AffectedCLINs:   []string{"CLIN-001"},
		Justification:   "test",
		SubmittedBy:     "user@test.com",
	}

	result, err := SubmitREAWithIDGen(contract, req, testIDGenerator)
	if err == nil {
		t.Fatal("expected error when contract has no CLINs, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid submission")
	}
}
