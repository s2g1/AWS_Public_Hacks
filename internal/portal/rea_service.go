package portal

import (
	"errors"
	"fmt"
	"time"
)

// REASubmissionRequest contains the data required to submit a new REA.
type REASubmissionRequest struct {
	RequestedAmount float64  `json:"requestedAmount"`
	AffectedCLINs   []string `json:"affectedClins"`
	Justification   string   `json:"justification"`
	SubmittedBy     string   `json:"submittedBy"`
}

// AuditEntry represents a record of an action taken on a portal entity.
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Actor       string    `json:"actor"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
}

// Notification represents a notification sent to a user.
type Notification struct {
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	SentAt    time.Time `json:"sentAt"`
}

// REASubmissionResult contains the REA and any side effects from submission.
type REASubmissionResult struct {
	REA           *REA           `json:"rea"`
	AuditEntries  []AuditEntry   `json:"auditEntries"`
	Notifications []Notification `json:"notifications"`
}

// IDGenerator is a function type that generates unique IDs.
type IDGenerator func() string

// defaultIDGenerator returns a simple timestamp-based ID.
func defaultIDGenerator() string {
	return fmt.Sprintf("REA-%d", time.Now().UnixNano())
}

// SubmitREA validates and creates a new REA submission for a contract.
// It validates that:
//   - The requested amount is positive
//   - At least one affected CLIN is specified
//   - All referenced CLINs exist on the contract
//
// On success, it creates an REA record with SUBMITTED status, generates
// a notification for the government CO, and logs an audit trail entry.
func SubmitREA(contract *Contract, req REASubmissionRequest) (*REASubmissionResult, error) {
	return SubmitREAWithIDGen(contract, req, defaultIDGenerator)
}

// SubmitREAWithIDGen is like SubmitREA but accepts a custom ID generator for testing.
func SubmitREAWithIDGen(contract *Contract, req REASubmissionRequest, idGen IDGenerator) (*REASubmissionResult, error) {
	// Validate requested amount is positive
	if req.RequestedAmount <= 0 {
		return nil, errors.New("requested amount must be positive")
	}

	// Validate at least one affected CLIN is specified
	if len(req.AffectedCLINs) == 0 {
		return nil, errors.New("at least one affected CLIN must be specified")
	}

	// Validate all referenced CLINs exist on the contract
	clinMap := make(map[string]bool)
	for _, clin := range contract.CLINs {
		clinMap[clin.CLINID] = true
	}
	for _, clinID := range req.AffectedCLINs {
		if !clinMap[clinID] {
			return nil, fmt.Errorf("CLIN %q does not exist on contract %s", clinID, contract.ContractID)
		}
	}

	// Create REA record with SUBMITTED status
	now := time.Now()
	rea := &REA{
		REAID:           idGen(),
		ContractID:      contract.ContractID,
		RequestedAmount: req.RequestedAmount,
		ApprovedAmount:  0,
		AffectedCLINs:   req.AffectedCLINs,
		Status:          REAStatusSubmitted,
		Justification:   req.Justification,
		SubmittedBy:     req.SubmittedBy,
		SubmittedAt:     now,
	}

	// Log audit trail entry
	auditEntry := AuditEntry{
		Timestamp:   now,
		Actor:       req.SubmittedBy,
		Action:      "REA_SUBMITTED",
		Description: fmt.Sprintf("REA submitted for contract %s requesting $%.2f affecting CLINs %v", contract.ContractID, req.RequestedAmount, req.AffectedCLINs),
	}

	// Notify government contracting officer
	notification := Notification{
		Recipient: "contracting_officer",
		Subject:   fmt.Sprintf("New REA Submitted for Contract %s", contract.ContractNumber),
		Body:      fmt.Sprintf("A new Request for Equitable Adjustment has been submitted by %s for contract %s requesting $%.2f. Justification: %s", req.SubmittedBy, contract.ContractNumber, req.RequestedAmount, req.Justification),
		SentAt:    now,
	}

	return &REASubmissionResult{
		REA:           rea,
		AuditEntries:  []AuditEntry{auditEntry},
		Notifications: []Notification{notification},
	}, nil
}
