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
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
	"github.com/banzaicloud/dast-operator/pkg/k8sutil"
	"github.com/banzaicloud/dast-operator/pkg/resources"
)

const (
	componentName = "analyzer"
)

var labelSelector = map[string]string{
	"app": "analyzer",
}

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

// New creates a new reconciler for analyzer
func New(client client.Client, dast *securityv1alpha1.Dast) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Dast:   dast,
		},
	}
}

// Reconcile implements the reconcile logic for analyzer
func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.V(1).Info("Reconciling")

	key := types.NamespacedName{
		Name:      r.Dast.Spec.ZapProxy.Name,
		Namespace: r.Dast.Namespace,
	}

	zapDeployment := appsv1.Deployment{}
	if err := r.Get(context.TODO(), key, &zapDeployment); err != nil {
		return emperror.Wrap(err, "failed to get zap deployment")
	}

	if func(deployment *appsv1.Deployment) bool {
		timeout := time.After(1 * time.Minute)
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-timeout:
				return false
			case <-ticker.C:
				r.Get(context.TODO(), key, deployment)
				if k8sutil.GetDeploymentStatusAvailable(deployment, log) {
					return true
				}
			}
		}
	}(&zapDeployment) {
		log.Info("deployment is available")
	}

	if r.Dast.Spec.Analyzer.Service != nil {
		key := types.NamespacedName{
			Name:      r.Dast.Spec.Analyzer.Service.GetName(),
			Namespace: r.Dast.Spec.Analyzer.Service.GetNamespace(),
		}

		service := corev1.Service{}
		if err := r.Get(context.TODO(), key, &service); err != nil {
			return emperror.Wrap(err, "failed to get service")
		}

		if func(service *corev1.Service) bool {
			timeout := time.After(1 * time.Minute)
			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-timeout:
					return false
				case <-ticker.C:
					r.Get(context.TODO(), key, service)
					if k8sutil.GetServiceStatus(service) {
						return true
					}
				}
			}
		}(&service) {
			log.Info("service is available")
		}
	}

	for _, res := range []resources.ResourceWithLogs{
		r.job,
	} {
		o := res(log)
		err := k8sutil.Reconcile(log, r.Client, o, r.Dast)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.V(1).Info("Reconciled")

	return nil
}
