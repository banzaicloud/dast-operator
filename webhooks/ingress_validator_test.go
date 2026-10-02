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
	"strings"
	"testing"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func newJob(conditions ...batchv1.JobCondition) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "test-service", Namespace: "zaproxy"},
		Status:     batchv1.JobStatus{Conditions: conditions},
	}
}

func TestGetScannerJobState(t *testing.T) {
	tests := []struct {
		name string
		objs []client.Object
		want scannerJobState
	}{
		{
			name: "not found",
			want: scannerJobNotFound,
		},
		{
			name: "running",
			objs: []client.Object{newJob()},
			want: scannerJobRunning,
		},
		{
			name: "completed",
			objs: []client.Object{newJob(batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionTrue})},
			want: scannerJobCompleted,
		},
		{
			name: "failed",
			objs: []client.Object{newJob(batchv1.JobCondition{Type: batchv1.JobFailed, Status: corev1.ConditionTrue})},
			want: scannerJobFailed,
		},
		{
			name: "condition not true",
			objs: []client.Object{newJob(batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionFalse})},
			want: scannerJobRunning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScannerJobState("test-service", "zaproxy", newFakeClient(t, tt.objs...))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckServicesDeniesUnfinishedScan(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test",
			Annotations: map[string]string{
				"dast.security.banzaicloud.io/zaproxy":           "dast-test",
				"dast.security.banzaicloud.io/zaproxy-namespace": "zaproxy",
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "dast-test", Namespace: "zaproxy"},
		Data:       map[string][]byte{"zap_api_key": []byte("key")},
	}
	job := newJob()

	tests := []struct {
		name string
		objs []client.Object
		want string
	}{
		{
			name: "job not found",
			objs: []client.Object{service, secret},
			want: "scanner job zaproxy/test-service is not found",
		},
		{
			name: "job running",
			objs: []client.Object{service, secret, job},
			want: "scanner job zaproxy/test-service is running",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := []map[string]string{{"name": "test-service", "port": "80"}}
			reason, err := checkServices(services, "test", logr.Discard(), newFakeClient(t, tt.objs...), map[string]int{})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(reason, tt.want) {
				t.Errorf("got %q, want %q", reason, tt.want)
			}
		})
	}
}
