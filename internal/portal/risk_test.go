package portal

import (
	"math"
	"testing"
	"time"
)

// --- DetermineRiskLevel Tests ---

func TestDetermineRiskLevel_RED_Overrun(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:     "CLIN-001",
		CLINStatus: CLINStatusActive,
		Ceiling:    100000.0,
		Obligated:  100000.0,
		Expended:   120000.0,
		EAC:        120000.0,
	}
	variance := &VarianceAnalysis{
		Overrun:                 20000.0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // in the past (already done)
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelRed {
		t.Errorf("expected RED for overrun case, got %s", result)
	}
}

func TestDetermineRiskLevel_RED_UnderRunOver40Percent(t *testing.T) {
	// Under-run > 40%: obligated=100000, EAC=50000 → under-run pct = 50%
	clin := &ContractLineItem{
		CLINID:     "CLIN-002",
		CLINStatus: CLINStatusActive,
		Ceiling:    200000.0,
		Obligated:  100000.0,
		Expended:   40000.0,
		EAC:        50000.0,
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelRed {
		t.Errorf("expected RED for under-run > 40%%, got %s", result)
	}
}

func TestDetermineRiskLevel_RED_ProjectedCompletionPastPoP(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:     "CLIN-003",
		CLINStatus: CLINStatusActive,
		Ceiling:    200000.0,
		Obligated:  200000.0,
		Expended:   100000.0,
		EAC:        190000.0, // under-run pct = 5% (not enough for yellow/red by under-run)
	}
	popEnd := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), // past PoP
	}

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelRed {
		t.Errorf("expected RED for projected completion past PoP end, got %s", result)
	}
}

