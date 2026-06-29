package portal

import "fmt"

// CheckAccess verifies that a user has permission to perform a given action on a contract.
// It enforces role-based access control:
//   - CO (Contracting Officer): can view all contracts in portfolio, respond to REA, exercise options, manage obligations
//   - PCO (Procuring Contracting Officer) / Contractor: can only view contracts in their organization,
//     submit REAs, update EAC, submit invoices
//   - Deny contractor access to non-associated contracts (user.OrganizationID != contract.OrganizationID)
//   - Deny actions outside role permissions
//
// Returns nil if access is allowed, or a descriptive error if denied.
func CheckAccess(user *PortalUser, contract *Contract, action Permission) error {
	if user == nil {
		return fmt.Errorf("access denied: user is nil")
	}
	if contract == nil {
		return fmt.Errorf("access denied: contract is nil")
	}

	// Step 1: Check organization-based access for non-CO roles.
	// CO can view all contracts in their portfolio regardless of organization.
	// All other roles (PCO, Contractor, COR, PM) can only access contracts
	// associated with their organization.
	if user.Role != PortalRoleContractingOfficer {
		if user.OrganizationID != contract.OrganizationID {
			return fmt.Errorf(
				"access denied: user %s (role %s, org %s) cannot access contract %s (org %s) - not associated with user's organization",
				user.UserID, user.Role, user.OrganizationID, contract.ContractID, contract.OrganizationID,
			)
		}
	}

	// Step 2: Check if the user's role grants the requested permission.
	if !user.HasPermission(action) {
		return fmt.Errorf(
			"access denied: user %s (role %s) does not have permission %s",
			user.UserID, user.Role, action,
		)
	}

	return nil
}
