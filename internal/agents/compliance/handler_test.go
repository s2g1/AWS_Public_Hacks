package compliance

import (
	"testing"

	"federal-payment-processing/internal/models"
)

// testSanctionsList returns a sanctions list for testing.
func testSanctionsList() SanctionsList {
	return NewInMemorySanctionsList([]string{
		"Osama Bin Laden",
		"Al-Qaeda Foundation",
		"Viktor Bout",
		"Wagner Group PMC",
	})
}

// testDebarmentList returns a debarment list for testing.
func testDebarmentList() DebarmentList {
	return NewInMemoryDebarmentList([]string{
		"Blackwater Security LLC",
		"Continental Defense Systems",
		"Pacific Rim Contractors Inc",
	})
}

// makeExtraction creates a minimal ExtractionResult with a payee name.
func makeExtraction(payeeName string) *models.ExtractionResult {
	return &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      payeeName,
				Normalized: payeeName,
				Confidence: 0.95,
			},
			"amount": {
				Value:      "5000.00",
				Normalized: "5000.00",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}
}

func TestCheckCompliance_ExactMatch_Blocking(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("Osama Bin Laden")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT status, got %s", result.Status)
	}
	if len(result.Flags) == 0 {
		t.Fatal("expected at least one compliance flag")
	}
	if result.Flags[0].Rule != "OFAC_SANCTIONS" {
		t.Errorf("expected OFAC_SANCTIONS rule, got %s", result.Flags[0].Rule)
	}
	if result.Flags[0].Severity != models.FlagSeverityBlocking {
		t.Errorf("expected BLOCKING severity, got %s", result.Flags[0].Severity)
	}
}

func TestCheckCompliance_FuzzyMatch_Blocking(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())

	// "Osama Bin Ladin" is a close fuzzy match to "Osama Bin Laden" (common misspelling).
	extraction := makeExtraction("Osama Bin Ladin")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for fuzzy match, got %s", result.Status)
	}
	if len(result.Flags) == 0 {
		t.Fatal("expected at least one compliance flag for fuzzy match")
	}
	if result.Flags[0].Rule != "OFAC_SANCTIONS" {
		t.Errorf("expected OFAC_SANCTIONS rule, got %s", result.Flags[0].Rule)
	}
	if result.Flags[0].Severity != models.FlagSeverityBlocking {
		t.Errorf("expected BLOCKING severity, got %s", result.Flags[0].Severity)
	}
}

func TestCheckCompliance_CaseInsensitive_Match(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("OSAMA BIN LADEN")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for case-insensitive match, got %s", result.Status)
	}
}

func TestCheckCompliance_NoMatch_Compliant(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("Acme Corporation")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT for non-match, got %s", result.Status)
	}
	if len(result.Flags) != 0 {
		t.Errorf("expected no flags for compliant result, got %d", len(result.Flags))
	}
}

func TestCheckCompliance_NoPayeeField_Compliant(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"amount": {
				Value:      "1000.00",
				Normalized: "1000.00",
				Confidence: 0.95,
			},
		},
		OverallConfidence: 0.95,
	}

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT when no payee field, got %s", result.Status)
	}
}

func TestCheckCompliance_CompletlyDifferentName_Compliant(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("John Smith Consulting LLC")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT for completely different name, got %s", result.Status)
	}
}

func TestCheckCompliance_UsesNormalizedField(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      "RAW Viktor Bout Data",
				Normalized: "Viktor Bout",
				Confidence: 0.90,
			},
		},
		OverallConfidence: 0.90,
	}

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT when normalized name matches, got %s", result.Status)
	}
}

func TestCheckCompliance_EmptyNormalized_FallsBackToValue(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := &models.ExtractionResult{
		DocumentType: models.DocumentTypeInvoice,
		Fields: map[string]models.ExtractedField{
			"payee": {
				Value:      "Viktor Bout",
				Normalized: "",
				Confidence: 0.90,
			},
		},
		OverallConfidence: 0.90,
	}

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT when falling back to Value field, got %s", result.Status)
	}
}

// --- Debarment Screening Tests ---

