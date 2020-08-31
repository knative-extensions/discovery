# Discovery API

| State          | Author(s)     | Last Update               |
| -------------- | ------------- | ------------------------- |
| In development | Scott Nichols | Last Updated: May 6, 2020 |

| Reviewer   | Role    | Approval Date |
| ---------- | ------- | ------------- |
| Serving    |         |               |
| Client     |         |               |
| Nacho Cano | Sources | 05/27/2020    |

In Knative, we are finding a need for additional metadata about the state of the
cluster and the resources that are installed there. The correct solution would
be to augment the CRDs (Custom Resource Definitions) but the amount of
customization supported by Kubernetes is limited for simplicity. The fallback we
have been using is adding annotations and labels onto the CRDs, and this has
some drawbacks, type-safety for one, and it does not help us understand the
built in types of Kubernetes as those resources have no CRD.

Another common usage of Annotations for CRDs is additional metadata that drives
UX. If we could provide a strongly typed resource to help drive CLIs and UIs for
resources, we can give a better, additive experience for operators and
developers of that cluster.

The Discovery API’s goal is to enable interactions with a cluster and its
resources: built in types, CRDs and COs. These interactions will begin with a
focus on allowing cluster level understanding of all the resources for which
interface they adhere to (ducktypes) and how they should be presented.

### ducktypes.discovery.knative.dev

Knative invented and is a heavy user of the ducktyping pattern for Kubernetes,
but to understand these types it takes several API calls and you are never 100%
sure you have found all the ducks. Today we label the CRDs for the ducks and
then collect the CRDs, and then query the api again for these known GVRs (Group
Version Resource) to find the COs (Custom Object, instances of CRDs). This model
works for Knative but to allow third parties to participate in the ducktyping
(even if they are not aware they are a duck) we would have to edit the CRD
today, but this DuckType resource is backed by a controller constantly
reconciling CRDs.

A PoC was created for this DuckType Kind:
[n3wscott/pod-discovery](https://github.com/n3wscott/poc-discovery)

It is the first piece of enabling interactions like “what is the status of all
sources on my cluster?” or “are all of my channels ready?”. Once you have a
DuckType CO of “sources.duck.knative.dev” then you can look inside the status
for the list of GroupVersionResources (GVRs) to query that are known to the
cluster to be a source.

The query is based on labels in the CRD, and by a static GVR list. Both lists
are merged together and presented in the status of the CO for that DuckType
kind.

We still need to solve the issue of CRDs adhering to a version of a DuckType at
a version of the CRD. A map of which version of the Duck to which version of the
CRD is also a goal.

### manuals.discovery.knative.dev

A kind Manual will be a rich set of descriptions and interactions that are
intended for a particular resource on the cluster. Openshift has a similar
concept but this is forced to be in the annotations of a CRD. I am proposing
that we make a strongly typed resource that is able to describe how to fill out
each field inside the spec of a resource. A good example could be in Knative, we
have a Trigger object, and this CO has a couple of object references that should
be addressable but are only validatable as addressable long after the operator
or developer has created the trigger resource. An interaction we could enable
with a Manual is the field could be given a type and the type could be the
addressable.duck.knative.dev DuckType. Now we can write UI and CLI that could
populate a list of valid COs that meet the requirement.

I am also suggesting this not be something that extends the Resource API we
already have. The intent is to enable rich interactions with any CRD that an
operator can install. This means a Manual is a separate Kind, independent of
even having the resource installed. A Manual could also give hints on where to
find the package to install to get the Kind you are requiring.

## Next Actions

There was talk of folding Discovery into Eventing directly but I feel it will be
independently useful for much more than just eventing. It will be a great
addition for things like Knative CLI, OpenShift UI generators, and GCP UI to
name a few.

This API will be owned and managed by the Eventing Source WG, it is under the
charter for sources in a cluster to be discoverable and easily installable.

To get started we need a github.com/knative/discovery repo to be created and we
can start over on API definition, starting with DuckType and Manual.

## Open Questions

1. Given the comments and perspective from the Client WG, should DuckType/Manual
   be a cluster or namespace scoped resource? Or both?
   - Work envs where only a namespace is provided to the user, and rights are
     very limited for that user, cluster scoped resources might not be best for
     them...

---

### Comments from a Knative CLI POV:

writing this in the main doc, as I think it’s easier to read & consume please
remove or move to comment anytime [rhuss]

One use case within the Knative Client kn is the management of unknown sources
in a typed way. typed in the sense that the mapping between CLI options and
fields to fill out in a source CO is detected dynamically by querying metadata.

Originally this was planned by querying all CRDs that have a certain label
(“type=source”) and then inspect the CRDs metadata (annotations of the resource
or annotations within the openApiSchema) for this mapping for a specific source.

It turns out that often a regular user is not allowed to query or get CRDs by
default which is a blocker for this implementation.

Discovery API helps a lot here as it allows us to query a specific CO (an object
of type `DuckType` with name “source”) and examine the status for the list of
source types available.

The next step would be then to query the CRD with a specific source type (e.g.
the KafkaSource CRD) and look into the meta-data provided. This is not possible
because of the restrictions, so I hope that the “Manual” suggestion above will
solve this use case (but not sure as I don’t fully understand it ;)
