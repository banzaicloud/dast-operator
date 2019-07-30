/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingress

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/zaproxy/zap-api-go/zap"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/dast-operator/pkg/k8sutil"
)

func validate(ar *admissionv1beta1.AdmissionReview, log logr.Logger, client client.Client) *admissionv1beta1.AdmissionResponse {
	req := ar.Request
	log.Info("AdmissionReview for", "Kind", req.Kind, "Namespsce", req.Namespace, "Resource", req.Resource, "UserInfo", req.UserInfo)

	switch req.Kind.Kind {
	case "Ingress":
		var ingress extv1beta1.Ingress
		if err := json.Unmarshal(req.Object.Raw, &ingress); err != nil {
			log.Error(err, "could not unmarshal raw object")
			return &admissionv1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		backendServices := k8sutil.GetIngressBackendServices(&ingress, log)
		log.Info("Services", "backend_services", backendServices)

		if !isServicesAnnoteted(backendServices, ingress.GetNamespace(), log, client) {
			return &admissionv1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Reason: "backend service isn't annotated",
				},
			}
		}
	}

	result := &metav1.Status{
		Reason: "validating result false",
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: false,
		Result:  result,
	}
}

func isServicesAnnoteted(services []map[string]string, namespace string, log logr.Logger, client client.Client) bool {
	for _, service := range services {

		_, err := k8sutil.GetServiceByName(service["name"], namespace, client)
		if err != nil {
			log.Error(err, "unable to get service")
		}
	}
	return false
}

func getServiceScanSummary(serviceName string, zapCore *zap.Core) map[string]string {
	// target := fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", service["name"], namespace, service["port"])
	// log.Info("Target", "url", target)
	return nil
}

func getIngressTresholds(ingress *extv1beta1.Ingress) map[string]string {
	return nil
}

func validateAgainstTreshold(summary, tershold map[string]string) {

}

func newZapClient(zapAddr string, apiKey string, log logr.Logger) (*zap.Core, error) {
	cfg := &zap.Config{
		Proxy:  zapAddr,
		APIKey: apiKey,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		return nil, emperror.Wrap(err, "annot create zap interface")
	}
	return client.Core(), nil
}
