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

package clusterducktype

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	discoveryv1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	cluserterducktypereconciler "knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
)

func New(ctx context.Context, cmw configmap.Watcher, dkv string, ctor DuckTypedControllerConstructor) cluserterducktypereconciler.Interface {
	dp := strings.Split(dkv, "/")
	if len(dp) != 2 {
		panic("expecting duck kind version to be <ClusterDuckTypeName>/<DuckVersion>")
	}
	r := &Reconciler{
		forDuck:     dp[0],
		duckVersion: dp[1],
		ctorDuck:    ctor,
		ogctx:       ctx,
		ogcmw:       cmw,
		lock:        sync.Mutex{},
	}
	return r
}

type runningController struct {
	gvk        schema.GroupVersionKind
	controller *controller.Impl
	cancel     context.CancelFunc
}

type DuckTypedControllerConstructor func(gvk schema.GroupVersionKind) injection.ControllerConstructor

type Reconciler struct {
	// forDuck filters ClusterDuckType names.
	forDuck string
	// duckVersion picks a version in the ClusterDuckType
	duckVersion string

	// ctorDuck is the constructor to use to make controller constructors based
	// on gvk found from the ClusterDuckType.
	ctorDuck DuckTypedControllerConstructor

	ogctx context.Context
	ogcmw configmap.Watcher

	// Local state

	controllers map[string]runningController
	lock        sync.Mutex
}

func (r *Reconciler) enabledFor(dt *discoveryv1alpha1.ClusterDuckType) bool {
	return dt.GetName() == r.forDuck
}

// ReconcileKind reconciles cluster duck types, and for each that are enabled,
// it manages an instance of a controller created from the
// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, dt *discoveryv1alpha1.ClusterDuckType) reconciler.Event {
	// Jump out if this controller creation reconciler is not configured/enabled for this duck type.
	if !r.enabledFor(dt) {
		return nil
	}

	if r.controllers == nil {
		r.controllers = make(map[string]runningController)
	}

	for _, v := range dt.Status.Ducks[r.duckVersion] {
		key := fmt.Sprintf("%s.%s", v.Kind, v.Group())

		if rc, found := r.controllers[key]; !found {
			gvk := schema.GroupVersionKind{
				Group:   v.Group(),
				Version: v.Version(),
				Kind:    v.Kind,
			}
			cc := r.ctorDuck(gvk)

			atctx, cancel := context.WithCancel(r.ogctx)

			impl := cc(atctx, r.ogcmw)

			rc = runningController{
				gvk:        gvk,
				controller: impl,
				cancel:     cancel,
			}

			r.lock.Lock()
			r.controllers[key] = rc
			r.lock.Unlock()

			logging.FromContext(ctx).Infof("starting addressable reconciler for gvk %q", rc.gvk.String())
			go func(c *controller.Impl) {
				if err := c.Run(2, atctx.Done()); err != nil {
					logging.FromContext(ctx).Errorf("unable to start addressable reconciler for gvk %q", rc.gvk.String())
				}
			}(rc.controller)
		}
	}

	logging.FromContext(ctx).Debugf("----- Meta-Reconciling -------")
	logging.FromContext(ctx).Debugf("%s@%s", r.forDuck, r.duckVersion)
	for k := range r.controllers {
		logging.FromContext(ctx).Debugf(" - %q", k)
	}
	logging.FromContext(ctx).Debugf("------------------------------")
	return nil
}
