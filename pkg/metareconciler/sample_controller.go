/*
Copyright 2021 The Knative Authors

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

package metareconciler

import (
	"context"
	"knative.dev/discovery/pkg/metareconciler/clusterducktype"
	"knative.dev/discovery/pkg/metareconciler/sample"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	clusterducktypeinformer "knative.dev/discovery/pkg/client/injection/informers/discovery/v1alpha1/clusterducktype"
	cluserterducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
)

// NewSampleController creates a Reconciler and returns the result of NewImpl.
// TODO document
func NewSampleController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	clusterducktypeInformer := clusterducktypeinformer.Get(ctx)

	r := clusterducktype.New(ctx, cmw, "addressables.duck.knative.dev/v1", sample.NewController)
	impl := cluserterducktypereconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	clusterducktypeInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
