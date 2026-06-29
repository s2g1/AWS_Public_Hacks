package compliance

import "federal-payment-processing/internal/models"

// DetermineComplianceStatus evaluates a set of compliance flags and returns the
// appropriate compliance status:
//   - Any BLOCKING flag → NON_COMPLIANT
//   - Any REQUIRES_REVIEW flag (no BLOCKING) → COMPLIANT_WITH_CONDITIONS
//   - No flags → COMPLIANT
//
// This function consolidates status determination logic so it can be reused
// across the compliance pipeline (OFAC, debarment, spending, FAR checks) and
// property-tested independently.
func DetermineComplianceStatus(flags []models.ComplianceFlag) models.ComplianceStatus {
	if len(flags) == 0 {
		return models.ComplianceStatusCompliant
	}

	for _, flag := range flags {
		if flag.Severity == models.FlagSeverityBlocking {
			return models.ComplianceStatusNonCompliant
		}
	}

	// At least one flag exists but none are BLOCKING — must be REQUIRES_REVIEW.
	return models.ComplianceStatusCompliantWithConditions
}
