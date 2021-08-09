package api

const (
	// These are reserved for use by system components and are NOT assignable to
	// Users and ServiceAccounts.

	// RoleObserver represents a system-level Role that enables principals to
	// update Worker and Job status based on observation of the underlying
	// workload execution substrate. This Role exists exclusively for use by
	// Brigade's Observer component.
	RoleObserver Role = "OBSERVER"

	// RoleScheduler represents a system-level Role that enables principals to
	// initiate execution of a Worker or Job on the underlying workload execution
	// substrate. This Role exists exclusively for use by Brigade's Scheduler
	// component.
	RoleScheduler Role = "SCHEDULER"

	// RoleWorker represents an event-level Role that enables principals to create
	// new Jobs, monitor the status of those Jobs, and access their logs. This
	// Role is exclusively for the use of Brigade Workers.
	RoleWorker Role = "WORKER"
)
