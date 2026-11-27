/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/devtron-labs/common-lib/utils/k8s"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

const (
	testNamespace = "devtron-ci"
	testJobName   = "test-wf-abcde"
)

// fakeKubeAPI serves the few batch/core endpoints the system executor calls and records every request.
type fakeKubeAPI struct {
	mu                sync.Mutex
	requests          []string
	deletePropagation string
	cmCreateErr       *k8sErrors.StatusError
	secretCreateErr   *k8sErrors.StatusError
	jobDeleteErr      *k8sErrors.StatusError
	jobGetErr         *k8sErrors.StatusError
	jobConditions     []batchV1.JobCondition
}

func (f *fakeKubeAPI) record(r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
}

func (f *fakeKubeAPI) calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

func writeJSON(w http.ResponseWriter, code int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(obj)
}

func writeStatusErr(w http.ResponseWriter, statusErr *k8sErrors.StatusError) {
	status := statusErr.ErrStatus
	status.TypeMeta = metaV1.TypeMeta{Kind: "Status", APIVersion: "v1"}
	writeJSON(w, int(status.Code), status)
}

func (f *fakeKubeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.record(r)
	jobsPath := "/apis/batch/v1/namespaces/" + testNamespace + "/jobs"
	jobPath := jobsPath + "/" + testJobName
	body, _ := io.ReadAll(r.Body)
	switch {
	case r.Method == http.MethodPost && r.URL.Path == jobsPath:
		job := &batchV1.Job{}
		_ = json.Unmarshal(body, job)
		job.Name, job.Namespace, job.UID = testJobName, testNamespace, "job-uid"
		writeJSON(w, http.StatusCreated, job)
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/namespaces/"+testNamespace+"/configmaps":
		if f.cmCreateErr != nil {
			writeStatusErr(w, f.cmCreateErr)
			return
		}
		cm := &coreV1.ConfigMap{}
		_ = json.Unmarshal(body, cm)
		writeJSON(w, http.StatusCreated, cm)
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/namespaces/"+testNamespace+"/secrets":
		if f.secretCreateErr != nil {
			writeStatusErr(w, f.secretCreateErr)
			return
		}
		secret := &coreV1.Secret{}
		_ = json.Unmarshal(body, secret)
		writeJSON(w, http.StatusCreated, secret)
	case r.Method == http.MethodPatch && r.URL.Path == jobPath:
		suspend := false
		writeJSON(w, http.StatusOK, &batchV1.Job{ObjectMeta: metaV1.ObjectMeta{Name: testJobName, Namespace: testNamespace}, Spec: batchV1.JobSpec{Suspend: &suspend}})
	case r.Method == http.MethodDelete && r.URL.Path == jobPath:
		// body may be json or protobuf; both carry the policy as a plain string
		if strings.Contains(string(body), string(metaV1.DeletePropagationForeground)) {
			f.mu.Lock()
			f.deletePropagation = string(metaV1.DeletePropagationForeground)
			f.mu.Unlock()
		}
		if f.jobDeleteErr != nil {
			writeStatusErr(w, f.jobDeleteErr)
			return
		}
		writeJSON(w, http.StatusOK, &metaV1.Status{TypeMeta: metaV1.TypeMeta{Kind: "Status", APIVersion: "v1"}, Status: metaV1.StatusSuccess})
	case r.Method == http.MethodGet && r.URL.Path == jobPath:
		if f.jobGetErr != nil {
			writeStatusErr(w, f.jobGetErr)
			return
		}
		writeJSON(w, http.StatusOK, &batchV1.Job{ObjectMeta: metaV1.ObjectMeta{Name: testJobName, Namespace: testNamespace}, Status: batchV1.JobStatus{Conditions: f.jobConditions}})
	default:
		writeStatusErr(w, k8sErrors.NewNotFound(schema.GroupResource{}, r.URL.Path))
	}
}

func newTestSystemExecutor(t *testing.T) *SystemWorkflowExecutorImpl {
	t.Helper()
	logger := zap.NewNop().Sugar()
	k8sUtil, err := k8s.NewK8sUtil(logger, &k8s.RuntimeConfig{})
	if err != nil {
		t.Fatalf("failed to build k8s util: %v", err)
	}
	return NewSystemWorkflowExecutorImpl(logger, k8sUtil)
}

