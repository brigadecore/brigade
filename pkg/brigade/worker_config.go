package brigade

import "fmt"

// WorkerConfig overrides what is specified under the `worker` key in brigade-wide config
type WorkerConfig struct {
	// Registry is the composition of:
	// - the docker registry hostname(e.g. quay.io for quay and none for dockerhub)
	// - docker repository(username or organizaton name)
	// parts of a docker image reference
	Registry string `json:"registry"`
	// Name is the name of a docker image. For example, `nginx` is the name of `nginx:latest`
	Name string `json:"name"`
	// Tag is the tag of a docker image. For example, `latest` is the tag of `nginx:latest`
	Tag string `json:"tag"`
	// PullPolicy specifies when you want to pull the docker image for brigade-worker
	PullPolicy string `json:"pullPolicy"`
}

func (c WorkerConfig) Image() string {
	image := c.Name
	if c.Registry != "" {
		image = fmt.Sprintf("%s/%s", c.Registry, image)
	}
	tag := c.Tag
	return fmt.Sprintf("%s:%s", image, tag)
}
