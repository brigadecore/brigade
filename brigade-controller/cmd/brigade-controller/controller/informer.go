package controller

import (
	"log"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) createIndexerInformer() {
	selector := "type=brigade.sh/build"
	c.indexer, c.informer = cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = selector
				return c.clientset.CoreV1().Secrets(c.Namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = selector
				return c.clientset.CoreV1().Secrets(c.Namespace).Watch(options)
			},
		},
		&v1.Secret{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
					log.Println("Adding to workqueue: ", key)
					c.queue.Add(key)
				}
			},
		},
		cache.Indexers{},
	)
}
