package system

import libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"

// RoleAdmin returns a system-level Role that enables principals to manage
// Users, ServiceAccounts, and system-level permissions for Users and
// ServiceAccounts.
func RoleAdmin() libAuthz.Role {
	return libAuthz.Role{
		Name: "ADMIN",
	}
}

// RoleReader returns a system-level Role that enables global read access.
func RoleReader() libAuthz.Role {
	return libAuthz.Role{
		Name: "READER",
	}
}
