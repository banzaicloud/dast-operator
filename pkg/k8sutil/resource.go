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
	"reflect"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// Reconcile reconciles K8S resources
func Reconcile(log logr.Logger, client runtimeClient.Client, desired runtimeClient.Object, cr *securityv1alpha1.Dast) error {
	desiredType := reflect.TypeOf(desired)
	current, ok := desired.DeepCopyObject().(runtimeClient.Object)
	if !ok {
		return emperror.With(errors.New("desired object is not a client.Object"), "kind", desiredType)
	}

	key := runtimeClient.ObjectKeyFromObject(current)
	log = log.WithValues("kind", desiredType, "name", key.Name)

	err := client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "kind", desiredType, "name", key.Name)
	}
	if apierrors.IsNotFound(err) {
		if err := client.Create(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
		}
		log.Info("resource created")
	}
	return nil
}
