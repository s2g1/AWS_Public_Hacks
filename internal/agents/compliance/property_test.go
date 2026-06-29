package compliance

import (
	"testing"

	"federal-payment-processing/internal/models"

	"pgregory.net/rapid"
)

// **Validates: Requirements 6.2, 7.2, 9.3**
// **Validates: Requirements 9.3, 9.4, 9.5**

// genFlagSeverity generates an arbitrary FlagSeverity value.
func genFlagSeverity() *rapid.Generator[models.FlagSeverity] {
	return rapid.Custom(func(t *rapid.T) models.FlagSeverity {
		severities := []models.FlagSeverity{
			models.FlagSeverityBlocking,
			models.FlagSeverityRequiresReview,
		}
		idx := rapid.IntRange(0, len(severities)-1).Draw(t, "severityIdx")
		return severities[idx]
	})
}

// genComplianceFlag generates an arbitrary ComplianceFlag.
func genComplianceFlag() *rapid.Generator[models.ComplianceFlag] {
	return rapid.Custom(func(t *rapid.T) models.ComplianceFlag {
		severity := genFlagSeverity().Draw(t, "severity")
		rule := rapid.StringMatching(`[A-Z_]{3,20}`).Draw(t, "rule")
		message := rapid.StringMatching(`[a-z ]{5,30}`).Draw(t, "message")
		return models.ComplianceFlag{
			Rule:     rule,
			Severity: severity,
			Message:  message,
		}
	})
}

// genRequiresReviewFlag generates a ComplianceFlag with REQUIRES_REVIEW severity.
func genRequiresReviewFlag() *rapid.Generator[models.ComplianceFlag] {
	return rapid.Custom(func(t *rapid.T) models.ComplianceFlag {
		rule := rapid.StringMatching(`[A-Z_]{3,20}`).Draw(t, "rule")
		message := rapid.StringMatching(`[a-z ]{5,30}`).Draw(t, "message")
		return models.ComplianceFlag{
			Rule:     rule,
			Severity: models.FlagSeverityRequiresReview,
			Message:  message,
		}
	})
}

// genBlockingFlag generates a ComplianceFlag with BLOCKING severity.
func genBlockingFlag() *rapid.Generator[models.ComplianceFlag] {
	return rapid.Custom(func(t *rapid.T) models.ComplianceFlag {
		rule := rapid.StringMatching(`[A-Z_]{3,20}`).Draw(t, "rule")
		message := rapid.StringMatching(`[a-z ]{5,30}`).Draw(t, "message")
		return models.ComplianceFlag{
			Rule:     rule,
			Severity: models.FlagSeverityBlocking,
			Message:  message,
		}
	})
}

// TestProperty_EmptyFlags_Compliant verifies that when no compliance flags are present,
// the status is always COMPLIANT.
func TestProperty_EmptyFlags_Compliant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Empty flags slice — no compliance issues
		flags := []models.ComplianceFlag{}

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusCompliant {
			t.Fatalf("expected COMPLIANT for empty flags, got %q", status)
		}
	})
}

// TestProperty_OnlyRequiresReviewFlags_CompliantWithConditions verifies that when
// all flags are REQUIRES_REVIEW (and no BLOCKING flags exist), the status is
// COMPLIANT_WITH_CONDITIONS.
func TestProperty_OnlyRequiresReviewFlags_CompliantWithConditions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate 1 or more REQUIRES_REVIEW flags only
		count := rapid.IntRange(1, 10).Draw(t, "flagCount")
		flags := make([]models.ComplianceFlag, count)
		for i := range flags {
			flags[i] = genRequiresReviewFlag().Draw(t, "reviewFlag")
		}

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusCompliantWithConditions {
			t.Fatalf("expected COMPLIANT_WITH_CONDITIONS for only REQUIRES_REVIEW flags (%d flags), got %q", count, status)
		}
	})
}

// TestProperty_AnyBlockingFlag_NonCompliant verifies that when any BLOCKING flag is
// present (regardless of other flag types), the status is always NON_COMPLIANT.
func TestProperty_AnyBlockingFlag_NonCompliant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a mix of flags, ensuring at least one BLOCKING flag
		blockingFlag := genBlockingFlag().Draw(t, "blockingFlag")

		// Optionally add other flags of any severity
		otherCount := rapid.IntRange(0, 10).Draw(t, "otherCount")
		flags := make([]models.ComplianceFlag, 0, otherCount+1)
		for i := 0; i < otherCount; i++ {
			flags = append(flags, genComplianceFlag().Draw(t, "otherFlag"))
		}

		// Insert the blocking flag at a random position
		insertIdx := rapid.IntRange(0, len(flags)).Draw(t, "insertIdx")
		flags = append(flags, models.ComplianceFlag{}) // grow slice
		copy(flags[insertIdx+1:], flags[insertIdx:])
		flags[insertIdx] = blockingFlag

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusNonCompliant {
			t.Fatalf("expected NON_COMPLIANT when BLOCKING flag present, got %q (flags: %v)", status, flags)
		}
	})
}

