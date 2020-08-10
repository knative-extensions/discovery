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

	ducktypeinformer "knative.dev/discovery/pkg/client/injection/informers/discovery/v1alpha1/clusterducktype"
	ducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	crdinformer "knative.dev/pkg/client/injection/apiextensions/informers/apiextensions/v1/customresourcedefinition"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	ducktypeInformer := ducktypeinformer.Get(ctx)
	crdInformer := crdinformer.Get(ctx)

	r := &Reconciler{
		client:    kubeclient.Get(ctx),
		crdLister: crdInformer.Lister(),
	}
	r.resyncResourceMapper(ctx)
	impl := ducktypereconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	ducktypeInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Watch custom resource definitions.
	grDt := func(obj interface{}) {
		r.resyncResourceMapper(ctx)
		impl.GlobalResync(ducktypeInformer.Informer())
	}
	crdInformer.Informer().AddEventHandler(controller.HandleAll(grDt))

	return impl
}

type ResourceMapping struct {
	ResourceToKind map[string]string
	KindToResource map[string]string
}
