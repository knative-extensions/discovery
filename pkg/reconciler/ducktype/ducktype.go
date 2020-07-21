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

package ducktype

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	discoveryv1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	ducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/ducktype"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason DuckTypeReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "DuckTypeReconciled", "DuckType reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements ducktypereconciler.Interface for
// DuckType resources.
type Reconciler struct {
}

// Check that our Reconciler implements Interface
var _ ducktypereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *discoveryv1alpha1.DuckType) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Debug("TODO: implement this.")

	return newReconciledNormal(o.Namespace, o.Name)
}
