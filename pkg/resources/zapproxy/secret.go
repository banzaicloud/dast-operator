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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// service return a secret for zapproxy
func (r *Reconciler) secret(log logr.Logger) runtime.Object {

	return newSecret(r.Dast)
}

func newSecret(dast *securityv1alpha1.Dast) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dast.Spec.ZapProxy.Name,
			Namespace: dast.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(dast, securityv1alpha1.GroupVersion.WithKind("Dast")),
			},
		},
		Data: map[string][]byte{
			"zap_api_key": []byte(dast.Spec.ZapProxy.APIKey),
		},
	}
}