func TestDetermineRiskLevel_YELLOW_ExpenditureRatioAbove90Active(t *testing.T) {
	// Expenditure ratio > 90%: expended/ceiling = 95000/100000 = 0.95
	clin := &ContractLineItem{
		CLINID:     "CLIN-004",
		CLINStatus: CLINStatusActive,
		Ceiling:    100000.0,
		Obligated:  100000.0,
		Expended:   95000.0,
		EAC:        99000.0, // under-run pct = 1% (not enough for yellow)
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // in the past
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelYellow {
		t.Errorf("expected YELLOW for expenditure ratio > 90%% on ACTIVE CLIN, got %s", result)
	}
}

func TestDetermineRiskLevel_YELLOW_ExpenditureRatioAbove90NotActive(t *testing.T) {
	// Same expenditure ratio but COMPLETED status — should NOT trigger YELLOW
	clin := &ContractLineItem{
		CLINID:     "CLIN-004b",
		CLINStatus: CLINStatusCompleted,
		Ceiling:    100000.0,
		Obligated:  100000.0,
		Expended:   95000.0,
		EAC:        99000.0,
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelGreen {
		t.Errorf("expected GREEN for expenditure ratio > 90%% on non-ACTIVE CLIN, got %s", result)
	}
}

func TestDetermineRiskLevel_YELLOW_UnderRun20to40(t *testing.T) {
	// Under-run 30%: obligated=100000, EAC=70000 → under-run pct = 30%
	clin := &ContractLineItem{
		CLINID:     "CLIN-005",
		CLINStatus: CLINStatusActive,
		Ceiling:    200000.0,
		Obligated:  100000.0,
		Expended:   50000.0, // expenditure ratio = 50000/200000 = 0.25 (not > 0.90)
		EAC:        70000.0,
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelYellow {
		t.Errorf("expected YELLOW for under-run 20-40%%, got %s", result)
	}
}

func TestDetermineRiskLevel_GREEN_HealthyCLIN(t *testing.T) {
	// Normal healthy CLIN: no overrun, low under-run, expenditure ratio well below 90%
	clin := &ContractLineItem{
		CLINID:     "CLIN-006",
		CLINStatus: CLINStatusActive,
		Ceiling:    500000.0,
		Obligated:  400000.0,
		Expended:   200000.0, // expenditure ratio = 200000/500000 = 0.40
		EAC:        380000.0, // under-run pct = (400000-380000)/400000 * 100 = 5%
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelGreen {
		t.Errorf("expected GREEN for healthy CLIN, got %s", result)
	}
}

// --- Edge Cases ---

func TestDetermineRiskLevel_ZeroObligated(t *testing.T) {
	// Zero obligated: under-run percentage should be 0, no RED/YELLOW from under-run
	clin := &ContractLineItem{
		CLINID:     "CLIN-EDGE-1",
		CLINStatus: CLINStatusActive,
		Ceiling:    100000.0,
		Obligated:  0,
		Expended:   50000.0, // expenditure ratio = 0.50
		EAC:        0,
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelGreen {
		t.Errorf("expected GREEN for zero obligated (no under-run risk), got %s", result)
	}
}

func TestDetermineRiskLevel_ZeroCeiling(t *testing.T) {
	// Zero ceiling: expenditure ratio should be 0, no YELLOW from expenditure
	clin := &ContractLineItem{
		CLINID:     "CLIN-EDGE-2",
		CLINStatus: CLINStatusActive,
		Ceiling:    0,
		Obligated:  100000.0,
		Expended:   0,
		EAC:        90000.0, // under-run pct = 10% (not enough for yellow)
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelGreen {
		t.Errorf("expected GREEN for zero ceiling edge case, got %s", result)
	}
}

func TestDetermineRiskLevel_ExactlyAt40PercentUnderRun(t *testing.T) {
	// Exactly at 40% boundary: obligated=100000, EAC=60000 → under-run = 40%
	// 40% is the boundary between YELLOW (20-40%) and RED (>40%)
	// At exactly 40%, it should be YELLOW (<=40 is yellow, >40 is red)
	clin := &ContractLineItem{
		CLINID:     "CLIN-EDGE-3",
		CLINStatus: CLINStatusActive,
		Ceiling:    200000.0,
		Obligated:  100000.0,
		Expended:   50000.0, // expenditure ratio = 0.25
		EAC:        60000.0, // under-run pct = (100000-60000)/100000 * 100 = 40%
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelYellow {
		t.Errorf("expected YELLOW at exactly 40%% under-run boundary, got %s", result)
	}
}

func TestDetermineRiskLevel_ExactlyAt20PercentUnderRun(t *testing.T) {
	// Exactly at 20% boundary: obligated=100000, EAC=80000 → under-run = 20%
	// 20% should trigger YELLOW (20-40% range)
	clin := &ContractLineItem{
		CLINID:     "CLIN-EDGE-4",
		CLINStatus: CLINStatusActive,
		Ceiling:    200000.0,
		Obligated:  100000.0,
		Expended:   50000.0,
		EAC:        80000.0, // under-run pct = (100000-80000)/100000 * 100 = 20%
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelYellow {
		t.Errorf("expected YELLOW at exactly 20%% under-run boundary, got %s", result)
	}
}

func TestDetermineRiskLevel_ExactlyAt90PercentExpenditureActive(t *testing.T) {
	// Expenditure ratio exactly at 90%: expended/ceiling = 90000/100000 = 0.90
	// Condition is > 90%, so exactly 90% should NOT trigger YELLOW
	clin := &ContractLineItem{
		CLINID:     "CLIN-EDGE-5",
		CLINStatus: CLINStatusActive,
		Ceiling:    100000.0,
		Obligated:  100000.0,
		Expended:   90000.0, // ratio = exactly 0.90
		EAC:        99000.0, // under-run pct = 1%
	}
	variance := &VarianceAnalysis{
		Overrun:                 0,
		ProjectedCompletionDate: time.Now().AddDate(0, -1, 0),
	}
	popEnd := time.Now().AddDate(1, 0, 0)

	result := DetermineRiskLevel(clin, variance, popEnd)
	if result != RiskLevelGreen {
		t.Errorf("expected GREEN at exactly 90%% expenditure ratio (not >90%%), got %s", result)
	}
}

// --- CalculateUnderRunPercentage Tests ---

func TestCalculateUnderRunPercentage_Normal(t *testing.T) {
	clin := &ContractLineItem{
		Obligated: 200000.0,
		EAC:       140000.0,
	}
	// (200000 - 140000) / 200000 * 100 = 30%
	result := CalculateUnderRunPercentage(clin)
	expected := 30.0
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected under-run percentage %.2f%%, got %.2f%%", expected, result)
	}
}

func TestCalculateUnderRunPercentage_ZeroObligated(t *testing.T) {
	clin := &ContractLineItem{
		Obligated: 0,
		EAC:       50000.0,
	}
	result := CalculateUnderRunPercentage(clin)
	if result != 0 {
		t.Errorf("expected 0 for zero obligated, got %.2f", result)
	}
}

func TestCalculateUnderRunPercentage_EACExceedsObligated(t *testing.T) {
	// EAC > Obligated means no under-run
	clin := &ContractLineItem{
		Obligated: 100000.0,
		EAC:       120000.0,
	}
	result := CalculateUnderRunPercentage(clin)
	if result != 0 {
		t.Errorf("expected 0 when EAC exceeds obligated, got %.2f", result)
	}
}

func TestCalculateUnderRunPercentage_FullUnderRun(t *testing.T) {
	// 100% under-run: EAC = 0
	clin := &ContractLineItem{
		Obligated: 100000.0,
		EAC:       0,
	}
	result := CalculateUnderRunPercentage(clin)
	expected := 100.0
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected 100%% under-run, got %.2f%%", result)
	}
}

// --- CalculateExpenditureRatio Tests ---

func TestCalculateExpenditureRatio_Normal(t *testing.T) {
	clin := &ContractLineItem{
		Expended: 75000.0,
		Ceiling:  100000.0,
	}
	result := CalculateExpenditureRatio(clin)
	expected := 0.75
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected expenditure ratio %.4f, got %.4f", expected, result)
	}
}

func TestCalculateExpenditureRatio_ZeroCeiling(t *testing.T) {
	clin := &ContractLineItem{
		Expended: 50000.0,
		Ceiling:  0,
	}
	result := CalculateExpenditureRatio(clin)
	if result != 0 {
		t.Errorf("expected 0 for zero ceiling, got %.4f", result)
	}
}

func TestCalculateExpenditureRatio_ZeroExpended(t *testing.T) {
	clin := &ContractLineItem{
		Expended: 0,
		Ceiling:  100000.0,
	}
	result := CalculateExpenditureRatio(clin)
	if result != 0 {
		t.Errorf("expected 0 for zero expended, got %.4f", result)
	}
}

func TestCalculateExpenditureRatio_OverCeiling(t *testing.T) {
	clin := &ContractLineItem{
		Expended: 150000.0,
		Ceiling:  100000.0,
	}
	result := CalculateExpenditureRatio(clin)
	expected := 1.5
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected expenditure ratio %.4f, got %.4f", expected, result)
	}
}
