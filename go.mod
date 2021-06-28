module knative.dev/discovery

go 1.16

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cucumber/godog v0.10.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/google/go-cmp v0.5.6
	github.com/google/licenseclassifier v0.0.0-20200708223521-3d09a0ea2f39
	github.com/sergi/go-diff v1.1.0 // indirect
	go.uber.org/zap v1.17.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/code-generator v0.20.7
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/hack/schema v0.0.0-20210622141627-e28525d8d260
	knative.dev/pkg v0.0.0-20210622173328-dd0db4b05c80
	knative.dev/reconciler-test v0.0.0-20210623134345-88c84739abd9
)

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
