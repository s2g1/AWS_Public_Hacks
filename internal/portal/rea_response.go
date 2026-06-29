package portal

import (
	"errors"
	"fmt"
	"time"
)

// REAResponseResult contains the results and side effects of responding to an REA.
type REAResponseResult struct {
	REA          *REA         `json:"rea"`
	Modification *Modification `json:"modification,omitempty"`
	AuditEntries []AuditEntry  `json:"auditEntries"`
}

// RespondToREA processes a contracting officer's response to an REA.
// It handles APPROVED, PARTIALLY_APPROVED, DENIED, and ADDITIONAL_INFO_REQUESTED responses.
func RespondToREA(contract *Contract, rea *REA, response REAResponse) (*REAResponseResult, error) {
	if rea == nil {
		return nil, errors.New("REA must not be nil")
	}
	if contract == nil {
		return nil, errors.New("contract must not be nil")
	}

	switch response.ResponseType {
	case REAStatusApproved:
		return handleApproved(contract, rea, response)
	case REAStatusPartiallyApproved:
		return handlePartiallyApproved(contract, rea, response)
	case REAStatusDenied:
		return handleDenied(rea, response)
	case REAStatusAdditionalInfoRequested:
		return handleAdditionalInfoRequested(rea, response)
	default:
		return nil, fmt.Errorf("unsupported response type: %s", response.ResponseType)
	}
}

func handleApproved(contract *Contract, rea *REA, response REAResponse) (*REAResponseResult, error) {
	now := response.RespondedAt
	if now.IsZero() {
		now = time.Now()
	}

	// Set REA status and approved amount
	rea.Status = REAStatusApproved
	rea.ApprovedAmount = rea.RequestedAmount
	rea.ResponseRationale = response.Rationale
	rea.ResolvedAt = &now

	// Create contract modification
	mod := Modification{
		ModificationID: fmt.Sprintf("MOD-%s", rea.REAID),
		Description:    fmt.Sprintf("REA %s approved: %s", rea.REAID, response.Rationale),
		Amount:         rea.RequestedAmount,
		CreatedAt:      now,
		CreatedBy:      response.RespondedBy,
	}
	contract.Modifications = append(contract.Modifications, mod)

	// Adjust affected CLIN ceilings by approved amount / number of CLINs
	adjustCLINCeilings(contract, rea.AffectedCLINs, rea.RequestedAmount)

	// Audit entry
	audit := AuditEntry{
		Timestamp:   now,
		Actor:       response.RespondedBy,
		Action:      "REA_APPROVED",
		Description: fmt.Sprintf("REA %s approved for $%.2f on contract %s", rea.REAID, rea.RequestedAmount, contract.ContractID),
	}

	return &REAResponseResult{
		REA:          rea,
		Modification: &mod,
		AuditEntries: []AuditEntry{audit},
	}, nil
}

func handlePartiallyApproved(contract *Contract, rea *REA, response REAResponse) (*REAResponseResult, error) {
	if response.ApprovedAmount <= 0 {
		return nil, errors.New("approved amount must be positive for partial approval")
	}
	if response.ApprovedAmount >= rea.RequestedAmount {
		return nil, errors.New("partial approval amount must be less than requested amount")
	}

	now := response.RespondedAt
	if now.IsZero() {
		now = time.Now()
	}

	// Set REA status and approved amount from response
	rea.Status = REAStatusPartiallyApproved
	rea.ApprovedAmount = response.ApprovedAmount
	rea.ResponseRationale = response.Rationale
	rea.ResolvedAt = &now

	// Create contract modification for partial amount
	mod := Modification{
		ModificationID: fmt.Sprintf("MOD-%s", rea.REAID),
		Description:    fmt.Sprintf("REA %s partially approved: %s", rea.REAID, response.Rationale),
		Amount:         response.ApprovedAmount,
		CreatedAt:      now,
		CreatedBy:      response.RespondedBy,
	}
	contract.Modifications = append(contract.Modifications, mod)

	// Adjust CLINs proportionally based on partial amount
	adjustCLINCeilings(contract, rea.AffectedCLINs, response.ApprovedAmount)

	// Audit entry
	audit := AuditEntry{
		Timestamp:   now,
		Actor:       response.RespondedBy,
		Action:      "REA_PARTIALLY_APPROVED",
		Description: fmt.Sprintf("REA %s partially approved for $%.2f (requested $%.2f) on contract %s", rea.REAID, response.ApprovedAmount, rea.RequestedAmount, contract.ContractID),
	}

	return &REAResponseResult{
		REA:          rea,
		Modification: &mod,
		AuditEntries: []AuditEntry{audit},
	}, nil
}

func handleDenied(rea *REA, response REAResponse) (*REAResponseResult, error) {
	now := response.RespondedAt
	if now.IsZero() {
		now = time.Now()
	}

	// Set REA status to DENIED and record rationale
	rea.Status = REAStatusDenied
	rea.ResponseRationale = response.Rationale
	rea.ResolvedAt = &now

	// Audit entry
	audit := AuditEntry{
		Timestamp:   now,
		Actor:       response.RespondedBy,
		Action:      "REA_DENIED",
		Description: fmt.Sprintf("REA %s denied: %s", rea.REAID, response.Rationale),
	}

	return &REAResponseResult{
		REA:          rea,
		Modification: nil,
		AuditEntries: []AuditEntry{audit},
	}, nil
}

func handleAdditionalInfoRequested(rea *REA, response REAResponse) (*REAResponseResult, error) {
	now := response.RespondedAt
	if now.IsZero() {
		now = time.Now()
	}

	// Set status without setting resolvedAt
	rea.Status = REAStatusAdditionalInfoRequested
	rea.ResponseRationale = response.Rationale
	// Do NOT set resolvedAt

	// Audit entry
	audit := AuditEntry{
		Timestamp:   now,
		Actor:       response.RespondedBy,
		Action:      "REA_ADDITIONAL_INFO_REQUESTED",
		Description: fmt.Sprintf("REA %s: additional information requested - %s", rea.REAID, response.Rationale),
	}

	return &REAResponseResult{
		REA:          rea,
		Modification: nil,
		AuditEntries: []AuditEntry{audit},
	}, nil
}

// adjustCLINCeilings distributes the approved amount equally across affected CLINs.
func adjustCLINCeilings(contract *Contract, affectedCLINIDs []string, totalAmount float64) {
	if len(affectedCLINIDs) == 0 {
		return
	}

	perCLINAmount := totalAmount / float64(len(affectedCLINIDs))

	affectedSet := make(map[string]bool, len(affectedCLINIDs))
	for _, id := range affectedCLINIDs {
		affectedSet[id] = true
	}

	for i := range contract.CLINs {
		if affectedSet[contract.CLINs[i].CLINID] {
			contract.CLINs[i].Ceiling += perCLINAmount
		}
	}

	// Also adjust the contract total ceiling
	contract.TotalCeiling += totalAmount
}
