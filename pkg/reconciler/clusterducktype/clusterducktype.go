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
	"github.com/go-openapi/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sort"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionslisters "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"

	v1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	ducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements ducktypereconciler.Interface for
// ClusterDuckType resources.
type Reconciler struct {
	crdLister apiextensionslisters.CustomResourceDefinitionLister
}

// Check that our Reconciler implements Interface
var _ ducktypereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface
func (r *Reconciler) ReconcileKind(ctx context.Context, dt *v1alpha1.ClusterDuckType) reconciler.Event {
	/// By query

	kinds := make(map[string]*apiextensionsv1.CustomResourceDefinition, 0)

	for _, st := range dt.Spec.Selectors {
		crds, err := r.getCRDsWith(st.LabelSelector)
		if err != nil {
			// TODO: this should be a condition that reports back that RBAC is incorrect for getting CRDs?
			return err
		}
		for _, crd := range crds {
			key := crd.Name
			if _, found := kinds[key]; !found {
				kinds[key] = crd
			}
		}
	}

	foundDucks := make([]v1alpha1.FoundDuck, 0)
	for _, crd := range kinds {
		foundDucks = append(foundDucks, CRDToFoundDuck("", crd))
	}

	/// By ref

	for _, dv := range dt.Spec.Versions {
		for _, ref := range dv.Refs {
			// TODO we should query and test that the Ref is installed and works on this cluster.
			foundDucks = append(foundDucks, v1alpha1.FoundDuck{
				DuckVersion: dv.Name,
				Ref:         ref,
			})
		}
	}

	// TODO: the above logic needs to pop out the label selected found ducks to
	// each version supported that is not overwritten. This will likely be
	// slightly complicated and needs a bunch of tests. Break it apart into
	// some smaller component that we can encapsulate those tests.

	// Sort and store.

	sort.Sort(ByFoundDuck(foundDucks))
	dt.Status.DuckList = foundDucks
	dt.Status.DuckCount = DuckCount(foundDucks)

	dt.Status.MarkReady()
	return nil
}

// getCRDsWith returns CRDs labeled as given.
// labelSelector should be in the form "duck.knative.dev/source=true"
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

// ByFoundDuck implements sort.Interface for []v1alpha1.FoundDuck based on
// the group and resource fields.
type ByFoundDuck []v1alpha1.ResourceMeta

func (a ByFoundDuck) Len() int      { return len(a) }
func (a ByFoundDuck) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFoundDuck) Less(i, j int) bool {
	// TODO: this needs to sort like this, but also group by duck version.
	keyI := fmt.Sprintf("%s-%s", a[i].APIVersion, a[i].Kind)
	keyJ := fmt.Sprintf("%s-%s", a[j].APIVersion, a[j].Kind)
	return keyI < keyJ
}

func DuckCount(ducks []v1alpha1.FoundDuck) int {
	// TODO: this is the wrong number, for now it is close enough.
	return len(ducks)
}
