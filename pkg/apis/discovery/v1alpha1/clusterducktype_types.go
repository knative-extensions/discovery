/*
Copyright 2020 The Knative Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genclient:nonNamespaced
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterDuckType is a query and identifier for Knative-style duck types installed in a cluster.
type ClusterDuckType struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the ClusterDuckType (from the client).
	// +optional
	Spec ClusterDuckTypeSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the ClusterDuckType (from the controller).
	// +optional
	Status ClusterDuckTypeStatus `json:"status,omitempty"`
}

var (
	// Check that ClusterDuckType can be validated and defaulted.
	_ apis.Validatable   = (*ClusterDuckType)(nil)
	_ apis.Defaultable   = (*ClusterDuckType)(nil)
	_ kmeta.OwnerRefable = (*ClusterDuckType)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*ClusterDuckType)(nil)
)

// ClusterDuckTypeSpec holds the desired state of the ClusterDuckType (from the client).
type ClusterDuckTypeSpec struct {
	// Group is the API group of the defined duck type.
	// Must match the name of the ClusterDuckType (in the form `<names.plural>.<group>`).
	Group string `json:"group"`

	// Names holds the naming conventions for this duck type.
	Names DuckTypeNames `json:"names"`

	// Versions holds the schema and printer column mappings for specific
	// versions fo duck types.
	Versions []DuckVersion `json:"versions" patchStrategy:"merge" patchMergeKey:"name"`

	// Selectors is a list of selectors for CustomResourceDefinitions to identify a duck type.
	// +optional
	Selectors []CustomResourceDefinitionSelector `json:"selectors,omitempty"`
}

// DuckTypeNames provides the naming rules for this duck type.
type DuckTypeNames struct {
	// Name is the serialized name of the resource. It is normally CamelCase and singular.
	Name string `json:"name"`

	// Plural is the plural name of the duck type.
	// Must match the name of the ClusterDuckType (in the form `<names.plural>.<group>`).
	// Must be all lowercase.
	Plural string `json:"plural"`

	// Singular is the singular name of the duck type. It must be all lowercase.
	// Defaults to lowercased `name`.
	Singular string `json:"singular"`
}

// DuckVersion
type DuckVersion struct {
	// Name is the name of this duck type version.
	Name string `json:"name"`

	// Refs is a list of ResourceRefs that implement this duck type.
	// Used for manual discovery.
	// +optional
	Refs []ResourceRef `json:"refs,omitempty"`

	// Custom Columns to be used to pretty print the duck type at this version.
	// +optional
	AdditionalPrinterColumns []apiextensionsv1.CustomResourceColumnDefinition `json:"additionalPrinterColumns,omitempty"`

	// Partial Schema of this version of the duck type.
	// +optional
	Schema *apiextensionsv1.CustomResourceValidation `json:"schema,omitempty"`
}

// CustomResourceDefinitionSelector
type CustomResourceDefinitionSelector struct {
	// Note: in the future we could also add a filtering component here to
	// limit the matches based on another field or annotation. Leaving this as
	// a sub-object for that future.

	// LabelSelector is a label selector used to find CRDs that associate with
	// the duck type.
	// Typically this will be in the form:
	//   `<group>/<names.singular>=true`
	// Annotations are used to map the versions of the CRD to the correct
	// ducktype. The annotation is expected to be in the form:
	//   `<names.plural>.<group>/<versions[x].name>=[CRD.Version]`
	// and results in `x = CRD.Version`.
	// The duck type version annotation can have several CRD versions that map:
	//   `<names.plural>.<group>/<versions[x].name>=[CRD.V1],[CRD.V2],[CRD.V3]`
	// this tells the interrupter to match x to all of V1, V2 and V3 versions.
	// If the version mapping annotation is missing, it is assumed this applies
	// as the match.
	// Must be a valid Kubernetes Label Selector.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

// ResourceScope is an enum defining the different scopes available to a custom resource
type ResourceScope string

const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

// ResourceRef points to a Kubernetes Resource kind.
type ResourceRef struct {
	// Group is the resource group.
	// +optional, must not be set if APIVersion is set.
	Group string `json:"group,omitempty"`
	// Version is the version the duck type applies to for the resource.
	// +optional
	// +optional, one of [Version, APIVersion] required.
	Version string `json:"version,omitempty"`

	// APIVersion is the group and version of the resource combined.
	// - if group is non-empty, `group/version`
	// - if group is empty, `version`
	// +optional, one of [Version, APIVersion] required.
	APIVersion string `json:"apiVersion,omitempty"`

	// Resource is the plural resource name.
	// +optional, one of [Resource, Kind] required.
	Resource string `json:"resource,omitempty"`
	// Kind is the CamelCased resource kind.
	// +optional, one of [Resource, Kind] required.
	Kind string `json:"kind,omitempty"`

	// Scope indicates whether the resource is cluster- or namespace-scoped.
	// +optional, allowed values are `Cluster` and `Namespaced`, defaults to "Namespaced".
	Scope ResourceScope `json:"scope"`
}

// APIVersion puts "group" and "version" into a single "group/version" string
// or returns APIVersion.
func (r *ResourceRef) GroupVersion() string {
	if len(r.APIVersion) > 0 {
		return r.APIVersion
	}
	if len(r.Group) > 0 {
		return r.Group + "/" + r.Version
	}
	return r.Version
}

// ResourceMeta is a resolved ResourceRef.
type ResourceMeta struct {
	// APIVersion is the group and version of the resource combined.
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind is the CamelCased resource kind.
	Kind string `json:"kind"`

	// Scope indicates whether the resource is cluster- or namespace-scoped.
	// Allowed values are `Cluster` and `Namespaced`.
	Scope ResourceScope `json:"scope"`
}

// Version inspects a ResourceMeta object and returns the correct version
// based on APIVersion.
func (r *ResourceMeta) Version() string {
	if strings.Contains(r.APIVersion, "/") {
		sp := strings.Split(r.APIVersion, "/")
		return sp[len(sp)-1]
	}
	return r.APIVersion
}

// Group inspects a ResourceMeta object and returns the correct group
// based on APIVersion.
func (r *ResourceMeta) Group() string {
	if strings.Contains(r.APIVersion, "/") {
		sp := strings.Split(r.APIVersion, "/")
		return sp[0]
	}
	return ""
}

const (
	// DuckTypeConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	DuckTypeConditionReady = apis.ConditionReady
)

// ClusterDuckTypeStatus communicates the observed state of the ClusterDuckType (from the controller).
type ClusterDuckTypeStatus struct {
	duckv1.Status `json:",inline"`

	// Ducks is a versioned mapping of the found resources that implement this duck.
	Ducks map[string][]ResourceMeta `json:"ducks,omitempty"`

	// DuckCount is the count of unique duck types found post-hunt.
	DuckCount int `json:"duckCount"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterDuckTypeList is a list of ClusterDuckType resources
type ClusterDuckTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterDuckType `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (dt *ClusterDuckType) GetStatus() *duckv1.Status {
	return &dt.Status.Status
}
