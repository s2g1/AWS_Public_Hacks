package portal

import (
	"math"
	"testing"
	"time"
)

func TestCalculateOverrun_PositiveOverrun(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:   "CLIN-001",
		Expended: 150000.0,
		Ceiling:  100000.0,
	}
	result := CalculateOverrun(clin)
	expected := 50000.0
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected overrun %.2f, got %.2f", expected, result)
	}
}

func TestCalculateOverrun_ZeroNoOverrun(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:   "CLIN-002",
		Expended: 80000.0,
		Ceiling:  100000.0,
	}
	result := CalculateOverrun(clin)
	if result != 0 {
		t.Errorf("expected no overrun (0), got %.2f", result)
	}
}

func TestCalculateOverrun_LargeExpended(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:   "CLIN-003",
		Expended: 5000000.0,
		Ceiling:  1000000.0,
	}
	result := CalculateOverrun(clin)
	expected := 4000000.0
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected overrun %.2f, got %.2f", expected, result)
	}
}

func TestCalculateUnderRun_PositiveUnderRun(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:    "CLIN-001",
		Obligated: 200000.0,
		EAC:       150000.0,
	}
	result := CalculateUnderRun(clin)
	expected := 50000.0
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected under-run %.2f, got %.2f", expected, result)
	}
}

func TestCalculateUnderRun_ZeroNoUnderRun(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:    "CLIN-002",
		Obligated: 100000.0,
		EAC:       120000.0,
	}
	result := CalculateUnderRun(clin)
	if result != 0 {
		t.Errorf("expected no under-run (0), got %.2f", result)
	}
}

func TestCalculateBurnRate_ExactlyThreeMonths(t *testing.T) {
	expenditures := []float64{10000.0, 12000.0, 14000.0}
	result := CalculateBurnRate(expenditures)
	expected := 12000.0 // (10000 + 12000 + 14000) / 3
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected burn rate %.2f, got %.2f", expected, result)
	}
}

func TestCalculateBurnRate_FewerThanThreeMonths(t *testing.T) {
	expenditures := []float64{8000.0, 12000.0}
	result := CalculateBurnRate(expenditures)
	expected := 10000.0 // (8000 + 12000) / 2
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected burn rate %.2f, got %.2f", expected, result)
	}
}

func TestCalculateBurnRate_MoreThanThreeMonths(t *testing.T) {
	// Should only use the last 3 months
	expenditures := []float64{5000.0, 6000.0, 9000.0, 12000.0, 15000.0}
	result := CalculateBurnRate(expenditures)
	expected := 12000.0 // (9000 + 12000 + 15000) / 3
	if math.Abs(result-expected) > epsilon {
		t.Errorf("expected burn rate %.2f, got %.2f", expected, result)
	}
}

func TestCalculateBurnRate_ZeroExpenditures(t *testing.T) {
	expenditures := []float64{0.0, 0.0, 0.0}
	result := CalculateBurnRate(expenditures)
	if result != 0 {
		t.Errorf("expected burn rate 0, got %.2f", result)
	}
}

func TestCalculateBurnRate_Empty(t *testing.T) {
	result := CalculateBurnRate([]float64{})
	if result != 0 {
		t.Errorf("expected burn rate 0 for empty input, got %.2f", result)
	}
}

func TestProjectCompletionDate_NormalProjection(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	remainingWork := 60000.0
	burnRate := 10000.0 // 10k per month → 6 months

	result := ProjectCompletionDate(remainingWork, burnRate, startDate)

	// 6 months * 30.44 days = 182.64, ceiling = 183 days
	expectedApprox := startDate.AddDate(0, 0, 183)
	if result != expectedApprox {
		t.Errorf("expected projection around %v, got %v", expectedApprox, result)
	}
}

func TestProjectCompletionDate_ZeroBurnRate(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	remainingWork := 60000.0
	burnRate := 0.0

	result := ProjectCompletionDate(remainingWork, burnRate, startDate)

	// Should return far future (100 years from start)
	farFuture := startDate.AddDate(100, 0, 0)
	if result != farFuture {
		t.Errorf("expected far-future date %v for zero burn rate, got %v", farFuture, result)
	}
}

func TestProjectCompletionDate_NoRemainingWork(t *testing.T) {
	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	remainingWork := 0.0
	burnRate := 10000.0

	result := ProjectCompletionDate(remainingWork, burnRate, startDate)

	// Already complete - should return startDate
	if result != startDate {
		t.Errorf("expected start date %v for no remaining work, got %v", startDate, result)
	}
}

func TestCalculateVariance_Integration(t *testing.T) {
	clin := &ContractLineItem{
		CLINID:    "CLIN-100",
		Ceiling:   500000.0,
		Obligated: 400000.0,
		Expended:  300000.0,
		EAC:       350000.0,
	}
	expenditures := []float64{25000.0, 30000.0, 35000.0}

	popEnd := time.Now().AddDate(1, 0, 0)
	result := CalculateVariance(clin, expenditures, popEnd)

	if result == nil {
		t.Fatal("expected non-nil VarianceAnalysis")
	}
	if result.CLINID != "CLIN-100" {
		t.Errorf("expected CLINID CLIN-100, got %s", result.CLINID)
	}
	// Overrun: max(0, 300000 - 500000) = 0
	if result.Overrun != 0 {
		t.Errorf("expected overrun 0, got %.2f", result.Overrun)
	}
	// Under-run: max(0, 400000 - 350000) = 50000
	expectedUnderRun := 50000.0
	if math.Abs(result.UnderRun-expectedUnderRun) > epsilon {
		t.Errorf("expected under-run %.2f, got %.2f", expectedUnderRun, result.UnderRun)
	}
	// Burn rate: (25000 + 30000 + 35000) / 3 = 30000
	expectedBurnRate := 30000.0
	if math.Abs(result.BurnRate-expectedBurnRate) > epsilon {
		t.Errorf("expected burn rate %.2f, got %.2f", expectedBurnRate, result.BurnRate)
	}
	// Risk level should default to GREEN (task 11.2 handles real determination)
	if result.RiskLevel != RiskLevelGreen {
		t.Errorf("expected default risk level GREEN, got %s", result.RiskLevel)
	}
}
