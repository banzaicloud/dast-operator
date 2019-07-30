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

package controllers

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
	"github.com/banzaicloud/dast-operator/pkg/k8sutil"
	"github.com/banzaicloud/dast-operator/pkg/resources"
	"github.com/banzaicloud/dast-operator/pkg/resources/analyzer"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=services,verbs=get;create;list;update;patch;watch

func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("service", req.NamespacedName)

	var service corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &service); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	annotations := service.GetAnnotations()
	if zapProxyName, ok := annotations["dast.security.banzaicloud.io/zapproxy"]; ok {
		zapProxyNameSpace, ok := annotations["dast.security.banzaicloud.io/zapproxy_namespace"]
		if !ok {
			log.Error(errors.New("missing zapproxy namespace"), "missing annotatons")
		}
		log.Info("service reconciler", "serrvice", service.Spec)

		var analyzerImage string
		if analyzerImage, ok = annotations["dast.security.banzaicloud.io/analyzer_image"]; !ok {
			analyzerImage = "banzaicloud/dast-analyzer:latest"
		}

		ann := securityv1alpha1.Dast{
			ObjectMeta: metav1.ObjectMeta{
				Name:      service.GetName(),
				Namespace: zapProxyNameSpace,
			},
			Spec: securityv1alpha1.DastSpec{
				ZapProxy: securityv1alpha1.ZapProxy{
					Name: zapProxyName,
				},
				Analyzer: securityv1alpha1.Analyzer{
					Image:   analyzerImage,
					Name:    service.GetName(),
					Target:  k8sutil.GetTargetService(&service),
					Service: &service,
				},
			},
		}

		reconcilers := []resources.ComponentReconciler{
			analyzer.New(r.Client, &ann),
		}

		for _, rec := range reconcilers {
			err := rec.Reconcile(log)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
