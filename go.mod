module knative.dev/discovery

go 1.14

require (
	github.com/google/go-cmp v0.5.1
	github.com/google/licenseclassifier v0.0.0-20200708223521-3d09a0ea2f39
	github.com/n3wscott/rigging v0.0.1
	go.uber.org/zap v1.14.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.0
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	knative.dev/pkg v0.0.0-20200723220257-e58afb06b774
	knative.dev/test-infra v0.0.0-20200723182457-517b66ba19c1
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)
