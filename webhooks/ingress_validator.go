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

package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"emperror.dev/emperror"
	"github.com/banzaicloud/dast-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"github.com/spf13/cast"
	"github.com/zaproxy/zap-api-go/zap"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/ingress,mutating=false,failurePolicy=fail,groups="extensions",resources=ingresses,verbs=create,versions=v1beta1,name=dast.security.banzaicloud.io

// NewIngressValidator creates new ingressValidator
func NewIngressValidator(client client.Client, log logr.Logger) IngressValidator {
	return &ingressValidator{
		Client: client,
		Log:    log,
	}
}

// IngressValidator implements Handler
type IngressValidator interface {
	Handle(context.Context, admission.Request) admission.Response
}

type ingressValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// ingressValidator mutates PersitentVolumeClaims.
func (a *ingressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &extv1beta1.Ingress{}

	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	tresholds := getIngressTresholds(ingress)

	backendServices := k8sutil.GetIngressBackendServices(ingress, a.Log)
	a.Log.Info("Services", "backend_services", backendServices)
	ok, err := checkServices(backendServices, ingress.GetNamespace(), a.Log, a.Client, tresholds)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !ok {
		return admission.Denied("scan results are above treshold")
	}

	return admission.Allowed("scan results are below treshold")

}

// InjectDecoder injects the decoder.
func (a *ingressValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func checkServices(services []map[string]string, namespace string, log logr.Logger, client client.Client, tresholds map[string]int) (bool, error) {
	for _, service := range services {
		k8sService, err := k8sutil.GetServiceByName(service["name"], namespace, client)
		if err != nil {
			return false, err
		}
		zaProxyCfg, err := k8sutil.GetServiceAnotations(k8sService, log)
		if err != nil {
			return false, err
		}
		secret, err := k8sutil.GetSercretByName(zaProxyCfg["name"], zaProxyCfg["namespace"], client, log)
		if err != nil {
			return false, err
		}

		// TODO check scan status and wait for end of progress
		// check the scanner job is running, completed or not exist

		zapCore, err := newZapClient(zaProxyCfg["name"], zaProxyCfg["namespace"], string(secret.Data["zap_api_key"]), log)
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
		return nil, emperror.Wrap(err, "failed to get service summary from ZaProxy")
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
