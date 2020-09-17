module github.com/n3wscott/rigging

go 1.14

require (
	github.com/google/go-cmp v0.5.2
	github.com/kelseyhightower/envconfig v1.4.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog v1.0.0
	knative.dev/pkg v0.0.0-20200909174550-69fcf75ee47d
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.18.8
	// WORKAROUND until k8s v1.18+ is in knative/pkg
	knative.dev/pkg => github.com/zroubalik/pkg v0.0.0-20200824111853-cf31d44b1119
)
