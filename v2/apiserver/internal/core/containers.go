package core

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

// DeepEquals returns a boolean value indicating whether or not a ContainerSpec
// is equivalent to another.
func (c ContainerSpec) DeepEquals(c2 ContainerSpec) bool {
	if c.Image != c2.Image ||
		c.ImagePullPolicy != c2.ImagePullPolicy {
		return false
	}

	sliceEquals := func(s1 []string, s2 []string) bool {
		if (s1 == nil) != (s2 == nil) {
			return false
		}

		if len(s2) != len(s2) {
			return false
		}

		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
		}
		return true
	}

	mapEquals := func(m1 map[string]string, m2 map[string]string) bool {
		if (m1 == nil) != (m2 == nil) {
			return false
		}

		if len(m1) != len(m2) {
			return false
		}

		for k, v := range m1 {
			if w, ok := m2[k]; !ok || v != w {
				return false
			}
		}
		return true
	}

	return sliceEquals(c.Command, c2.Command) &&
		sliceEquals(c.Arguments, c2.Arguments) &&
		mapEquals(c.Environment, c2.Environment)
}
