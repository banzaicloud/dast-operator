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
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/banzaicloud/dast-operator/api/v1alpha1"
	"github.com/banzaicloud/dast-operator/pkg/k8sutil"
	"github.com/banzaicloud/dast-operator/pkg/resources"
)

const (
	componentName = "zapproxy"
)

var labelSelector = map[string]string{
	"app": "zapproxy",
}

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

// New creates a new reconciler for Zapproxy
func New(client client.Client, dast *securityv1alpha1.Dast) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Dast:   dast,
		},
	}
}

// Reconcile implements the reconcile logic for Zapproxy
func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.V(1).Info("Reconciling")

	for _, res := range []resources.ResourceWithLogs{
		r.deployment,
		r.service,
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
