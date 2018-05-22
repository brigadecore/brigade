package apicache

import (
	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type podStoreFactory struct{}

// return a new cached store for secrets
func (podStoreFactory) new(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, synced chan struct{}) cache.Store {
	return newListStore(client, storeConfig{
		resource:     "pods",
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		expectedType: &v1.Pod{},
		listFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(namespace).List(options)
		},
		watchFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(namespace).Watch(options)
		},
	}, synced)
}

// GetPodsFilteredBy returns all pods filtered by a label selector
// e.g. for 'heritage=brigade,component=build,project=%s'
// map[string]string{
//	"heritage":  "brigade",
//	"component": "build",
//	"project":   proj.ID,
// }
func (a *apiCache) GetPodsFilteredBy(selectors map[string]string) []v1.Pod {

	var filteredSecrets []v1.Pod

	for _, raw := range a.podStore.List() {

		secret, ok := raw.(*v1.Pod)
		if !ok {
			continue
		}

		// skip if the maps don't match
		if !stringMapsMatch(secret.Labels, selectors) {
			continue
		}

		filteredSecrets = append(filteredSecrets, *secret)
	}

	return filteredSecrets
}
