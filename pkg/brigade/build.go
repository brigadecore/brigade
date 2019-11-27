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
	// ShortTitle is an optional field for a short (and not necessarily unique)
	// string value that can be added to a build by a gateway to ascribe context
	// that may be meaningful to human users. For instance, the GitHub gateway
	// COULD label a build triggered by a pull request with the title or number of
	// that pull request.
	ShortTitle string `json:"short_title"`
	// LongTitle is an optional field for a longer (and not necessarily unique)
	// string value that can be added to a build by a gateway to ascribe context
	// that may be meaningful to human users. For instance, the GitHub gateway
	// COULD label a build triggered by a pull request with the title or number of
	// that pull request.
	LongTitle string `json:"long_title"`
	// CloneURL is the URL at which the repository can be cloned.
	// This is optional at the build-level. If set, it overrides the same setting
	// at the projet-level.
	CloneURL string `json:"clone_url"`
	// Revision describes a vcs revision.
	Revision *Revision `json:"revision"`
	// Payload is the raw data as received by the webhook.
	Payload []byte `json:"payload"`
	// Script is the brigadeJS to be executed.
	Script []byte `json:"script"`
	// Config is a JSON file representing Brigade configuration,
	// including JS dependencies and other information
	Config []byte `json:"config"`
	// Worker is the master job that is running this build.
	// The Worker's properties (creation time, state, exit code, and so on)
	// reflect a "roll-up" of the job.
	// This property is not guaranteed to be set, and may be nil.
	Worker *Worker `json:"worker"`
	// LogLevel determines what level of logging from the Javascript
	// to print to console.
	LogLevel string `json:"log_level,omitempty"`
}

// Revision describes a vcs revision.
type Revision struct {
	// Commit is the ID of the VCS version, such as the Git commit SHA.
	Commit string `json:"commit"`
	// Ref is the symbolic ref name. (refs/heads/master, refs/pull/12/head, refs/tags/v0.1.0)
	Ref string `json:"ref"`
}
