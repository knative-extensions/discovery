/*
Copyright 2020 The Knative Authors.

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

package v1alpha1

import (
	"context"
	"fmt"
	"log"
	"reflect"

	kubev1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeapiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubev1lister "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	discoveryv1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	fakesampleclientset "knative.dev/discovery/pkg/client/clientset/versioned/fake"
	discoverylister "knative.dev/discovery/pkg/client/listers/discovery/v1alpha1"
	"knative.dev/pkg/reconciler/testing"
)

func AddConverters(scheme *runtime.Scheme) error {
	return scheme.AddConversionFunc(&unstructured.Unstructured{}, &discoveryv1alpha1.ClusterDuckType{}, func(a, b interface{}, scope conversion.Scope) error {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(a.(*unstructured.Unstructured).Object, b); err != nil {
			log.Fatalf("Error DefaultUnstructuredConverter.FromUnstructured. %v", err)
		}
		return nil
	})
}

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakesampleclientset.AddToScheme,
	fakeapiextensionsclientset.AddToScheme,
	AddConverters,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewListers(objs []runtime.Object) Listers {
	scheme := NewScheme()

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(ToKnownObjects(objs)...)

	return ls
}

func ToKnownObjects(objs []runtime.Object) []runtime.Object {
	scheme := NewScheme()

	known := make([]runtime.Object, 0)

	for _, obj := range objs {
		if reflect.TypeOf(obj) == reflect.TypeOf(&unstructured.Unstructured{}) { // I am sure there is a better way...
			kind := obj.GetObjectKind()
			if scheme.Recognizes(kind.GroupVersionKind()) {
				// Try to pop the kind out of unstructured.Unstructured.
				like, err := scheme.New(kind.GroupVersionKind())
				if err != nil {
					panic(err)
				}
				if err := scheme.Convert(obj, like, context.TODO()); err != nil {
					panic(err)
				}
				like.GetObjectKind().SetGroupVersionKind(kind.GroupVersionKind())
				known = append(known, like)

			} else {
				panic(fmt.Errorf("unregistered kind: %s", kind.GroupVersionKind().String()))
			}
		} else {
			known = append(known, obj)
		}
	}

	return known
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}
	return scheme
}

func (*Listers) NewScheme() *runtime.Scheme {
	return NewScheme()
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l *Listers) GetAPIExtentionsObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeapiextensionsclientset.AddToScheme)
}

func (l *Listers) GetCustomResourceDefinitionLister() kubev1lister.CustomResourceDefinitionLister {

	return kubev1lister.NewCustomResourceDefinitionLister(l.IndexerFor(&kubev1.CustomResourceDefinition{}))
}

func (l *Listers) GetDiscoveryObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakesampleclientset.AddToScheme)
}

func (l *Listers) GetClusterDuckTypeLister() discoverylister.ClusterDuckTypeLister {
	return discoverylister.NewClusterDuckTypeLister(l.IndexerFor(&discoveryv1alpha1.ClusterDuckType{}))
}
