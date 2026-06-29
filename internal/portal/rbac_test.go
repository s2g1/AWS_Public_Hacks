package portal

import (
	"testing"
)

func newRBACTestContract(orgID string) *Contract {
	return &Contract{
		ContractID:     "CONTRACT-001",
		ContractNumber: "FA8721-20-C-0001",
		ContractType:   ContractTypeFFP,
		TotalCeiling:   1000000.0,
		TotalObligated: 500000.0,
		TotalExpended:  250000.0,
		EAC:            600000.0,
		Status:         ContractStatusActive,
		OrganizationID: orgID,
	}
}

func TestCheckAccess_CO_CanPerformAllActions(t *testing.T) {
	co := &PortalUser{
		UserID:         "user-co-001",
		Name:           "Jane Smith",
		Email:          "jane@gov.mil",
		Role:           PortalRoleContractingOfficer,
		OrganizationID: "ORG-GOV-001",
	}

	// CO can access any contract regardless of organization
	contract := newRBACTestContract("ORG-CONTRACTOR-999")

	actions := []Permission{
		PermissionViewContracts,
		PermissionRespondREA,
		PermissionExerciseOptions,
		PermissionManageObligations,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			err := CheckAccess(co, contract, action)
			if err != nil {
				t.Errorf("CO should be allowed to %s, got error: %v", action, err)
			}
		})
	}
}

func TestCheckAccess_Contractor_DeniedAccessToNonAssociatedContract(t *testing.T) {
	contractor := &PortalUser{
		UserID:         "user-contractor-001",
		Name:           "Bob Builder",
		Email:          "bob@contractor.com",
		Role:           PortalRoleContractor,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	// Contract belongs to a different organization
	contract := newRBACTestContract("ORG-CONTRACTOR-999")

	err := CheckAccess(contractor, contract, PermissionViewContracts)
	if err == nil {
		t.Error("contractor should be denied access to non-associated contract")
	}
}

func TestCheckAccess_Contractor_DeniedRestrictedActions(t *testing.T) {
	contractor := &PortalUser{
		UserID:         "user-contractor-001",
		Name:           "Bob Builder",
		Email:          "bob@contractor.com",
		Role:           PortalRoleContractor,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	// Contract in same org so org check passes
	contract := newRBACTestContract("ORG-CONTRACTOR-001")

	restrictedActions := []Permission{
		PermissionRespondREA,
		PermissionExerciseOptions,
		PermissionManageObligations,
	}

	for _, action := range restrictedActions {
		t.Run(string(action), func(t *testing.T) {
			err := CheckAccess(contractor, contract, action)
			if err == nil {
				t.Errorf("contractor should be denied action %s", action)
			}
		})
	}
}

func TestCheckAccess_Contractor_AllowedOwnOrgActions(t *testing.T) {
	contractor := &PortalUser{
		UserID:         "user-contractor-001",
		Name:           "Bob Builder",
		Email:          "bob@contractor.com",
		Role:           PortalRoleContractor,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	contract := newRBACTestContract("ORG-CONTRACTOR-001")

	allowedActions := []Permission{
		PermissionViewContracts,
		PermissionSubmitREA,
		PermissionUpdateEAC,
		PermissionSubmitInvoices,
	}

	for _, action := range allowedActions {
		t.Run(string(action), func(t *testing.T) {
			err := CheckAccess(contractor, contract, action)
			if err != nil {
				t.Errorf("contractor should be allowed %s on own org contract, got error: %v", action, err)
			}
		})
	}
}

func TestCheckAccess_PCO_CanViewOwnOrgAndSubmitREA(t *testing.T) {
	pco := &PortalUser{
		UserID:         "user-pco-001",
		Name:           "Alice PCO",
		Email:          "alice@contractor.com",
		Role:           PortalRoleProcuringContractingOfficer,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	contract := newRBACTestContract("ORG-CONTRACTOR-001")

	// PCO can view contracts in own org
	err := CheckAccess(pco, contract, PermissionViewContracts)
	if err != nil {
		t.Errorf("PCO should be allowed to view own org contracts, got error: %v", err)
	}

	// PCO can submit REA
	err = CheckAccess(pco, contract, PermissionSubmitREA)
	if err != nil {
		t.Errorf("PCO should be allowed to submit REA, got error: %v", err)
	}

	// PCO can update EAC
	err = CheckAccess(pco, contract, PermissionUpdateEAC)
	if err != nil {
		t.Errorf("PCO should be allowed to update EAC, got error: %v", err)
	}

	// PCO can submit invoices
	err = CheckAccess(pco, contract, PermissionSubmitInvoices)
	if err != nil {
		t.Errorf("PCO should be allowed to submit invoices, got error: %v", err)
	}
}

func TestCheckAccess_PCO_DeniedAccessToOtherOrgContract(t *testing.T) {
	pco := &PortalUser{
		UserID:         "user-pco-001",
		Name:           "Alice PCO",
		Email:          "alice@contractor.com",
		Role:           PortalRoleProcuringContractingOfficer,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	// Contract in different org
	contract := newRBACTestContract("ORG-OTHER-999")

	err := CheckAccess(pco, contract, PermissionViewContracts)
	if err == nil {
		t.Error("PCO should be denied access to contract in different organization")
	}
}

func TestCheckAccess_PCO_DeniedCOActions(t *testing.T) {
	pco := &PortalUser{
		UserID:         "user-pco-001",
		Name:           "Alice PCO",
		Email:          "alice@contractor.com",
		Role:           PortalRoleProcuringContractingOfficer,
		OrganizationID: "ORG-CONTRACTOR-001",
	}

	contract := newRBACTestContract("ORG-CONTRACTOR-001")

	// PCO cannot respond to REA (that's a CO action)
	err := CheckAccess(pco, contract, PermissionRespondREA)
	if err == nil {
		t.Error("PCO should be denied RESPOND_REA permission")
	}

	// PCO cannot exercise options
	err = CheckAccess(pco, contract, PermissionExerciseOptions)
	if err == nil {
		t.Error("PCO should be denied EXERCISE_OPTIONS permission")
	}

	// PCO cannot manage obligations
	err = CheckAccess(pco, contract, PermissionManageObligations)
	if err == nil {
		t.Error("PCO should be denied MANAGE_OBLIGATIONS permission")
	}
}

func TestCheckAccess_NilUser(t *testing.T) {
	contract := newRBACTestContract("ORG-001")
	err := CheckAccess(nil, contract, PermissionViewContracts)
	if err == nil {
		t.Error("should return error for nil user")
	}
}

func TestCheckAccess_NilContract(t *testing.T) {
	user := &PortalUser{
		UserID:         "user-001",
		Name:           "Test User",
		Email:          "test@gov.mil",
		Role:           PortalRoleContractingOfficer,
		OrganizationID: "ORG-001",
	}
	err := CheckAccess(user, nil, PermissionViewContracts)
	if err == nil {
		t.Error("should return error for nil contract")
	}
}
