package sdk

const (
	// RoleAdmin is the name of a system-level Role that enables principals to
	// manage Users, ServiceAccounts, and system-level permissions for Users and
	// ServiceAccounts.
	RoleAdmin Role = "ADMIN"

	// RoleEventCreator is the name of a system-level Role that enables principals
	// to create Events for all Projects-- provided the Events have a specific
	// value in the Source field. This is useful for Event gateways, which should
	// be able to create Events for all Projects, but should NOT be able to
	// impersonate other gateways.
	RoleEventCreator Role = "EVENT_CREATOR"

	// RoleProjectCreator is the name of a system-level Role that enables
	// principals to create new Projects.
	RoleProjectCreator Role = "PROJECT_CREATOR"

	// RoleReader is the name of a system-level Role that enables global read
	// access.
	RoleReader Role = "READER"
)
