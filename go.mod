module knative.dev/discovery

go 1.15

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cucumber/godog v0.10.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/google/go-cmp v0.5.5
	github.com/google/licenseclassifier v0.0.0-20200708223521-3d09a0ea2f39
	github.com/sergi/go-diff v1.1.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/code-generator v0.19.7
	k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6
	knative.dev/hack v0.0.0-20210309141825-9b73a256fd9a
	knative.dev/pkg v0.0.0-20210315160101-6a33a1ab29ac
	knative.dev/reconciler-test v0.0.0-20210317002242-cfb4db568965
)

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
