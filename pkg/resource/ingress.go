package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/api/networking/v1beta1"
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
			r.RestClient = func(clientset *kubernetes.Clientset) *rest.Request {
				return clientset.NetworkingV1beta1().RESTClient().Get().Resource(r.ResourceName)
			}
			r.ResourceType = &v1beta1.Ingress{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1beta1.Ingress)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
