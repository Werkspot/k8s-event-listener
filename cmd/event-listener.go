package cmd

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	// Load oidc auth library
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

const (
	// ADD event is called when an object is added.
	ADD = "add"

	// UPDATE is called when an object is modified. Note that oldObj is the
	//      last known state of the object-- it is possible that several changes
	//      were combined together, so you can't use this to see every single
	//      change. OnUpdate is also called when a re-list happens, and it will
	//      get called even if nothing changed. This is useful for periodically
	//      evaluating or syncing something.
	UPDATE = "update"

	// DELETE will get the final state of the item if it is known, otherwise
	//      it will get an object of type DeletedFinalStateUnknown. This can
	//      happen if the watch is closed and misses the delete event and we don't
	//      notice the deletion until the subsequent re-list.
	DELETE = "delete"
)

// EventListener allow us to listen on a kubernetes cluster for events
type EventListener struct {
	kubeConfig, kubeContext string
	clientSet               *kubernetes.Clientset
	errHandler              func(error)
}

// Event holds an event info
type Event struct {
	Key, Action string
}

// NewEventListener returns a pointer to EventListener
func NewEventListener(kubeConfig, kubeContext string, errHandler func(error)) *EventListener {
	return &EventListener{
		kubeConfig:  kubeConfig,
		kubeContext: kubeContext,
		errHandler:  errHandler,
	}
}

// Init event listener
func (e *EventListener) Init() (err error) {
	utilruntime.ErrorHandlers = []func(error){
		e.errHandler,
	}

	config, err := e.getKubeConfig()
	if err != nil {
		return
	}

	e.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	return e.checkConn()
}

// Listen for incoming events from a kubernetes instance
func (e *EventListener) Listen(resource *Resource) (err error) {
	listWatcher := e.newFilteredListWatchFromClient(e.clientSet.CoreV1().RESTClient(), resource.resourceName, fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(listWatcher, resource.resourceType, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(Event{Key: key, Action: ADD})
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(Event{Key: key, Action: UPDATE})
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(Event{Key: key, Action: DELETE})
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer, resource.callback)

	stop := make(chan struct{})
	defer close(stop)

	go controller.Run(1, stop)

	select {}
}

func (e *EventListener) checkConn() (err error) {
	return
}

func (e *EventListener) getKubeConfig() (config *rest.Config, err error) {
	if e.kubeConfig == "" {
		return rest.InClusterConfig()
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: e.kubeConfig},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{}, CurrentContext: e.kubeContext},
	).ClientConfig()
}

func (e *EventListener) newFilteredListWatchFromClient(c cache.Getter, resource string, fieldSelector fields.Selector) *cache.ListWatch {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fieldSelector.String()
	}

	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		optionsModifier(&options)
		return c.Get().
			Resource(resource).
			VersionedParams(&options, metav1.ParameterCodec).
			Do().
			Get()
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		options.Watch = true
		optionsModifier(&options)
		return c.Get().
			Resource(resource).
			VersionedParams(&options, metav1.ParameterCodec).
			Watch()
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}
