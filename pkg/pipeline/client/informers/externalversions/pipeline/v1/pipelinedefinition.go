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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"

	versioned "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned"
	internalinterfaces "github.com/Azure/brigade/pkg/pipeline/client/informers/externalversions/internalinterfaces"
	v1 "github.com/Azure/brigade/pkg/pipeline/client/listers/pipeline/v1"
	pipelinev1 "github.com/Azure/brigade/pkg/pipeline/v1"
)

// PipelineDefinitionInformer provides access to a shared informer and lister for
// PipelineDefinitions.
type PipelineDefinitionInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.PipelineDefinitionLister
}

type pipelineDefinitionInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPipelineDefinitionInformer constructs a new informer for PipelineDefinition type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPipelineDefinitionInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPipelineDefinitionInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPipelineDefinitionInformer constructs a new informer for PipelineDefinition type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPipelineDefinitionInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PipelineV1().PipelineDefinitions(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PipelineV1().PipelineDefinitions(namespace).Watch(options)
			},
		},
		&pipelinev1.PipelineDefinition{},
		resyncPeriod,
		indexers,
	)
}

func (f *pipelineDefinitionInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPipelineDefinitionInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *pipelineDefinitionInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&pipelinev1.PipelineDefinition{}, f.defaultInformer)
}

func (f *pipelineDefinitionInformer) Lister() v1.PipelineDefinitionLister {
	return v1.NewPipelineDefinitionLister(f.Informer().GetIndexer())
}
