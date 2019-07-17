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
	"crypto/tls"
	"flag"
	"fmt"
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

type serverConfig struct {
	certFile string
	keyFile  string
	port     int
}

var parameters serverConfig

func init() {
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
}

func (r *IngressWH) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(r)
}

func (r *IngressWH) Start(<-chan struct{}) error {
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		r.Log.Error(err, "Failed to load key pair")
	}

	ln, _ := net.Listen("tcp", fmt.Sprintf(":%v", parameters.port))
	httpServer := &http.Server{
		Handler:   ingress.NewApp(r.Log),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
	r.Log.Info("starting the webhook.")

	return httpServer.ServeTLS(ln, "", "")
}
