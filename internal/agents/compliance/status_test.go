package compliance

import (
	"testing"

	"federal-payment-processing/internal/models"
)

func TestDetermineComplianceStatus_NoFlags_Compliant(t *testing.T) {
	status := DetermineComplianceStatus([]models.ComplianceFlag{})
	if status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT for no flags, got %s", status)
	}
}

func TestDetermineComplianceStatus_NilFlags_Compliant(t *testing.T) {
	status := DetermineComplianceStatus(nil)
	if status != models.ComplianceStatusCompliant {
		t.Errorf("expected COMPLIANT for nil flags, got %s", status)
	}
}

func TestDetermineComplianceStatus_SingleBlocking_NonCompliant(t *testing.T) {
	flags := []models.ComplianceFlag{
		{Rule: "OFAC_SANCTIONS", Severity: models.FlagSeverityBlocking, Message: "match"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for BLOCKING flag, got %s", status)
	}
}

func TestDetermineComplianceStatus_MultipleBlocking_NonCompliant(t *testing.T) {
	flags := []models.ComplianceFlag{
		{Rule: "OFAC_SANCTIONS", Severity: models.FlagSeverityBlocking, Message: "ofac match"},
		{Rule: "DEBARMENT", Severity: models.FlagSeverityBlocking, Message: "debarment match"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT for multiple BLOCKING flags, got %s", status)
	}
}

func TestDetermineComplianceStatus_SingleRequiresReview_CompliantWithConditions(t *testing.T) {
	flags := []models.ComplianceFlag{
		{Rule: "THRESHOLD_EXCEEDED", Severity: models.FlagSeverityRequiresReview, Message: "over threshold"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusCompliantWithConditions {
		t.Errorf("expected COMPLIANT_WITH_CONDITIONS for REQUIRES_REVIEW flag, got %s", status)
	}
}

func TestDetermineComplianceStatus_MultipleRequiresReview_CompliantWithConditions(t *testing.T) {
	flags := []models.ComplianceFlag{
		{Rule: "THRESHOLD_EXCEEDED", Severity: models.FlagSeverityRequiresReview, Message: "threshold"},
		{Rule: "ANNUAL_LIMIT", Severity: models.FlagSeverityRequiresReview, Message: "annual limit"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusCompliantWithConditions {
		t.Errorf("expected COMPLIANT_WITH_CONDITIONS for multiple REQUIRES_REVIEW flags, got %s", status)
	}
}

func TestDetermineComplianceStatus_MixedBlockingAndReview_NonCompliant(t *testing.T) {
	flags := []models.ComplianceFlag{
		{Rule: "THRESHOLD_EXCEEDED", Severity: models.FlagSeverityRequiresReview, Message: "threshold"},
		{Rule: "OFAC_SANCTIONS", Severity: models.FlagSeverityBlocking, Message: "ofac match"},
		{Rule: "ANNUAL_LIMIT", Severity: models.FlagSeverityRequiresReview, Message: "annual limit"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT when BLOCKING is present among REQUIRES_REVIEW, got %s", status)
	}
}

func TestDetermineComplianceStatus_BlockingAfterReview_NonCompliant(t *testing.T) {
	// BLOCKING flag appears after REQUIRES_REVIEW flags — should still be NON_COMPLIANT.
	flags := []models.ComplianceFlag{
		{Rule: "ANNUAL_LIMIT", Severity: models.FlagSeverityRequiresReview, Message: "annual limit"},
		{Rule: "FAR_VIOLATION", Severity: models.FlagSeverityBlocking, Message: "far violation"},
	}
	status := DetermineComplianceStatus(flags)
	if status != models.ComplianceStatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT when BLOCKING follows REQUIRES_REVIEW, got %s", status)
	}
}