// TestProperty_StatusesMutuallyExclusiveAndExhaustive verifies that for any arbitrary
// set of flags, the returned status is always exactly one of the three defined statuses
// (COMPLIANT, COMPLIANT_WITH_CONDITIONS, NON_COMPLIANT).
func TestProperty_StatusesMutuallyExclusiveAndExhaustive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate an arbitrary number of flags (including zero)
		count := rapid.IntRange(0, 15).Draw(t, "flagCount")
		flags := make([]models.ComplianceFlag, count)
		for i := range flags {
			flags[i] = genComplianceFlag().Draw(t, "flag")
		}

		status := DetermineComplianceStatus(flags)

		validStatuses := map[models.ComplianceStatus]bool{
			models.ComplianceStatusCompliant:              true,
			models.ComplianceStatusNonCompliant:           true,
			models.ComplianceStatusCompliantWithConditions: true,
		}

		if !validStatuses[status] {
			t.Fatalf("status %q is not one of the three valid statuses", status)
		}
	})
}

// --- Property 6: Compliance Blocking Enforcement ---
// Additional property tests validating OFAC and debarment blocking always yield NON_COMPLIANT.

// TestProperty_ComplianceBlocking_OFACAlwaysNonCompliant verifies that an OFAC
// BLOCKING match always results in NON_COMPLIANT regardless of other flags present.
func TestProperty_ComplianceBlocking_OFACAlwaysNonCompliant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Create an OFAC BLOCKING flag.
		ofacFlag := models.ComplianceFlag{
			Rule:     "OFAC_SANCTIONS",
			Severity: models.FlagSeverityBlocking,
			Message:  rapid.StringMatching(`[a-z ]{5,30}`).Draw(t, "ofacMsg"),
		}

		// Generate 0-5 random REQUIRES_REVIEW flags alongside the OFAC flag.
		extraCount := rapid.IntRange(0, 5).Draw(t, "extraCount")
		flags := []models.ComplianceFlag{ofacFlag}
		for i := 0; i < extraCount; i++ {
			flags = append(flags, genRequiresReviewFlag().Draw(t, "otherFlag"))
		}

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusNonCompliant {
			t.Fatalf("OFAC BLOCKING flag must yield NON_COMPLIANT, got %s (flags: %+v)", status, flags)
		}
	})
}

// TestProperty_ComplianceBlocking_DebarmentAlwaysNonCompliant verifies that a
// DEBARMENT BLOCKING match always results in NON_COMPLIANT regardless of other flags.
func TestProperty_ComplianceBlocking_DebarmentAlwaysNonCompliant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Create a DEBARMENT BLOCKING flag.
		debarmentFlag := models.ComplianceFlag{
			Rule:     "DEBARMENT",
			Severity: models.FlagSeverityBlocking,
			Message:  rapid.StringMatching(`[a-z ]{5,30}`).Draw(t, "debarmentMsg"),
		}

		// Generate 0-5 random REQUIRES_REVIEW flags alongside the DEBARMENT flag.
		extraCount := rapid.IntRange(0, 5).Draw(t, "extraCount")
		flags := []models.ComplianceFlag{debarmentFlag}
		for i := 0; i < extraCount; i++ {
			flags = append(flags, genRequiresReviewFlag().Draw(t, "otherFlag"))
		}

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusNonCompliant {
			t.Fatalf("DEBARMENT BLOCKING flag must yield NON_COMPLIANT, got %s (flags: %+v)", status, flags)
		}
	})
}

// TestProperty_ComplianceBlocking_MixedBlockingDominates verifies that a random
// mix of BLOCKING and REQUIRES_REVIEW flags always results in NON_COMPLIANT
// (BLOCKING dominates).
func TestProperty_ComplianceBlocking_MixedBlockingDominates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate at least one BLOCKING and at least one REQUIRES_REVIEW.
		blockingFlag := genBlockingFlag().Draw(t, "blockingFlag")
		reviewFlag := genRequiresReviewFlag().Draw(t, "reviewFlag")

		// Generate 0-4 additional flags of mixed severity.
		extraCount := rapid.IntRange(0, 4).Draw(t, "extraCount")
		flags := []models.ComplianceFlag{reviewFlag, blockingFlag}
		for i := 0; i < extraCount; i++ {
			flags = append(flags, genComplianceFlag().Draw(t, "extraFlag"))
		}

		status := DetermineComplianceStatus(flags)
		if status != models.ComplianceStatusNonCompliant {
			t.Fatalf("BLOCKING must dominate over REQUIRES_REVIEW, expected NON_COMPLIANT, got %s (flags: %+v)", status, flags)
		}
	})
}
