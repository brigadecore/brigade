package core

import "context"

// Substrate is an interface for components that permit services to coordinate
// with Brigade's underlying workload execution substrate, i.e. Kubernetes.
type Substrate interface {
	// CreateProject prepares the substrate to host Project workloads. The
	// provided Project argument may be amended with substrate-specific details
	// and returned, so this function be called prior to a Project being initially
	// persisted so that substrate-specific details will be included.
	CreateProject(context.Context, Project) (Project, error)
	// DeleteProject removes all Project-related resources from the substrate.
	DeleteProject(context.Context, Project) error
}
