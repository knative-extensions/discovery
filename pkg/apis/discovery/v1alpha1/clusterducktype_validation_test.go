/*
Copyright 2020 The Knative Authors

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
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestDuckTypeValidation(t *testing.T) {
	tests := map[string]struct {
		in   *ClusterDuckType
		want *apis.FieldError
	}{
		"empty": {
			in: &ClusterDuckType{},
			want: (&apis.FieldError{
				Message: "invalid value: ",
				Paths:   []string{"name"},
			}).Also(
				&apis.FieldError{
					Message: "missing field(s)",
					Paths:   []string{"spec.group", "spec.names.name", "spec.names.plural", "spec.names.singular", "spec.versions"},
				}),
		},
		"missing versions": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "thisduck",
						Singular: "thisduck",
						Plural:   "thisducks",
					},
				},
			},
			want: &apis.FieldError{
				Message: "missing field(s)",
				Paths:   []string{"spec.versions"},
			},
		},
		"invalid name": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "ThisDucks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}},
				}},
			want: &apis.FieldError{
				Message: "invalid value: ThisDucks.example.com",
				Paths:   []string{"name"},
			},
		},
		"plural not lowercase": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "ThisDucks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "ThisDucks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}},
				}},
			want: &apis.FieldError{
				Message: "invalid value: ThisDucks",
				Paths:   []string{"spec.names.plural"},
			},
		},
		"singular not lowercase": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "ThisDuck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}},
				}},
			want: &apis.FieldError{
				Message: "invalid value: ThisDuck",
				Paths:   []string{"spec.names.singular"},
			},
		},
		"dup versions": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}, {
						Name: "v1",
					}},
				}},
			want: &apis.FieldError{
				Message: "duplicate entry found: v1",
				Paths:   []string{"spec.versions[1].name"},
			},
		},
		"version with no name": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{}},
				}},
			want: &apis.FieldError{
				Message: "missing field(s)",
				Paths:   []string{"spec.versions[0].name"},
			},
		},
		"version with invalid ref, no kind or resource": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Version: "v2",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got neither",
				Paths:   []string{"spec.versions[0].refs[0].kind, spec.versions[0].refs[0].resource"},
			},
		},
		"version with invalid ref, both kind and resource": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Version:  "v2",
							Kind:     "Foo",
							Resource: "bars",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].kind, spec.versions[0].refs[0].resource"},
			},
		},
		"version with invalid ref, missing version": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Kind: "Foo",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got neither",
				Paths:   []string{"spec.versions[0].refs[0].apiVersion, spec.versions[0].refs[0].version"},
			},
		},
		"bad selector": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}},
					Selectors: []CustomResourceDefinitionSelector{{
						LabelSelector: "turn down for duck",
					}},
				},
			},
			want: &apis.FieldError{
				Message: "invalid value: turn down for duck",
				Paths:   []string{"spec.selectors[0].labelSelector"},
			},
		},
		"valid": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
					}},
				}},
		},
		"valid - GVR": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Group:    "a.group",
							Version:  "v2",
							Resource: "bills",
						}},
					}},
				}},
		},
		"valid - GVK": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Group:   "a.group",
							Version: "v2",
							Kind:    "Bill",
						}},
					}},
				}},
		},
		"valid - AK": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Kind:       "Bill",
						}},
					}},
				}},
		},
		"valid - AR": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Resource:   "bills",
						}},
					}},
				}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in.Validate(context.Background())
			if !cmp.Equal(tc.want.Error(), got.Error()) {
				t.Errorf("Validate (-want, +got) = %v",
					cmp.Diff(tc.want.Error(), got.Error()))
			}
		})
	}

}

func TestDuckTypeValidation_Refs(t *testing.T) {
	tests := map[string]struct {
		in   *ClusterDuckType
		want *apis.FieldError
	}{
		"invalid - GVR+K": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							Group:    "a.group",
							Version:  "v2",
							Resource: "bills",
							Kind:     "Bills",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].kind, spec.versions[0].refs[0].resource"},
			},
		},
		"invalid - GVR+A": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Group:      "a.group",
							Version:    "v2",
							Resource:   "bills",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].apiVersion, spec.versions[0].refs[0].group, spec.versions[0].refs[0].version"},
			},
		},
		"invalid - GVR+A+K": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Group:      "a.group",
							Version:    "v2",
							Resource:   "bills",
							Kind:       "Bill",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].apiVersion, spec.versions[0].refs[0].group, spec.versions[0].refs[0].kind, spec.versions[0].refs[0].resource, spec.versions[0].refs[0].version"},
			},
		},
		"invalid - AK+G": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Kind:       "Bill",
							Group:      "a.group",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].apiVersion, spec.versions[0].refs[0].group"},
			},
		},
		"invalid - AK+V": {
			in: &ClusterDuckType{
				ObjectMeta: v1.ObjectMeta{
					Name: "thisducks.example.com",
				},
				Spec: ClusterDuckTypeSpec{
					Group: "example.com",
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Plural:   "thisducks",
						Singular: "thisduck",
					},
					Versions: []DuckVersion{{
						Name: "v1",
						Refs: []ResourceRef{{
							APIVersion: "a.group/v2",
							Kind:       "Bill",
							Version:    "v2",
						}},
					}},
				}},
			want: &apis.FieldError{
				Message: "expected exactly one, got both",
				Paths:   []string{"spec.versions[0].refs[0].apiVersion, spec.versions[0].refs[0].version"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in.Validate(context.Background())
			if !cmp.Equal(tc.want.Error(), got.Error()) {
				t.Errorf("Validate (-want, +got) = %v",
					cmp.Diff(tc.want.Error(), got.Error()))
			}
		})
	}

}
