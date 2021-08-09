package api

const (
	// RoleAdmin represents a system-level Role that enables principals to manage
	// Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleAdmin Role = "ADMIN"

	// RoleEventCreator represents a system-level Role that enables principals to
	// create Events for all Projects.
	RoleEventCreator Role = "EVENT_CREATOR"

	// RoleProjectCreator represents a system-level Role that enables principals
	// to create new Projects.
	RoleProjectCreator Role = "PROJECT_CREATOR"

	// RoleReader represents a system-level Role that enables global read access.
	RoleReader Role = "READER"
)
