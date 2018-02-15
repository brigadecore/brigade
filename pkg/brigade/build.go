package brigade

// Build represents an invocation of an event in Brigade.
//
// Each build has a unique ID, and is tied to a project, as well as an event type.
type Build struct {
	// ID is the unique ID for a webhook event.
	ID string `json:"id"`
	// ProjectID is the computed name of the project (brigade-aeff2343a3234ff)
	ProjectID string `json:"project_id"`
	// Type is the event type (push, pull_request, tag, etc.)
	Type string `json:"type"`
	// Provider is the name of the service that caused the event (github, vsts, cron, ...)
	Provider string `json:"provider"`
	// Revision describes a vcs revision.
	Revision *Revision `json:"revision"`
	// Payload is the raw data as received by the webhook.
	Payload []byte `json:"payload,omitempty"`
	// Script is the brigadeJS to be executed.
	Script []byte `json:"script,omityempty"`
	// Worker is the master job that is running this build.
	// The Worker's properties (creation time, state, exit code, and so on)
	// reflect a "roll-up" of the job.
	// This property is not guaranteed to be set, and may be nil.
	Worker *Worker `json:"worker,omitempty"`
}

// Revision describes a vcs revision.
type Revision struct {
	// Commit is the ID of the VCS version, such as the Git commit SHA.
	Commit string `json:"commit"`
	// Ref is the symbolic ref name. (refs/heads/master, refs/pull/12/head, refs/tags/v0.1.0)
	Ref string `json:"ref"`
}
