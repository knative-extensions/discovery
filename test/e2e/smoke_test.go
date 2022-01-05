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

package e2e

import (
	"context"

	"knative.dev/discovery/test/e2e/config/smoke"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
)

func ClusterDuckTypeSmoke() *feature.Feature {
	f := new(feature.Feature)

	f.Setup("install a simple ClusterDuckType", smoke.Install())

	f.Alpha("for a single ClusterDuckType").
		Must("goes ready", AllGoReady)

	return f
}

func AllGoReady(ctx context.Context, t feature.T) {
	env := environment.FromContext(ctx)
	for _, ref := range env.References() {
		if err := k8s.WaitForReadyOrDone(ctx, t, ref); err != nil {
			t.Fatalf("failed to wait for ready or done, %s", err)
		}
	}
}
