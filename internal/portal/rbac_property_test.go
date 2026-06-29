package portal

import (
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 20.3, 20.4**
// Property 23: Role-Based Access Enforcement
// 1. CO role always has access regardless of organization (given valid permission)
// 2. Contractor/PCO with different OrganizationID than contract always gets denied
// 3. Any user attempting an action NOT in their role's permission set always gets denied
// 4. Any user attempting an action IN their role's permission set AND matching org always succeeds

// allRoles is the complete set of portal roles for generator use.
var allRoles = []PortalRole{
	PortalRoleContractingOfficer,
	PortalRoleCOR,
	PortalRoleProcuringContractingOfficer,
	PortalRoleProgramManager,
	PortalRoleContractor,
}

// allPermissions is the complete set of permissions for generator use.
var allPermissions = []Permission{
	PermissionViewContracts,
	PermissionSubmitREA,
	PermissionRespondREA,
	PermissionExerciseOptions,
	PermissionManageObligations,
	PermissionUpdateEAC,
	PermissionSubmitInvoices,
}

// nonCORoles are roles that are NOT Contracting Officer.
var nonCORoles = []PortalRole{
	PortalRoleCOR,
	PortalRoleProcuringContractingOfficer,
	PortalRoleProgramManager,
	PortalRoleContractor,
}

// permissionsNotInRole returns the set of permissions NOT granted to the given role.
func permissionsNotInRole(role PortalRole) []Permission {
	granted := make(map[Permission]bool)
	for _, p := range RolePermissions[role] {
		granted[p] = true
	}
	var missing []Permission
	for _, p := range allPermissions {
		if !granted[p] {
			missing = append(missing, p)
		}
	}
	return missing
}

func TestProperty23_COAlwaysHasAccessRegardlessOfOrganization(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a CO user with a random org
		userOrg := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "userOrg")
		contractOrg := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "contractOrg")

		// Pick a permission that the CO role has
		coPerms := RolePermissions[PortalRoleContractingOfficer]
		permIdx := rapid.IntRange(0, len(coPerms)-1).Draw(t, "permIdx")
		action := coPerms[permIdx]

		user := &PortalUser{
			UserID:         "user-co-test",
			Name:           "Test CO",
			Email:          "co@test.gov",
			Role:           PortalRoleContractingOfficer,
			OrganizationID: userOrg,
		}

		contract := &Contract{
			ContractID:     "CONTRACT-TEST",
			OrganizationID: contractOrg,
		}

		err := CheckAccess(user, contract, action)

		// Property: CO always has access regardless of organization mismatch
		if err != nil {
			t.Fatalf("CO should always have access regardless of org (userOrg=%s, contractOrg=%s, action=%s): got error %v",
				userOrg, contractOrg, action, err)
		}
	})
}

func TestProperty23_ContractorPCOWithDifferentOrgAlwaysDenied(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Pick a non-CO role (Contractor, PCO, COR, PM)
		roleIdx := rapid.IntRange(0, len(nonCORoles)-1).Draw(t, "roleIdx")
		role := nonCORoles[roleIdx]

		// Generate two DIFFERENT org IDs
		userOrg := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "userOrg")
		contractOrg := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "contractOrg")

		// Ensure they're actually different
		if userOrg == contractOrg {
			contractOrg = contractOrg + "-DIFF"
		}

		// Pick any permission
		permIdx := rapid.IntRange(0, len(allPermissions)-1).Draw(t, "permIdx")
		action := allPermissions[permIdx]

		user := &PortalUser{
			UserID:         "user-nonco-test",
			Name:           "Test Non-CO",
			Email:          "nonco@test.gov",
			Role:           role,
			OrganizationID: userOrg,
		}

		contract := &Contract{
			ContractID:     "CONTRACT-TEST",
			OrganizationID: contractOrg,
		}

		err := CheckAccess(user, contract, action)

		// Property: non-CO user with different org than contract always gets denied
		if err == nil {
			t.Fatalf("non-CO user (role=%s, org=%s) accessing contract (org=%s) should be denied, but got nil error",
				role, userOrg, contractOrg)
		}
	})
}

func TestProperty23_ActionNotInRolePermissionsAlwaysDenied(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Pick any role
		roleIdx := rapid.IntRange(0, len(allRoles)-1).Draw(t, "roleIdx")
		role := allRoles[roleIdx]

		// Get permissions NOT in this role
		missingPerms := permissionsNotInRole(role)
		if len(missingPerms) == 0 {
			// This role has all permissions, skip
			return
		}

		permIdx := rapid.IntRange(0, len(missingPerms)-1).Draw(t, "permIdx")
		action := missingPerms[permIdx]

		// Use matching org so org-check passes
		org := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "org")

		user := &PortalUser{
			UserID:         "user-perm-test",
			Name:           "Test User",
			Email:          "user@test.gov",
			Role:           role,
			OrganizationID: org,
		}

		contract := &Contract{
			ContractID:     "CONTRACT-TEST",
			OrganizationID: org, // Same org so only permission check matters
		}

		err := CheckAccess(user, contract, action)

		// Property: action not in role's permission set always gets denied
		if err == nil {
			t.Fatalf("user (role=%s) should be denied action=%s which is not in their permission set, but got nil error",
				role, action)
		}
	})
}

func TestProperty23_ActionInPermissionsAndMatchingOrgAlwaysSucceeds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Pick any role
		roleIdx := rapid.IntRange(0, len(allRoles)-1).Draw(t, "roleIdx")
		role := allRoles[roleIdx]

		// Pick a permission that IS in this role's set
		rolePerms := RolePermissions[role]
		if len(rolePerms) == 0 {
			return
		}
		permIdx := rapid.IntRange(0, len(rolePerms)-1).Draw(t, "permIdx")
		action := rolePerms[permIdx]

		// Use matching org
		org := rapid.StringMatching(`[A-Z]{3}-[0-9]{4}`).Draw(t, "org")

		user := &PortalUser{
			UserID:         "user-success-test",
			Name:           "Test User",
			Email:          "user@test.gov",
			Role:           role,
			OrganizationID: org,
		}

		contract := &Contract{
			ContractID:     "CONTRACT-TEST",
			OrganizationID: org, // Matching org
		}

		err := CheckAccess(user, contract, action)

		// Property: action in permission set + matching org always succeeds
		if err != nil {
			t.Fatalf("user (role=%s, org=%s) with action=%s in permission set and matching org should succeed, but got error: %v",
				role, org, action, err)
		}
	})
}
