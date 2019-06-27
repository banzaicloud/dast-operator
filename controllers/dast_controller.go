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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
	"github.com/banzaicloud/dast-operator/pkg/resources"
	"github.com/banzaicloud/dast-operator/pkg/resources/analyzer"
	"github.com/banzaicloud/dast-operator/pkg/resources/zapproxy"
)

// DastReconciler reconciles a Dast object
type DastReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.banzaicloud.io,resources=dasts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.banzaicloud.io,resources=dasts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;create;list;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;create;list;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;create;list;update;patch

func (r *DastReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("dast", req.NamespacedName)

	var dast securityv1alpha1.Dast
	if err := r.Get(ctx, req.NamespacedName, &dast); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	reconcilers := []resources.ComponentReconciler{
		zapproxy.New(r.Client, &dast),
		analyzer.New(r.Client, &dast),
	}

	for _, rec := range reconcilers {
		err := rec.Reconcile(log)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

var (
	ownerKey = ".metadata.controller"
	apiGVStr = securityv1alpha1.GroupVersion.String()
)

func (r *DastReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Dast{}).
		Complete(r)
}
