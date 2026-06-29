package portal

import "time"

// CLINType represents the type of a contract line item.
type CLINType string

const (
	CLINTypeFFP    CLINType = "FFP"
	CLINTypeCPFF   CLINType = "CPFF"
	CLINTypeCPIF   CLINType = "CPIF"
	CLINTypeTAndM  CLINType = "T_AND_M"
	CLINTypeOption CLINType = "OPTION"
)

// CLINStatus represents the lifecycle status of a CLIN.
type CLINStatus string

const (
	CLINStatusActive       CLINStatus = "ACTIVE"
	CLINStatusExercised    CLINStatus = "EXERCISED"
	CLINStatusCompleted    CLINStatus = "COMPLETED"
	CLINStatusExpired      CLINStatus = "EXPIRED"
	CLINStatusNotExercised CLINStatus = "NOT_EXERCISED"
)

// ContractLineItem represents a distinct element of work or supply on a contract.
type ContractLineItem struct {
	CLINID                 string     `json:"clinId"`
	CLINNumber             string     `json:"clinNumber"`
	Description            string     `json:"description"`
	CLINType               CLINType   `json:"clinType"`
	CLINStatus             CLINStatus `json:"clinStatus"`
	Ceiling                float64    `json:"ceiling"`
	Obligated              float64    `json:"obligated"`
	Expended               float64    `json:"expended"`
	EAC                    float64    `json:"eac"`
	OptionExerciseDeadline *time.Time `json:"optionExerciseDeadline,omitempty"`
}
