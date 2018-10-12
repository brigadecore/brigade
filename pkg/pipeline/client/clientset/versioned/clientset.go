/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	pipelinev1 "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned/typed/pipeline/v1"
=======
	radixv1 "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned/typed/pipeline/v1"
>>>>>>> 0d0313d... added crd types
=======
	pipelinev1 "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned/typed/pipeline/v1"
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.
=======
>>>>>>> 6bb4934... fixed linting issues
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"

	pipelinev1 "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned/typed/pipeline/v1"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
<<<<<<< HEAD
<<<<<<< HEAD
	PipelineV1() pipelinev1.PipelineV1Interface
	// Deprecated: please explicitly pick a version if possible.
	Pipeline() pipelinev1.PipelineV1Interface
=======
	RadixV1() radixv1.RadixV1Interface
	// Deprecated: please explicitly pick a version if possible.
	Radix() radixv1.RadixV1Interface
>>>>>>> 0d0313d... added crd types
=======
	PipelineV1() pipelinev1.PipelineV1Interface
	// Deprecated: please explicitly pick a version if possible.
	Pipeline() pipelinev1.PipelineV1Interface
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
<<<<<<< HEAD
<<<<<<< HEAD
	pipelineV1 *pipelinev1.PipelineV1Client
}

// PipelineV1 retrieves the PipelineV1Client
func (c *Clientset) PipelineV1() pipelinev1.PipelineV1Interface {
	return c.pipelineV1
}

// Deprecated: Pipeline retrieves the default version of PipelineClient.
// Please explicitly pick a version.
func (c *Clientset) Pipeline() pipelinev1.PipelineV1Interface {
	return c.pipelineV1
=======
	radixV1 *radixv1.RadixV1Client
=======
	pipelineV1 *pipelinev1.PipelineV1Client
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.
}

// PipelineV1 retrieves the PipelineV1Client
func (c *Clientset) PipelineV1() pipelinev1.PipelineV1Interface {
	return c.pipelineV1
}

// Deprecated: Pipeline retrieves the default version of PipelineClient.
// Please explicitly pick a version.
<<<<<<< HEAD
func (c *Clientset) Radix() radixv1.RadixV1Interface {
	return c.radixV1
>>>>>>> 0d0313d... added crd types
=======
func (c *Clientset) Pipeline() pipelinev1.PipelineV1Interface {
	return c.pipelineV1
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
<<<<<<< HEAD
<<<<<<< HEAD
	cs.pipelineV1, err = pipelinev1.NewForConfig(&configShallowCopy)
=======
	cs.radixV1, err = radixv1.NewForConfig(&configShallowCopy)
>>>>>>> 0d0313d... added crd types
=======
	cs.pipelineV1, err = pipelinev1.NewForConfig(&configShallowCopy)
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
<<<<<<< HEAD
<<<<<<< HEAD
	cs.pipelineV1 = pipelinev1.NewForConfigOrDie(c)
=======
	cs.radixV1 = radixv1.NewForConfigOrDie(c)
>>>>>>> 0d0313d... added crd types
=======
	cs.pipelineV1 = pipelinev1.NewForConfigOrDie(c)
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
<<<<<<< HEAD
<<<<<<< HEAD
	cs.pipelineV1 = pipelinev1.New(c)
=======
	cs.radixV1 = radixv1.New(c)
>>>>>>> 0d0313d... added crd types
=======
	cs.pipelineV1 = pipelinev1.New(c)
>>>>>>> ccd1e53... started on brig pipeline functionality. more types work.

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
