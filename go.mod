module github.com/banzaicloud/dast-operator

go 1.14

require (
	emperror.dev/emperror v0.32.0
	github.com/go-logr/logr v0.1.0
	github.com/goph/emperror v0.17.2
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.10.0
	github.com/spf13/cast v1.3.0
	github.com/zaproxy/zap-api-go v0.0.0-20200721180916-5fc7048efb18
	istio.io/pkg v0.0.0-20200603210349-955e16c6198a
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)
