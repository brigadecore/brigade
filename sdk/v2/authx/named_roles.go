package authx

const (
	// RoleNameAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin RoleName = "ADMIN"

	// RoleNameReader is the name of a system-level Role that enables global read
	// access.
	RoleNameReader RoleName = "READER"
)

// RoleAdmin returns a system-level Role that enables principals to manage
// Users, ServiceAccounts, and system-level permissions for Users and
// ServiceAccounts.
func RoleAdmin() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameAdmin,
	}
}

// RoleReader returns a system-level Role that enables global read access.
func RoleReader() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: RoleNameReader,
	}
}
