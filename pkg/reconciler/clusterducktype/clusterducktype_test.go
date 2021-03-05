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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/discovery/pkg/collection"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"knative.dev/discovery/pkg/client/injection/client"
	"knative.dev/discovery/pkg/reconciler/testing/featured"
	. "knative.dev/discovery/pkg/reconciler/testing/v1alpha1"
)

// hardcode this for now... we can get this from CRDs in the system
var apiGroups = []*metav1.APIResourceList{
	{
		GroupVersion: "central.america/v2",
		APIResources: []metav1.APIResource{{
			Name:       "monkeys",
			Namespaced: true,
			Kind:       "Monkey",
		}},
	}, {
		GroupVersion: "north.america/v1alpha1",
		APIResources: []metav1.APIResource{{
			Name:       "ducks",
			Namespaced: true,
			Kind:       "Ducks",
		}},
	}, {
		GroupVersion: "north.america/v1alpha2",
		APIResources: []metav1.APIResource{{
			Name:       "ducks",
			Namespaced: true,
			Kind:       "Ducks",
		}},
	}, {
		GroupVersion: "north.america/v1beta1",
		APIResources: []metav1.APIResource{{
			Name:       "ducks",
			Namespaced: true,
			Kind:       "Ducks",
		}},
	}, {
		GroupVersion: "north.america/v2",
		APIResources: []metav1.APIResource{{
			Name:       "gilamonsters",
			Namespaced: false,
			Kind:       "GilaMonster",
		}},
	}, {
		GroupVersion: "australia/v1alpha1",
		APIResources: []metav1.APIResource{{
			Name:       "platypi",
			Namespaced: true,
			Kind:       "Platypus",
		}},
	}, {
		GroupVersion: "australia/v1alpha2",
		APIResources: []metav1.APIResource{{
			Name:       "platypi",
			Namespaced: true,
			Kind:       "Platypus",
		}},
	}, {
		GroupVersion: "australia/v1beta1",
		APIResources: []metav1.APIResource{{
			Name:       "platypi",
			Namespaced: true,
			Kind:       "Platypus",
		}},
	}, {
		GroupVersion: "australia/v1",
		APIResources: []metav1.APIResource{{
			Name:       "platypi",
			Namespaced: true,
			Kind:       "Platypus",
		}},
	},
}

func TestMain(m *testing.M) {
	featured.Run(m)
}

func TestReconcileKind(t *testing.T) {
	featured.TestReconcileKind(t, "ClusterDuckType", MakeFactory(func(ctx context.Context, listers *Listers, watcher configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			client:         fakekubeclient.Get(ctx),
			crdLister:      listers.GetCustomResourceDefinitionLister(),
			resourceMapper: collection.NewResourceMapper(apiGroups),
		}
		return clusterducktype.NewReconciler(ctx, logging.FromContext(ctx),
			client.Get(ctx), listers.GetClusterDuckTypeLister(),
			controller.GetEventRecorder(ctx), r)
	}))
}
