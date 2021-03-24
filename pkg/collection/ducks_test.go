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
	"sort"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
)

func makeCRD(group, kind string, versions map[string]bool) *apiextensionsv1.CustomResourceDefinition {
	gvk := schema.GroupVersion{
		Group:   group,
		Version: "",
	}.WithKind(kind)

	plural, singular := meta.UnsafeGuessKindToResource(gvk)

	crd := apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: gvk.Group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     gvk.Kind,
				Plural:   plural.Resource,
				Singular: singular.Resource,
			},
			Scope:                 "Namespaced",
			Versions:              []apiextensionsv1.CustomResourceDefinitionVersion{},
			Conversion:            nil,
			PreserveUnknownFields: false,
		},
	}

	for name, served := range versions {
		crd.Spec.Versions = append(crd.Spec.Versions, apiextensionsv1.CustomResourceDefinitionVersion{
			Name:   name,
			Served: served,
		})
	}
	return &crd
}

func makeCRDAnnotated(group, kind string, versions map[string]bool, labels, annotations map[string]string) *apiextensionsv1.CustomResourceDefinition {
	crd := makeCRD(group, kind, versions)

	crd.Annotations = annotations
	crd.Labels = labels

	return crd
}

func TestNewDuckHunter(t *testing.T) {
	tests := map[string]struct {
		mapper   ResourceMapper
		versions []v1alpha1.DuckVersion
		want     *duckHunter
	}{
		"no defaultVersions": {
			want: &duckHunter{
				mapper:                  NewResourceMapper(nil),
				defaultVersions:         []string{},
				ducks:                   map[string][]v1alpha1.ResourceMeta{},
				accesbileGroupresources: map[string]bool{},
			},
		},
		"one defaultVersions": {
			versions: []v1alpha1.DuckVersion{
				{Name: "v1"},
			},
			want: &duckHunter{
				mapper:          NewResourceMapper(nil),
				defaultVersions: []string{"v1"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
				},
				accesbileGroupresources: map[string]bool{},
			},
		},
		"non nil mapper": {
			versions: []v1alpha1.DuckVersion{
				{Name: "v1"},
			},
			mapper: NewResourceMapper([]*metav1.APIResourceList{{
				GroupVersion: "teach.me.how/v2",
				APIResources: []metav1.APIResource{{
					Kind:       "Ducky",
					Name:       "duckies",
					Namespaced: false,
				}},
			}}),
			want: &duckHunter{
				mapper: NewResourceMapper([]*metav1.APIResourceList{{
					GroupVersion: "teach.me.how/v2",
					APIResources: []metav1.APIResource{{
						Kind:       "Ducky",
						Name:       "duckies",
						Namespaced: false,
					}},
				}}),
				defaultVersions: []string{"v1"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
				},
				accesbileGroupresources: map[string]bool{},
			},
		},
		"three defaultVersions": {
			versions: []v1alpha1.DuckVersion{{Name: "v1"}, {Name: "v2"}, {Name: "v3"}},
			want: &duckHunter{
				mapper:          NewResourceMapper(nil),
				defaultVersions: []string{"v1", "v2", "v3"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
					"v2": {},
					"v3": {},
				},
				accesbileGroupresources: map[string]bool{},
			},
		},
		"overlapping defaultVersions": {
			versions: []v1alpha1.DuckVersion{
				{Name: "v1"}, {Name: "v2"}, {Name: "v2"},
			},
			want: &duckHunter{
				mapper:          NewResourceMapper(nil),
				defaultVersions: []string{"v1", "v2"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
					"v2": {},
				},
				accesbileGroupresources: map[string]bool{},
			},
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewDuckHunter(tc.mapper, tc.versions, nil, nil); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewDuckHunter() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_DuckHunter_AddCRD(t *testing.T) {
	tests := map[string]struct {
		dh   DuckHunter
		crd  *apiextensionsv1.CustomResourceDefinition
		want map[string][]v1alpha1.ResourceMeta
	}{
		"one duck version, one crd version": {
			dh:  NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"clusterrole that can access duck": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"teach.me.how"},
					Resources: []string{"Duckys"},
				}},
			}),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: true,
				}},
			},
		},
		"clusterrole with wildcard Resource can access duck": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"teach.me.how"},
					Resources: []string{"*"},
				}},
			}),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: true,
				}},
			},
		},
		"clusterrole with wildcard APIGroups can access duck": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"*"},
					Resources: []string{"Duckys"},
				}},
			}),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: true,
				}},
			},
		},
		"clusterrole with wildcard Resources and APIGroups can access duck": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"*"},
					Resources: []string{"*"},
				}},
			}),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: true,
				}},
			},
		},
		"clusterrole that CANNOT access duck": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{"get"},
					APIGroups: []string{"teach.me.how"},
					Resources: []string{"Duckys"},
				}},
			}),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: false,
				}},
			},
		},
		"one duck version, two crd version": {
			dh:  NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true, "v3": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}, {
					APIVersion: "teach.me.how/v3",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"three duck defaultVersions, one crd version": {
			dh:  NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}, {Name: "v2"}, {Name: "v3"}}, nil, nil),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
				"v2": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
				"v3": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.dh.AddCRD(tc.crd)
			if got := tc.dh.Ducks(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_DuckHunter_AddCRD_filtered(t *testing.T) {
	tests := map[string]struct {
		dh   DuckHunter
		crd  *apiextensionsv1.CustomResourceDefinition
		want map[string][]v1alpha1.ResourceMeta
	}{
		"match filter": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v2": true},
				map[string]string{"teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"multi-match filter": {
			dh: NewDuckHunter(nil, []v1alpha1.DuckVersion{{Name: "v1"}, {Name: "v2"}}, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"red": true, "blue": true, "green": true},
				map[string]string{"teach.me.how/ducky": "true"},
				map[string]string{"duckies.teach.me.how/v1": "red,blue", "duckies.teach.me.how/v2": "green"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/blue",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}, {
					APIVersion: "teach.me.how/red",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
				"v2": {{
					APIVersion: "teach.me.how/green",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"match second label": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v2": true},
				map[string]string{"you.know.how": "whatever", "teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"reject match filter": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v2": true},
				map[string]string{"teach.me.how/ducky": "false"}, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: nil,
		},
		"no duck label": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v2": true},
				map[string]string{"totes": "unrelated"}, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"no labels": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v2": true},
				nil, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"one matching version of two": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v1": true, "v2": true},
				map[string]string{"teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v2"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"twp matching version of two": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v1": true, "v2": true},
				map[string]string{"teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v2", "duckies.teach.me.how/v1swag": "v1"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
				"v1swag": {{
					APIVersion: "teach.me.how/v1",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"no match": {
			dh: NewDuckHunter(nil, nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}, nil),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v1": true, "v2": true},
				map[string]string{"teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v3"}),
			want: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.dh.AddCRD(tc.crd)
			if got := tc.dh.Ducks(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_DuckHunter_AddRef(t *testing.T) {
	mapper := NewResourceMapper([]*metav1.APIResourceList{
		{
			GroupVersion: "teach.me.how/v2",
			APIResources: []metav1.APIResource{{
				Kind:       "Ducky",
				Name:       "duckies",
				Namespaced: false,
			}},
		}})

	tests := map[string]struct {
		dh          DuckHunter
		duckVersion string
		ref         v1alpha1.ResourceRef
		want        map[string][]v1alpha1.ResourceMeta
		wantErr     bool
	}{
		"GVK, no default duck type version": {
			dh:          NewDuckHunter(mapper, nil, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:   "teach.me.how",
				Version: "v2",
				Kind:    "Ducky",
				Scope:   "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"GVK, one duck type version": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:   "teach.me.how",
				Version: "v2",
				Kind:    "Ducky",
				Scope:   "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"ClusterRole that can access the duck": {
			dh: NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"teach.me.how"},
					Resources: []string{"Duckys"},
				}},
			}),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:   "teach.me.how",
				Version: "v2",
				Kind:    "Ducky",
				Scope:   "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion:               "teach.me.how/v2",
					Kind:                     "Ducky",
					Scope:                    "Namespaced",
					AccessibleViaClusterRole: true,
				}},
			},
		},
		"ClusterRole that CANNOT access the duck": {
			dh: NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{"watch"},
					APIGroups: []string{"teach.me.how"},
					Resources: []string{"Duckys"},
				}},
			}),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:   "teach.me.how",
				Version: "v2",
				Kind:    "Ducky",
				Scope:   "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"GVR, one duck type version": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:    "teach.me.how",
				Version:  "v2",
				Resource: "duckies",
				Scope:    "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"AK, one duck type version": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				APIVersion: "teach.me.how/v2",
				Kind:       "Ducky",
				Scope:      "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"AR, one duck type version": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				APIVersion: "teach.me.how/v2",
				Resource:   "duckies",
				Scope:      "Namespaced",
			},
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"GVK, unknown ref": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:   "already.know.how",
				Version: "v2",
				Kind:    "Ducky",
				Scope:   "Namespaced",
			},
			wantErr: true,
		},
		"GVR, known group, unknown resource": {
			dh:          NewDuckHunter(mapper, []v1alpha1.DuckVersion{{Name: "v1"}}, nil, nil),
			duckVersion: "v1",
			ref: v1alpha1.ResourceRef{
				Group:    "teach.me.how",
				Version:  "v2",
				Resource: "geesey",
				Scope:    "Namespaced",
			},
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.dh.AddRef(tc.duckVersion, tc.ref)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("expected error calling dh.AddRef: %v", err)
				}
			} else if got := tc.dh.Ducks(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_crdToResourceMeta(t *testing.T) {
	tests := map[string]struct {
		crd  *apiextensionsv1.CustomResourceDefinition
		want []v1alpha1.ResourceMeta
	}{
		"one crd version": {
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: []v1alpha1.ResourceMeta{{
				APIVersion: "teach.me.how/v2",
				Kind:       "Ducky",
				Scope:      "Namespaced",
			}},
		},
		"two crd version": {
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true, "v3": true}),
			want: []v1alpha1.ResourceMeta{{
				APIVersion: "teach.me.how/v2",
				Kind:       "Ducky",
				Scope:      "Namespaced",
			}, {
				APIVersion: "teach.me.how/v3",
				Kind:       "Ducky",
				Scope:      "Namespaced",
			}},
		},
		"three crd version, only one served": {
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v1": false, "v2": true, "v3": false}),
			want: []v1alpha1.ResourceMeta{{
				APIVersion: "teach.me.how/v2",
				Kind:       "Ducky",
				Scope:      "Namespaced",
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := crdToResourceMeta(tc.crd)

			sort.Sort(ByResourceMeta(got))
			sort.Sort(ByResourceMeta(tc.want))

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_accessibleGroupResources(t *testing.T) {
	tests := map[string]struct {
		cr            *rbacv1.ClusterRole
		expectedVerbs []string
		want          map[string]bool
	}{
		"nil clusterrole": {
			cr:   nil,
			want: map[string]bool{},
		},
		"clusterrole with no rules": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{},
			},
			want: map[string]bool{},
		},
		"clusterrole with rule satisfying list of verbs": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{"get", "list"},
					APIGroups: []string{"animal"},
					Resources: []string{"fox"},
				}},
			},
			want: map[string]bool{
				"animal:fox": true,
			},
			expectedVerbs: []string{"get", "list"},
		},
		"clusterrole with rule NOT satisfying list of verbs": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{"list"},
					APIGroups: []string{"animal"},
					Resources: []string{"fox"},
				}},
			},
			want:          map[string]bool{},
			expectedVerbs: []string{"get", "list"},
		},
		"clusterrole with multiple rules satisfying list of verbs": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"animal"},
						Resources: []string{"fox"},
					},
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"mammal"},
						Resources: []string{"deer"},
					},
					{
						Verbs:     []string{"get"},
						APIGroups: []string{"animal"},
						Resources: []string{"chicken"},
					},
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"cell"},
					},
				},
			},
			want: map[string]bool{
				"animal:fox":  true,
				"mammal:deer": true,
				"*:cell":      true,
			},
			expectedVerbs: []string{"get", "list"},
		},
		"clusterrole with rule with multiple apiGroups and resources": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{"list", "get", "watch"},
					APIGroups: []string{"animal", "mammal"},
					Resources: []string{"fox", "deer"},
				}},
			},
			want: map[string]bool{
				"animal:fox":  true,
				"animal:deer": true,
				"mammal:fox":  true,
				"mammal:deer": true,
			},
			expectedVerbs: []string{"get", "list"},
		},
		"clusterrole with rule with wildcard verbAll": {
			cr: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{{
					Verbs:     []string{rbacv1.VerbAll},
					APIGroups: []string{"animal"},
					Resources: []string{"fox"},
				}},
			},
			want: map[string]bool{
				"animal:fox": true,
			},
			expectedVerbs: []string{"get", "list"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := accessibleGroupResources(tc.expectedVerbs, tc.cr)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("accessibleGroupResources() = %v, want %v", got, tc.want)
			}
		})
	}
}
