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
	"context"
	"fmt"
	"strconv"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetServiceStatus(service *corev1.Service) bool {
	// TODO improve service status check
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

func GetIngressBackendServices(ingress *unstructured.Unstructured, log logr.Logger) ([]map[string]string, error) {
	log.Info("ingress", "ingress", ingress)
	backends := []map[string]string{}
	rules, ok, _ := unstructured.NestedSlice(ingress.Object, "spec", "rules")
	if !ok {
		return backends, errors.New("value not found: rules")
	}
	for _, rule := range rules {
		paths, ok, _ := unstructured.NestedSlice(rule.(map[string]interface{}), "http", "paths")
		if !ok {
			return backends, errors.New("value not found: paths")
		}
		for _, path := range paths {
			backend := map[string]string{}
			backend["name"], ok, _ = unstructured.NestedString(path.(map[string]interface{}), "backend", "serviceName")
			if !ok {
				backend["name"], ok, _ = unstructured.NestedString(path.(map[string]interface{}), "backend", "service", "name")
				if !ok {
					return backends, errors.New("value not found: service name")
				}
			}
			portNum, ok, _ := unstructured.NestedFieldCopy(path.(map[string]interface{}), "backend", "servicePort")
			if !ok {
				portNum, ok, _ = unstructured.NestedFieldCopy(path.(map[string]interface{}), "backend", "service", "port", "number")
				if !ok {
					return backends, errors.New("value not found: service port")
				}
			}

			backend["port"] = fmt.Sprintf("%v", portNum)
			backends = append(backends, backend)
		}
	}

	return backends, nil
}

func GetServiceByName(name, namespace string, client client.Client) (*corev1.Service, error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var service corev1.Service
	if err := client.Get(context.TODO(), key, &service); err != nil {
		return nil, emperror.Wrap(err, "cannot get service by name")
	}

	return &service, nil
}

func GetServiceAnotations(service *corev1.Service, log logr.Logger) (map[string]string, error) {
	annotations := service.GetAnnotations()
	zaProxyCfg := map[string]string{}
	if zaproxyName, ok := annotations["dast.security.banzaicloud.io/zaproxy"]; ok {
		zaProxyCfg["name"] = zaproxyName
		zaProxyCfg["namespace"], ok = annotations["dast.security.banzaicloud.io/zaproxy-namespace"]
		if !ok {
			zaProxyCfg["namespace"] = service.GetNamespace()
			log.Info("missing zaproxy namespace annotation, using service namespace", "ns_name", zaProxyCfg["namespace"])
		}
		zaProxyCfg["analyzer_image"], ok = annotations["dast.security.banzaicloud.io/analyzer_image"]
		if !ok {
			zaProxyCfg["analyzer_image"] = "ghcr.io/banzaicloud/dast-analyzer:latest"
			log.Info("missing zaproxy analyzer image annotation, using ", "analyzer_image", zaProxyCfg["analyzer_image"])
		}
		return zaProxyCfg, nil
	}

	return nil, errors.New("service isn't annotated")
}
