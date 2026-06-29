package compliance

import (
	"fmt"
	"testing"

	"federal-payment-processing/internal/models"
)

// mockSpendingStore implements SpendingStore for testing.
type mockSpendingStore struct {
	spendByPayee map[string]float64
	err          error
}

func newMockSpendingStore(spend map[string]float64) *mockSpendingStore {
	return &mockSpendingStore{spendByPayee: spend}
}

func (m *mockSpendingStore) GetCumulativeSpend(payee string, fiscalYear int) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.spendByPayee[payee], nil
}

// makeExtractionWithAmount creates a minimal ExtractionResult with payee and amount.
func makeExtractionWithAmount(payeeName string, amount string) *models.ExtractionResult {
	return &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      payeeName,
				Normalized: payeeName,
				Confidence: 0.95,
			},
			"amount": {
				Value:      amount,
				Normalized: amount,
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
}

func TestCheckSpendingThresholds_SingleTransactionExceeded(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 10000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 15000.00, thresholds, store)

	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Rule != "THRESHOLD_EXCEEDED" {
		t.Errorf("expected THRESHOLD_EXCEEDED rule, got %s", flags[0].Rule)
	}
	if flags[0].Severity != models.FlagSeverityRequiresReview {
		t.Errorf("expected REQUIRES_REVIEW severity, got %s", flags[0].Severity)
	}
}

func TestCheckSpendingThresholds_SingleTransactionUnder(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 10000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 5000.00, thresholds, store)

	if len(flags) != 0 {
		t.Errorf("expected no flags for amount under threshold, got %d", len(flags))
	}
}

func TestCheckSpendingThresholds_AnnualLimitExceeded(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{
		"Acme Corp": 480000.00,
	})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 100000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 25000.00, thresholds, store)

	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Rule != "ANNUAL_LIMIT" {
		t.Errorf("expected ANNUAL_LIMIT rule, got %s", flags[0].Rule)
	}
	if flags[0].Severity != models.FlagSeverityRequiresReview {
		t.Errorf("expected REQUIRES_REVIEW severity, got %s", flags[0].Severity)
	}
}

func TestCheckSpendingThresholds_AnnualLimitNotExceeded(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{
		"Acme Corp": 100000.00,
	})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 100000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 25000.00, thresholds, store)

	if len(flags) != 0 {
		t.Errorf("expected no flags when under annual limit, got %d", len(flags))
	}
}

func TestCheckSpendingThresholds_BothThresholdsExceeded(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{
		"Acme Corp": 480000.00,
	})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 10000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 25000.00, thresholds, store)

	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}

	rules := map[string]bool{}
	for _, f := range flags {
		rules[f.Rule] = true
	}
	if !rules["THRESHOLD_EXCEEDED"] {
		t.Error("expected THRESHOLD_EXCEEDED flag")
	}
	if !rules["ANNUAL_LIMIT"] {
		t.Error("expected ANNUAL_LIMIT flag")
	}
}

func TestCheckSpendingThresholds_StoreError_NoAnnualFlag(t *testing.T) {
	store := &mockSpendingStore{
		spendByPayee: nil,
		err:          fmt.Errorf("database error"),
	}
	thresholds := SpendingThresholds{
		SingleTransactionMax: 100000.00,
		AnnualMax:            500000.00,
	}

	// Amount exceeds single transaction max but store errors on cumulative check
	flags := checkSpendingThresholds("Acme Corp", 150000.00, thresholds, store)

	// Should still flag single transaction exceeded even if store errors
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag (single transaction only), got %d", len(flags))
	}
	if flags[0].Rule != "THRESHOLD_EXCEEDED" {
		t.Errorf("expected THRESHOLD_EXCEEDED rule, got %s", flags[0].Rule)
	}
}

func TestCheckSpendingThresholds_NilStore_NoAnnualCheck(t *testing.T) {
	thresholds := SpendingThresholds{
		SingleTransactionMax: 10000.00,
		AnnualMax:            500000.00,
	}

	flags := checkSpendingThresholds("Acme Corp", 5000.00, thresholds, nil)

	if len(flags) != 0 {
		t.Errorf("expected no flags with nil store and under single max, got %d", len(flags))
	}
}

// --- Integration tests: spending thresholds through CheckCompliance ---

