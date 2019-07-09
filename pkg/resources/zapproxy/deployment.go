// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zapproxy

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// deployment return a deployment for zapproxy
func (r *Reconciler) deployment(log logr.Logger) runtime.Object {

	return newDeployment(r.Dast)
}

func newDeployment(dast *securityv1alpha1.Dast) *appsv1.Deployment {
	labels := map[string]string{
		"app":        componentName,
		"controller": dast.Name,
	}

	var zapImage string
	if zapImage := dast.Spec.ZapProxy.Image; zapImage == "" {
		zapImage = "owasp/zap2docker-live"
	}

	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dast.Spec.ZapProxy.Name,
			Namespace: dast.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(dast, securityv1alpha1.GroupVersion.WithKind("Dast")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "zap-proxy",
							Image:   zapImage,
							Command: []string{"zap.sh"},
							Args: []string{
								"-daemon",
								"-host",
								"0.0.0.0",
								"-port",
								"8080",
								"-config",
								"api.key=" + dast.Spec.ZapProxy.APIKey,
								"-config",
								"api.addrs.addr.name=.*",
								"-config",
								"api.addrs.addr.regex=true",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      "TCP",
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.IntOrString{IntVal: 8080},
										HTTPHeaders: []corev1.HTTPHeader{
											{
												Name:  "X-ZAP-API-Key",
												Value: dast.Spec.ZapProxy.APIKey,
											},
										},
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
							},
						},
					},
				},
			},
		},
	}
}
