package portal

import "time"

// RiskLevel represents the financial risk classification for a CLIN.
type RiskLevel string

const (
	RiskLevelRed    RiskLevel = "RED"
	RiskLevelYellow RiskLevel = "YELLOW"
	RiskLevelGreen  RiskLevel = "GREEN"
)

// VarianceAnalysis contains the computed financial variance metrics for a CLIN.
type VarianceAnalysis struct {
	CLINID                 string    `json:"clinId"`
	Overrun                float64   `json:"overrun"`
	UnderRun               float64   `json:"underRun"`
	BurnRate               float64   `json:"burnRate"`
	ProjectedCompletionDate time.Time `json:"projectedCompletionDate"`
	RiskLevel              RiskLevel `json:"riskLevel"`
}
