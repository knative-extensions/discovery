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
	"sort"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
)

// DuckHunter is used to collect, sort and bucket Kubernetes mappings. This is
// based on adding CRDs and inspecting labels and annotations, or by adding
// references directly. The DuckHunter will determine which version of the
// duck type should be used for each kind of Add function. The current collection
// of ducks is retrieved by the Ducks() call.
type DuckHunter interface {
	// AddCRDs, add a list of CRDs, sorts based on the configuration of the duck hunter.
	AddCRDs(crds []*apiextensionsv1.CustomResourceDefinition)

	// AddCRD, add a single CRD, sorts based on the configuration of the duck hunter.
	AddCRD(crd *apiextensionsv1.CustomResourceDefinition)

	// AddRef is used to insert built-in resource types or types that are not
	// directly in control of the DuckType author but apply to the duck type.
	// Refs are used to override the label and annotation based discovery
	// mechanisms.
	// TODO: if the ref is a CRD, load the CRD and pass that CRD to AddCRD.
	AddRef(duckVersion string, ref v1alpha1.ResourceRef)

	// Ducks returns the current mapped collection of ducks added to the hunter.
	Ducks() map[string][]v1alpha1.ResourceMeta
}

type DuckFilters struct {
	// DuckLabel is expected to be in the form:
	//	 `<group>/<names.singular>`
	DuckLabel string
	// DuckVersionFormat is expected to be in the form:
	//   `<names.plural>.<group>`
	// and will be used to assemble `<names.plural>.<group>/<duckVersion>`.
	// This will group duck type versions based on the given annotation base.
	DuckVersionPrefix string
}

// NewDuckHunter
// versions are used to default all the DuckType versions that apply to unfiltered CRDs.
func NewDuckHunter(mapper ResourceMapper, versions []v1alpha1.DuckVersion, filters *DuckFilters) DuckHunter {
	dh := &duckHunter{
		mapper:   mapper,
		filters:  filters,
		versions: make([]string, 0),
		ducks:    make(map[string][]v1alpha1.ResourceMeta, len(versions)),
	}

	for _, v := range versions {
		if _, found := dh.ducks[v.Name]; !found {
			dh.versions = append(dh.versions, v.Name)
			dh.ducks[v.Name] = make([]v1alpha1.ResourceMeta, 0)
		}
	}

	return dh
}

// duckHunter is an internal implementation of the duck hunter logic.
type duckHunter struct {
	// mapper is used to convert between resource and kind, and validate
	// resources exist on the cluster.
	mapper ResourceMapper

	filters *DuckFilters
	// versions are the versions to use for unfiltered CRDs being added to the ducks map.
	versions []string
	ducks    map[string][]v1alpha1.ResourceMeta
}

// AddCRDs implements DuckHunter.AddCRDs
func (dh *duckHunter) AddCRDs(crds []*apiextensionsv1.CustomResourceDefinition) {
	for _, crd := range crds {
		dh.AddCRD(crd)
	}
}

// AddCRD implements DuckHunter.AddCRD
func (dh *duckHunter) AddCRD(crd *apiextensionsv1.CustomResourceDefinition) {
	if crd == nil {
		return
	}
	if metas := crdToResourceMeta(crd); len(metas) > 0 {
		dh.collectVersionsByFilter(crd)
		for _, meta := range metas {
			if !dh.addHandledWithFilters(crd, meta) {
				// TODO: here is where the label filters would be tested for duck version annotation matching.
				for _, v := range dh.versions {
					dh.ducks[v] = append(dh.ducks[v], meta)
				}
			}
		}
	}
}

// addHandledWithFilters attempts to add the CRD and resource meta to the ducks
// collection based on the current filters.
// Returning true means the handler handled the CRD.
func (dh *duckHunter) addHandledWithFilters(crd *apiextensionsv1.CustomResourceDefinition, meta v1alpha1.ResourceMeta) bool {
	if dh.filters == nil {
		return false
	}
	for k, v := range crd.Labels {
		if dh.filters.DuckLabel != "" {
			if k == dh.filters.DuckLabel {
				if v == "true" {
					return dh.insertHandledDuckByVersionFilter(crd, meta)
				}
				// Returning true here means the CRD was handled, but this
				// instance happen to not match the filters set. This CRD
				// instance should be skipped.
				return true
			}
		} else {
			return dh.insertHandledDuckByVersionFilter(crd, meta)
		}
	}
	return false
}

