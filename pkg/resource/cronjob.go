package resource

import (
	"k8s-event-listener/pkg/eventlistener"

	v1 "k8s.io/api/batch/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	resources = append(resources, getCronjob())
}

func getCronjob() resourceType {
	return resourceType{
		name: []string{"cj", "cronjob", "cronjobs"},
		fn: func(callback string) (r *eventlistener.Resource, e error) {
			r = &eventlistener.Resource{}
			r.ResourceName = "cronjobs"
			r.RestClient = func(clientset *kubernetes.Clientset) rest.Interface {
				return clientset.BatchV1().RESTClient()
			}
			r.ResourceType = &v1.CronJob{}
			r.Callback = createCallbackFn(
				callback,
				r.ResourceName,
				func(obj interface{}, meta *callBackMeta) {
					if obj != nil {
						objType := obj.(*v1.CronJob)
						meta.namespace = objType.GetNamespace()
						meta.name = objType.GetName()
					}
				},
			)

			return
		},
	}
}
