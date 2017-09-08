package api

import (
	"github.com/deis/acid/pkg/storage"
)

// API represents the rest api handlers.
type API struct {
	store storage.Store
}

// New creates a new api handler.
func New(s storage.Store) API {
	return API{store: s}
}

// Project returns a handler for projects.
func (api API) Project() Project {
	return Project{store: api.store}
}

// Build returns a handler for builds.
func (api API) Build() Build {
	return Build{store: api.store}
}

// Job returns a handler for jobs.
func (api API) Job() Job {
	return Job{store: api.store}
}
