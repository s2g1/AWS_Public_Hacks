package portal

import "time"

// DetermineRiskLevel evaluates the financial risk level for a CLIN based on
// variance analysis and period of performance constraints.
//
// Risk levels are assigned with the following priority:
//   - RED: any overrun (expended > ceiling), under-run > 40% of obligated,
//     or projected completion date exceeds PoP end date
//   - YELLOW: expenditure ratio > 90% of ceiling AND CLIN is ACTIVE,
//     or under-run between 20-40% of obligated
//   - GREEN: otherwise
func DetermineRiskLevel(clin *ContractLineItem, variance *VarianceAnalysis, popEnd time.Time) RiskLevel {
	// RED conditions (check first — RED takes priority over YELLOW)
	if variance.Overrun > 0 {
		return RiskLevelRed
	}

	underRunPct := CalculateUnderRunPercentage(clin)
	if underRunPct > 40.0 {
		return RiskLevelRed
	}

	if variance.ProjectedCompletionDate.After(popEnd) {
		return RiskLevelRed
	}

	// YELLOW conditions
	expenditureRatio := CalculateExpenditureRatio(clin)
	if expenditureRatio > 0.90 && clin.CLINStatus == CLINStatusActive {
		return RiskLevelYellow
	}

	if underRunPct >= 20.0 && underRunPct <= 40.0 {
		return RiskLevelYellow
	}

	// GREEN: no risk indicators triggered
	return RiskLevelGreen
}

// CalculateUnderRunPercentage returns the under-run as a percentage of the
// obligated amount: (obligated - EAC) / obligated * 100.
// Returns 0 if obligated is zero (avoids division by zero).
func CalculateUnderRunPercentage(clin *ContractLineItem) float64 {
	if clin.Obligated == 0 {
		return 0
	}
	underRun := clin.Obligated - clin.EAC
	if underRun <= 0 {
		return 0
	}
	return (underRun / clin.Obligated) * 100.0
}

// CalculateExpenditureRatio returns the ratio of expended to ceiling.
// Returns 0 if ceiling is zero (avoids division by zero).
func CalculateExpenditureRatio(clin *ContractLineItem) float64 {
	if clin.Ceiling == 0 {
		return 0
	}
	return clin.Expended / clin.Ceiling
}
