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

package k8sutil

import (
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

func GetServiceStatus(service *corev1.Service) bool {
	// TODO improve service statucs check
	if service.Spec.ClusterIP != "" {
		return true
	}
	return false
}

func GetTargetService(service *corev1.Service) string {
	var portNR string
	// TODO handle multiple port
	for _, port := range service.Spec.Ports {
		portNR = strconv.Itoa(int(port.Port))
	}
	// TODO handle protocol
	return "http://" + service.GetName() + "." + service.GetNamespace() + ".svc.cluster.local:" + portNR
}

func GetIngressBackendServices(ingress *extv1beta1.Ingress, log logr.Logger) []map[string]string {
	log.Info("ingress", "ingress", ingress)
	backends := []map[string]string{}
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			backend := map[string]string{}
			backend["name"] = path.Backend.ServiceName
			backend["port"] = path.Backend.ServicePort.String()
			backends = append(backends, backend)
		}
	}
	return backends
}
