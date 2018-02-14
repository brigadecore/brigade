package controller

import (
	"fmt"
	"log"
	"time"

	"k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Config is config for setting Controller
type Config struct {
	Namespace            string
	WorkerImage          string
	WorkerPullPolicy     string
	WorkerServiceAccount string
}

// Controller listens for new brigade builds and starts the worker pods.
type Controller struct {
	*Config
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	clientset kubernetes.Interface
}

// NewController creates a new Controller.
func NewController(clientset kubernetes.Interface, config *Config) *Controller {
	c := &Controller{
		clientset: clientset,
		Config:    config,
		queue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	c.createIndexerInformer()
	return c
}

// getSecret gets the Secret we are interested in
func (c *Controller) getSecret(key string) (*v1.Secret, bool, error) {
	obj, exists, err := c.indexer.GetByKey(key)

	if exists {
		return obj.(*v1.Secret), exists, err
	}
	return nil, exists, err
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two secrets with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.sync(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// HasSynced returns true if the controller has synced.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// sync is the business logic of the controller.
// The retry logic should not be part of the business logic.
func (c *Controller) sync(key string) error {
	secret, exists, err := c.getSecret(key)
	if err != nil {
		log.Printf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		fmt.Printf("Secret %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Secret was recreated with the same name
		log.Printf("Executing on Secret: %s\n", secret.GetName())
		return c.syncSecret(secret)
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Printf("Error syncing secret %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	utilruntime.HandleError(err)
	log.Printf("Dropping secret %q out of the queue: %v", key, err)
}

// Run executes the controller.
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Print("Starting Secret controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Print("Stopping Secret controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}
