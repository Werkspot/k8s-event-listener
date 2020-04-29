package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/api/certificates/v1beta1"
)

func init() {
	resources = append(resources, getCertificateSigningRequest())
}

func getCertificateSigningRequest() resourceType {
	return resourceType{
		name: []string{"csr", "certificatesigningrequest", "certificatesigningrequests"},
		fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "certificatesigningrequests"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.CertificatesV1beta1().RESTClient()
			}
			r.ResourceType = &v1beta1.CertificateSigningRequest{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1beta1.CertificateSigningRequest)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
