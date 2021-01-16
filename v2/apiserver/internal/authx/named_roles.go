package authx

// RoleAdmin returns a system-level Role that enables principals to manage
// Users, ServiceAccounts, and system-level permissions for Users and
// ServiceAccounts.
func RoleAdmin() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: "ADMIN",
	}
}

// RoleReader returns a system-level Role that enables global read access.
func RoleReader() Role {
	return Role{
		Type: RoleTypeSystem,
		Name: "READER",
	}
}
