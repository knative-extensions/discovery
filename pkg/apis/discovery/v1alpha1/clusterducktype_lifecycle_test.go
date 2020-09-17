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
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestDuckTypeDuckTypes(t *testing.T) {
	tests := []struct {
		name string
		t    duck.Implementable
	}{{
		name: "conditions",
		t:    &duckv1.Conditions{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := duck.VerifyType(&ClusterDuckType{}, test.t)
			if err != nil {
				t.Errorf("VerifyType(ClusterDuckType, %T) = %v", test.t, err)
			}
		})
	}
}

func TestDuckTypeGetConditionSet(t *testing.T) {
	r := &ClusterDuckType{}

	if got, want := r.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestDuckTypeGetGroupVersionKind(t *testing.T) {
	r := &ClusterDuckType{}
	want := schema.GroupVersionKind{
		Group:   "discovery.knative.dev",
		Version: "v1alpha1",
		Kind:    "ClusterDuckType",
	}
	if got := r.GetGroupVersionKind(); got != want {
		t.Errorf("GVK: %v, want: %v", got, want)
	}
}

func TestDuckTypeInitializeConditions(t *testing.T) {
	rs := &ClusterDuckTypeStatus{}
	rs.InitializeConditions()

	types := make([]string, 0, len(rs.Conditions))
	for _, cond := range rs.Conditions {
		types = append(types, string(cond.Type))
	}

	// These are already sorted.
	expected := []string{
		string(DuckTypeConditionReady),
	}

	sort.Strings(types)

	if diff := cmp.Diff(expected, types); diff != "" {
		t.Error("Conditions(-want,+got):\n", diff)
	}
}

func TestDuckTypeMarkReady(t *testing.T) {
	rs := &ClusterDuckTypeStatus{}
	rs.MarkReady()

	c := rs.GetCondition(DuckTypeConditionReady)
	if c == nil || c.Status != corev1.ConditionTrue {
		t.Errorf("expected Ready to be true, got %v\n", c)
	}
}
