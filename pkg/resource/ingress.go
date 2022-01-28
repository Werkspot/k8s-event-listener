package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	resources = append(resources, getIngress())
}

func getIngress() resourceType {
	return resourceType{
		name: []string{"i", "ingress", "ingresses"},
		fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "ingresses"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.NetworkingV1().RESTClient()
			}
			r.ResourceType = &v1.Ingress{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.Ingress)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