func startFakeKubeAPI(t *testing.T, api *fakeKubeAPI) *rest.Config {
	t.Helper()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	return &rest.Config{Host: server.URL}
}

func testWorkflowTemplate(restConfig *rest.Config, configMaps, secrets []apiBean.ConfigSecretMap) bean.WorkflowTemplate {
	return bean.WorkflowTemplate{
		WorkflowRunnerId:   7080,
		WorkflowNamePrefix: "test-wf",
		Namespace:          testNamespace,
		ClusterConfig:      restConfig,
		ConfigMaps:         configMaps,
		Secrets:            secrets,
	}
}

func testCm(name string) apiBean.ConfigSecretMap {
	return apiBean.ConfigSecretMap{Name: name, Type: "environment", Data: json.RawMessage(`{"KEY":"value"}`)}
}

func testSecret(name string) apiBean.ConfigSecretMap {
	return apiBean.ConfigSecretMap{Name: name, Type: "environment", Data: json.RawMessage(`{"KEY":"dmFsdWU="}`)}
}

const (
	reqCreateJob    = "POST /apis/batch/v1/namespaces/" + testNamespace + "/jobs"
	reqCreateCm     = "POST /api/v1/namespaces/" + testNamespace + "/configmaps"
	reqCreateSecret = "POST /api/v1/namespaces/" + testNamespace + "/secrets"
	reqUnsuspendJob = "PATCH /apis/batch/v1/namespaces/" + testNamespace + "/jobs/" + testJobName
	reqDeleteJob    = "DELETE /apis/batch/v1/namespaces/" + testNamespace + "/jobs/" + testJobName
)

func forbidden(resource, name string) *k8sErrors.StatusError {
	return k8sErrors.NewForbidden(schema.GroupResource{Resource: resource}, name, fmt.Errorf("user cannot create resource %q", resource))
}

func alreadyExists(resource, name string) *k8sErrors.StatusError {
	return k8sErrors.NewAlreadyExists(schema.GroupResource{Resource: resource}, name)
}

