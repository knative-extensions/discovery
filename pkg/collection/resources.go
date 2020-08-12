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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceMapper can be used to convert between Resource and Kind or validate
// a Resource or Kind at a GroupVersion exists on the cluster.
type ResourceMapper interface {
	// KindExists returns true if the given kind is known for the given
	// groupVersion.
	KindExists(groupVersion, kind string) bool
	// ResourceExists returns true if the given resource is known for the given
	// GroupVersion.
	ResourceExists(groupVersion, resource string) bool
	// KindFor returns the Kind for the given resource in the given
	// GroupVersion.
	KindFor(groupVersion, resource string) (string, error)
	// ResourceFor returns the Kind for the given kind in the given
	// GroupVersion.
	ResourceFor(groupVersion, kind string) (string, error)

	// DeepCopy returns a copy of the ResourceMapper.
	DeepCopy() ResourceMapper
}

// NewResourceMapper processes a list of APIResourceLists and creates a
// ResourceMapper that can be used to convert between Resource and Kind or
// validate a Resource or Kind at a GroupVersion exists on the cluster.
func NewResourceMapper(apiGroups []*metav1.APIResourceList) ResourceMapper {
	mappings := make(map[string]mapping)

	for _, apiGroup := range apiGroups {
		mappings[apiGroup.GroupVersion] = mapping{
			r2k: make(map[string]string, 0),
			k2r: make(map[string]string, 0),
		}

		for _, apiResource := range apiGroup.APIResources {
			// the API groups contains all the endpoints, we don't want the
			//sub mappings for now. We will skip.
			if strings.Contains(apiResource.Name, "/") {
				continue
			}
			mappings[apiGroup.GroupVersion].k2r[apiResource.Kind] = apiResource.Name
			mappings[apiGroup.GroupVersion].r2k[apiResource.Name] = apiResource.Kind
		}
	}
	return &resourceMapper{mappings: mappings}
}

type resourceMapper struct {
	mappings map[string]mapping
}

type mapping struct {
	r2k map[string]string
	k2r map[string]string
}

// KindExists implements ResourceMapper.KindExists
func (rm *resourceMapper) KindExists(groupVersion, kind string) bool {
	m, found := rm.mappings[groupVersion]
	if !found {
		return false
	}
	_, found = m.k2r[kind]
	return found
}

// ResourceExists implements ResourceMapper.ResourceExists
func (rm *resourceMapper) ResourceExists(groupVersion, resource string) bool {
	m, found := rm.mappings[groupVersion]
	if !found {
		return false
	}
	_, found = m.r2k[resource]
	return found
}

// KindFor implements ResourceMapper.KindFor
func (rm *resourceMapper) KindFor(groupVersion, resource string) (string, error) {
	m, found := rm.mappings[groupVersion]
	if !found {
		return "", fmt.Errorf("kind not found for %s in %s", resource, groupVersion)
	}
	kind, found := m.r2k[resource]
	if !found {
		return "", fmt.Errorf("kind not found for %s in %s", resource, groupVersion)
	}
	return kind, nil
}

// ResourceFor implements ResourceMapper.ResourceFor
func (rm *resourceMapper) ResourceFor(groupVersion, kind string) (string, error) {
	m, found := rm.mappings[groupVersion]
	if !found {
		return "", fmt.Errorf("resource not found for %s in %s", kind, groupVersion)
	}
	resource, found := m.k2r[kind]
	if !found {
		return "", fmt.Errorf("resource not found for %s in %s", kind, groupVersion)
	}
	return resource, nil

}

// DeepCopy implements ResourceMapper.DeepCopy
func (rm *resourceMapper) DeepCopy() ResourceMapper {
	mappings := make(map[string]mapping, len(rm.mappings))
	for mk, mv := range rm.mappings {
		mappings[mk] = mapping{
			r2k: make(map[string]string, len(mv.r2k)),
			k2r: make(map[string]string, len(mv.k2r)),
		}
		for k, v := range mv.r2k {
			mappings[mk].r2k[k] = v
		}
		for k, v := range mv.k2r {
			mappings[mk].k2r[k] = v
		}
	}
	return &resourceMapper{mappings: mappings}
}
