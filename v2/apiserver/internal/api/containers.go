package api

// ImagePullPolicy represents a policy for whether container hosts already
// having a certain OCI image should attempt to re-pull that image prior to
// launching a new container based on that image.
type ImagePullPolicy string

// ContainerSpec represents the technical details of an OCI container.
type ContainerSpec struct {
	// Image specifies the OCI image on which the container should be based.
	Image string `json:"image,omitempty" bson:"image,omitempty"`
	// ImagePullPolicy specifies whether a container host already having the
	// specified OCI image should attempt to re-pull that image prior to launching
	// a new container.
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy,omitempty" bson:"imagePullPolicy,omitempty"` // nolint: lll
	// Command specifies the command to be executed by the OCI container. This
	// can be used to optionally override the default command specified by the OCI
	// image itself.
	Command []string `json:"command,omitempty" bson:"command,omitempty"`
	// Arguments specifies arguments to the command executed by the OCI container.
	Arguments []string `json:"arguments,omitempty" bson:"arguments,omitempty"`
	// Environment is a map of key/value pairs that specify environment variables
	// to be set within the OCI container.
	Environment map[string]string `json:"environment,omitempty" bson:"environment,omitempty"` // nolint: lll
}

func (cs ContainerSpec) EqualTo(cs2 ContainerSpec) bool {
	// Compare Command slices
	if len(cs.Command) != len(cs2.Command) {
		return false
	}
	for i, command := range cs.Command {
		if command != cs2.Command[i] {
			return false
		}
	}

	// Compare Arguments slices
	if len(cs.Arguments) != len(cs2.Arguments) {
		return false
	}
	for i, argument := range cs.Arguments {
		if argument != cs2.Arguments[i] {
			return false
		}
	}

	// Compare Environment maps
	if len(cs.Environment) != len(cs2.Environment) {
		return false
	}
	for name, value := range cs.Environment {
		if value != cs2.Environment[name] {
			return false
		}
	}

	// Return remaining field equivalence
	return cs.Image == cs2.Image &&
		cs.ImagePullPolicy == cs2.ImagePullPolicy
}
