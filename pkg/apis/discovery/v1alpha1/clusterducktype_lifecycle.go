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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var duckTypeCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (*ClusterDuckType) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ClusterDuckType")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*ClusterDuckType) GetConditionSet() apis.ConditionSet {
	return duckTypeCondSet
}

// InitializeConditions sets the initial values to the conditions.
func (dts *ClusterDuckTypeStatus) InitializeConditions() {
	duckTypeCondSet.Manage(dts).InitializeConditions()
}

// TODO: temp function, delete later as we develop the reconciler.
func (dts *ClusterDuckTypeStatus) MarkReady() {
	duckTypeCondSet.Manage(dts).MarkTrue(DuckTypeConditionReady)
}
