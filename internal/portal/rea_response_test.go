package portal

import (
	"testing"
	"time"
)

func newTestREA() *REA {
	return &REA{
		REAID:           "REA-001",
		ContractID:      "CONTRACT-001",
		RequestedAmount: 60000.00,
		ApprovedAmount:  0,
		AffectedCLINs:   []string{"CLIN-001", "CLIN-002"},
		Status:          REAStatusSubmitted,
		Justification:   "Scope increase",
		SubmittedBy:     "contractor@example.com",
		SubmittedAt:     time.Now().Add(-24 * time.Hour),
	}
}

func TestRespondToREA_Approved(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()
	respondedAt := time.Now()

	originalCLIN1Ceiling := contract.CLINs[0].Ceiling
	originalCLIN2Ceiling := contract.CLINs[1].Ceiling
	originalTotalCeiling := contract.TotalCeiling

	response := REAResponse{
		ResponseType:   REAStatusApproved,
		ApprovedAmount: 0, // ignored for full approval
		Rationale:      "Scope change justified",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    respondedAt,
	}

	result, err := RespondToREA(contract, rea, response)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify REA status
	if result.REA.Status != REAStatusApproved {
		t.Errorf("expected status APPROVED, got %s", result.REA.Status)
	}

	// Verify approved amount equals requested amount
	if result.REA.ApprovedAmount != 60000.00 {
		t.Errorf("expected approved amount 60000.00, got %f", result.REA.ApprovedAmount)
	}

	// Verify resolvedAt is set
	if result.REA.ResolvedAt == nil {
		t.Fatal("expected resolvedAt to be set")
	}
	if !result.REA.ResolvedAt.Equal(respondedAt) {
		t.Errorf("expected resolvedAt=%v, got %v", respondedAt, *result.REA.ResolvedAt)
	}

	// Verify contract modification created
	if result.Modification == nil {
		t.Fatal("expected modification to be created")
	}
	if result.Modification.Amount != 60000.00 {
		t.Errorf("expected modification amount 60000.00, got %f", result.Modification.Amount)
	}
	if result.Modification.CreatedBy != "co@gov.example.com" {
		t.Errorf("expected modification created by co@gov.example.com, got %s", result.Modification.CreatedBy)
	}

	// Verify modification added to contract
	found := false
	for _, mod := range contract.Modifications {
		if mod.ModificationID == result.Modification.ModificationID {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected modification to be appended to contract.Modifications")
	}

	// Verify CLIN ceilings adjusted: 60000 / 2 CLINs = 30000 each
	expectedCLIN1 := originalCLIN1Ceiling + 30000.00
	expectedCLIN2 := originalCLIN2Ceiling + 30000.00
	if contract.CLINs[0].Ceiling != expectedCLIN1 {
		t.Errorf("expected CLIN-001 ceiling %f, got %f", expectedCLIN1, contract.CLINs[0].Ceiling)
	}
	if contract.CLINs[1].Ceiling != expectedCLIN2 {
		t.Errorf("expected CLIN-002 ceiling %f, got %f", expectedCLIN2, contract.CLINs[1].Ceiling)
	}

	// Verify total ceiling adjusted
	expectedTotalCeiling := originalTotalCeiling + 60000.00
	if contract.TotalCeiling != expectedTotalCeiling {
		t.Errorf("expected total ceiling %f, got %f", expectedTotalCeiling, contract.TotalCeiling)
	}

	// Verify audit entry
	if len(result.AuditEntries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(result.AuditEntries))
	}
	if result.AuditEntries[0].Action != "REA_APPROVED" {
		t.Errorf("expected action REA_APPROVED, got %s", result.AuditEntries[0].Action)
	}
}

