package apicache

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"time"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"github.com/Azure/brigade/pkg/merge"
)

type ApiCache interface {
	// get cached secrets filtered by label selectors k/v pairs
	GetSecretsFilteredBy(labelSelectors map[string]string) []v1.Secret
	// get cached pods filtered by label selectors k/v pairs
	GetPodsFilteredBy(labelSelectors map[string]string) []v1.Pod
	// blocks until the api cache is populated
	// returns true when cache is synced or false when timeout reached
	BlockUntilApiCacheSynced(waitUntil <-chan time.Time) bool
}

type apiCache struct {
	// the kubernetes client
	client             kubernetes.Interface
	// a ready to use cache.Store for secrets
	secretStore        cache.Store
	// a ready to use cache.Store for pods
	podStore           cache.Store
	// a chan which is going to be closed after the ApiCache has initially synced all cache.Store's
	hasSyncedInitially <-chan struct{}
}

type storeConfig struct {
	// the kubernetes resource to listen to (e.g. 'pods', 'secrets', etc.)
	resource     string
	// the kubernetes namespace to filter by
	namespace    string
	// how often to re-sync
	resyncPeriod time.Duration
	// which type to expect when new synced objects arrive
	expectedType runtime.Object
	// implement the method invokind the kubernetes.Interface to return
	// a List of the expected runtime.Object type
	listFunc     func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error)
	// implement the method invoking the kubernetes.Interface to return
	// a watch.Interface that returns the expected runtime.Object type
	watchFunc    func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error)
}

func New(client kubernetes.Interface, namespace string, resyncPeriod time.Duration) ApiCache {

	secretsSynced := make(chan struct{})
	podsSynced := make(chan struct{})

	merged := merge.Channels(secretsSynced, podsSynced)

	return &apiCache{
		hasSyncedInitially: merged,
		client:             client,
		secretStore:        secretStoreFactory{}.new(client, namespace, resyncPeriod, secretsSynced),
		podStore:           podStoreFactory{}.new(client, namespace, resyncPeriod, podsSynced),
	}
}

// Blocks until all cache.Store's are synced
// returns true if all channels are closed
// returns false if the timeout was reached (in case waitUntil is not nil)
func (a *apiCache) BlockUntilApiCacheSynced(waitUntil <-chan time.Time) bool {
	if waitUntil == nil {
		<-a.hasSyncedInitially
		return true
	} else {
		select {
		case <-waitUntil:
			return false
		case _, ok := <-a.hasSyncedInitially:
			// ok == false when all merged channels are closed
			// therefore return !ok to indicate every listener is synced if that's the case
			return !ok
		}
	}
}