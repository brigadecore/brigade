package system

import libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin libAuthz.RoleName = "ADMIN"

	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader libAuthz.RoleName = "READER"
)

// RoleAdmin returns a system-level Role that enables principals to manage
// Users, ServiceAccounts, and system-level permissions for Users and
// ServiceAccounts.
func RoleAdmin() libAuthz.Role {
	return libAuthz.Role{
		Type: RoleTypeSystem,
		Name: RoleNameAdmin,
	}
}

// RoleReader returns a system-level Role that enables global read access.
func RoleReader() libAuthz.Role {
	return libAuthz.Role{
		Type: RoleTypeSystem,
		Name: RoleNameReader,
	}
}
