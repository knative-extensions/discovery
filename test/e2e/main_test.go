//go:build e2e
// +build e2e

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
	"flag"
	"os"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"knative.dev/pkg/injection"
	"knative.dev/reconciler-test/pkg/environment"
)

var global environment.GlobalEnvironment

func init() {
	environment.InitFlags(flag.CommandLine)
}

func TestMain(m *testing.M) {
	flag.Parse()

	ctx, startInformers := injection.EnableInjectionOrDie(nil, nil) //nolint
	startInformers()

	global = environment.NewGlobalEnvironment(ctx)

	os.Exit(m.Run())
}

// TestSmoke makes sure a ClusterDuckType goes ready.
func TestSmoke(t *testing.T) {
	t.Parallel()
	ctx, env := global.Environment()
	env.Test(ctx, t, ClusterDuckTypeSmoke())
	env.Finish()
}

// TestClusterRole makes sure a ClusterRole with an AggregationRule can be specified on a ClusterDuckType
func TestClusterRole(t *testing.T) {
	t.Parallel()
	ctx, env := global.Environment()
	env.Test(ctx, t, ClusterRole())
	env.Finish()
}
