package routing

import (
	"testing"
	"time"

	"federal-payment-processing/internal/models"
)

// helper to build an ExtractionResult with a given amount string.
func extractionWithAmount(amount string) *models.ExtractionResult {
	return &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      amount,
				Normalized: amount,
				Confidence: 0.95,
			},
			"payee": {
				Value:      "Test Vendor",
				Normalized: "Test Vendor",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
}

// helper to build an ExtractionResult with amount and date fields.
func extractionWithAmountAndDate(amount, date string) *models.ExtractionResult {
	return &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      amount,
				Normalized: amount,
				Confidence: 0.95,
			},
			"payee": {
				Value:      "Test Vendor",
				Normalized: "Test Vendor",
				Confidence: 0.95,
			},
			"date": {
				Value:      date,
				Normalized: date,
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
}

func compliantResult() *models.ComplianceResult {
	return &models.ComplianceResult{
		Status: models.ComplianceStatusCompliant,
		Flags:  []models.ComplianceFlag{},
	}
}

func compliantWithConditionsResult() *models.ComplianceResult {
	return &models.ComplianceResult{
		Status: models.ComplianceStatusCompliantWithConditions,
		Flags: []models.ComplianceFlag{
			{Rule: "THRESHOLD_EXCEEDED", Severity: models.FlagSeverityRequiresReview, Message: "Amount exceeds threshold"},
		},
	}
}

