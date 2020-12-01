module knative.dev/discovery

go 1.14

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cucumber/godog v0.10.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/google/go-cmp v0.5.2
	github.com/google/licenseclassifier v0.0.0-20200708223521-3d09a0ea2f39
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.12
	k8s.io/apiextensions-apiserver v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.12
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	knative.dev/hack v0.0.0-20201125230335-c46a6498e9ed
	knative.dev/pkg v0.0.0-20201130192436-e5346d90e980
	knative.dev/reconciler-test v0.0.0-20201124190335-83a44efcdfef
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)
