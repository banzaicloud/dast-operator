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

package analyzer

import (
	"github.com/go-logr/logr"
	"istio.io/pkg/log"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// job return a job for analyzer
func (r *Reconciler) job(log logr.Logger) runtime.Object {

	return newAnalyzerJob(r.Dast)
}

func newAnalyzerJob(dast *securityv1alpha1.Dast) *batchv1.Job {
	var ownerReferences []metav1.OwnerReference
	if dast.Spec.Analyzer.Service != nil {
		ownerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(dast.Spec.Analyzer.Service, schema.GroupVersion{Group: "app", Version: "v1"}.WithKind("Service"))}
	} else {
		ownerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(dast, securityv1alpha1.GroupVersion.WithKind("Dast"))}
	}

	annotations := dast.Spec.Analyzer.Service.GetAnnotations()

	command := []string{
		"/dynamic-analyzer",
		"scanner",
		"-t",
		dast.Spec.Analyzer.Target,
		"-p",
		// TODO use https
		"http://" + dast.Spec.ZaProxy.Name + ":8080",
	}

	apiScan, ok := annotations["dast.security.banzaicloud.io/apiscan"]
	if ok {
		if apiScan == "true" {
			log.Info("apiscan enabled")
			openapiURL, ok := annotations["dast.security.banzaicloud.io/openapi-url"]
			if ok {
				log.Info("openapi url is defined")
				command = []string{
					"/dynamic-analyzer",
					"apiscan",
					"-t",
					dast.Spec.Analyzer.Target,
					"-o",
					openapiURL,
					"-p",
					// TODO use https
					"http://" + dast.Spec.ZaProxy.Name + ":8080",
				}
			} else {
				log.Info("openapi url is missing")
			}
		}
	}

	backofflimit := int32(5)
	completion := int32(1)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dast.Spec.Analyzer.Name,
			Namespace:       dast.Namespace,
			OwnerReferences: ownerReferences,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			Completions:  &completion,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:            dast.Spec.Analyzer.Name,
							Image:           dast.Spec.Analyzer.Image,
							ImagePullPolicy: "IfNotPresent",
							Command:         command,
							Env:             withEnv(dast),
						},
					},
				},
			},
		},
	}
}

func withEnv(dast *securityv1alpha1.Dast) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name: "ZAPAPIKEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dast.Spec.ZaProxy.Name,
					},
					Key: "zap_api_key",
				},
			},
		},
	}
	return env
}
