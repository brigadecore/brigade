package controller

import (
	"log"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) createIndexerInformer() {
	selector := fields.OneTermEqualSelector("type", "brigade.sh/build")
	watcher := cache.NewListWatchFromClient(c.clientset.CoreV1().RESTClient(), "secrets", c.Namespace, selector)

	c.indexer, c.informer = cache.NewIndexerInformer(watcher, &v1.Secret{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				log.Println("Adding to workqueue: ", key)
				c.queue.Add(key)
			}
		},
	}, cache.Indexers{})

}
