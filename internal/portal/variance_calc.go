package portal

import (
	"math"
	"time"
)

// CalculateOverrun returns the overrun amount for a CLIN.
// Overrun is defined as max(0, expended - ceiling).
func CalculateOverrun(clin *ContractLineItem) float64 {
	return math.Max(0, clin.Expended-clin.Ceiling)
}

// CalculateUnderRun returns the under-run amount for a CLIN.
// Under-run is defined as max(0, obligated - EAC).
func CalculateUnderRun(clin *ContractLineItem) float64 {
	return math.Max(0, clin.Obligated-clin.EAC)
}

// CalculateBurnRate computes the average monthly expenditure from the last 3 months.
// If fewer than 3 months are provided, it averages what's available.
// Returns 0 if no expenditures are provided.
func CalculateBurnRate(monthlyExpenditures []float64) float64 {
	if len(monthlyExpenditures) == 0 {
		return 0
	}

	// Take the last 3 months (or fewer if not enough data)
	count := 3
	if len(monthlyExpenditures) < count {
		count = len(monthlyExpenditures)
	}

	sum := 0.0
	start := len(monthlyExpenditures) - count
	for i := start; i < len(monthlyExpenditures); i++ {
		sum += monthlyExpenditures[i]
	}

	return sum / float64(count)
}

// ProjectCompletionDate projects the completion date based on remaining work and
// monthly burn rate. If burnRate is zero or negative, returns a far-future date
// to indicate that completion cannot be projected.
func ProjectCompletionDate(remainingWork float64, burnRate float64, startDate time.Time) time.Time {
	if burnRate <= 0 {
		// Cannot project completion with zero or negative burn rate.
		// Return a date far in the future (100 years from start).
		return startDate.AddDate(100, 0, 0)
	}

	if remainingWork <= 0 {
		// Already complete
		return startDate
	}

	// Calculate months to complete
	monthsToComplete := remainingWork / burnRate

	// Convert months to duration (approximate: 30.44 days per month)
	daysToComplete := int(math.Ceil(monthsToComplete * 30.44))
	return startDate.AddDate(0, 0, daysToComplete)
}

// CalculateVariance combines all variance calculations into a VarianceAnalysis struct
// and determines the risk level based on overrun, under-run, and projected completion.
func CalculateVariance(clin *ContractLineItem, monthlyExpenditures []float64, popEnd time.Time) *VarianceAnalysis {
	overrun := CalculateOverrun(clin)
	underRun := CalculateUnderRun(clin)
	burnRate := CalculateBurnRate(monthlyExpenditures)

	// Remaining work is the difference between the ceiling and what's been expended
	remainingWork := math.Max(0, clin.Ceiling-clin.Expended)

	// Project completion from now
	projectedCompletion := ProjectCompletionDate(remainingWork, burnRate, time.Now())

	variance := &VarianceAnalysis{
		CLINID:                  clin.CLINID,
		Overrun:                 overrun,
		UnderRun:                underRun,
		BurnRate:                burnRate,
		ProjectedCompletionDate: projectedCompletion,
	}

	variance.RiskLevel = DetermineRiskLevel(clin, variance, popEnd)

	return variance
}
