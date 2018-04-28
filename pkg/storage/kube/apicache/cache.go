package apicache

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/fields"
	"time"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/davecgh/go-spew/spew"
)

type ApiCache interface {
	GetSecretsFilteredBy(labelSelectors map[string]string) []v1.Secret
	GetPodsFilteredBy(labelSelectors map[string]string) []v1.Pod
}

type apiCache struct {
	client      kubernetes.Interface
	secretStore cache.Store
	podStore    cache.Store
}

func New(kubernetesClient kubernetes.Interface, namespace string, resyncPeriod time.Duration) ApiCache {
	return &apiCache{
		client: kubernetesClient,
		secretStore: newListStore(kubernetesClient, storeConfig{
			resource:     "secrets",
			namespace:    namespace,
			resyncPeriod: resyncPeriod,
			expectedType: &v1.Secret{},
		}),
		podStore: newListStore(kubernetesClient, storeConfig{
			resource:     "pods",
			namespace:    namespace,
			resyncPeriod: resyncPeriod,
			expectedType: &v1.Pod{},
		}),
	}
}

type storeConfig struct {
	resource     string
	namespace    string
	resyncPeriod time.Duration
	expectedType runtime.Object
}

func newListStore(client kubernetes.Interface, config storeConfig) cache.Store {

	listWatch := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		config.resource,
		config.namespace,
		fields.Everything())

	store, ctr := cache.NewInformer(
		listWatch,
		config.expectedType,
		config.resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				spew.Dump("object added",obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				spew.Dump("object updated",oldObj,newObj)
			},
			DeleteFunc: func(obj interface{}) {
				spew.Dump("object deleted",obj)
			},
		})

	go ctr.Run(nil)
	return store
}
