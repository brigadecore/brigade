package js

// Event describes a Runner event.
type Event struct {
	// Type is the event type (push, pull_request, tag, etc.)
	Type string `json:"type"`
	// Provider is the name of the service that caused the event (github, vsts, cron, ...)
	Provider string `json:"provider"`
	// Commit is the ID of the VCS version, such as the Git commit SHA.
	Commit string `json:"commit"`
	// Repo describes the repository where the source code is stored.
	Payload interface{} `json:"payload"`
}
