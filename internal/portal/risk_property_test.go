package portal

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// **Validates: Requirements 17.3, 17.4, 17.5, 17.6, 17.7**
// Property 18: Risk Level Determination
// Tests that DetermineRiskLevel assigns the correct risk level based on
// overrun, under-run percentage, expenditure ratio, and projected completion.

// Property 18.1: Any CLIN with overrun > 0 (expended > ceiling) always gets RED risk level.
func TestProperty18_OverrunAlwaysRed(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a CLIN where expended > ceiling (overrun > 0)
		ceiling := rapid.Float64Range(1, 10_000_000).Draw(t, "ceiling")
		// Expended must exceed ceiling to produce overrun
		overrun := rapid.Float64Range(0.01, 5_000_000).Draw(t, "overrun")
		expended := ceiling + overrun

		obligated := rapid.Float64Range(expended, expended+5_000_000).Draw(t, "obligated")
		eac := rapid.Float64Range(0, obligated).Draw(t, "eac")

		clin := &ContractLineItem{
			CLINID:     "CLIN-TEST",
			Ceiling:    ceiling,
			Obligated:  obligated,
			Expended:   expended,
			EAC:        eac,
			CLINStatus: CLINStatusActive,
		}

		variance := &VarianceAnalysis{
			CLINID:                  clin.CLINID,
			Overrun:                 overrun,
			UnderRun:                0,
			BurnRate:                1000,
			ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // in the past (before PoP end)
		}

		popEnd := time.Now().AddDate(1, 0, 0) // far future

		risk := DetermineRiskLevel(clin, variance, popEnd)

		if risk != RiskLevelRed {
			t.Fatalf("expected RED risk for overrun=%f (expended=%f > ceiling=%f), got %s",
				overrun, expended, ceiling, risk)
		}
	})
}

// Property 18.2: Any CLIN with under-run > 40% of obligated always gets RED risk level.
func TestProperty18_UnderRunOver40PctAlwaysRed(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a CLIN with under-run percentage > 40%
		// under-run% = (obligated - EAC) / obligated * 100
		// We need (obligated - EAC) / obligated > 0.40
		// So EAC < 0.60 * obligated
		obligated := rapid.Float64Range(100, 10_000_000).Draw(t, "obligated")
		// EAC must be less than 60% of obligated for under-run > 40%
		maxEAC := obligated * 0.5999 // strictly less than 60%
		eac := rapid.Float64Range(0, maxEAC).Draw(t, "eac")

		// No overrun: expended <= ceiling
		ceiling := rapid.Float64Range(1_000_000, 20_000_000).Draw(t, "ceiling")
		expended := rapid.Float64Range(0, ceiling).Draw(t, "expended")

		clin := &ContractLineItem{
			CLINID:     "CLIN-TEST",
			Ceiling:    ceiling,
			Obligated:  obligated,
			Expended:   expended,
			EAC:        eac,
			CLINStatus: CLINStatusActive,
		}

		variance := &VarianceAnalysis{
			CLINID:                  clin.CLINID,
			Overrun:                 0, // no overrun
			UnderRun:                obligated - eac,
			BurnRate:                1000,
			ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // before PoP end
		}

		popEnd := time.Now().AddDate(1, 0, 0) // far future

		risk := DetermineRiskLevel(clin, variance, popEnd)

		underRunPct := (obligated - eac) / obligated * 100.0
		if risk != RiskLevelRed {
			t.Fatalf("expected RED risk for under-run pct=%.2f%% (obligated=%f, eac=%f), got %s",
				underRunPct, obligated, eac, risk)
		}
	})
}

// Property 18.3: Any CLIN with expenditure ratio > 90% of ceiling AND ACTIVE status
// always gets at least YELLOW (could be RED if other conditions also trigger).
func TestProperty18_HighExpenditureRatioActiveAtLeastYellow(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a CLIN with expenditure ratio > 90% of ceiling
		// expenditureRatio = expended / ceiling > 0.90
		// But expended <= ceiling (no overrun) to isolate this condition
		ceiling := rapid.Float64Range(1000, 10_000_000).Draw(t, "ceiling")
		// expended between 90.01% and 100% of ceiling
		minExpended := ceiling * 0.9001
		expended := rapid.Float64Range(minExpended, ceiling).Draw(t, "expended")

		// Ensure no under-run > 40% to avoid RED from that condition
		// under-run% = (obligated - EAC) / obligated * 100
		// Keep under-run% < 20% so we only test the expenditure ratio condition
		obligated := rapid.Float64Range(1000, 10_000_000).Draw(t, "obligated")
		// EAC should be close to obligated: EAC >= 0.81 * obligated means under-run < 19%
		minEAC := obligated * 0.81
		eac := rapid.Float64Range(minEAC, obligated).Draw(t, "eac")

		clin := &ContractLineItem{
			CLINID:     "CLIN-TEST",
			Ceiling:    ceiling,
			Obligated:  obligated,
			Expended:   expended,
			EAC:        eac,
			CLINStatus: CLINStatusActive,
		}

		variance := &VarianceAnalysis{
			CLINID:                  clin.CLINID,
			Overrun:                 0, // no overrun
			UnderRun:                0,
			BurnRate:                1000,
			ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // before PoP end
		}

		popEnd := time.Now().AddDate(1, 0, 0) // far future

		risk := DetermineRiskLevel(clin, variance, popEnd)

		// Must be at least YELLOW (YELLOW or RED are acceptable)
		if risk == RiskLevelGreen {
			t.Fatalf("expected at least YELLOW for expenditure ratio=%.4f (expended=%f, ceiling=%f) with ACTIVE status, got GREEN",
				expended/ceiling, expended, ceiling)
		}
	})
}