func TestSystemExecutorExecuteWorkflow(t *testing.T) {
	cases := []struct {
		name            string
		configMaps      []apiBean.ConfigSecretMap
		secrets         []apiBean.ConfigSecretMap
		cmCreateErr     *k8sErrors.StatusError
		secretCreateErr *k8sErrors.StatusError
		jobDeleteErr    *k8sErrors.StatusError
		wantErr         func(error) bool
		wantCalls       []string
		wantItems       int
	}{
		{
			name:       "cm and secret created, job resumed",
			configMaps: []apiBean.ConfigSecretMap{testCm("cm-1")},
			secrets:    []apiBean.ConfigSecretMap{testSecret("secret-1")},
			wantCalls:  []string{reqCreateJob, reqCreateCm, reqCreateSecret, reqUnsuspendJob},
			wantItems:  3,
		},
		{
			name:      "no cm or secret, job resumed",
			wantCalls: []string{reqCreateJob, reqUnsuspendJob},
			wantItems: 1,
		},
		{
			name:       "external cm and secret are not created",
			configMaps: []apiBean.ConfigSecretMap{{Name: "ext-cm", External: true}},
			secrets:    []apiBean.ConfigSecretMap{{Name: "ext-secret", External: true}},
			wantCalls:  []string{reqCreateJob, reqUnsuspendJob},
			wantItems:  1,
		},
		{
			name:        "forbidden cm fails the trigger and deletes the suspended job",
			configMaps:  []apiBean.ConfigSecretMap{testCm("cm-1"), testCm("cm-2")},
			secrets:     []apiBean.ConfigSecretMap{testSecret("secret-1")},
			cmCreateErr: forbidden("configmaps", "cm-1"),
			wantErr:     k8sErrors.IsForbidden,
			wantCalls:   []string{reqCreateJob, reqCreateCm, reqDeleteJob},
		},
		{
			name:            "forbidden secret fails the trigger and deletes the suspended job",
			configMaps:      []apiBean.ConfigSecretMap{testCm("cm-1")},
			secrets:         []apiBean.ConfigSecretMap{testSecret("secret-1"), testSecret("secret-2")},
			secretCreateErr: forbidden("secrets", "secret-1"),
			wantErr:         k8sErrors.IsForbidden,
			wantCalls:       []string{reqCreateJob, reqCreateCm, reqCreateSecret, reqDeleteJob},
		},
		{
			name:        "server error on cm fails the trigger",
			configMaps:  []apiBean.ConfigSecretMap{testCm("cm-1")},
			cmCreateErr: k8sErrors.NewInternalError(fmt.Errorf("etcd unavailable")),
			wantErr:     k8sErrors.IsInternalError,
			wantCalls:   []string{reqCreateJob, reqCreateCm, reqDeleteJob},
		},
		{
			name:        "invalid cm fails the trigger",
			configMaps:  []apiBean.ConfigSecretMap{testCm("cm-1")},
			cmCreateErr: k8sErrors.NewInvalid(schema.GroupKind{Kind: "ConfigMap"}, "cm-1", nil),
			wantErr:     k8sErrors.IsInvalid,
			wantCalls:   []string{reqCreateJob, reqCreateCm, reqDeleteJob},
		},
		{
			name:        "cm that already exists is tolerated",
			configMaps:  []apiBean.ConfigSecretMap{testCm("cm-1")},
			secrets:     []apiBean.ConfigSecretMap{testSecret("secret-1")},
			cmCreateErr: alreadyExists("configmaps", "cm-1"),
			wantCalls:   []string{reqCreateJob, reqCreateCm, reqCreateSecret, reqUnsuspendJob},
			wantItems:   3,
		},
		{
			name:            "secret that already exists is tolerated",
			configMaps:      []apiBean.ConfigSecretMap{testCm("cm-1")},
			secrets:         []apiBean.ConfigSecretMap{testSecret("secret-1")},
			secretCreateErr: alreadyExists("secrets", "secret-1"),
			wantCalls:       []string{reqCreateJob, reqCreateCm, reqCreateSecret, reqUnsuspendJob},
			wantItems:       3,
		},
		{
			name:         "cleanup failure still returns the original cm error",
			configMaps:   []apiBean.ConfigSecretMap{testCm("cm-1")},
			cmCreateErr:  forbidden("configmaps", "cm-1"),
			jobDeleteErr: k8sErrors.NewInternalError(fmt.Errorf("delete failed")),
			wantErr:      k8sErrors.IsForbidden,
			wantCalls:    []string{reqCreateJob, reqCreateCm, reqDeleteJob},
		},
		{
			name:         "job already gone during cleanup still returns the original cm error",
			configMaps:   []apiBean.ConfigSecretMap{testCm("cm-1")},
			cmCreateErr:  forbidden("configmaps", "cm-1"),
			jobDeleteErr: k8sErrors.NewNotFound(schema.GroupResource{Group: "batch", Resource: "jobs"}, testJobName),
			wantErr:      k8sErrors.IsForbidden,
			wantCalls:    []string{reqCreateJob, reqCreateCm, reqDeleteJob},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeKubeAPI{cmCreateErr: tc.cmCreateErr, secretCreateErr: tc.secretCreateErr, jobDeleteErr: tc.jobDeleteErr}
			restConfig := startFakeKubeAPI(t, api)
			executor := newTestSystemExecutor(t)

			list, err := executor.ExecuteWorkflow(testWorkflowTemplate(restConfig, tc.configMaps, tc.secrets))

			if tc.wantErr != nil {
				if err == nil || !tc.wantErr(err) {
					t.Fatalf("expected matching error, got %v", err)
				}
				if list != nil {
					t.Errorf("expected nil list on error, got %d items", len(list.Items))
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if len(list.Items) != tc.wantItems {
				t.Errorf("expected %d items, got %d", tc.wantItems, len(list.Items))
			}
			if got := api.calls(); strings.Join(got, "\n") != strings.Join(tc.wantCalls, "\n") {
				t.Errorf("unexpected api calls\n got: %v\nwant: %v", got, tc.wantCalls)
			}
			for _, call := range api.calls() {
				if call == reqDeleteJob && api.deletePropagation != string(metaV1.DeletePropagationForeground) {
					t.Errorf("expected foreground delete, got %q", api.deletePropagation)
				}
			}
		})
	}
}

func jobCondition(conditionType batchV1.JobConditionType, status coreV1.ConditionStatus) batchV1.JobCondition {
	return batchV1.JobCondition{Type: conditionType, Status: status}
}

const (
	jobSuccessCriteriaMet batchV1.JobConditionType = "SuccessCriteriaMet"
	jobFailureTarget      batchV1.JobConditionType = "FailureTarget"
)

func TestGetJobStatus(t *testing.T) {
	cases := []struct {
		name       string
		conditions []batchV1.JobCondition
		want       string
	}{
		{name: "no conditions", want: ""},
		{name: "complete only", conditions: []batchV1.JobCondition{jobCondition(batchV1.JobComplete, coreV1.ConditionTrue)}, want: "Complete"},
		{name: "failed only", conditions: []batchV1.JobCondition{jobCondition(batchV1.JobFailed, coreV1.ConditionTrue)}, want: "Failed"},
		{name: "still suspended", conditions: []batchV1.JobCondition{jobCondition(batchV1.JobSuspended, coreV1.ConditionTrue)}, want: "Suspended"},
		{name: "resumed, running", conditions: []batchV1.JobCondition{jobCondition(batchV1.JobSuspended, coreV1.ConditionFalse)}, want: ""},
		{name: "resumed then complete", conditions: []batchV1.JobCondition{
			jobCondition(batchV1.JobSuspended, coreV1.ConditionFalse),
			jobCondition(jobSuccessCriteriaMet, coreV1.ConditionTrue),
			jobCondition(batchV1.JobComplete, coreV1.ConditionTrue),
		}, want: "Complete"},
		{name: "resumed then failed", conditions: []batchV1.JobCondition{
			jobCondition(batchV1.JobSuspended, coreV1.ConditionFalse),
			jobCondition(jobFailureTarget, coreV1.ConditionTrue),
			jobCondition(batchV1.JobFailed, coreV1.ConditionTrue),
		}, want: "Failed"},
		{name: "complete with interim condition first", conditions: []batchV1.JobCondition{
			jobCondition(jobSuccessCriteriaMet, coreV1.ConditionTrue),
			jobCondition(batchV1.JobComplete, coreV1.ConditionTrue),
		}, want: "Complete"},
		{name: "failed with interim condition first", conditions: []batchV1.JobCondition{
			jobCondition(jobFailureTarget, coreV1.ConditionTrue),
			jobCondition(batchV1.JobFailed, coreV1.ConditionTrue),
		}, want: "Failed"},
		{name: "failure decided, pods still terminating", conditions: []batchV1.JobCondition{
			jobCondition(batchV1.JobSuspended, coreV1.ConditionFalse),
			jobCondition(jobFailureTarget, coreV1.ConditionTrue),
		}, want: "FailureTarget"},
		{name: "unknown status is ignored", conditions: []batchV1.JobCondition{
			jobCondition(batchV1.JobFailed, coreV1.ConditionUnknown),
		}, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getJobStatus(tc.conditions); got != tc.want {
				t.Errorf("getJobStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSystemExecutorGetWorkflowStatus(t *testing.T) {
	api := &fakeKubeAPI{jobConditions: []batchV1.JobCondition{
		jobCondition(batchV1.JobSuspended, coreV1.ConditionFalse),
		jobCondition(jobFailureTarget, coreV1.ConditionTrue),
		jobCondition(batchV1.JobFailed, coreV1.ConditionTrue),
	}}
	restConfig := startFakeKubeAPI(t, api)
	status, err := newTestSystemExecutor(t).GetWorkflowStatus(testJobName, testNamespace, restConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != "Failed" {
		t.Errorf("expected Failed, got %q", status.Status)
	}

	missing := &fakeKubeAPI{jobGetErr: k8sErrors.NewNotFound(schema.GroupResource{Group: "batch", Resource: "jobs"}, testJobName)}
	_, err = newTestSystemExecutor(t).GetWorkflowStatus(testJobName, testNamespace, startFakeKubeAPI(t, missing))
	if err == nil || err.Error() != "cannot find workflow "+testJobName {
		t.Errorf("expected not found error, got %v", err)
	}
}
