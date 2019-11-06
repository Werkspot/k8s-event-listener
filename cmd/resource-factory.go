package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-cmd/cmd"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource to be listened
type Resource struct {
	resourceName string
	resourceType runtime.Object
	callback     callbackFn
}

type callBackMeta struct {
	callback, resourceName, action, namespace, name string
}

// NewResource returns pointer to Resource and/or error
func NewResource(resourceName, callback string) (*Resource, error) {
	r := &Resource{}

	switch strings.ToLower(resourceName) {
	case "p", "pod", "pods":
		r.resourceName = "pods"
		r.resourceType = &v1.Pod{}
		r.callback = createCallbackFn(
			callback,
			r.resourceName,
			func(obj interface{}, meta *callBackMeta) {
				if obj != nil {
					objType := obj.(*v1.Pod)
					meta.namespace = objType.GetNamespace()
					meta.name = objType.GetName()
				}
			},
		)
	case "sa", "serviceaccount", "serviceaccounts":
		r.resourceName = "serviceaccounts"
		r.resourceType = &v1.ServiceAccount{}
		r.callback = createCallbackFn(
			callback,
			r.resourceName,
			func(obj interface{}, meta *callBackMeta) {
				if obj != nil {
					objType := obj.(*v1.ServiceAccount)
					meta.namespace = objType.GetNamespace()
					meta.name = objType.GetName()
				}
			},
		)
	default:
		return nil, fmt.Errorf("unknown resource %s", resourceName)
	}

	return r, nil
}

func createCallbackFn(callback, resourceName string, closure func(interface{}, *callBackMeta)) callbackFn {
	return func(event Event, i interface{}) error {
		cm := callBackMeta{
			callback:     callback,
			resourceName: resourceName,
			name:         event.Key,
			action:       event.Action,
		}

		closure(i, &cm)

		return createCallback(cm)
	}
}

func createCallback(meta callBackMeta) error {
	findCmd := cmd.NewCmd(meta.callback, meta.resourceName, meta.action, meta.namespace, meta.name)
	statusChan := findCmd.Start()
	status := <-statusChan

	log.Printf("stdout: %s\n", status.Stdout)
	log.Printf("stderr: %s\n", status.Stderr)

	return status.Error
}
