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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ingressValidate = "/ingress"

// NewApp returns HTTPHandler
func NewApp(log logr.Logger, client client.Client) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(ingressValidate, newHTTPHandler(log, client))
	return mux
}

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	defaulter     = runtime.ObjectDefaulter(runtimeScheme)
)

// HTTPController collects the greeting use cases and exposes them as HTTP handlers.
type HTTPController struct {
	Logger logr.Logger
	Client client.Client
}

// NewHTTPHandler returns a new HTTP handler for the greeter.
func newHTTPHandler(log logr.Logger, client client.Client) http.Handler {
	mux := http.NewServeMux()
	controller := NewHTTPController(log, client)
	mux.HandleFunc(ingressValidate, controller.webhookCTRL)
	return mux
}

// NewHTTPController returns a new HTTPController instance.
func NewHTTPController(log logr.Logger, client client.Client) *HTTPController {
	return &HTTPController{
		Logger: log,
		Client: client,
	}
}

func (a *HTTPController) webhookCTRL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "reading request body failed", http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var admissionResponse *admissionv1beta1.AdmissionResponse
	ar := admissionv1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		a.Logger.Error(err, "Can't decode body")
		admissionResponse = &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == ingressValidate {
			admissionResponse = validate(&ar, a.Logger, a.Client)
		}
	}

	admissionReview := admissionv1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		a.Logger.Error(err, "Can't encode response")
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	a.Logger.Info("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		a.Logger.Error(err, "Can't write response")
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
