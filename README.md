# Knative Discovery API

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/knative.dev/discovery)
[![Go Report Card](https://goreportcard.com/badge/knative.dev/discovery)](https://goreportcard.com/report/knative.dev/discovery)
[![Releases](https://img.shields.io/github/release-pre/knative-sandbox/discovery.svg)](https://github.com/knative-sandbox/discovery/releases)
[![LICENSE](https://img.shields.io/github/license/knative-sandbox/discovery.svg)](https://github.com/knative-sandbox/discovery/blob/master/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://knative.slack.com)
[![TestGrid](https://img.shields.io/badge/testgrid-discovery-informational)](https://testgrid.knative.dev/discovery)

This is Work in Progress. It is based on the [Discovery API](./docs/proposal.md)
design doc.

## Install

TODO: instructions on installing nightly and a release.

## Development

Install, 

```shell script
ko apply -f ./config
```

## ClusterDuckType:discovery.knative.dev/v1alpha1

The goal is to have a custom type that is use installable to help a developer,
cluster admin, or tooling to better understand the duck types that are installed
in the cluster. This information could be used to understand which Kinds could
fulfill a role for another resource.

### demos.example.com.yaml

```yaml
apiVersion: discovery.knative.dev/v1alpha1
kind: ClusterDuckType
metadata:
  name: demos.example.com
spec:
  # selectors is a list of CRD label selectors to find CRDs that have been
  # labeled as the given duck type.
  selectors:
    - labelSelector: "example.com/demo=true"

  # Names allows us to give a short name to the duck type.
  names:
    name: "Demo"
    plural: "demos"
    singular: "demo"

  # Versions are to allow the definition of a single duck type with multiple
  # versions, useful if the duck type API shape changes.
  versions:
    - name: "v1"
      # refs allows for adding native types, or crds directly as the ducks via
      # Group/Version/Kind/Resource
      refs:
        - group: "demo.example.com"
          version: "v1"
          kind: "Demo"
      # additionalPrinterColumns is intended to understand what printer columns
      # should be used for the custom objects.
      additionalPrinterColumns:
        - name: Ready
          type: string
          jsonPath: ".status.conditions[?(@.type=='Ready')].status"
        - name: Reason
          type: string
          jsonPath: ".status.conditions[?(@.type=='Ready')].reason"
        - name: Demo
          type: string
          jsonPath: .status.demo
      # schema is the partial schema of the duck type.
      schema:
        openAPIV3Schema:
          properties:
            status:
              type: object
              properties:
                address:
                  type: object
                  properties:
                    demo:
                      type: string
  group: example.com
```

### Demo

Using [`addressables.duck.knative.dev.yaml`](./config/knative/addressables.duck.knative.dev.yaml), we will apply it,

```shell
kubectl apply -f ./config/knative/addressables.duck.knative.dev.yaml
```

```text
clusterducktype.discovery.knative.dev/addressables.duck.knative.dev created
```

After applying this, you can fetch it:

```shell
kubectl get ducktypes addressables.duck.knative.dev
```

```text
NAME                            SHORT NAME    DUCKS   READY   REASON
addressables.duck.knative.dev   addressable   6       True
```

And get the full DuckType `addressable.duck.knative.dev` resource:

```shell
kubectl get clusterducktypes addressable.duck.knative.dev -oyaml
```

```yaml
apiVersion: discovery.knative.dev/v1alpha1
kind: CluserDuckType
metadata:
  generation: 2
  name: addressables.duck.knative.dev
spec:
  ...
status:
  conditions:
  - lastTransitionTime: "2020-08-11T22:21:57Z"
    status: "True"
    type: Ready
  duckCount: 6
  ducks:
    v1:
    - apiVersion: eventing.knative.dev/v1
      kind: Broker
      scope: Namespaced
    - apiVersion: eventing.knative.dev/v1beta1
      kind: Broker
      scope: Namespaced
    - apiVersion: flows.knative.dev/v1
      kind: Parallel
      scope: Namespaced
    - apiVersion: flows.knative.dev/v1
      kind: Sequence
      scope: Namespaced
    - apiVersion: flows.knative.dev/v1beta1
      kind: Parallel
      scope: Namespaced
    - apiVersion: flows.knative.dev/v1beta1
      kind: Sequence
      scope: Namespaced
    - apiVersion: messaging.knative.dev/v1
      kind: Channel
      scope: Namespaced
    - apiVersion: messaging.knative.dev/v1beta1
      kind: Channel
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1
      kind: Route
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1
      kind: Service
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1alpha1
      kind: Route
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1alpha1
      kind: Service
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1beta1
      kind: Route
      scope: Namespaced
    - apiVersion: serving.knative.dev/v1beta1
      kind: Service
      scope: Namespaced
  observedGeneration: 2
```

## Knative Duck Types

If the `./config/knative` directory is applied (via
`kubectl apply -f config/knative`), a quick view of the duck types used by Knative in this cluster
becomes easier to find:

```shell
kubectl get clusterducktypes
```

```text
NAME                            SHORT NAME    DUCKS   READY   REASON
addressables.duck.knative.dev   Addressable   6       True
bindings.duck.knative.dev       Binding       1       True
channelables.duck.knative.dev   Channelable   0       True
podspecables.duck.knative.dev   PodSpecable   7       True
sources.duck.knative.dev        Source        4       True
```

_Note_: there is also a short name: `cducks`

```shell
kubectl get cducks
``` 