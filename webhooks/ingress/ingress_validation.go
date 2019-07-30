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
	"fmt"

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

		ok, err := checkServices(backendServices, ingress.GetNamespace(), log, client)
		// TODO reason error or failed check
		if err != nil {
			return &admissionv1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Reason: metav1.StatusReason(err.Error()),
				},
			}
		}
		if !ok {
			return &admissionv1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Reason: "not OK",
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

func checkServices(services []map[string]string, namespace string, log logr.Logger, client client.Client) (bool, error) {
	for _, service := range services {
		k8sService, err := k8sutil.GetServiceByName(service["name"], namespace, client)
		if err != nil {
			return false, err
		}
		zapProxyCfg, err := k8sutil.GetServiceAnotations(k8sService, log)
		if err != nil {
			return false, err
		}
		secret, err := k8sutil.GetSercretByName(zapProxyCfg["name"], zapProxyCfg["namespace"], client, log)
		if err != nil {
			return false, err
		}
		zapCore, err := newZapClient(zapProxyCfg["name"], string(secret.Data["zap_api_key"]), log)
		if err != nil {
			return false, err
		}
		getServiceScanSummary(service, namespace, zapCore, log)
	}
	return true, nil
}

func getServiceScanSummary(service map[string]string, namespace string, zapCore *zap.Core, log logr.Logger) map[string]string {
	target := fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", service["name"], namespace, service["port"])
	log.Info("Target", "url", target)

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
