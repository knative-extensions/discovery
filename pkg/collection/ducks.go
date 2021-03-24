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
	"fmt"
	"sort"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
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
	AddRef(duckVersion string, ref v1alpha1.ResourceRef) error

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
	// This will group duck type defaultVersions based on the given annotation base.
	DuckVersionPrefix string
}

// NewDuckHunter
// defaultVersions are used to default all the DuckType defaultVersions that apply to unfiltered CRDs.
func NewDuckHunter(mapper ResourceMapper, defaultVersions []v1alpha1.DuckVersion, filters *DuckFilters, clusterRole *rbacv1.ClusterRole) DuckHunter {
	if mapper == nil {
		mapper = NewResourceMapper(nil)
	}

	expectedVerbsForAccess := []string{"get", "watch", "list"}

	dh := &duckHunter{
		mapper:                  mapper,
		filters:                 filters,
		defaultVersions:         make([]string, 0),
		ducks:                   make(map[string][]v1alpha1.ResourceMeta, len(defaultVersions)),
		accesbileGroupresources: accessibleGroupResources(expectedVerbsForAccess, clusterRole),
	}

	for _, v := range defaultVersions {
		if _, found := dh.ducks[v.Name]; !found {
			dh.defaultVersions = append(dh.defaultVersions, v.Name)
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
	// defaultVersions are the defaultVersions to use for unfiltered CRDs being added to the ducks map.
	defaultVersions         []string
	ducks                   map[string][]v1alpha1.ResourceMeta
	accesbileGroupresources map[string]bool
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
				// If not handled within the filter aware handler, then apply
				// this resource to all the default duck versions.
				for _, v := range dh.defaultVersions {
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
		}
	}
	// Getting here means there is no labels or no matching duck label.
	return dh.insertHandledDuckByVersionFilter(crd, meta)
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
	for k, versions := range crd.Annotations {
		if strings.HasPrefix(k, dh.filters.DuckVersionPrefix+"/") {
			// If we see any duck version prefix, then assume it is handled.
			// It could be not matching this meta version.
			handled = true
			// If the annotation key contains the meta version for the duck version, map it.
			for _, v := range strings.Split(versions, ",") {
				v = strings.TrimSpace(v)
				if v == version(meta) {
					duckVersion := strings.TrimPrefix(k, dh.filters.DuckVersionPrefix+"/")
					dh.ducks[duckVersion] = append(dh.ducks[duckVersion], meta)
				}
			}
		}
	}
	return
}

// AddRef implements DuckHunter.AddRef
func (dh *duckHunter) AddRef(duckVersion string, ref v1alpha1.ResourceRef) error {
	// Use ref.Kind or look up the kind based on ref.Resource.
	kind := ref.Kind
	if kind == "" {
		var err error
		kind, err = dh.mapper.KindFor(ref.GroupVersion(), ref.Resource)
		if err != nil {
			return err
		}
	}

	rm := v1alpha1.ResourceMeta{
		APIVersion: ref.GroupVersion(),
		Kind:       kind,
		Scope:      ref.Scope,
	}

	// Validate that the resource exists in this cluster.
	if !dh.mapper.KindExists(rm.APIVersion, rm.Kind) {
		return fmt.Errorf("resource \"%s %s\" not known to the cluster", rm.Kind, rm.APIVersion)
	}

	// Save the resource at the given duck type version, making sure there is
	// a place to store it.
	if _, found := dh.ducks[duckVersion]; !found {
		dh.ducks[duckVersion] = []v1alpha1.ResourceMeta{rm}
	} else {
		dh.ducks[duckVersion] = append(dh.ducks[duckVersion], rm)
	}

	return nil
}

// duckCopy makes a deep copy of the ducks map
func duckCopy(d map[string][]v1alpha1.ResourceMeta) map[string][]v1alpha1.ResourceMeta {
	ducks := make(map[string][]v1alpha1.ResourceMeta, len(d))
	for k, v := range d {
		vc := make([]v1alpha1.ResourceMeta, len(v))
		copy(vc, v)
		ducks[k] = vc
	}
	return ducks
}

// Ducks implements DuckHunter.Ducks
func (dh *duckHunter) Ducks() map[string][]v1alpha1.ResourceMeta {
	ducks := duckCopy(dh.ducks)
	for k := range ducks {
		if len(ducks[k]) == 0 {
			delete(ducks, k)
		} else {
			sort.Sort(ByResourceMeta(ducks[k]))
			setAccessibleViaClusterRole(dh.accesbileGroupresources, ducks[k])
		}
	}
	if len(ducks) == 0 {
		return nil
	}
	return ducks
}

// setAccessibleViaClusterRole sets the AccessibleViaClusterRole flag on each duck if
//   the ClusterRole can preform the expected verbs on the duck
func setAccessibleViaClusterRole(accessibleGroupResources map[string]bool, metas []v1alpha1.ResourceMeta) {
	for index, meta := range metas {
		// TODO: it would be nice if ResourceMeta had a version-free unique hash to do this.
		key := strings.ToLower(fmt.Sprintf("%s:%ss", group(meta), meta.Kind))
		wildcardKey := "*:*"
		wildcardAPIGroupKey := strings.ToLower(fmt.Sprintf("*:%ss", meta.Kind))
		wildcardKindKey := strings.ToLower(fmt.Sprintf("%s:*", group(meta)))

		if _, ok := accessibleGroupResources[key]; ok {
			metas[index].AccessibleViaClusterRole = true
		} else if _, ok := accessibleGroupResources[wildcardKey]; ok {
			metas[index].AccessibleViaClusterRole = true
		} else if _, ok := accessibleGroupResources[wildcardAPIGroupKey]; ok {
			metas[index].AccessibleViaClusterRole = true
		} else if _, ok := accessibleGroupResources[wildcardKindKey]; ok {
			metas[index].AccessibleViaClusterRole = true
		}

	}
}

// group returns the correct group for a ResourceMeta.
// TODO: it might be better to have meta be a GVK, and display the APIVersion.
func group(meta v1alpha1.ResourceMeta) string {
	if strings.Contains(meta.APIVersion, "/") {
		sp := strings.Split(meta.APIVersion, "/")
		group := strings.TrimSuffix(meta.APIVersion, "/"+sp[len(sp)-1])
		return group
	}
	return ""
}

// crdToResourceMeta takes in a CRD and converts it to a set of ResourceMeta.
func crdToResourceMeta(crd *apiextensionsv1.CustomResourceDefinition) []v1alpha1.ResourceMeta {
	metas := make([]v1alpha1.ResourceMeta, 0)
	for _, v := range crd.Spec.Versions {
		if !v.Served {
			continue
		}

		// TODO: this will have issues for unaccepted CRDs
		// We need to look at the combo of crd.spec and crd.status
		// for this metadata.
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

// accessibleGroupResources finds the rules in the ClusterRole that satisfy the expectedVerbs
// and returns a map of
//   key: apiGroup + ":" + resource
//   value: true if Rule satisfies the expectedVerbs, false otherwise
func accessibleGroupResources(expectedVerbs []string, clusterRole *rbacv1.ClusterRole) map[string]bool {
	groupResources := map[string]bool{}
	if clusterRole == nil {
		return groupResources
	}
	for _, rule := range clusterRole.Rules {
		if isSubset(expectedVerbs, rule.Verbs) {
			for _, apiGroup := range rule.APIGroups {
				for _, resource := range rule.Resources {
					key := strings.ToLower(fmt.Sprintf("%s:%s", apiGroup, resource))
					groupResources[key] = true
				}
			}
		}
	}
	return groupResources
}

// Returns true if first is a fully contained subset of second
//   returns false otherwise
//   supports "*" as a wildcard
func isSubset(first, second []string) bool {
	set := map[string]bool{}
	for _, value := range second {
		// Support "*" verbAll wildcard
		if value == "*" {
			return true
		}
		set[value] = true
	}
	for _, value := range first {
		if !set[value] {
			return false
		}
	}
	return true
}
