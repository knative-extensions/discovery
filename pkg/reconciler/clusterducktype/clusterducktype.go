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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionslisters "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/discovery/pkg/collection"

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
	hunter := collection.NewDuckHunter(nil, &collection.DuckFilters{
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
			hunter.AddRef(dv.Name, ref)
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

func DuckCount(ducks map[string][]v1alpha1.ResourceMeta) (count int) {
	kinds := make(map[string]bool, 0)
	for _, metas := range ducks {
		for _, meta := range metas {
			key := fmt.Sprintf("%s-%s", meta.APIVersion, meta.Kind)
			if _, found := kinds[key]; !found {
				kinds[key] = true
				count++
			}
		}
	}
	return count
}
