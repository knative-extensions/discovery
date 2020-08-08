package collection

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	"reflect"
	"testing"
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

func TestNewDuckCollector(t *testing.T) {
	tests := map[string]struct {
		versions []v1alpha1.DuckVersion
		want     *duckHunter
	}{
		"one versions": {
			versions: []v1alpha1.DuckVersion{
				{Name: "v1"},
			},
			want: &duckHunter{
				versions: []string{"v1"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
				},
			},
		}, "three versions": {
			versions: []v1alpha1.DuckVersion{{Name: "v1"}, {Name: "v2"}, {Name: "v3"}},
			want: &duckHunter{
				versions: []string{"v1", "v2", "v3"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
					"v2": {},
					"v3": {},
				},
			},
		},
		"overlapping versions": {
			versions: []v1alpha1.DuckVersion{
				{Name: "v1"}, {Name: "v2"}, {Name: "v2"},
			},
			want: &duckHunter{
				versions: []string{"v1", "v2"},
				ducks: map[string][]v1alpha1.ResourceMeta{
					"v1": {},
					"v2": {},
				},
			},
		}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewDuckHunter(tc.versions, nil); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewDuckHunter() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_duckCollector_AddCRD(t *testing.T) {
	tests := map[string]struct {
		dc   DuckHunter
		crd  *apiextensionsv1.CustomResourceDefinition
		want map[string][]v1alpha1.ResourceMeta
	}{
		"one duck version, one crd version": {
			dc:  NewDuckHunter([]v1alpha1.DuckVersion{{Name: "v1"}}, nil),
			crd: makeCRD("teach.me.how", "Ducky", map[string]bool{"v2": true}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {{
					APIVersion: "teach.me.how/v2",
					Kind:       "Ducky",
					Scope:      "Namespaced",
				}},
			},
		},
		"one duck version, two crd version": {
			dc:  NewDuckHunter([]v1alpha1.DuckVersion{{Name: "v1"}}, nil),
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
		"three duck versions, one crd version": {
			dc:  NewDuckHunter([]v1alpha1.DuckVersion{{Name: "v1"}, {Name: "v2"}, {Name: "v3"}}, nil),
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
			tc.dc.AddCRD(tc.crd)
			if got := tc.dc.Ducks(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_duckCollector_AddCRD_filtered(t *testing.T) {
	tests := map[string]struct {
		dc   DuckHunter
		crd  *apiextensionsv1.CustomResourceDefinition
		want map[string][]v1alpha1.ResourceMeta
	}{
		"match filter": {
			dc: NewDuckHunter(nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}),
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
		"one matching version of two": {
			dc: NewDuckHunter(nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}),
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
			dc: NewDuckHunter(nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}),
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
			dc: NewDuckHunter(nil, &DuckFilters{
				DuckLabel:         "teach.me.how/ducky",
				DuckVersionPrefix: "duckies.teach.me.how",
			}),
			crd: makeCRDAnnotated("teach.me.how", "Ducky", map[string]bool{"v1": true, "v2": true},
				map[string]string{"teach.me.how/ducky": "true"}, map[string]string{"duckies.teach.me.how/v1": "v3"}),
			want: map[string][]v1alpha1.ResourceMeta{
				"v1": {},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.dc.AddCRD(tc.crd)
			if got := tc.dc.Ducks(); !reflect.DeepEqual(got, tc.want) {
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
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Ducks() = %v, want %v", got, tc.want)
			}
		})
	}
}