func TestCheckCompliance_SpendingThresholdExceeded_CompliantWithConditions(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 5000.00,
		AnnualMax:            500000.00,
	}
	handler := NewComplianceHandlerWithSpending(testSanctionsList(), testDebarmentList(), store, thresholds)
	extraction := makeExtractionWithAmount("Acme Corporation", "10000.00")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliantWithConditions {
		t.Errorf("expected COMPLIANT_WITH_CONDITIONS, got %s", result.Status)
	}
	if len(result.Flags) == 0 {
		t.Fatal("expected at least one flag")
	}
	found := false
	for _, f := range result.Flags {
		if f.Rule == "THRESHOLD_EXCEEDED" {
			found = true
			if f.Severity != models.FlagSeverityRequiresReview {
				t.Errorf("expected REQUIRES_REVIEW severity, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected THRESHOLD_EXCEEDED flag")
	}
}

func TestCheckCompliance_AnnualLimitExceeded_CompliantWithConditions(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{
		"Acme Corporation": 490000.00,
	})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 100000.00,
		AnnualMax:            500000.00,
	}
	handler := NewComplianceHandlerWithSpending(testSanctionsList(), testDebarmentList(), store, thresholds)
	extraction := makeExtractionWithAmount("Acme Corporation", "15000.00")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliantWithConditions {
		t.Errorf("expected COMPLIANT_WITH_CONDITIONS, got %s", result.Status)
	}
	found := false
	for _, f := range result.Flags {
		if f.Rule == "ANNUAL_LIMIT" {
			found = true
			if f.Severity != models.FlagSeverityRequiresReview {
				t.Errorf("expected REQUIRES_REVIEW severity, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected ANNUAL_LIMIT flag")
	}
}

func TestCheckCompliance_SpendingUnderThresholds_Compliant(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{
		"Acme Corporation": 10000.00,
	})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 100000.00,
		AnnualMax:            500000.00,
	}
	handler := NewComplianceHandlerWithSpending(testSanctionsList(), testDebarmentList(), store, thresholds)
	extraction := makeExtractionWithAmount("Acme Corporation", "5000.00")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT when under thresholds, got %s", result.Status)
	}
	if len(result.Flags) != 0 {
		t.Errorf("expected no flags when under thresholds, got %d", len(result.Flags))
	}
}

func TestCheckCompliance_OFACBlocksBeforeSpendingCheck(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 1000.00,
		AnnualMax:            5000.00,
	}
	handler := NewComplianceHandlerWithSpending(testSanctionsList(), testDebarmentList(), store, thresholds)
	// Payee matches OFAC list and also exceeds spending thresholds
	extraction := makeExtractionWithAmount("Viktor Bout", "50000.00")

	result := handler.CheckCompliance(extraction)

	// OFAC should block immediately, no spending flags
	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for OFAC match, got %s", result.Status)
	}
	if len(result.Flags) != 1 {
		t.Fatalf("expected exactly 1 flag (OFAC), got %d", len(result.Flags))
	}
	if result.Flags[0].Rule != "OFAC_SANCTIONS" {
		t.Errorf("expected OFAC_SANCTIONS rule, got %s", result.Flags[0].Rule)
	}
}

func TestCheckCompliance_DebarmentBlocksBeforeSpendingCheck(t *testing.T) {
	store := newMockSpendingStore(map[string]float64{})
	thresholds := SpendingThresholds{
		SingleTransactionMax: 1000.00,
		AnnualMax:            5000.00,
	}
	handler := NewComplianceHandlerWithSpending(testSanctionsList(), testDebarmentList(), store, thresholds)
	// Payee matches debarment list and also exceeds spending thresholds
	extraction := makeExtractionWithAmount("Blackwater Security LLC", "50000.00")

	result := handler.CheckCompliance(extraction)

	// Debarment should block immediately, no spending flags
	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for debarment match, got %s", result.Status)
	}
	if len(result.Flags) != 1 {
		t.Fatalf("expected exactly 1 flag (DEBARMENT), got %d", len(result.Flags))
	}
	if result.Flags[0].Rule != "DEBARMENT" {
		t.Errorf("expected DEBARMENT rule, got %s", result.Flags[0].Rule)
	}
}

func TestCheckCompliance_NoSpendingStoreConfigured_StillCompliant(t *testing.T) {
	// Using the basic NewComplianceHandler (no spending store)
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtractionWithAmount("Acme Corporation", "999999.00")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT with no spending store configured, got %s", result.Status)
	}
}
