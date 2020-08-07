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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (dt *ClusterDuckType) Validate(ctx context.Context) (errs *apis.FieldError) {
	if dt.Name != fmt.Sprintf("%s.%s", dt.Spec.Names.Plural, dt.Spec.Group) {
		errs = errs.Also(apis.ErrInvalidValue(dt.Name, "name"))
	}

	return errs.Also(dt.Spec.Validate(ctx).ViaField("spec"))
}

// Validate implements apis.Validatable
func (dts *ClusterDuckTypeSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	if dts.Group == "" {
		errs = errs.Also(apis.ErrMissingField("group"))
	}
	if len(dts.Versions) == 0 {
		errs = errs.Also(apis.ErrMissingField("versions"))
	}

	errs = errs.Also(dts.Names.Validate(ctx).ViaField("names"))

	seenVersionNames := make(map[string]string, 0)
	for i, v := range dts.Versions {
		if _, found := seenVersionNames[v.Name]; found {
			errs = errs.Also((&apis.FieldError{
				Message: fmt.Sprintf("duplicate entry found: %s", v.Name),
				Paths:   []string{"name"},
			}).ViaFieldIndex("versions", i))
		}
		seenVersionNames[v.Name] = v.Name
		errs = errs.Also(v.Validate(ctx).ViaFieldIndex("versions", i))
	}

	for i, st := range dts.Selectors {
		_, err := labels.Parse(st.LabelSelector)
		if err != nil {
			errs = errs.Also(apis.ErrInvalidValue(st.LabelSelector, "labelSelector").ViaFieldIndex("selectors", i))
		}
	}

	return errs
}

// Validate implements apis.Validatable
func (dtn *DuckTypeNames) Validate(ctx context.Context) (errs *apis.FieldError) {
	if dtn.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}
	if dtn.Plural == "" {
		errs = errs.Also(apis.ErrMissingField("plural"))
	} else if dtn.Plural != strings.ToLower(dtn.Plural) {
		errs = errs.Also(apis.ErrInvalidValue(dtn.Plural, "plural"))
	}
	if dtn.Singular == "" {
		errs = errs.Also(apis.ErrMissingField("singular"))
	} else if dtn.Singular != strings.ToLower(dtn.Singular) {
		errs = errs.Also(apis.ErrInvalidValue(dtn.Singular, "singular"))
	}
	return errs
}

// Validate implements apis.Validatable
func (dv *DuckVersion) Validate(ctx context.Context) (errs *apis.FieldError) {
	if dv.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}
	for i, ref := range dv.Refs {
		errs = errs.Also(ref.Validate(ctx).ViaFieldIndex("refs", i))
	}
	return errs
}

// Validate implements apis.Validatable
func (g *ResourceRef) Validate(ctx context.Context) (errs *apis.FieldError) {
	// Version OR APIVersion
	if g.Version != "" && g.APIVersion != "" {
		errs = errs.Also(apis.ErrMultipleOneOf("version", "apiVersion"))
	} else if g.Version == "" && g.APIVersion == "" {
		errs = errs.Also(apis.ErrMissingOneOf("version", "apiVersion"))
	}

	// Kind OR Resource
	if g.Kind != "" && g.Resource != "" {
		errs = errs.Also(apis.ErrMultipleOneOf("kind", "resource"))
	} else if g.Kind == "" && g.Resource == "" {
		errs = errs.Also(apis.ErrMissingOneOf("kind", "resource"))
	}

	// If Group, then APIVersion should not be set.
	if g.Group != "" && g.APIVersion != "" {
		errs = errs.Also(apis.ErrMultipleOneOf("group", "apiVersion"))
	}
	return errs
}
