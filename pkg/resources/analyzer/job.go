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

package analyzer

import (
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// job return a job for analyzer
func (r *Reconciler) job(log logr.Logger) runtime.Object {

	return newAnalyzerJob(r.Dast)
}

func newAnalyzerJob(dast *securityv1alpha1.Dast) *batchv1.Job {
	backofflimit := int32(5)
	completion := int32(1)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dast.Spec.DeploymentName,
			Namespace: dast.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(dast, securityv1alpha1.GroupVersion.WithKind("Dast")),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			Completions:  &completion,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:            "analyzer",
							Image:           "test2:latest",
							ImagePullPolicy: "IfNotPresent",
							Command: []string{
								"/dynamic-analyzer",
								"scanner",
								"-t",
								dast.Spec.Target,
								"-a",
								"abcd1234",
								"-p",
								"zap-proxy:8080",
							},
						},
					},
				},
			},
		},
	}
}
