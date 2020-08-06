# Knative Discovery API

This is Work in Progress. It is based on the
[Discovery API](https://docs.google.com/document/d/1CVMI0IpGqZH7j64-KQ6_8hJVUkeYEM0pdXslJ10a69E/edit)
design doc.

## Install

```shell script
ko apply -f ./config
```

## ClusterDuckType:discovery.knative.dev/v1alpha1

The goal is to have a custom type that is use installable to help a developer,
cluster admin, or tooling to better understand the duck types that are installed
in the cluster. This information could be used to understand which Kinds could
fulfill a role for another resource.

### example.yaml

```yaml
apiVersion: discovery.knative.dev/v1alpha1
kind: ClusterDuckType
metadata:
  name: addressables.duck.knative.dev
spec:
  # selectors is a list of CRD label selectors to find CRDs that have been
  # labeled as the given duck type.
  selectors:
    - labelSelector: "duck.knative.dev/addressable=true"

  # Names allows us to give a short name to the duck type.
  names:
    name: "Addressable"
    plural: "addressables"
    singular: "addressable"

  # Versions are to allow the definition of a single duck type with multiple
  # versions, useful if the duck type API shape changes.
  versions:
    - name: "v1"
      # refs allows for adding native types, or crds directly as the ducks via
      # Group/Version/Kind/Resource
      refs:
        - version: v1
          resource: services
          kind: Service

      # additionalPrinterColumns is intended to understand what printer columns
      # should be used for the custom objects.
      additionalPrinterColumns:
        - name: Ready
          type: string
          JSONPath: ".status.conditions[?(@.type=='Ready')].status"
        - name: Reason
          type: string
          JSONPath: ".status.conditions[?(@.type=='Ready')].reason"
        - name: URL
          type: string
          JSONPath: .status.address.url
        - name: Age
          type: date
          JSONPath: .metadata.creationTimestamp

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
                    url:
                      type: string

  group: duck.knative.dev
```

After applying this, you can fetch it:

```shell
$  kubectl get ducktypes addressables.duck.knative.dev
NAME                           SHORT NAME    DUCKS   READY   REASON
addressables.duck.knative.dev   addressable   11      True
```

And get the full DuckType `addressable.duck.knative.dev` resource:

_(note this is pending finalization)_

```shell
$ kubectl get clusterducktypes addressable.duck.knative.dev -oyaml
apiVersion: discovery.knative.dev/v1alpha1
kind: CluserDuckType
metadata:
  generation: 1
  name: addressables.duck.knative.dev
spec:
  ...
status:
  conditions:
  - lastTransitionTime: "2020-01-16T22:13:24Z"
    status: "True"
    type: Ready
  duckCount: 10
  ducks:
  - group: eventing.knative.dev
    kind: Broker
    resource: brokers
    version: v1
  - group: flows.knative.dev
    kind: Parallel
    resource: parallels
    version: v1
  - group: flows.knative.dev
    kind: Sequence
    resource: sequences
    version: v1
  - group: messaging.knative.dev
    kind: Channel
    resource: channels
    version: v1
  - group: messaging.knative.dev
    kind: InMemoryChannel
    resource: inmemorychannels
    version: v1
  - group: messaging.knative.dev
    kind: Parallel
    resource: parallels
    version: v1
  - group: messaging.knative.dev
    kind: Sequence
    resource: sequences
    version: v1
  - kind: Service
    resource: services
    version: v1
  - group: serving.knative.dev
    kind: Route
    resource: routes
    version: v1alpha1
  - group: serving.knative.dev
    kind: Service
    resource: services
    version: v1
  observedGeneration: 1
```

## Knative Duck Types

If the `./config/knative` directory is applied (via
`kubectl apply -f config/knative`), a quick view of the duck types that are on
the cluster becomes easier to get:

_(note this is pending finalization)_

```shell
$ kubectl get clusterducktypes
NAME                                  DUCK NAME      DUCKS   READY   REASON
addressables.duck.knative.dev         Addressable    10      True
bindings.duck.knative.dev             Binding        2       True
podspecables.duck.knative.dev         PodSpecable    7       True
sources.duck.knative.dev              Source         7       True
subscribables.messaging.knative.dev   Subscribable   2       True
```
