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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
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

func (r *DastReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("dast", req.NamespacedName)

	var dast securityv1alpha1.Dast
	if err := r.Get(ctx, req.NamespacedName, &dast); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			log.V(1).Info("test", "test", "lofasz")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var deployments appsv1.DeploymentList
	err := r.List(ctx, &deployments, client.InNamespace(req.Namespace), client.MatchingField(ownerKey, req.Name))
	if err != nil {
		log.Error(err, "unable to list deployments")
		return ctrl.Result{}, err
	}
	if len(deployments.Items) == 0 {
		err = r.Create(ctx, newDeployment(&dast))
		if err != nil {
			log.Error(err, "unable to create deployment", "deployment", dast.Spec.DeploymentName)
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
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Dast" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Dast{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func newDeployment(dast *securityv1alpha1.Dast) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "dast",
		"controller": dast.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dast.Spec.DeploymentName,
			Namespace: dast.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(dast, securityv1alpha1.GroupVersion.WithKind("Dast")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: dast.Spec.Replicas,
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
							Image:   dast.Spec.ImageRepo,
							Command: []string{"zap.sh"},
							Args: []string{
								"-daemon",
								"-host",
								"0.0.0.0",
								"-port",
								"8080",
								"-config",
								"api.key=abcd1234",
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
												Value: "abcd1234",
											},
										},
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}
}
