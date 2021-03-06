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
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	pkgtest "knative.dev/pkg/reconciler/testing"

	"knative.dev/discovery/pkg/client/injection/client"
	"knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/discovery/pkg/reconciler/testing/featured"

	. "knative.dev/discovery/pkg/reconciler/testing/v1alpha1"
)

func TestMain(m *testing.M) {
	featured.Run(m)
}

func TestReconcileKind(t *testing.T) {
	var r clusterducktype.Interface

	featured.TestReconcileKind(t, "ClusterDuckType", nil, featured.Step{
		Expr: `^a "([^"]*)" metareconciler$`,
		StepFuncCtor: func(rt *featured.ReconcilerTest) interface{} {
			return func(dkv string) error {
				fmt.Println("using duck type", dkv)
				rt.Factory = MakeFactory(func(ctx context.Context, listers *Listers, watcher configmap.Watcher) controller.Reconciler {
					r = New(ctx, watcher, dkv, newSampleController)

					return clusterducktype.NewReconciler(ctx, logging.FromContext(ctx),
						client.Get(ctx), listers.GetClusterDuckTypeLister(),
						controller.GetEventRecorder(ctx), r)
				})
				return nil
			}
		},
	}, featured.Step{
		Expr: `^expect (\d) controllers$`,
		StepFuncCtor: func(rt *featured.ReconcilerTest) interface{} {
			return func(want int) error {
				rt.Row.PostConditions = append(rt.Row.PostConditions, func(t *testing.T, row *pkgtest.TableRow) {
					dt := r.(*Reconciler)
					if got := len(dt.controllers); got != want {
						t.Errorf("want %d controllers, got %d", want, got)
					}
				})
				return nil
			}
		},
	})
}

func newSampleController(gvk schema.GroupVersionKind) injection.ControllerConstructor {
	return func(ctx context.Context,
		cmw configmap.Watcher,
	) *controller.Impl {
		r := &sampleReconciler{}
		return controller.NewImplFull(r, controller.ControllerOptions{
			WorkQueueName: "sampleReconciler" + gvk.String(),
			Logger:        logging.FromContext(ctx),
		})
	}
}

type sampleReconciler struct{}

func (r *sampleReconciler) Reconcile(ctx context.Context, key string) error {
	logging.FromContext(ctx).Infof("Here I could do work with key: %+v\n", key)
	return nil
}
