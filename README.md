# K8s Event listener
Listen for changes on specified resources and invokes the callback script

Currently supporting:
- pods
- serviceaccounts
- ingresses

## Usage

```
$ k8s-event-listener --help
Listen for specific kubernetes events

Usage:
  k8s-event-listener [flags]

Flags:
  -c, --callback string       Callback to be executed
  -h, --help                  help for k8s-event-listener
      --kube-config string    Path to kubeconfig file
      --kube-context string   Context to use
  -r, --resource string       K8s resource to listen
```

This application has been designed to live inside the cluster, it uses the injected service-account tokens to interact 
with the API.
Optionally can be executed from outside the cluster, `--kube-config` and `--kube-context` are mandatory on those cases.

Matching resource events will be sent to callback script as arguments:
- resourceType
- action (`add`, `update` or `delete`)
- namespace
- name 

Is important to mention that this application uses a cache system that needs to be populated, so during bootstrap, 
all matching resources will be evaluated as insertions.

### Example

```
$ k8s-event-listener -r pod -c ./test.sh
```

Script will start receiving instructions like:
```
pods add instapro-client ic-nl-tests-acc-26937c1b-webserver-7ff884c575-dvwkb
pods add instapro-gd-city-filter r-travaux-com-db-queue-worker-7585fd84c-4gxzc
pods add monitoring datadog-5swgj
```