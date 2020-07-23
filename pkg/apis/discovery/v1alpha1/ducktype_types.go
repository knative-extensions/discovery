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

// DuckType is a query and identifier for Knative-style duck types installed in a cluster.
type DuckType struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the DuckType (from the client).
	// +optional
	Spec DuckTypeSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the DuckType (from the controller).
	// +optional
	Status DuckTypeStatus `json:"status,omitempty"`
}

var (
	// Check that DuckType can be validated and defaulted.
	_ apis.Validatable   = (*DuckType)(nil)
	_ apis.Defaultable   = (*DuckType)(nil)
	_ kmeta.OwnerRefable = (*DuckType)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*DuckType)(nil)
)

// DuckTypeSpec holds the desired state of the DuckType (from the client).
type DuckTypeSpec struct {
	// Group is the API group of the defined duck type.
	// Must match the name of the DuckType (in the form `<names.plural>.<group>`).
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
	// Must match the name of the DuckType (in the form `<names.plural>.<group>`).
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

	// Refs is a list of GVRKs that implement this duck type.
	// Used for manual discovery.
	// +optional
	Refs []GroupVersionResourceKind `json:"refs,omitempty"`

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
	// If the version mapping annotation is missing, it is assumed this applies
	// as the match.
	// Must be a valid Kubernetes Label Selector.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

// GroupVersionResourceKind
type GroupVersionResourceKind struct {
	// Group is the resource group.
	// +optional
	Group string `json:"group,omitempty"`
	// Version is the version the duck type applies to for the resource.
	Version string `json:"version,omitempty"`
	// Resource is the plural resource name.
	// +optional, one of [Resource, Kind] required.
	Resource string `json:"resource,omitempty"`
	// Kind is the CamelCased resource kind.
	// +optional, one of [Resource, Kind] required.
	Kind string `json:"kind,omitempty"`
}

// FoundDuckVersion
type FoundDuck struct {
	// Version is the version of the duck found.
	DuckVersion string `json:"duckVersion"`
	// Ref is the GVRK that adheres to this duck type version.
	Ref GroupVersionResourceKind `json:"ref,omitempty"`
}

const (
	// DuckTypeConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	DuckTypeConditionReady = apis.ConditionReady
)

// DuckTypeStatus communicates the observed state of the DuckType (from the controller).
type DuckTypeStatus struct {
	duckv1.Status `json:",inline"`

	// DuckList is a list of GVRK's that implement this duck.
	DuckList []FoundDuck `json:"ducks,omitempty"`

	// DuckCount
	DuckCount int `json:"duckCount,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DuckTypeList is a list of DuckType resources
type DuckTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DuckType `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (dt *DuckType) GetStatus() *duckv1.Status {
	return &dt.Status.Status
}
