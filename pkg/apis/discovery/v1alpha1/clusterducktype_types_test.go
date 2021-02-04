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
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

	rbacv1 "k8s.io/api/rbac/v1"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestClusterDuckTypeGetStatus(t *testing.T) {
	status := &duckv1.Status{}
	config := ClusterDuckType{
		Status: ClusterDuckTypeStatus{
			Status: *status,
		},
	}

	if !cmp.Equal(config.GetStatus(), status) {
		t.Errorf("GetStatus did not retrieve status. Got=%v Want=%v", config.GetStatus(), status)
	}
}

func TestClusterDuckTypeRoundTrips_YAML(t *testing.T) {
	y := `
spec:
  group: example.com
  names:
    name: ThisDuck
    singular: thisduck
    plural: thisducks
  versions:
  - name: v1
    refs:
    - group: foo.com
      kind: Bar
      version: v2
  selectors:
  - labelSelector: "example.com/thisduck=true"
  role:
    roleref:
      kind: someKind
      name: someName
      apigroup: someAPIGroup
`

	want := &ClusterDuckType{
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
					Group:   "foo.com",
					Version: "v2",
					Kind:    "Bar",
				}},
			}},
			Selectors: []CustomResourceDefinitionSelector{{
				LabelSelector: "example.com/thisduck=true",
			}},
			Role: &Role{
				RoleRef: &rbacv1.RoleRef{
					Kind:     "someKind",
					Name:     "someName",
					APIGroup: "someAPIGroup",
				},
			},
		}}

	got := &ClusterDuckType{}
	if err := yaml.Unmarshal([]byte(y), got); err != nil {
		t.Fail()
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("From YAML (-want, +got) = %v",
			diff)
	}
}

func TestResourceRef_GroupVersion(t *testing.T) {
	tests := map[string]struct {
		in   ResourceRef
		want string
	}{
		"empty": {
			in:   ResourceRef{},
			want: "",
		},
		"version and group set": {
			in: ResourceRef{
				Group:   "this.group",
				Version: "v1",
			},
			want: "this.group/v1",
		},
		"only version": {
			in: ResourceRef{
				Version: "v1",
			},
			want: "v1",
		},
		"only group": {
			in: ResourceRef{
				Group: "this.group",
			},
			want: "this.group/", // version is required normally on these objects.
		},
		"only apiVersion": {
			in: ResourceRef{
				APIVersion: "this.group/v1",
			},
			want: "this.group/v1",
		},
		"apiVersion+group+version": {
			in: ResourceRef{
				Group:      "that.group",
				Version:    "v0",
				APIVersion: "this.group/v1",
			},
			want: "this.group/v1",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in.GroupVersion()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GroupVersion (-want, +got) = %v",
					diff)
			}
		})
	}
}

func TestResourceMeta_Group(t *testing.T) {
	tests := map[string]struct {
		in   ResourceMeta
		want string
	}{
		"empty": {
			in:   ResourceMeta{},
			want: "",
		},
		"version and group set": {
			in: ResourceMeta{
				APIVersion: "this.group/v1",
			},
			want: "this.group",
		},
		"only version": {
			in: ResourceMeta{
				APIVersion: "v1",
			},
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in.Group()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Group (-want, +got) = %v",
					diff)
			}
		})
	}
}

func TestResourceMeta_Version(t *testing.T) {
	tests := map[string]struct {
		in   ResourceMeta
		want string
	}{
		"empty": {
			in:   ResourceMeta{},
			want: "",
		},
		"version and group set": {
			in: ResourceMeta{
				APIVersion: "this.group/v1",
			},
			want: "v1",
		},
		"only version": {
			in: ResourceMeta{
				APIVersion: "v1",
			},
			want: "v1",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in.Version()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Version (-want, +got) = %v",
					diff)
			}
		})
	}
}