// Property 18.4: Any CLIN with under-run between 20-40% always gets at least YELLOW.
func TestProperty18_UnderRun20To40PctAtLeastYellow(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a CLIN with under-run% strictly between 20% and 40%
		// under-run% = (obligated - EAC) / obligated * 100
		// We need 20 <= underRunPct <= 40
		// Use obligated values large enough to avoid floating-point boundary issues
		obligated := rapid.Float64Range(10_000, 10_000_000).Draw(t, "obligated")
		// Generate under-run percentage comfortably within 20-40 range
		underRunPct := rapid.Float64Range(20.5, 39.5).Draw(t, "underRunPct")
		eac := obligated * (1.0 - underRunPct/100.0)

		// No overrun: expended <= ceiling
		ceiling := rapid.Float64Range(1_000_000, 20_000_000).Draw(t, "ceiling")
		// Keep expenditure ratio <= 89% to avoid triggering that yellow condition
		maxExpended := ceiling * 0.89
		expended := rapid.Float64Range(0, maxExpended).Draw(t, "expended")

		clin := &ContractLineItem{
			CLINID:     "CLIN-TEST",
			Ceiling:    ceiling,
			Obligated:  obligated,
			Expended:   expended,
			EAC:        eac,
			CLINStatus: CLINStatusActive,
		}

		variance := &VarianceAnalysis{
			CLINID:                  clin.CLINID,
			Overrun:                 0,
			UnderRun:                obligated - eac,
			BurnRate:                1000,
			ProjectedCompletionDate: time.Now().AddDate(0, -1, 0), // before PoP end
		}

		popEnd := time.Now().AddDate(1, 0, 0) // far future

		risk := DetermineRiskLevel(clin, variance, popEnd)

		// Must be at least YELLOW (YELLOW or RED are acceptable)
		if risk == RiskLevelGreen {
			t.Fatalf("expected at least YELLOW for under-run pct=%.2f%% (obligated=%f, eac=%f), got GREEN",
				underRunPct, obligated, eac)
		}
	})
}

// Property 18.5: A CLIN with no overrun, under-run < 20%, expenditure ratio <= 90%,
// and projected completion before PoP end always gets GREEN.
func TestProperty18_NoRiskIndicatorsAlwaysGreen(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// No overrun: expended <= ceiling, with expenditure ratio strictly <= 89%
		ceiling := rapid.Float64Range(10_000, 10_000_000).Draw(t, "ceiling")
		// Keep expenditure ratio well below 90% to avoid boundary issues
		maxExpended := ceiling * 0.89
		expended := rapid.Float64Range(0, maxExpended).Draw(t, "expended")

		// Under-run < 20%: (obligated - EAC) / obligated * 100 < 20
		// Keep EAC > 81% of obligated to stay well below 20% under-run
		obligated := rapid.Float64Range(10_000, 10_000_000).Draw(t, "obligated")
		minEAC := obligated * 0.81
		eac := rapid.Float64Range(minEAC, obligated).Draw(t, "eac")

		clin := &ContractLineItem{
			CLINID:     "CLIN-TEST",
			Ceiling:    ceiling,
			Obligated:  obligated,
			Expended:   expended,
			EAC:        eac,
			CLINStatus: CLINStatusActive,
		}

		// Projected completion well before PoP end
		popEnd := time.Now().AddDate(1, 0, 0)
		projectedCompletion := time.Now().AddDate(0, -1, 0) // well before PoP end

		variance := &VarianceAnalysis{
			CLINID:                  clin.CLINID,
			Overrun:                 0,
			UnderRun:                0,
			BurnRate:                1000,
			ProjectedCompletionDate: projectedCompletion,
		}

		risk := DetermineRiskLevel(clin, variance, popEnd)

		if risk != RiskLevelGreen {
			t.Fatalf("expected GREEN for no-risk CLIN (ceiling=%f, expended=%f, obligated=%f, eac=%f, ratio=%.4f, underRun%%=%.2f), got %s",
				ceiling, expended, obligated, eac, expended/ceiling,
				(obligated-eac)/obligated*100.0, risk)
		}
	})
}
