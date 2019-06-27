module github.com/banzaicloud/dast-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/goph/emperror v0.17.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/cobra v0.0.5
	github.com/zaproxy/zap-api-go v0.0.0-20180130105416-8779ab35e992
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)
