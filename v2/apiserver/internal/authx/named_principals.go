package authx

var (
	// Root is a singleton that represents Brigade's "root" user.
	Root = &root{}
	// Scheduler is a singleton that represents Brigade's scheduler component.
	Scheduler = &scheduler{}
	// Observer is a singleton that represents Brigade's observer component.
	Observer = &observer{}
)

// root is an implementation of the Principal interface for the "root" user.
type root struct{}

func (r *root) Roles() []Role {
	return []Role{
		RoleAdmin(),
		RoleEventCreator(RoleScopeGlobal),
		RoleProjectAdmin(RoleScopeGlobal),
		RoleProjectCreator(),
		RoleProjectDeveloper(RoleScopeGlobal),
		RoleProjectUser(RoleScopeGlobal),
		RoleReader(),
	}
}

// scheduler is an implementation of the Principal interface that represents the
// scheduler component, which is a special class of user because, although it
// cannot do much, it has the UNIQUE ability to launch Workers and Jobs.
type scheduler struct{}

func (s *scheduler) Roles() []Role {
	return []Role{
		RoleScheduler(),
	}
}

// observer is an implementation of the Principal interface that represents the
// observer component, which is a special class of user because, although it
// cannot do much, it has the UNIQUE ability to update Worker and Job statuses.
type observer struct{}

func (o *observer) Roles() []Role {
	return []Role{
		RoleObserver(),
	}
}

// worker is an implementation of the Principal interface that represents an
// Event's Worker, which is a special class of user because, although it cannot
// do much, it has the UNIQUE ability to create new Jobs.
type worker struct {
	eventID string
}

// Worker returns a Principal that represents the specified Event's Worker.
func Worker(eventID string) Principal {
	return &worker{
		eventID: eventID,
	}
}

func (w *worker) Roles() []Role {
	return []Role{
		RoleWorker(w.eventID),
	}
}
