package portal

// PortalRole represents a user role within the Contract Financial Management Portal.
type PortalRole string

const (
	PortalRoleContractingOfficer          PortalRole = "CONTRACTING_OFFICER"
	PortalRoleCOR                         PortalRole = "COR"
	PortalRoleProcuringContractingOfficer PortalRole = "PROCURING_CONTRACTING_OFFICER"
	PortalRoleProgramManager              PortalRole = "PROGRAM_MANAGER"
	PortalRoleContractor                  PortalRole = "CONTRACTOR"
)

// Permission represents a specific action a user can perform in the portal.
type Permission string

const (
	PermissionViewContracts     Permission = "VIEW_CONTRACTS"
	PermissionSubmitREA         Permission = "SUBMIT_REA"
	PermissionRespondREA        Permission = "RESPOND_REA"
	PermissionExerciseOptions   Permission = "EXERCISE_OPTIONS"
	PermissionManageObligations Permission = "MANAGE_OBLIGATIONS"
	PermissionUpdateEAC         Permission = "UPDATE_EAC"
	PermissionSubmitInvoices    Permission = "SUBMIT_INVOICES"
)

// RolePermissions defines the permission sets for each portal role.
var RolePermissions = map[PortalRole][]Permission{
	PortalRoleContractingOfficer: {
		PermissionViewContracts,
		PermissionRespondREA,
		PermissionExerciseOptions,
		PermissionManageObligations,
	},
	PortalRoleCOR: {
		PermissionViewContracts,
	},
	PortalRoleProcuringContractingOfficer: {
		PermissionViewContracts,
		PermissionSubmitREA,
		PermissionUpdateEAC,
		PermissionSubmitInvoices,
	},
	PortalRoleProgramManager: {
		PermissionViewContracts,
		PermissionSubmitREA,
		PermissionUpdateEAC,
	},
	PortalRoleContractor: {
		PermissionViewContracts,
		PermissionSubmitREA,
		PermissionUpdateEAC,
		PermissionSubmitInvoices,
	},
}

// PortalUser represents a user of the Contract Financial Management Portal.
type PortalUser struct {
	UserID         string     `json:"userId"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Role           PortalRole `json:"role"`
	OrganizationID string     `json:"organizationId"`
}

// HasPermission checks whether the user's role grants the specified permission.
func (u *PortalUser) HasPermission(perm Permission) bool {
	permissions, ok := RolePermissions[u.Role]
	if !ok {
		return false
	}
	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}