func TestRespondToREA_PartiallyApproved(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()
	respondedAt := time.Now()

	originalCLIN1Ceiling := contract.CLINs[0].Ceiling
	originalCLIN2Ceiling := contract.CLINs[1].Ceiling
	originalTotalCeiling := contract.TotalCeiling

	response := REAResponse{
		ResponseType:   REAStatusPartiallyApproved,
		ApprovedAmount: 40000.00,
		Rationale:      "Partial scope change approved",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    respondedAt,
	}

	result, err := RespondToREA(contract, rea, response)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify REA status
	if result.REA.Status != REAStatusPartiallyApproved {
		t.Errorf("expected status PARTIALLY_APPROVED, got %s", result.REA.Status)
	}

	// Verify approved amount is from response (not requested)
	if result.REA.ApprovedAmount != 40000.00 {
		t.Errorf("expected approved amount 40000.00, got %f", result.REA.ApprovedAmount)
	}

	// Verify resolvedAt is set
	if result.REA.ResolvedAt == nil {
		t.Fatal("expected resolvedAt to be set")
	}

	// Verify contract modification created for partial amount
	if result.Modification == nil {
		t.Fatal("expected modification to be created")
	}
	if result.Modification.Amount != 40000.00 {
		t.Errorf("expected modification amount 40000.00, got %f", result.Modification.Amount)
	}

	// Verify CLIN ceilings adjusted proportionally: 40000 / 2 CLINs = 20000 each
	expectedCLIN1 := originalCLIN1Ceiling + 20000.00
	expectedCLIN2 := originalCLIN2Ceiling + 20000.00
	if contract.CLINs[0].Ceiling != expectedCLIN1 {
		t.Errorf("expected CLIN-001 ceiling %f, got %f", expectedCLIN1, contract.CLINs[0].Ceiling)
	}
	if contract.CLINs[1].Ceiling != expectedCLIN2 {
		t.Errorf("expected CLIN-002 ceiling %f, got %f", expectedCLIN2, contract.CLINs[1].Ceiling)
	}

	// Verify total ceiling adjusted by partial amount
	expectedTotalCeiling := originalTotalCeiling + 40000.00
	if contract.TotalCeiling != expectedTotalCeiling {
		t.Errorf("expected total ceiling %f, got %f", expectedTotalCeiling, contract.TotalCeiling)
	}

	// Verify audit entry
	if len(result.AuditEntries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(result.AuditEntries))
	}
	if result.AuditEntries[0].Action != "REA_PARTIALLY_APPROVED" {
		t.Errorf("expected action REA_PARTIALLY_APPROVED, got %s", result.AuditEntries[0].Action)
	}
}

func TestRespondToREA_Denied(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()
	respondedAt := time.Now()

	originalTotalCeiling := contract.TotalCeiling
	originalCLIN1Ceiling := contract.CLINs[0].Ceiling

	response := REAResponse{
		ResponseType:   REAStatusDenied,
		ApprovedAmount: 0,
		Rationale:      "Scope change not justified per contract terms",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    respondedAt,
	}

	result, err := RespondToREA(contract, rea, response)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify REA status
	if result.REA.Status != REAStatusDenied {
		t.Errorf("expected status DENIED, got %s", result.REA.Status)
	}

	// Verify rationale recorded
	if result.REA.ResponseRationale != "Scope change not justified per contract terms" {
		t.Errorf("expected rationale recorded, got %s", result.REA.ResponseRationale)
	}

	// Verify resolvedAt is set
	if result.REA.ResolvedAt == nil {
		t.Fatal("expected resolvedAt to be set")
	}

	// Verify no modification created
	if result.Modification != nil {
		t.Error("expected no modification for denied REA")
	}

	// Verify CLIN ceilings unchanged
	if contract.CLINs[0].Ceiling != originalCLIN1Ceiling {
		t.Errorf("expected CLIN ceiling unchanged, got %f", contract.CLINs[0].Ceiling)
	}

	// Verify total ceiling unchanged
	if contract.TotalCeiling != originalTotalCeiling {
		t.Errorf("expected total ceiling unchanged, got %f", contract.TotalCeiling)
	}

	// Verify audit entry
	if len(result.AuditEntries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(result.AuditEntries))
	}
	if result.AuditEntries[0].Action != "REA_DENIED" {
		t.Errorf("expected action REA_DENIED, got %s", result.AuditEntries[0].Action)
	}
}

