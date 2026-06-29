package portal

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 17.1, 17.2**
// Property 17: Variance Calculation Correctness
// For any CLIN with ceiling, obligated, expended, and EAC values:
// overrun equals max(0, expended - ceiling), under-run equals max(0, obligated - EAC),
// and both values are never negative.

func TestProperty17_OverrunIsMaxZeroExpendedMinusCeiling(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		expended := rapid.Float64Range(0, 100_000_000).Draw(t, "expended")
		ceiling := rapid.Float64Range(0, 100_000_000).Draw(t, "ceiling")

		clin := &ContractLineItem{
			CLINID:   "CLIN-TEST",
			Expended: expended,
			Ceiling:  ceiling,
		}

		overrun := CalculateOverrun(clin)

		// Property: overrun == max(0, expended - ceiling)
		expectedOverrun := expended - ceiling
		if expectedOverrun < 0 {
			expectedOverrun = 0
		}

		if overrun != expectedOverrun {
			t.Fatalf("overrun mismatch: got %f, expected max(0, %f - %f) = %f",
				overrun, expended, ceiling, expectedOverrun)
		}

		// Property: overrun is never negative
		if overrun < 0 {
			t.Fatalf("overrun is negative: %f (expended=%f, ceiling=%f)",
				overrun, expended, ceiling)
		}
	})
}

func TestProperty17_UnderRunIsMaxZeroObligatedMinusEAC(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		obligated := rapid.Float64Range(0, 100_000_000).Draw(t, "obligated")
		eac := rapid.Float64Range(0, 100_000_000).Draw(t, "eac")

		clin := &ContractLineItem{
			CLINID:    "CLIN-TEST",
			Obligated: obligated,
			EAC:       eac,
		}

		underRun := CalculateUnderRun(clin)

		// Property: under-run == max(0, obligated - EAC)
		expectedUnderRun := obligated - eac
		if expectedUnderRun < 0 {
			expectedUnderRun = 0
		}

		if underRun != expectedUnderRun {
			t.Fatalf("under-run mismatch: got %f, expected max(0, %f - %f) = %f",
				underRun, obligated, eac, expectedUnderRun)
		}

		// Property: under-run is never negative
		if underRun < 0 {
			t.Fatalf("under-run is negative: %f (obligated=%f, eac=%f)",
				underRun, obligated, eac)
		}
	})
}

func TestProperty17_OverrunAndUnderRunAlwaysNonNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ceiling := rapid.Float64Range(0, 100_000_000).Draw(t, "ceiling")
		obligated := rapid.Float64Range(0, 100_000_000).Draw(t, "obligated")
		expended := rapid.Float64Range(0, 100_000_000).Draw(t, "expended")
		eac := rapid.Float64Range(0, 100_000_000).Draw(t, "eac")

		clin := &ContractLineItem{
			CLINID:    "CLIN-TEST",
			Ceiling:   ceiling,
			Obligated: obligated,
			Expended:  expended,
			EAC:       eac,
		}

		overrun := CalculateOverrun(clin)
		underRun := CalculateUnderRun(clin)

		// Property: both overrun and under-run are always >= 0
		if overrun < 0 {
			t.Fatalf("overrun is negative: %f for CLIN (ceiling=%f, expended=%f)",
				overrun, ceiling, expended)
		}
		if underRun < 0 {
			t.Fatalf("under-run is negative: %f for CLIN (obligated=%f, eac=%f)",
				underRun, obligated, eac)
		}
	})
}
