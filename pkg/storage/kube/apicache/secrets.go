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

type secretStoreFactory struct{}

// return a new cached store for secrets
func (secretStoreFactory) new(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, synced chan struct{}) cache.Store {
	return newListStore(client, storeConfig{
		resource:     "secrets",
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		expectedType: &v1.Secret{},
		listFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Secrets(namespace).List(options)
		},
		watchFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Secrets(namespace).Watch(options)
		},
	}, synced)
}

// GetSecretsFilteredBy returns all secrets filtered by a label selector
// e.g. for 'heritage=brigade,component=build,project=%s'
// map[string]string{
//	"heritage":  "brigade",
//	"component": "build",
//	"project":   proj.ID,
// }
func (a *apiCache) GetSecretsFilteredBy(selectors map[string]string) []v1.Secret {

	var filteredSecrets []v1.Secret

	for _, raw := range a.secretStore.List() {

		secret, ok := raw.(*v1.Secret)
		if !ok {
			continue
		}

		// skip if the string maps don't match
		if !stringMapsMatch(secret.Labels, selectors) {
			continue
		}

		filteredSecrets = append(filteredSecrets, *secret)
	}

	return filteredSecrets
}
