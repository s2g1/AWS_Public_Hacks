package models

import "time"

// ComplianceStatus represents the overall compliance determination.
type ComplianceStatus string

const (
	ComplianceStatusCompliant              ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant           ComplianceStatus = "NON_COMPLIANT"
	ComplianceStatusCompliantWithConditions ComplianceStatus = "COMPLIANT_WITH_CONDITIONS"
)

// FlagSeverity represents the severity of a compliance flag.
type FlagSeverity string

const (
	FlagSeverityBlocking       FlagSeverity = "BLOCKING"
	FlagSeverityRequiresReview FlagSeverity = "REQUIRES_REVIEW"
)

// ComplianceFlag represents a specific compliance concern identified during evaluation.
type ComplianceFlag struct {
	Rule     string       `json:"rule"`
	Severity FlagSeverity `json:"severity"`
	Message  string       `json:"message"`
}

// ComplianceResult contains the complete output of the compliance agent.
type ComplianceResult struct {
	Status    ComplianceStatus `json:"status"`
	Flags     []ComplianceFlag `json:"flags"`
	Rules     []string         `json:"rules"`
	CheckedAt time.Time        `json:"checkedAt"`
}
