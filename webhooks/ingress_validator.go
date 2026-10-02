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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/ingress,mutating=false,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1,groups="networking.k8s.io",resources=ingresses,verbs=create,versions=v1,name=dast.security.banzaicloud.io

// NewIngressValidator creates new ingressValidator
func NewIngressValidator(client client.Client, log logr.Logger) IngressValidator {
	return &ingressValidator{
		Client:  client,
		decoder: admission.NewDecoder(client.Scheme()),
		Log:     log,
	}
}

// IngressValidator implements Handle
type IngressValidator interface {
	Handle(context.Context, admission.Request) admission.Response
}

type ingressValidator struct {
	Client  client.Client
	decoder admission.Decoder
	Log     logr.Logger
}

// ingressValidator validates ingress.
func (a *ingressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &unstructured.Unstructured{}

	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	tresholds := getIngressTresholds(ingress)

	backendServices, err := k8sutil.GetIngressBackendServices(ingress, a.Log)
	if err != nil {
		return admission.Errored(http.StatusNotImplemented, err)
	}
	a.Log.Info("Services", "backend_services", backendServices)
	reason, err := checkServices(backendServices, ingress.GetNamespace(), a.Log, a.Client, tresholds)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if reason != "" {
		return admission.Denied(reason)
	}

	return admission.Allowed("scan results are below treshold")

}

type scannerJobState string

const (
	scannerJobNotFound  scannerJobState = "not found"
	scannerJobRunning   scannerJobState = "running"
	scannerJobFailed    scannerJobState = "failed"
	scannerJobCompleted scannerJobState = "completed"
)

func getScannerJobState(name, namespace string, c client.Client) (scannerJobState, error) {
	job := &batchv1.Job{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, job); err != nil {
		if apierrors.IsNotFound(err) {
			return scannerJobNotFound, nil
		}
		return "", emperror.Wrap(err, "cannot get scanner job")
	}
	for _, condition := range job.Status.Conditions {
		if condition.Status != corev1.ConditionTrue {
			continue
		}
		switch condition.Type {
		case batchv1.JobComplete:
			return scannerJobCompleted, nil
		case batchv1.JobFailed:
			return scannerJobFailed, nil
		}
	}
	return scannerJobRunning, nil
}

func checkServices(services []map[string]string, namespace string, log logr.Logger, client client.Client, tresholds map[string]int) (string, error) {
	for _, service := range services {
		k8sService, err := k8sutil.GetServiceByName(service["name"], namespace, client)
		if err != nil {
			return "", err
		}
		zaProxyCfg, err := k8sutil.GetServiceAnotations(k8sService, log)
		if err != nil {
			return "", err
		}
		secret, err := k8sutil.GetSercretByName(zaProxyCfg["name"], zaProxyCfg["namespace"], client, log)
		if err != nil {
			return "", err
		}

		state, err := getScannerJobState(service["name"], zaProxyCfg["namespace"], client)
		if err != nil {
			return "", err
		}
		if state != scannerJobCompleted {
			return fmt.Sprintf("scanner job %s/%s is %s", zaProxyCfg["namespace"], service["name"], state), nil
		}

		zapCore, err := newZapClient(zaProxyCfg["name"], zaProxyCfg["namespace"], string(secret.Data["zap_api_key"]), log)
		if err != nil {
			return "", err
		}
		summary, err := getServiceScanSummary(service, namespace, zapCore, log)
		if err != nil {
			return "", err
		}

		s, err := cast.ToStringMapIntE(summary["alertsSummary"])
		if err != nil {
			return "", err
		}
		for key, value := range s {
			if value > tresholds[key] {
				return "scan results are above treshold", nil
			}
		}
	}
	return "", nil
}

func getIngressTresholds(ingress *unstructured.Unstructured) map[string]int {
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
