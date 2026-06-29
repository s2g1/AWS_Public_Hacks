package portal

import "time"

// ContractType represents the type of federal contract.
type ContractType string

const (
	ContractTypeFFP   ContractType = "FFP"
	ContractTypeCPFF  ContractType = "CPFF"
	ContractTypeCPIF  ContractType = "CPIF"
	ContractTypeTAndM ContractType = "T_AND_M"
	ContractTypeIDIQ  ContractType = "IDIQ"
)

// ContractStatus represents the lifecycle status of a contract.
type ContractStatus string

const (
	ContractStatusActive    ContractStatus = "ACTIVE"
	ContractStatusCompleted ContractStatus = "COMPLETED"
	ContractStatusTerminated ContractStatus = "TERMINATED"
)

// Contract represents a federal contract with its financial data.
type Contract struct {
	ContractID               string             `json:"contractId"`
	ContractNumber           string             `json:"contractNumber"`
	ContractType             ContractType       `json:"contractType"`
	TotalCeiling             float64            `json:"totalCeiling"`
	TotalObligated           float64            `json:"totalObligated"`
	TotalExpended            float64            `json:"totalExpended"`
	EAC                      float64            `json:"eac"`
	PeriodOfPerformanceStart time.Time          `json:"periodOfPerformanceStart"`
	PeriodOfPerformanceEnd   time.Time          `json:"periodOfPerformanceEnd"`
	Status                   ContractStatus     `json:"status"`
	CLINs                    []ContractLineItem `json:"clins"`
	Modifications            []Modification     `json:"modifications"`
	OrganizationID           string             `json:"organizationId"`
}

// Modification represents a contract modification record.
type Modification struct {
	ModificationID string    `json:"modificationId"`
	Description    string    `json:"description"`
	Amount         float64   `json:"amount"`
	CreatedAt      time.Time `json:"createdAt"`
	CreatedBy      string    `json:"createdBy"`
}
