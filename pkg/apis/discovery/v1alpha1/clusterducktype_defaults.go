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
	"context"
	"strings"

	"knative.dev/pkg/apis"
)

// SetDefaults implements apis.Defaultable
func (dt *ClusterDuckType) SetDefaults(ctx context.Context) {
	ctx = apis.WithinParent(ctx, dt.ObjectMeta)
	dt.Spec.SetDefaults(apis.WithinSpec(ctx))
}

// SetDefaults implements apis.Defaultable
func (dts *ClusterDuckTypeSpec) SetDefaults(ctx context.Context) {
	// names.singular defaults to lowercase names.name if not set.
	if dts.Names.Singular == "" {
		dts.Names.Singular = strings.ToLower(dts.Names.Name)
	}
	for v := range dts.Versions {
		dts.Versions[v].SetDefaults(ctx)
	}
}

// SetDefaults implements apis.Defaultable
func (dv *DuckVersion) SetDefaults(ctx context.Context) {
	for r := range dv.Refs {
		dv.Refs[r].SetDefaults(ctx)
	}
}

func (rr *ResourceRef) SetDefaults(ctx context.Context) {
	if rr.Scope == "" {
		rr.Scope = NamespaceScoped
	}
}
