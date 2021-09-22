# K8s Event listener

[![Build Status](https://travis-ci.com/Werkspot/k8s-event-listener.svg?branch=master)](https://travis-ci.com/Werkspot/k8s-event-listener)

Listen for changes on specified resources and invokes the callback script

Currently supporting:
- certificatesigningrequests
- cronjobs
- ingresses
- pods
- serviceaccounts
- nodes

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
  -p, --probe-port string     HTTP port to listen for liveness/readiness probes (default "8080")
  -r, --resource string       K8s resource to listen
  -v, --verbose string        Verbose level (default "0")
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

## Healthcheck
An HTTP server will be created listening to 8080 (can be overwritten via -p flag) with two available probe URLs

- `/live` to be used in the liveness configuration
- `/ready` to be used in the readiness configuration

Currently only checks for kube-api connectivity.

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
