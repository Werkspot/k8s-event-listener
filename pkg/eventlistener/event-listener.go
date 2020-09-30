package eventlistener

import (
	"context"
	"flag"
	"net"
	"net/url"
	"time"

	"github.com/heptiolabs/healthcheck"

	"k8s.io/klog"

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
	kubeAPIConfig           *rest.Config
	clientSet               *kubernetes.Clientset
	errHandler              func(error)
	ctx                     context.Context
	logLevel                string
	healthHandler           healthcheck.Handler
}

// Event holds an event info
type Event struct {
	Key, Action string
}

// Resource to be listened
type Resource struct {
	ResourceName string
	ResourceType runtime.Object
	RestClient   RestClientFn
	Callback     CallbackFn
}

// CallbackFn will be invoked when a matching event will be found
type CallbackFn func(Event, interface{}) error

// RestClientFn will be used to get a REST request
type RestClientFn func(clientset *kubernetes.Clientset) rest.Interface

// NewEventListener returns a pointer to EventListener
func NewEventListener(
	ctx context.Context,
	kubeConfig, kubeContext string,
	errHandler func(error),
	logLevel string,
	healthHandler healthcheck.Handler,
) *EventListener {
	return &EventListener{
		kubeConfig:    kubeConfig,
		kubeContext:   kubeContext,
		errHandler:    errHandler,
		ctx:           ctx,
		logLevel:      logLevel,
		healthHandler: healthHandler,
	}
}

// Init event listener
func (e *EventListener) Init() (err error) {
	utilruntime.ErrorHandlers = []func(error){
		e.errHandler,
	}

	klogFlags := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(klogFlags)
	err = klogFlags.Set("v", e.logLevel)
	if err != nil {
		return
	}

	e.kubeAPIConfig, err = e.getKubeConfig()
	if err != nil {
		return
	}

	e.clientSet, err = kubernetes.NewForConfig(e.kubeAPIConfig)
	if err != nil {
		return
	}

	return e.checkConn()
}

// Listen for incoming events from a kubernetes instance
func (e *EventListener) Listen(resource *Resource) (err error) {
	listWatcher := e.newFilteredListWatchFromClient(resource.RestClient(e.clientSet), resource.ResourceName, fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(listWatcher, resource.ResourceType, 0, cache.ResourceEventHandlerFuncs{
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

	stop := make(chan struct{})
	c := NewController(queue, indexer, informer, resource.Callback, stop)

	go func() {
		defer close(stop)
		go c.Run(1)
		<-e.ctx.Done()
	}()

	return
}

func (e *EventListener) checkConn() (err error) {
	addr, err := e.getKubeAPIAddr()
	if err != nil {
		return err
	}

	check := healthcheck.TCPDialCheck(addr, 1*time.Second)
	e.healthHandler.AddLivenessCheck(
		"kubeapi",
		check,
	)

	return check()
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
			Do(context.TODO()).
			Get()
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		options.Watch = true
		optionsModifier(&options)
		return c.Get().
			Resource(resource).
			VersionedParams(&options, metav1.ParameterCodec).
			Watch(context.TODO())
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

func (e *EventListener) getKubeAPIAddr() (string, error) {
	u, err := url.Parse(e.kubeAPIConfig.Host)
	if err != nil {
		return "", err
	}

	host := u.Hostname()
	port := func() string {
		p := u.Port()
		if p != "" {
			return p
		}

		return func() string {
			switch u.Scheme {
			case "https":
				return "443"
			}

			return "80"
		}()
	}()

	return net.JoinHostPort(host, port), nil
}
