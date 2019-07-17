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

package ingress

import (
	"encoding/json"

	"github.com/go-logr/logr"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validate(ar *admissionv1beta1.AdmissionReview, log logr.Logger) *admissionv1beta1.AdmissionResponse {
	req := ar.Request
	log.Info("AdmissionReview for", "Kind", req.Kind, "Namespsce", req.Namespace, "Resource", req.Resource, "UserInfo", req.UserInfo)

	switch req.Kind.Kind {
	case "Ingress":
		var ingress extv1beta1.Ingress
		if err := json.Unmarshal(req.Object.Raw, &ingress); err != nil {
			log.Error(err, "Could not unmarshal raw object")
			return &admissionv1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
	}

	result := &metav1.Status{
		Reason: "validating result false",
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: false,
		Result:  result,
	}
}
