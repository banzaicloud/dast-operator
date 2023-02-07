module github.com/banzaicloud/dast-operator

go 1.16

require (
	emperror.dev/emperror v0.33.0
	emperror.dev/errors v0.8.0
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/spf13/cast v1.3.0
	github.com/zaproxy/zap-api-go v0.0.0-20200721180916-5fc7048efb18
	istio.io/pkg v0.0.0-20200603210349-955e16c6198a
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
	sigs.k8s.io/controller-runtime v0.6.4
)
