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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"

	discoveryv1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	ducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements ducktypereconciler.Interface for
// ClusterDuckType resources.
type Reconciler struct {
	crdLister apiextensionsv1.CustomResourceDefinitionLister
}

// Check that our Reconciler implements Interface
var _ ducktypereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *discoveryv1alpha1.ClusterDuckType) reconciler.Event {
	logger := logging.FromContext(ctx)

	crd, err := r.crdLister.Get("channels.messaging.knative.dev") // TODO generalize past testing
	if err != nil {
		logger.Errorf("error getting crd: %q", err)
	} else {
		logger.Errorf("found crd: %q", crd)
	}

	// TODO: remove this when there is real logic here.
	o.Status.MarkReady()

	return nil
}
