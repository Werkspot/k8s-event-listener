package cmd

import (
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type callbackFn func(Event, interface{}) error

// Controller for a queue
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	callback callbackFn
}

// NewController returns a pointer to Controller
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, callback callbackFn) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		callback: callback,
	}
}

func (c *Controller) processNextItem() bool {
	event, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(event)

	err := c.sync(event.(Event))

	c.handleErr(err, event)
	return true
}

func (c *Controller) sync(event Event) (err error) {
	obj, exists, err := c.indexer.GetByKey(event.Key)
	if err != nil {
		return err
	}

	if !exists && event.Action != DELETE {
		return fmt.Errorf("object %s does not exists in %s event", event.Key, event.Action)
	}

	return c.callback(event, obj)
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		log.Println(fmt.Sprintf("Error syncing %v: %v", key, err))

		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)

	runtime.HandleError(err)

	log.Println(fmt.Sprintf("Dropping %q out of the queue: %v", key, err))
}

// Run start processing queue events
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	log.Println("Starting controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Println("Stopping controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}