func TestRespondToREA_AdditionalInfoRequested(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()
	respondedAt := time.Now()

	originalTotalCeiling := contract.TotalCeiling
	originalCLIN1Ceiling := contract.CLINs[0].Ceiling

	response := REAResponse{
		ResponseType:   REAStatusAdditionalInfoRequested,
		ApprovedAmount: 0,
		Rationale:      "Please provide detailed cost breakdown",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    respondedAt,
	}

	result, err := RespondToREA(contract, rea, response)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify REA status
	if result.REA.Status != REAStatusAdditionalInfoRequested {
		t.Errorf("expected status ADDITIONAL_INFO_REQUESTED, got %s", result.REA.Status)
	}

	// Verify resolvedAt is NOT set
	if result.REA.ResolvedAt != nil {
		t.Errorf("expected resolvedAt to be nil for info request, got %v", result.REA.ResolvedAt)
	}

	// Verify rationale recorded
	if result.REA.ResponseRationale != "Please provide detailed cost breakdown" {
		t.Errorf("expected rationale recorded, got %s", result.REA.ResponseRationale)
	}

	// Verify no modification created
	if result.Modification != nil {
		t.Error("expected no modification for info request")
	}

	// Verify CLIN ceilings unchanged
	if contract.CLINs[0].Ceiling != originalCLIN1Ceiling {
		t.Errorf("expected CLIN ceiling unchanged, got %f", contract.CLINs[0].Ceiling)
	}

	// Verify total ceiling unchanged
	if contract.TotalCeiling != originalTotalCeiling {
		t.Errorf("expected total ceiling unchanged, got %f", contract.TotalCeiling)
	}

	// Verify audit entry
	if len(result.AuditEntries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(result.AuditEntries))
	}
	if result.AuditEntries[0].Action != "REA_ADDITIONAL_INFO_REQUESTED" {
		t.Errorf("expected action REA_ADDITIONAL_INFO_REQUESTED, got %s", result.AuditEntries[0].Action)
	}
}

func TestRespondToREA_PartiallyApproved_InvalidAmount_Zero(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()

	response := REAResponse{
		ResponseType:   REAStatusPartiallyApproved,
		ApprovedAmount: 0,
		Rationale:      "test",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    time.Now(),
	}

	result, err := RespondToREA(contract, rea, response)
	if err == nil {
		t.Fatal("expected error for zero approved amount in partial approval")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestRespondToREA_PartiallyApproved_AmountExceedsRequested(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()

	response := REAResponse{
		ResponseType:   REAStatusPartiallyApproved,
		ApprovedAmount: 70000.00, // more than requested 60000
		Rationale:      "test",
		RespondedBy:    "co@gov.example.com",
		RespondedAt:    time.Now(),
	}

	result, err := RespondToREA(contract, rea, response)
	if err == nil {
		t.Fatal("expected error when partial approval amount >= requested")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestRespondToREA_NilREA(t *testing.T) {
	contract := newREATestContract()

	response := REAResponse{
		ResponseType: REAStatusApproved,
		Rationale:    "test",
		RespondedBy:  "co@gov.example.com",
		RespondedAt:  time.Now(),
	}

	result, err := RespondToREA(contract, nil, response)
	if err == nil {
		t.Fatal("expected error for nil REA")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestRespondToREA_NilContract(t *testing.T) {
	rea := newTestREA()

	response := REAResponse{
		ResponseType: REAStatusApproved,
		Rationale:    "test",
		RespondedBy:  "co@gov.example.com",
		RespondedAt:  time.Now(),
	}

	result, err := RespondToREA(nil, rea, response)
	if err == nil {
		t.Fatal("expected error for nil contract")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestRespondToREA_UnsupportedResponseType(t *testing.T) {
	contract := newREATestContract()
	rea := newTestREA()

	response := REAResponse{
		ResponseType: REAStatus("INVALID_TYPE"),
		Rationale:    "test",
		RespondedBy:  "co@gov.example.com",
		RespondedAt:  time.Now(),
	}

	result, err := RespondToREA(contract, rea, response)
	if err == nil {
		t.Fatal("expected error for unsupported response type")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestRespondToREA_Approved_SingleCLIN(t *testing.T) {
	contract := newREATestContract()
	rea := &REA{
		REAID:           "REA-002",
		ContractID:      "CONTRACT-001",
		RequestedAmount: 30000.00,
		ApprovedAmount:  0,
		AffectedCLINs:   []string{"CLIN-003"},
		Status:          REAStatusSubmitted,
		Justification:   "Travel cost increase",
		SubmittedBy:     "contractor@example.com",
		SubmittedAt:     time.Now().Add(-12 * time.Hour),
	}

	originalCLIN3Ceiling := contract.CLINs[2].Ceiling

	response := REAResponse{
		ResponseType: REAStatusApproved,
		Rationale:    "Travel increase justified",
		RespondedBy:  "co@gov.example.com",
		RespondedAt:  time.Now(),
	}

	result, err := RespondToREA(contract, rea, response)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// With 1 CLIN, entire amount goes to that CLIN
	expectedCLIN3 := originalCLIN3Ceiling + 30000.00
	if contract.CLINs[2].Ceiling != expectedCLIN3 {
		t.Errorf("expected CLIN-003 ceiling %f, got %f", expectedCLIN3, contract.CLINs[2].Ceiling)
	}

	if result.REA.ApprovedAmount != 30000.00 {
		t.Errorf("expected approved amount 30000.00, got %f", result.REA.ApprovedAmount)
	}
}
