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

package clusterducktype

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionslisters "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"knative.dev/discovery/pkg/collection"

	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	ducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements ducktypereconciler.Interface for
// ClusterDuckType resources.
type Reconciler struct {
	client    kubernetes.Interface
	crdLister apiextensionslisters.CustomResourceDefinitionLister

	resourceMapper collection.ResourceMapper
	rmx            sync.Mutex
}

// Check that our Reconciler implements Interface
var _ ducktypereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface
func (r *Reconciler) ReconcileKind(ctx context.Context, dt *v1alpha1.ClusterDuckType) reconciler.Event {
	// Make a safe copy of the resource mapper.
	r.rmx.Lock()
	rm := r.resourceMapper.DeepCopy()
	r.rmx.Unlock()

	// Set up this instance of a duck hunter.
	hunter := collection.NewDuckHunter(rm, nil, &collection.DuckFilters{
		DuckLabel:         fmt.Sprintf("%s/%s", dt.Spec.Group, dt.Spec.Names.Singular),
		DuckVersionPrefix: fmt.Sprintf("%s.%s", dt.Spec.Names.Plural, dt.Spec.Group),
	})

	/// By query

	for _, st := range dt.Spec.Selectors {
		crds, err := r.getCRDsWith(st.LabelSelector)
		if err != nil {
			// TODO: this should be a condition that reports back that RBAC is incorrect for getting CRDs?
			return err
		}
		hunter.AddCRDs(crds)
	}

	/// By ref

	for _, dv := range dt.Spec.Versions {
		for _, ref := range dv.Refs {
			// TODO we should query and test that the Ref is installed and works on this cluster.
			if err := hunter.AddRef(dv.Name, ref); err != nil {
				logging.FromContext(ctx).Warnw("unable to add resource ref: %s", zap.Error(err))
			}
		}
	}

	dt.Status.Ducks = hunter.Ducks()
	dt.Status.DuckCount = DuckCount(dt.Status.Ducks)
	dt.Status.MarkReady()
	return nil
}

// getCRDsWith returns CRDs labeled as given.
// labelSelector should be in the form "<group>/<names.singular>=true"
func (r *Reconciler) getCRDsWith(labelSelector string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	ls, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}

	list, err := r.crdLister.List(ls)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// resyncResourceMapper will make a call to the Kubernetes APIServer to request
// the full list of resources on this cluster and then process the list to
// create a lookup table between GroupVersions, Kinds and Resources.
func (r *Reconciler) resyncResourceMapper(ctx context.Context) {
	_, apiResources, err := r.client.Discovery().ServerGroupsAndResources()
	if err != nil {
		logging.FromContext(ctx).Errorf("Failed to resync resource mapper.", zap.Error(err))
		return
	}

	r.rmx.Lock()
	r.resourceMapper = collection.NewResourceMapper(apiResources)
	r.rmx.Unlock()
}

// DuckCount de-dupes the number of ducks inside the mapped collection of found
// duck types. Some resources could apply to several duck types, throwing the
// count off in the status of ClusterDuckType.
func DuckCount(ducks map[string][]v1alpha1.ResourceMeta) int {
	count := 0
	kinds := make(map[string]bool, 0)
	for _, metas := range ducks {
		for _, meta := range metas {
			// TODO: it would be nice if ResourceMeta had a version-free unique hash to do this.
			key := strings.ToLower(fmt.Sprintf("%s-%s", group(meta), meta.Kind))
			if _, found := kinds[key]; !found {
				kinds[key] = true
				count++
			}
		}
	}
	return count
}

// group returns the correct group for a ResourceMeta.
// TODO: it might be better to have meta be a GVK, and display the APIVersion.
func group(meta v1alpha1.ResourceMeta) string {
	if strings.Contains(meta.APIVersion, "/") {
		sp := strings.Split(meta.APIVersion, "/")
		group := strings.TrimSuffix(meta.APIVersion, sp[len(sp)-1])
		return group
	}
	return ""
}
