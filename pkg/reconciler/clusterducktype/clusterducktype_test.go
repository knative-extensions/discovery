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

package clusterducktype

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/n3wscott/rigging/pkg/installer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/cucumber/messages-go/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	client "knative.dev/discovery/pkg/client/injection/client"
	"knative.dev/discovery/pkg/client/injection/reconciler/discovery/v1alpha1/clusterducktype"
	"knative.dev/discovery/pkg/collection"
	. "knative.dev/discovery/pkg/reconciler/testing/v1alpha1"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	pkgtest "knative.dev/pkg/reconciler/testing"
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
}

var testStatus int

func TestMain(m *testing.M) {
	flag.Parse()

	if len(flag.Args()) > 0 {
		opt.Paths = flag.Args()
	} else {
		opt.Paths = []string{
			"./testdata/features/",
		}
	}

	format := "progress"
	for _, arg := range os.Args[1:] {
		if arg == "-test.v=true" { // go test transforms -v option
			format = "pretty"
			break
		}
	}

	opt.Format = format

	os.Exit(m.Run())
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
}

func TestReconcile(t *testing.T) {
	status := godog.TestSuite{
		Name:                 "ClusterDuckType",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			ClusterDuckTypeFeatureContext(t, s)
		},
		Options: &opt,
	}.Run()

	if status != 0 {
		t.Fail()
	}
}

func ClusterDuckTypeFeatureContext(t *testing.T, s *godog.ScenarioContext) {
	ctx := context.Background()

	rt := &ReconcilerTest{
		t: t,
		row: pkgtest.TableRow{
			Name: "ClusterDuckType",
			Ctx:  ctx,
			//OtherTestData:           nil,
			//Objects:                 nil,
			//Key:                     "",
			//WantErr:                 false,
			//WantCreates:             nil,
			//WantUpdates:             nil,
			//WantStatusUpdates:       nil,
			//WantDeletes:             nil,
			//WantDeleteCollections:   nil,
			//WantPatches:             nil,
			//WantEvents:              nil,
			//WithReactors:            nil,
			//SkipNamespaceValidation: false,
			//PostConditions:          nil,
			//Reconciler:              nil,
		}}

	s.Step(`^the following objects:$`, rt.theFollowingObjects)
	s.Step(`^the following objects \(from file\):$`, rt.theFollowingObjectFiles)
	s.Step(`^a ClusterDuckType reconciler$`, rt.aClusterDuckTypeReconciler)
	s.Step(`^reconciling "([^"]*)"$`, rt.reconcilingKey)
	s.Step(`^expect nothing$`, rt.expectNothing)
	s.Step(`^expect status updates:$`, rt.expectStatusUpdates)
	s.Step(`^expect status updates \(from file\):$`, rt.expectStatusUpdateFiles)
	s.Step(`^expect Kubernetes Events:$`, rt.expectKubernetesEvents)

	// hardcode this for now... we can get this from CRDs in the system
	rt.apiGroups = []*metav1.APIResourceList{
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

	s.AfterScenario(func(pickle *messages.Pickle, err error) {
		originObjects := make([]runtime.Object, 0, len(rt.row.Objects))
		for _, obj := range rt.row.Objects {
			originObjects = append(originObjects, obj.DeepCopyObject())
		}

		if rt.factory == nil {
			rt.t.Fatalf("factory is not set for test %s", rt.row.Name)
		}

		// Reconcile.
		rt.row.Name = pickle.Name
		rt.t.Run(pickle.Name, func(t *testing.T) {
			t.Helper()
			rt.row.Test(t, rt.factory)
		})

		// Validate cached objects do not get soiled after controller loops
		if diff := cmp.Diff(originObjects, rt.row.Objects, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
			rt.t.Errorf("Unexpected objects in test %s (-want, +got): %v", rt.row.Name, diff)
		}
	})

	_ = rt
}

type ReconcilerTest struct {
	t   *testing.T
	row pkgtest.TableRow

	apiGroups []*metav1.APIResourceList

	factory pkgtest.Factory
}

func (rt *ReconcilerTest) theFollowingObjectFiles(y *messages.PickleStepArgument_PickleTable) error {
	keys := make([]string, 0)
	for row, v := range y.Rows {
		var file string
		config := make(map[string]interface{}, 0)

		for i, c := range v.Cells {
			if row == 0 {
				if i == 0 {
					continue
				}
				keys = append(keys, c.Value)
				continue
			}

			switch i {
			case 0:
				file = "testdata/" + c.Value
			default:
				config[keys[i-1]] = c.Value
			}
		}

		// Leverage the installer to parse the template files.
		// TODO: move this code around to let it be reusable
		// in a more clear way.
		if file != "" {
			files := installer.ParseTemplates(file, config)
			list, err := ioutil.ReadDir(files)
			if err != nil {
				return err
			}
			// len zero would be an invalid or missing file.
			if len(list) == 0 {
				return fmt.Errorf("expected to read a yaml file from %q but found none", file)
			}

			for _, f := range list {
				name := path.Join(files, f.Name())
				if !f.IsDir() {
					ff, err := os.Open(name)
					if err != nil {
						return err
					}
					objs, err := ParseYAML(bufio.NewReader(ff))
					if err != nil {
						return err
					}

					rt.addObjects(objs)
				}
			}
		}
	}
	return nil
}

func (rt *ReconcilerTest) theFollowingObjects(y *messages.PickleStepArgument_PickleDocString) error {
	objs, err := ParseYAML(strings.NewReader(y.Content))
	if err != nil {
		return err
	}
	rt.addObjects(objs)

	return nil
}

func (rt *ReconcilerTest) addObjects(objs []unstructured.Unstructured) {
	if rt.row.Objects == nil {
		rt.row.Objects = make([]runtime.Object, 0, len(objs))
	}

	for _, obj := range objs {
		rt.row.Objects = append(rt.row.Objects, obj.DeepCopyObject())
	}

	rt.row.Objects = ToKnownObjects(rt.row.Objects)
}

func (rt *ReconcilerTest) aClusterDuckTypeReconciler() error {

	rt.factory = MakeFactory(func(ctx context.Context, listers *Listers, watcher configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			client:         fakekubeclient.Get(ctx),
			crdLister:      listers.GetCustomResourceDefinitionLister(),
			resourceMapper: collection.NewResourceMapper(rt.apiGroups),
		}

		return clusterducktype.NewReconciler(ctx, logging.FromContext(ctx),
			client.Get(ctx), listers.GetClusterDuckTypeLister(),
			controller.GetEventRecorder(ctx), r)
	})

	return nil
}

