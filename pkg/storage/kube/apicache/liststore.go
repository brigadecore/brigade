package apicache

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// convenience method to create listStores
func newListStore(client kubernetes.Interface, config storeConfig, hasSynced chan struct{}) cache.Store {

	listWatch := cache.ListWatch{
		ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
			return config.listFunc(client, config.namespace, options)
		},
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return config.watchFunc(client, config.namespace, options)
		},
	}

	store, ctr := cache.NewInformer(
		&listWatch,
		config.expectedType,
		config.resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		})

	// run the controller in a new goroutine, else this operation would block
	// we currently don't supply a close chan as there is no need to
	// this might change
	go ctr.Run(nil)

	// loop until the store is in sync, then close the chan to signal
	go func(ctr cache.Controller) {
		for {
			if hasSynced == nil {
				break
			}

			if ctr.HasSynced() {
				// if this panics you might unintentionally closed this channel from outside
				// please fix your code as this behavior is not not expected
				close(hasSynced)
				hasSynced = nil
				break
			}
		}
	}(ctr)

	return store
}

// returns true under these conditions:
// 1. all keys from the expected map exist on the actual map
// 2. all values from the expected map match those from the actual map
// else false
func stringMapsMatch(actual, expected map[string]string) bool {
	for key, expected := range expected {
		actual, exists := actual[key]
		if !exists || actual != expected {
			return false
		}
	}

	return true
}
