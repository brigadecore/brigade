package system

import libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"

const (
	// RoleAdmin represents a system-level Role that enables principals to manage
	// Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleAdmin libAuthz.Role = "ADMIN"

	// RoleReader represents a system-level Role that enables global read access.
	RoleReader libAuthz.Role = "READER"
)
