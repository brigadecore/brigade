package system

import libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"

const (
	// RoleAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleAdmin libAuthz.Role = "ADMIN"

	// RoleReader is the name of a system-level Role that enables global read
	// access.
	RoleReader libAuthz.Role = "READER"
)
