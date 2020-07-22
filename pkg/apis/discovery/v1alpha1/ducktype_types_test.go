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

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestDuckTypeGetStatus(t *testing.T) {
	status := &duckv1.Status{}
	config := DuckType{
		Status: DuckTypeStatus{
			Status: *status,
		},
	}

	if !cmp.Equal(config.GetStatus(), status) {
		t.Errorf("GetStatus did not retrieve status. Got=%v Want=%v", config.GetStatus(), status)
	}
}

func TestDuckTypeRoundTrips_YAML(t *testing.T) {
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

`

	want := &DuckType{
		Spec: DuckTypeSpec{
			Group: "example.com",
			Names: DuckTypeNames{
				Name:     "ThisDuck",
				Plural:   "thisducks",
				Singular: "thisduck",
			},
			Versions: []DuckVersion{{
				Name: "v1",
				Refs: []GroupVersionResourceKind{{
					Group:   "foo.com",
					Version: "v2",
					Kind:    "Bar",
				}},
			}},
			Selectors: []CustomResourceDefinitionSelector{{
				LabelSelector: "example.com/thisduck=true",
			}},
		}}

	got := &DuckType{}
	if err := yaml.Unmarshal([]byte(y), got); err != nil {
		t.Fail()
	}

	if !cmp.Equal(got, want) {
		t.Errorf("From YAML (-want, +got) = %v",
			cmp.Diff(want, got))
	}
}
