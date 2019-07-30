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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DastSpec defines the desired state of Dast
type DastSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ZapProxy ZapProxy `json:"zapproxy"`
	Analyzer Analyzer `json:"analyzer,omitempty"`
}

type ZapProxy struct {
	Image     string `json:"image,omitempty"`
	Name      string `json:"name"`
	NameSpace string `json:"namespace,omitempty"`
	APIKey    string `json:"apikey,omitempty"`
}

type Analyzer struct {
	Image   string          `json:"image"`
	Name    string          `json:"name"`
	Target  string          `json:"target,omitempty"`
	Service *corev1.Service `json:"service,omitempty"`
}

// DastStatus defines the observed state of Dast
type DastStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Dast is the Schema for the dasts API
type Dast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DastSpec   `json:"spec"`
	Status DastStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DastList contains a list of Dast
type DastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dast `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dast{}, &DastList{})
}
