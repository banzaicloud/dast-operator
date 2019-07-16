module github.com/banzaicloud/dast-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/goph/emperror v0.17.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9 // indirect
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/sys v0.0.0-20181205085412-a5c9d58dba9a // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
)

replace github.com/banzaicloud/dast-operator/cmd/dynamic-analyzer => ./cmd/dynamic-analyzer
