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

package collection

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewResourceMapper(t *testing.T) {
	tests := map[string]struct {
		apis []*metav1.APIResourceList
		want *resourceMapper
	}{
		"one api, one resource": {
			apis: []*metav1.APIResourceList{
				{
					GroupVersion: "swan.lake/v1",
					APIResources: []metav1.APIResource{{
						Name:       "ballets",
						Namespaced: false,
						Kind:       "Ballet",
					}},
				}},
			want: &resourceMapper{
				mappings: map[string]mapping{
					"swan.lake/v1": {
						r2k: map[string]string{"ballets": "Ballet"},
						k2r: map[string]string{"Ballet": "ballets"},
					},
				},
			},
		},
		"one api, one resource, one sub resource": {
			apis: []*metav1.APIResourceList{
				{
					GroupVersion: "swan.lake/v1",
					APIResources: []metav1.APIResource{{
						Name:       "ballets",
						Namespaced: false,
						Kind:       "Ballet",
					}, {
						Name:       "ballets/status",
						Namespaced: false,
						Kind:       "Ballet",
					}},
				}},
			want: &resourceMapper{
				mappings: map[string]mapping{
					"swan.lake/v1": {
						r2k: map[string]string{"ballets": "Ballet"},
						k2r: map[string]string{"Ballet": "ballets"},
					},
				},
			},
		},
		"one api, two resources": {
			apis: []*metav1.APIResourceList{
				{
					GroupVersion: "swan.lake/v1",
					APIResources: []metav1.APIResource{{
						Name: "ballets",
						Kind: "Ballet",
					}, {
						Name: "shoes",
						Kind: "Shoe",
					}},
				}},
			want: &resourceMapper{
				mappings: map[string]mapping{
					"swan.lake/v1": {
						r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
						k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
					},
				},
			},
		},
		"two api, three resources": {
			apis: []*metav1.APIResourceList{
				{
					GroupVersion: "swan.lake/v1",
					APIResources: []metav1.APIResource{{
						Name: "ballets",
						Kind: "Ballet",
					}, {
						Name: "shoes",
						Kind: "Shoe",
					}},
				}, {
					GroupVersion: "duck.lake/v2",
					APIResources: []metav1.APIResource{{
						Name: "breads",
						Kind: "Bread",
					}},
				}},
			want: &resourceMapper{
				mappings: map[string]mapping{
					"duck.lake/v2": {
						r2k: map[string]string{"breads": "Bread"},
						k2r: map[string]string{"Bread": "breads"},
					},
					"swan.lake/v1": {
						r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
						k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
					},
				},
			},
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewResourceMapper(tc.apis); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewDuckHunter() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResourceFor(t *testing.T) {
	rm := &resourceMapper{
		mappings: map[string]mapping{
			"swan.lake/v1": {
				r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
				k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
			},
			"duck.lake/v2": {
				r2k: map[string]string{"breads": "Bread"},
				k2r: map[string]string{"Bread": "breads"},
			},
		}}

	tests := map[string]struct {
		rm           ResourceMapper
		groupVersion string
		kind         string
		want         string
		wantErr      bool
	}{
		"duck lake Bread->breads": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "Bread",
			want:         "breads",
		},
		"wrong kind casing breads->Bread": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "breads",
			wantErr:      true,
		},
		"duck lake Ballet->error": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "Ballet",
			wantErr:      true,
		},
		"unknown Shoe->error": {
			rm:           rm,
			groupVersion: "data.lake/v3",
			kind:         "Shoe",
			wantErr:      true,
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tc.rm.ResourceFor(tc.groupVersion, tc.kind)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("expected error calling rm.ResourceFor(%q, %q): %v", tc.groupVersion, tc.kind, err)
				}
			} else if got != tc.want {
				t.Errorf("rm.ResourceFor(%q, %q) = %v, want %v", tc.groupVersion, tc.kind, got, tc.want)
			}
		})
	}
}

func TestKindExists(t *testing.T) {
	rm := &resourceMapper{
		mappings: map[string]mapping{
			"swan.lake/v1": {
				r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
				k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
			},
			"duck.lake/v2": {
				r2k: map[string]string{"breads": "Bread"},
				k2r: map[string]string{"Bread": "breads"},
			},
		}}

	tests := map[string]struct {
		rm           ResourceMapper
		groupVersion string
		kind         string
		want         bool
	}{
		"duck lake Bread->breads": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "Bread",
			want:         true,
		},
		"wrong kind casing breads->Bread": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "breads",
			want:         false,
		},
		"duck lake Ballet->error": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			kind:         "Ballet",
			want:         false,
		},
		"unknown Shoe->error": {
			rm:           rm,
			groupVersion: "data.lake/v3",
			kind:         "Shoe",
			want:         false,
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.rm.KindExists(tc.groupVersion, tc.kind); got != tc.want {
				t.Errorf("rm.KindExists(%q, %q) = %v, want %v", tc.groupVersion, tc.kind, got, tc.want)
			}
		})
	}
}

func TestKindFor(t *testing.T) {
	rm := &resourceMapper{
		mappings: map[string]mapping{
			"swan.lake/v1": {
				r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
				k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
			},
			"duck.lake/v2": {
				r2k: map[string]string{"breads": "Bread"},
				k2r: map[string]string{"Bread": "breads"},
			},
		}}

	tests := map[string]struct {
		rm           ResourceMapper
		groupVersion string
		resource     string
		want         string
		wantErr      bool
	}{
		"duck lake breads->Bread": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "breads",
			want:         "Bread",
		},
		"wrong resource casing Bread->breads": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "Bread",
			wantErr:      true,
		},
		"duck lake ballets->error": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "ballet",
			wantErr:      true,
		},
		"unknown Shoe->error": {
			rm:           rm,
			groupVersion: "data.lake/v3",
			resource:     "shoe",
			wantErr:      true,
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tc.rm.KindFor(tc.groupVersion, tc.resource)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("expected error calling rm.KindFor(%q, %q): %v", tc.groupVersion, tc.resource, err)
				}
			} else if got != tc.want {
				t.Errorf("rm.KindFor(%q, %q) = %v, want %v", tc.groupVersion, tc.resource, got, tc.want)
			}
		})
	}
}

func TestResourceExists(t *testing.T) {
	rm := &resourceMapper{
		mappings: map[string]mapping{
			"swan.lake/v1": {
				r2k: map[string]string{"ballets": "Ballet", "shoes": "Shoe"},
				k2r: map[string]string{"Ballet": "ballets", "Shoe": "shoes"},
			},
			"duck.lake/v2": {
				r2k: map[string]string{"breads": "Bread"},
				k2r: map[string]string{"Bread": "breads"},
			},
		}}

	tests := map[string]struct {
		rm           ResourceMapper
		groupVersion string
		resource     string
		want         bool
	}{
		"duck lake breads": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "breads",
			want:         true,
		},
		"wrong resource casing Bread": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "Bread",
			want:         false,
		},
		"duck lake ballets->error": {
			rm:           rm,
			groupVersion: "duck.lake/v2",
			resource:     "ballets",
			want:         false,
		},
		"unknown Shoe->error": {
			rm:           rm,
			groupVersion: "data.lake/v3",
			resource:     "shoe",
			want:         false,
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.rm.ResourceExists(tc.groupVersion, tc.resource); got != tc.want {
				t.Errorf("rm.ResourceExists(%q, %q) = %v, want %v", tc.groupVersion, tc.resource, got, tc.want)
			}
		})
	}
}
