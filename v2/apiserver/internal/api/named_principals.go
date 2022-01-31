package api

// RootPrincipal is an implementation of the Principal interface for the "root"
// user.
type RootPrincipal struct{}

func (r *RootPrincipal) RoleAssignments() []RoleAssignment {
	return []RoleAssignment{
		{Role: RoleAdmin},
		{Role: RoleReader},
		{
			Role:  RoleEventCreator,
			Scope: RoleScopeGlobal,
		},
		{Role: RoleProjectCreator},
	}
}

func (r *RootPrincipal) ProjectRoleAssignments() []ProjectRoleAssignment {
	return []ProjectRoleAssignment{
		{
			ProjectID: ProjectRoleScopeGlobal,
			Role:      RoleProjectAdmin,
		},
		{
			ProjectID: ProjectRoleScopeGlobal,
			Role:      RoleProjectDeveloper,
		},
		{
			ProjectID: ProjectRoleScopeGlobal,
			Role:      RoleProjectUser,
		},
	}
}

// SchedulerPrincipal is an implementation of the Principal interface that
// represents the scheduler component, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to launch Workers and
// Jobs.
type SchedulerPrincipal struct{}

func (s *SchedulerPrincipal) RoleAssignments() []RoleAssignment {
	return []RoleAssignment{
		{Role: RoleReader},
		{Role: RoleScheduler},
	}
}

// ObserverPrincipal is an implementation of the Principal interface that
// represents the observer component, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to update Worker and
// Job statuses.
type ObserverPrincipal struct{}

func (o *ObserverPrincipal) RoleAssignments() []RoleAssignment {
	return []RoleAssignment{
		{Role: RoleReader},
		{Role: RoleObserver},
	}
}

// WorkerPrincipal is an implementation of the Principal interface that
// represents an Event's Worker, which is a special class of user because,
// although it cannot do much, it has the UNIQUE ability to create new Jobs.
type WorkerPrincipal struct {
	eventID string
}

func (w *WorkerPrincipal) RoleAssignments() []RoleAssignment {
	return []RoleAssignment{
		{Role: RoleReader},
		{
			Role:  RoleWorker,
			Scope: w.eventID,
		},
	}
}

// GetWorkerPrincipal returns an Principal that represents the specified Event's
// Worker.
func GetWorkerPrincipal(eventID string) *WorkerPrincipal {
	return &WorkerPrincipal{
		eventID: eventID,
	}
}
