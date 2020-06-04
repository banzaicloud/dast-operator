module github.com/banzaicloud/dast-operator/cmd/dynamic-analyzer

go 1.13

require (
	github.com/spf13/cobra v1.0.0
	github.com/zaproxy/zap-api-go v0.0.0-20180130105416-8779ab35e992
)

replace github.com/zaproxy/zap-api-go => github.com/pbalogh-sa/zap-api-go v0.0.0-20200603214217-3acd33985b93