// collectVersionsByFilter this makes sure that there is a slice for each found version of the duck.
func (dh *duckHunter) collectVersionsByFilter(crd *apiextensionsv1.CustomResourceDefinition) {
	if dh.filters == nil || dh.filters.DuckVersionPrefix == "" {
		return
	}
	for k := range crd.Annotations {
		if strings.HasPrefix(k, dh.filters.DuckVersionPrefix+"/") {
			version := strings.TrimPrefix(k, dh.filters.DuckVersionPrefix+"/")
			if _, found := dh.ducks[version]; !found {
				dh.ducks[version] = make([]v1alpha1.ResourceMeta, 0)
			}
		}
	}
}

// insertHandledDuckByVersionFilter holds the logic to map a CRD to a duck type
// version based on the duck type version annotations, if present.
func (dh *duckHunter) insertHandledDuckByVersionFilter(crd *apiextensionsv1.CustomResourceDefinition, meta v1alpha1.ResourceMeta) (handled bool) {
	if dh.filters == nil || dh.filters.DuckVersionPrefix == "" {
		return false
	}
	for k, v := range crd.Annotations {
		if strings.HasPrefix(k, dh.filters.DuckVersionPrefix+"/") {
			// If we see any duck version prefix, then assume it is handled.
			// It could be not matching this meta version.
			handled = true
			// If the annotation says to use this meta version for the duck version, map it.
			if v == version(meta) {
				duckVersion := strings.TrimPrefix(k, dh.filters.DuckVersionPrefix+"/")
				dh.ducks[duckVersion] = append(dh.ducks[duckVersion], meta)
			}
		}
	}
	return
}

// AddRef implements DuckHunter.AddRef
func (dh *duckHunter) AddRef(duckVersion string, ref v1alpha1.ResourceRef) {
	if _, found := dh.ducks[duckVersion]; !found {
		dh.ducks[duckVersion] = make([]v1alpha1.ResourceMeta, 0)
	}
	// TODO: before this ref is added, we should validate the cluster understands this kind of resource at the version given.

	// To enable this, look at https://github.com/kubernetes/kubernetes/pull/42873/files and discoveryclient.ServerPreferredResources()

	dh.ducks[duckVersion] = append(dh.ducks[duckVersion], v1alpha1.ResourceMeta{
		APIVersion: v1alpha1.APIVersion(ref),
		Kind:       ref.Kind + ref.Resource, // TODO: FIXME this is a shortcoming of the api, we do not know Kind always. Might need to make a mapping of resource to Kind and keep it as a singleton lookup.
		Scope:      ref.Scope,
	})
}

// Ducks implements DuckHunter.Ducks
func (dh *duckHunter) Ducks() map[string][]v1alpha1.ResourceMeta {
	for v := range dh.ducks {
		sort.Sort(ByResourceMeta(dh.ducks[v]))
	}
	return dh.ducks
}

// crdToResourceMeta takes in a CRD and converts it to a set of ResourceMeta.
func crdToResourceMeta(crd *apiextensionsv1.CustomResourceDefinition) []v1alpha1.ResourceMeta {
	metas := make([]v1alpha1.ResourceMeta, 0)
	for _, v := range crd.Spec.Versions {
		if !v.Served {
			continue
		}

		metas = append(metas, v1alpha1.ResourceMeta{
			APIVersion: apiVersion(crd.Spec.Group, v.Name),
			Kind:       crd.Spec.Names.Kind,
			Scope:      v1alpha1.ResourceScope(crd.Spec.Scope),
		})
	}
	return metas
}

// apiVersion converts group and version to an APIVersion.
// TODO: might upstream this somewhere common.
func apiVersion(group, version string) string {
	if len(group) > 0 {
		return group + "/" + version
	}
	return version
}

// version inspects a ResourceMeta object and returns the correct APIVersion.
func version(meta v1alpha1.ResourceMeta) string {
	if strings.Contains(meta.APIVersion, "/") {
		sp := strings.Split(meta.APIVersion, "/")
		return sp[len(sp)-1]
	}
	return meta.APIVersion
}
