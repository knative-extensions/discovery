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

package sample

import (
	"context"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"

	addressaleinformer "knative.dev/pkg/client/injection/ducks/duck/v1/addressable"
)

const (
	// ReconcilerName is the name of the reconciler.
	ReconcilerName = "SampleDuck"
)

// NewController returns a function that initializes the controller and
// Registers event handlers to enqueue events
func NewController(_ context.Context, gvk schema.GroupVersionKind) injection.ControllerConstructor {
	return func(ctx context.Context,
		cmw configmap.Watcher,
	) *controller.Impl {
		logger := logging.FromContext(ctx)

		// Get the addressable informer for duck types.
		addressableduckInformer := addressaleinformer.Get(ctx)
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		addressableInformer, addressableLister, err := addressableduckInformer.Get(ctx, gvr)
		if err != nil {
			logger.Errorw("Error getting source informer", zap.String("GVR", gvr.String()), zap.Error(err))
			return nil
		}

		_ = addressableInformer

		r := &Reconciler{
			lister: addressableLister,
		}

		impl := controller.NewImplFull(r, controller.ControllerOptions{WorkQueueName: ReconcilerName + gvr.String(), Logger: logger})

		logger.Info("Setting up event handlers")

		// Watch for all updates for the addressable.
		addressableInformer.AddEventHandler(controller.HandleAll(impl.Enqueue))

		return nil
	}
}
