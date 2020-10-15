package core

// LogLevel represents the desired granularity of Worker log output.
type LogLevel string

// WorkerSpec is the technical blueprint for a Worker.
type WorkerSpec struct {
	// Container specifies the details of an OCI container that forms the
	// cornerstone of the Worker.
	Container *ContainerSpec `json:"container,omitempty" bson:"container,omitempty"` // nolint: lll
	// UseWorkspace indicates whether the Worker and/or any Jobs it may spawn
	// requires access to a shared workspace. When false, no such workspace is
	// provisioned prior to Worker creation. This is a generally useful feature,
	// but by opting out of it (or rather, not opting-in), Job results can be made
	// cacheable and Jobs resumable/retriable-- something which cannot be done
	// otherwise since managing the state of the shared volume would require a
	// layered file system that we currently do not have.
	UseWorkspace bool `json:"useWorkspace" bson:"useWorkspace"`
	// WorkspaceSize specifies the size of a volume that will be provisioned as
	// a shared workspace for the Worker and any Jobs it spawns.
	// The value can be expressed in bytes (as a plain integer) or as a
	// fixed-point integer using one of these suffixes: E, P, T, G, M, K.
	// Power-of-two equivalents may also be used: Ei, Pi, Ti, Gi, Mi, Ki.
	WorkspaceSize string `json:"workspaceSize,omitempty" bson:"workspaceSize,omitempty"` // nolint: lll
	// Git contains git-specific Worker details.
	Git *GitConfig `json:"git,omitempty"`
	// Kubernetes contains Kubernetes-specific Worker details.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
	// JobPolicies specifies policies for any Jobs spawned by the Worker.
	JobPolicies *JobPolicies `json:"jobPolicies,omitempty" bson:"jobPolicies,omitempty"` // nolint: lll
	// LogLevel specifies the desired granularity of Worker log output.
	LogLevel LogLevel `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
	// ConfigFilesDirectory specifies a directory within the Worker's workspace
	// where any relevant configuration files (e.g. brigade.json, brigade.js,
	// etc.) can be located.
	ConfigFilesDirectory string `json:"configFilesDirectory,omitempty" bson:"configFilesDirectory,omitempty"` // nolint: lll
	// DefaultConfigFiles is a map of configuration file names to configuration
	// file content. This is useful for Workers that do not integrate with any
	// source control system and would like to embed configuration (e.g.
	// brigade.json) or scripts (e.g. brigade.js) directly within the WorkerSpec.
	DefaultConfigFiles map[string]string `json:"defaultConfigFiles,omitempty" bson:"defaultConfigFiles,omitempty"` // nolint: lll
}

// GitConfig represents git-specific Worker details.
type GitConfig struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Commit specifies a commit (by SHA) to be checked out.
	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
	Ref string `json:"ref,omitempty" bson:"ref,omitempty"`
	// InitSubmodules indicates whether to clone the repository's submodules.
	InitSubmodules bool `json:"initSubmodules" bson:"initSubmodules"`
}

// KubernetesConfig represents Kubernetes-specific Worker or Job configuration.
type KubernetesConfig struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which a Worker's or Job's container is based.
	// This field only needs to be utilized in the case of private, custom Worker
	// or Job images. The image pull secrets in question must be created
	// out-of-band by a sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"` // nolint: lll
}

// JobPolicies represents policies for any Jobs spawned by a Worker.
type JobPolicies struct {
	// AllowPrivileged specifies whether the Worker is permitted to launch Jobs
	// that utilize privileged containers.
	AllowPrivileged bool `json:"allowPrivileged" bson:"allowPrivileged"`
	// AllowDockerSocketMount specifies whether the Worker is permitted to launch
	// Jobs that mount the underlying host's Docker socket into its own file
	// system.
	AllowDockerSocketMount bool `json:"allowDockerSocketMount" bson:"allowDockerSocketMount"` // nolint: lll
}