func TestCheckCompliance_DebarmentExactMatch_Blocking(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("Blackwater Security LLC")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for debarment match, got %s", result.Status)
	}
	if len(result.Flags) == 0 {
		t.Fatal("expected at least one compliance flag for debarment match")
	}
	if result.Flags[0].Rule != "DEBARMENT" {
		t.Errorf("expected DEBARMENT rule, got %s", result.Flags[0].Rule)
	}
	if result.Flags[0].Severity != models.FlagSeverityBlocking {
		t.Errorf("expected BLOCKING severity, got %s", result.Flags[0].Severity)
	}
}

func TestCheckCompliance_DebarmentFuzzyMatch_Blocking(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	// Close fuzzy match to "Continental Defense Systems"
	extraction := makeExtraction("Continental Defence Systems")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for debarment fuzzy match, got %s", result.Status)
	}
	if len(result.Flags) == 0 {
		t.Fatal("expected at least one compliance flag for debarment fuzzy match")
	}
	if result.Flags[0].Rule != "DEBARMENT" {
		t.Errorf("expected DEBARMENT rule, got %s", result.Flags[0].Rule)
	}
	if result.Flags[0].Severity != models.FlagSeverityBlocking {
		t.Errorf("expected BLOCKING severity, got %s", result.Flags[0].Severity)
	}
}

func TestCheckCompliance_DebarmentNoMatch_Passes(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), testDebarmentList())
	extraction := makeExtraction("Honest Federal Services Inc")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT when not on debarment list, got %s", result.Status)
	}
	if len(result.Flags) != 0 {
		t.Errorf("expected no flags for non-debarred entity, got %d", len(result.Flags))
	}
}

func TestCheckCompliance_OFACTakesPrecedenceOverDebarment(t *testing.T) {
	// Create a handler where the payee matches both lists
	sanctions := NewInMemorySanctionsList([]string{"Shared Bad Entity"})
	debarment := NewInMemoryDebarmentList([]string{"Shared Bad Entity"})
	handler := NewComplianceHandler(sanctions, debarment)
	extraction := makeExtraction("Shared Bad Entity")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT, got %s", result.Status)
	}
	// OFAC should be checked first and return immediately
	if len(result.Flags) != 1 {
		t.Fatalf("expected exactly 1 flag (OFAC takes precedence), got %d", len(result.Flags))
	}
	if result.Flags[0].Rule != "OFAC_SANCTIONS" {
		t.Errorf("expected OFAC_SANCTIONS rule (checked first), got %s", result.Flags[0].Rule)
	}
}

func TestCheckCompliance_NilDebarmentList_Passes(t *testing.T) {
	handler := NewComplianceHandler(testSanctionsList(), nil)
	extraction := makeExtraction("Acme Corporation")

	result := handler.CheckCompliance(extraction)

	if result.Status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT with nil debarment list, got %s", result.Status)
	}
}

// --- Jaro-Winkler Tests ---

func TestJaroWinklerSimilarity_ExactMatch(t *testing.T) {
	sim := JaroWinklerSimilarity("hello", "hello")
	if sim != 1.0 {
		t.Errorf("expected 1.0 for exact match, got %f", sim)
	}
}

func TestJaroWinklerSimilarity_EmptyStrings(t *testing.T) {
	if sim := JaroWinklerSimilarity("", "hello"); sim != 0.0 {
		t.Errorf("expected 0.0 for empty s1, got %f", sim)
	}
	if sim := JaroWinklerSimilarity("hello", ""); sim != 0.0 {
		t.Errorf("expected 0.0 for empty s2, got %f", sim)
	}
}

func TestJaroWinklerSimilarity_HighSimilarity(t *testing.T) {
	// "laden" vs "ladin" should have high similarity.
	sim := JaroWinklerSimilarity("osama bin laden", "osama bin ladin")
	if sim < OFACMatchThreshold {
		t.Errorf("expected similarity >= %f for close match, got %f", OFACMatchThreshold, sim)
	}
}

func TestJaroWinklerSimilarity_LowSimilarity(t *testing.T) {
	sim := JaroWinklerSimilarity("acme corporation", "osama bin laden")
	if sim >= OFACMatchThreshold {
		t.Errorf("expected similarity < %f for dissimilar strings, got %f", OFACMatchThreshold, sim)
	}
}