func TestDetermineRoute_PurchaseCard_AtThreshold(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("2500.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelPurchaseCard {
		t.Errorf("expected PURCHASE_CARD, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityLow {
		t.Errorf("expected LOW priority, got %s", decision.Priority)
	}
	if decision.Status != models.RoutingStatusRouted {
		t.Errorf("expected ROUTED status, got %s", decision.Status)
	}
}

func TestDetermineRoute_PurchaseCard_BelowThreshold(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("100.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelPurchaseCard {
		t.Errorf("expected PURCHASE_CARD, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityLow {
		t.Errorf("expected LOW priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_Supervisor_JustAbovePurchaseCard(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("2500.01")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_Supervisor_AtThreshold(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("25000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_ContractingOfficer_JustAboveSupervisor(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("25000.01")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelContractingOfficer {
		t.Errorf("expected CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_ContractingOfficer_AtThreshold(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("250000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelContractingOfficer {
		t.Errorf("expected CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_SeniorContractingOfficer_JustAboveCO(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("250000.01")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSeniorContractingOfficer {
		t.Errorf("expected SENIOR_CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityHigh {
		t.Errorf("expected HIGH priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_SeniorContractingOfficer_AtThreshold(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("1000000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSeniorContractingOfficer {
		t.Errorf("expected SENIOR_CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityHigh {
		t.Errorf("expected HIGH priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_AgencyHead_JustAboveSeniorCO(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("1000000.01")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelAgencyHead {
		t.Errorf("expected AGENCY_HEAD, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_AgencyHead_LargeAmount(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("5000000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelAgencyHead {
		t.Errorf("expected AGENCY_HEAD, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority, got %s", decision.Priority)
	}
}

func TestDetermineRoute_CurrencyFormatWithDollarSign(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("$12,345.67")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR, got %s", decision.ApprovalLevel)
	}
}

func TestDetermineRoute_UsesNormalizedValue(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      "raw garbage",
				Normalized: "500.00",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelPurchaseCard {
		t.Errorf("expected PURCHASE_CARD, got %s", decision.ApprovalLevel)
	}
}

func TestDetermineRoute_FallsBackToValueWhenNormalizedEmpty(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      "30000.00",
				Normalized: "",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelContractingOfficer {
		t.Errorf("expected CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
}

func TestDetermineRoute_UsesTotalAmountField(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypePurchaseOrder,
		Fields: map[string]models.ExtractedField{
			"totalAmount": {
				Value:      "50000.00",
				Normalized: "50000.00",
				Confidence: 0.90,
			},
		},
		OverallConfidence: 0.90,
	}
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelContractingOfficer {
		t.Errorf("expected CONTRACTING_OFFICER, got %s", decision.ApprovalLevel)
	}
}

func TestDetermineRoute_ErrorNoAmountField(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      "Test Vendor",
				Normalized: "Test Vendor",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
	_, err := handler.DetermineRoute(extraction, compliantResult())
	if err == nil {
		t.Fatal("expected error for missing amount field, got nil")
	}
}

func TestDetermineRoute_ErrorInvalidAmountFormat(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("not-a-number")
	_, err := handler.DetermineRoute(extraction, compliantResult())
	if err == nil {
		t.Fatal("expected error for invalid amount format, got nil")
	}
}

func TestDetermineRoute_RationaleIncludesAmount(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("5000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Rationale == "" {
		t.Error("expected non-empty rationale")
	}
}

func TestDetermineRoute_ZeroAmount(t *testing.T) {
	handler := NewRoutingHandler()
	extraction := extractionWithAmount("0.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Zero is at or below $2,500
	if decision.ApprovalLevel != models.ApprovalLevelPurchaseCard {
		t.Errorf("expected PURCHASE_CARD for zero amount, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityLow {
		t.Errorf("expected LOW priority for zero amount, got %s", decision.Priority)
	}
}


// --- Tests for Task 6.2: Compliance condition elevation and urgency override ---

func TestDetermineRoute_CompliantWithConditions_ElevatesLevelAndPriority(t *testing.T) {
	handler := NewRoutingHandler()
	// Base: $5,000 → SUPERVISOR / NORMAL
	// With conditions: → CONTRACTING_OFFICER / HIGH
	extraction := extractionWithAmount("5000.00")
	decision, err := handler.DetermineRoute(extraction, compliantWithConditionsResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelContractingOfficer {
		t.Errorf("expected CONTRACTING_OFFICER after elevation, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityHigh {
		t.Errorf("expected HIGH priority after elevation, got %s", decision.Priority)
	}
}

func TestDetermineRoute_CompliantWithConditions_ElevatesFromPurchaseCard(t *testing.T) {
	handler := NewRoutingHandler()
	// Base: $1,000 → PURCHASE_CARD / LOW
	// With conditions: → SUPERVISOR / NORMAL
	extraction := extractionWithAmount("1000.00")
	decision, err := handler.DetermineRoute(extraction, compliantWithConditionsResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR after elevation, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority after elevation, got %s", decision.Priority)
	}
}

func TestDetermineRoute_CompliantWithConditions_MaxLevelStaysAtMax(t *testing.T) {
	handler := NewRoutingHandler()
	// Base: $2,000,000 → AGENCY_HEAD / URGENT (already at max)
	// With conditions: → AGENCY_HEAD / URGENT (stays at max)
	extraction := extractionWithAmount("2000000.00")
	decision, err := handler.DetermineRoute(extraction, compliantWithConditionsResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelAgencyHead {
		t.Errorf("expected AGENCY_HEAD to stay at max, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT to stay at max, got %s", decision.Priority)
	}
}

func TestDetermineRoute_CompliantWithConditions_HighPriorityElevatesToUrgent(t *testing.T) {
	handler := NewRoutingHandler()
	// Base: $500,000 → SENIOR_CONTRACTING_OFFICER / HIGH
	// With conditions: → AGENCY_HEAD / URGENT
	extraction := extractionWithAmount("500000.00")
	decision, err := handler.DetermineRoute(extraction, compliantWithConditionsResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelAgencyHead {
		t.Errorf("expected AGENCY_HEAD after elevation, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT after elevation, got %s", decision.Priority)
	}
}

func TestDetermineRoute_DueDateWithin3Days_SetsUrgent(t *testing.T) {
	handler := NewRoutingHandler()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
	// Due date 2 days from now
	dueDate := now.AddDate(0, 0, 2).Format("2006-01-02")
	// Base: $5,000 → SUPERVISOR / NORMAL
	extraction := extractionWithAmountAndDate("5000.00", dueDate)
	decision, err := handler.DetermineRouteWithTime(extraction, compliantResult(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority for due date within 3 days, got %s", decision.Priority)
	}
	// Approval level should remain SUPERVISOR (no compliance elevation)
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR, got %s", decision.ApprovalLevel)
	}
}

func TestDetermineRoute_DueDateExactly3Days_SetsUrgent(t *testing.T) {
	handler := NewRoutingHandler()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
	// Due date exactly 3 days from now
	dueDate := now.AddDate(0, 0, 3).Format("2006-01-02")
	extraction := extractionWithAmountAndDate("5000.00", dueDate)
	decision, err := handler.DetermineRouteWithTime(extraction, compliantResult(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority for due date exactly 3 days away, got %s", decision.Priority)
	}
}

func TestDetermineRoute_DueDateMoreThan3Days_DoesNotOverride(t *testing.T) {
	handler := NewRoutingHandler()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
	// Due date 10 days from now
	dueDate := now.AddDate(0, 0, 10).Format("2006-01-02")
	// Base: $5,000 → SUPERVISOR / NORMAL
	extraction := extractionWithAmountAndDate("5000.00", dueDate)
	decision, err := handler.DetermineRouteWithTime(extraction, compliantResult(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority for due date > 3 days, got %s", decision.Priority)
	}
}

func TestDetermineRoute_DueDateInPast_SetsUrgent(t *testing.T) {
	handler := NewRoutingHandler()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
	// Due date was yesterday (past due is within 3 days, actually negative)
	dueDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	extraction := extractionWithAmountAndDate("5000.00", dueDate)
	decision, err := handler.DetermineRouteWithTime(extraction, compliantResult(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority for past-due date, got %s", decision.Priority)
	}
}

func TestDetermineRoute_BothConditionsAndUrgentDueDate(t *testing.T) {
	handler := NewRoutingHandler()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
	// Due date 1 day from now
	dueDate := now.AddDate(0, 0, 1).Format("2006-01-02")
	// Base: $1,000 → PURCHASE_CARD / LOW
	// With conditions: → SUPERVISOR / NORMAL
	// With urgency: → SUPERVISOR / URGENT
	extraction := extractionWithAmountAndDate("1000.00", dueDate)
	decision, err := handler.DetermineRouteWithTime(extraction, compliantWithConditionsResult(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.ApprovalLevel != models.ApprovalLevelSupervisor {
		t.Errorf("expected SUPERVISOR after compliance elevation, got %s", decision.ApprovalLevel)
	}
	if decision.Priority != models.PriorityUrgent {
		t.Errorf("expected URGENT priority from due date override, got %s", decision.Priority)
	}
}

func TestDetermineRoute_NoDateField_NoPriorityOverride(t *testing.T) {
	handler := NewRoutingHandler()
	// No date field in extraction
	extraction := extractionWithAmount("5000.00")
	decision, err := handler.DetermineRoute(extraction, compliantResult())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should keep base priority NORMAL since no date to trigger urgency
	if decision.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority with no date field, got %s", decision.Priority)
	}
}
