package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
)

func init() {
	resources = append(resources, getServiceAccount())
}

func getServiceAccount() resourceType {
	return resourceType{
		name: []string{"sa", "serviceaccount", "serviceaccounts"},
		fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "serviceaccounts"
			r.RestClient = func(clientset *kubernetes.Clientset) *rest.Request {
				return clientset.CoreV1().RESTClient().Get().Resource(r.ResourceName)
			}
			r.ResourceType = &v1.ServiceAccount{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.ServiceAccount)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
