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
	"log"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	"knative.dev/discovery/test/e2e/config/clusterrole"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
)

func ClusterRole() *feature.Feature {
	f := new(feature.Feature)

	f.Setup("install a simple ClusterDuckType with a role ref", clusterrole.Install())

	f.Alpha("for a single ClusterDuckType").
		Must("assertions are satisfied", clusterRoleAssert)

	return f
}

func clusterRoleAssert(ctx context.Context, t feature.T) {
	env := environment.FromContext(ctx)
	for _, ref := range env.References() {
		if err := k8s.WaitForReadyOrDone(ctx, ref, interval, timeout); err != nil {
			t.Fatalf("failed to wait for ready or done, %s", err)
		}

		k := ref.GroupVersionKind()
		gvr, _ := meta.UnsafeGuessKindToResource(k)
		like := &v1alpha1.ClusterDuckType{}
		client := dynamicclient.Get(ctx)

		us, err := client.Resource(gvr).Namespace(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("failed to get ref, %s", ref.Name)
		}

		obj := like.DeepCopy()
		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(us.Object, obj); err != nil {
			log.Fatalf("Error DefaultUnstructuree.Dynamiconverter. %v", err)
		}

		if _, ok := obj.Status.ClusterRoleAggregationRule.ClusterRoleSelectors[0].MatchLabels["rbac.authorization.k8s.io/aggregate-to-view"]; !ok {
			t.Fatalf("ClusterDuckType status doesn't include ClusterRole's AggregationRule %#v", obj.Status.ClusterRoleAggregationRule)
		}

		if obj.Status.Ducks["v1"][0].AccessibleViaClusterRole != false {
			t.Fatalf("ClusterDuckType's duck CRD shouldn't be accessibleByClusterRole %#v", obj.Status.Ducks["v1"][0])
		}
	}
}