func (rt *ReconcilerTest) reconcilingKey(key string) error {
	// Set the reconciler key
	rt.row.Key = key

	return nil
}

func (rt *ReconcilerTest) expectNothing() error {
	return nil
}

func (rt *ReconcilerTest) expectStatusUpdates(y *messages.PickleStepArgument_PickleDocString) error {
	objs, err := ParseYAML(strings.NewReader(y.Content))
	if err != nil {
		return err
	}

	rt.addWantStatusUpdates(objs)
	return nil
}

func (rt *ReconcilerTest) expectStatusUpdateFiles(y *messages.PickleStepArgument_PickleTable) error {
	keys := make([]string, 0)
	for row, v := range y.Rows {
		var file string
		config := make(map[string]interface{}, 0)

		for i, c := range v.Cells {
			if row == 0 {
				if i == 0 {
					continue
				}
				keys = append(keys, c.Value)
				continue
			}

			switch i {
			case 0:
				file = "testdata/" + c.Value
			default:
				config[keys[i-1]] = c.Value
			}
		}

		// Leverage the installer to parse the template files.
		// TODO: move this code around to let it be reusable
		// in a more clear way.
		files := installer.ParseTemplates(file, config)
		if file != "" {
			list, err := ioutil.ReadDir(files)
			if err != nil {
				return err
			}

			for _, f := range list {
				name := path.Join(files, f.Name())
				if !f.IsDir() {
					ff, err := os.Open(name)
					if err != nil {
						return err
					}
					o, err := ParseYAML(bufio.NewReader(ff))
					if err != nil {
						return err
					}

					rt.addWantStatusUpdates(o)
				}
			}
		}
	}
	return nil
}

func (rt *ReconcilerTest) addWantStatusUpdates(objs []unstructured.Unstructured) {
	if rt.row.WantStatusUpdates == nil {
		rt.row.WantStatusUpdates = make([]clientgotesting.UpdateActionImpl, 0)
	}

	updates := make([]runtime.Object, 0, len(objs))
	for _, obj := range objs {
		updates = append(updates, obj.DeepCopyObject())
	}
	updates = ToKnownObjects(updates)
	for _, u := range updates {
		updateAction := clientgotesting.UpdateActionImpl{
			Object: u,
		}
		rt.row.WantStatusUpdates = append(rt.row.WantStatusUpdates, updateAction)
	}
}

func (rt *ReconcilerTest) expectKubernetesEvents(attributes *messages.PickleStepArgument_PickleTable) error {

	rt.row.WantEvents = make([]string, 0)

	for _, row := range attributes.Rows {
		eventType := row.Cells[0].Value
		reason := row.Cells[1].Value
		message := row.Cells[2].Value

		if eventType == "Type" {
			// ignore the headers
			continue
		}

		rt.row.WantEvents = append(rt.row.WantEvents, fmt.Sprintf(eventType+" "+reason+" "+message))
	}
	return nil
}

var (
	safeDeployDiff = cmpopts.IgnoreUnexported(resource.Quantity{})
)
