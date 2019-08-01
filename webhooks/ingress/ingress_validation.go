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
	"strconv"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/spf13/cast"
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
		tresholds := getIngressTresholds(&ingress)

		backendServices := k8sutil.GetIngressBackendServices(&ingress, log)
		log.Info("Services", "backend_services", backendServices)

		ok, err := checkServices(backendServices, ingress.GetNamespace(), log, client, tresholds)
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
					Reason: "scan results are above treshold",
				},
			}
		}
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: true,
		Result: &metav1.Status{
			Reason: "scan results are below treshold",
		},
	}
}

func checkServices(services []map[string]string, namespace string, log logr.Logger, client client.Client, tresholds map[string]int) (bool, error) {
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

		// TODO check scan status and wait for end of progress
		// check the scanner job is running, completed or not exist

		zapCore, err := newZapClient(zapProxyCfg["name"], zapProxyCfg["namespace"], string(secret.Data["zap_api_key"]), log)
		if err != nil {
			return false, err
		}
		summary, err := getServiceScanSummary(service, namespace, zapCore, log)
		if err != nil {
			return false, err
		}

		s, err := cast.ToStringMapIntE(summary["alertsSummary"])
		if err != nil {
			return false, err
		}
		for key, value := range s {
			if value > tresholds[key] {
				return false, nil
			}
		}
	}
	return true, nil
}

func getIngressTresholds(ingress *extv1beta1.Ingress) map[string]int {
	annotations := ingress.GetAnnotations()
	treshold := map[string]int{
		"High":          0,
		"Medium":        0,
		"Low":           0,
		"Informational": 0,
	}
	if high, ok := annotations["dast.security.banzaicloud.io/high"]; ok {
		treshold["High"], _ = strconv.Atoi(high)
	}
	if medium, ok := annotations["dast.security.banzaicloud.io/medium"]; ok {
		treshold["Medium"], _ = strconv.Atoi(medium)
	}
	if low, ok := annotations["dast.security.banzaicloud.io/low"]; ok {
		treshold["Low"], _ = strconv.Atoi(low)
	}
	if informational, ok := annotations["dast.security.banzaicloud.io/informational"]; ok {
		treshold["Informational"], _ = strconv.Atoi(informational)
	}
	return treshold
}

// TODO refactor to pkg
func getServiceScanSummary(service map[string]string, namespace string, zapCore *zap.Core, log logr.Logger) (map[string]interface{}, error) {
	target := fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", service["name"], namespace, service["port"])
	log.Info("Target", "url", target)
	summary, err := zapCore.AlertsSummary(target)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to get service summary from ZapProxy")
	}
	log.Info("Tresholds", "summary", summary)
	return summary, nil
}

func newZapClient(zapAddr, zapNamespace, apiKey string, log logr.Logger) (*zap.Core, error) {
	// TODO use https
	cfg := &zap.Config{
		Proxy:  "http://" + zapAddr + "." + zapNamespace + ".svc.cluster.local:8080",
		APIKey: apiKey,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to create zap interface")
	}
	return client.Core(), nil
}
