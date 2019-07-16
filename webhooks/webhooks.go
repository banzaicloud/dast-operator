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

package webhooks

import (
	"net"
	"net/http"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/dast-operator/webhooks/ingress"
)

type IngressWH struct {
	Client client.Client
	Log    logr.Logger
}

func (r *IngressWH) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(r)
}

func (r *IngressWH) Start(<-chan struct{}) error {
	ln, _ := net.Listen("tcp", ":5555")
	httpServer := &http.Server{Handler: ingress.NewApp(r.Log)}
	r.Log.Info("Starting the HTTP server.")

	return httpServer.Serve(ln)
}
