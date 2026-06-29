package portal

import "time"

// REAStatus represents the lifecycle status of a Request for Equitable Adjustment.
type REAStatus string

const (
	REAStatusSubmitted             REAStatus = "SUBMITTED"
	REAStatusApproved              REAStatus = "APPROVED"
	REAStatusPartiallyApproved     REAStatus = "PARTIALLY_APPROVED"
	REAStatusDenied                REAStatus = "DENIED"
	REAStatusAdditionalInfoRequested REAStatus = "ADDITIONAL_INFO_REQUESTED"
)

// REA represents a Request for Equitable Adjustment submitted by a contractor.
type REA struct {
	REAID             string    `json:"reaId"`
	ContractID        string    `json:"contractId"`
	RequestedAmount   float64   `json:"requestedAmount"`
	ApprovedAmount    float64   `json:"approvedAmount"`
	AffectedCLINs     []string  `json:"affectedClins"`
	Status            REAStatus `json:"status"`
	Justification     string    `json:"justification"`
	SubmittedBy       string    `json:"submittedBy"`
	SubmittedAt       time.Time `json:"submittedAt"`
	ResolvedAt        *time.Time `json:"resolvedAt,omitempty"`
	ResponseRationale string    `json:"responseRationale,omitempty"`
}

// REAResponse represents a contracting officer's response to an REA.
type REAResponse struct {
	ResponseType   REAStatus `json:"responseType"`
	ApprovedAmount float64   `json:"approvedAmount"`
	Rationale      string    `json:"rationale"`
	RespondedBy    string    `json:"respondedBy"`
	RespondedAt    time.Time `json:"respondedAt"`
}
