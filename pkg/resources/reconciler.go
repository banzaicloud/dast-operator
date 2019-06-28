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

package resources

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
)

// Reconciler holds client and CR for Dast
type Reconciler struct {
	client.Client
	Dast *securityv1alpha1.Dast
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

// Resource simple function without parameter
type Resource func() runtime.Object

// ResourceWithLogs function with log parameter
type ResourceWithLogs func(log logr.Logger) runtime.Object
