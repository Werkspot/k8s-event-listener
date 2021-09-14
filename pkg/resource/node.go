package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
)

func init() {
	resources = append(resources, getNode())
}

func getNode() resourceType {
	return resourceType{
		name: []string{"n", "node", "nodes"},
		fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "nodes"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.CoreV1().RESTClient()
			}
			r.ResourceType = &v1.Node{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.Node)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
