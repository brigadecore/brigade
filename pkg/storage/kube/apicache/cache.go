package apicache

import (
	"errors"
	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/Azure/brigade/pkg/merge"
)

const defaultCacheSyncTimeout = 30 * time.Second

// APICache continuously queries the kubernetes api and caches secret and pod information in the current implementation
// this way dashboards can be build faster
type APICache interface {
	// get cached secrets filtered by label selectors k/v pairs
	GetSecretsFilteredBy(labelSelectors map[string]string) ([]v1.Secret, error)
	// get cached pods filtered by label selectors k/v pairs
	GetPodsFilteredBy(labelSelectors map[string]string) ([]v1.Pod, error)
}

type apiCache struct {
	// the kubernetes client
	client kubernetes.Interface
	// a ready to use cache.Store for secrets
	secretStore cache.Store
	// a ready to use cache.Store for pods
	podStore cache.Store
	// a chan which is going to be closed after the APICache has initially synced all cache.Store's
	hasSyncedInitially <-chan struct{}
}

type storeConfig struct {
	// the kubernetes resource to listen to (e.g. 'pods', 'secrets', etc.)
	resource string
	// the kubernetes namespace to filter by
	namespace string
	// how often to re-sync
	resyncPeriod time.Duration
	// which type to expect when new synced objects arrive
	expectedType runtime.Object
	// implement the method invokind the kubernetes.Interface to return
	// a List of the expected runtime.Object type
	listFunc func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error)
	// implement the method invoking the kubernetes.Interface to return
	// a watch.Interface that returns the expected runtime.Object type
	watchFunc func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error)
}

// New returns a new APICache
func New(client kubernetes.Interface, namespace string, resyncPeriod time.Duration) APICache {

	secretsSynced := make(chan struct{})
	podsSynced := make(chan struct{})

	return &apiCache{
		client:             client,
		hasSyncedInitially: merge.Channels(secretsSynced, podsSynced),
		secretStore:        newSecretStore(client, namespace, resyncPeriod, secretsSynced),
		podStore:           newPodStore(client, namespace, resyncPeriod, podsSynced),
	}
}

// blockUntilAPICacheSynced blocks until all cache.Store's are synced
// returns nil if all channels are closed
// returns error if the timeout was reached (in case timeout > 0)
func (a *apiCache) blockUntilAPICacheSynced(timeout time.Duration) error {
	if timeout < 0 {
		<-a.hasSyncedInitially
		return nil
	}

	select {
	case <-time.After(timeout):
		return errors.New("Timed out waiting for cache to sync")
	case <-a.hasSyncedInitially:
		// ok == false when all merged channels are closed
		// therefore return !ok to indicate every listener is synced if that's the case
		return nil
	}
}
