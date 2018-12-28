package apicache

import (
	"sort"
	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// return a new cached store for secrets
func newSecretStore(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, synced chan struct{}) cache.Store {
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
func (a *apiCache) GetSecretsFilteredBy(selectors map[string]string) ([]v1.Secret, error) {
	var filteredSecrets []v1.Secret

	if err := a.blockUntilAPICacheSynced(defaultCacheSyncTimeout); err != nil {
		return filteredSecrets, err
	}

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
	sort.Sort(ByCreation(filteredSecrets))
	return filteredSecrets, nil
}

// ByCreation sorts secrets by their creation timestamp.
type ByCreation []v1.Secret

// Len returns the length of the secrets slice.
func (b ByCreation) Len() int {
	return len(b)
}

// Swap swaps the position of two indices.
func (b ByCreation) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less tests that i is less than j.
func (b ByCreation) Less(i, j int) bool {
	jj := b[j].ObjectMeta.CreationTimestamp.Time
	ii := b[i].ObjectMeta.CreationTimestamp.Time
	return ii.After(jj)
}
