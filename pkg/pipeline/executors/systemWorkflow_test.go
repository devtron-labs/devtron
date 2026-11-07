package executors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/devtron-labs/common-lib/utils/k8s"
	api "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"go.uber.org/zap"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestSystemWorkflowDependentCreationErrors(t *testing.T) {
	for _, deniedResource := range []string{"", "configmaps", "secrets"} {
		deny := deniedResource != ""
		var mu sync.Mutex
		calls := []string{}
		job := batch.Job{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, r.Method+" "+r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			if r.Header.Get("Authorization") != "Bearer synthetic-fixture-only" {
				http.Error(w, "unexpected identity", 401)
				return
			}
			if !strings.Contains(r.URL.Path, "/namespaces/executor-fixture/") {
				http.Error(w, "outside namespace", 403)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/jobs") && r.Method == "POST" {
				if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
					panic(err)
				}
				job.Name = "fixture-job"
				job.Namespace = "executor-fixture"
				job.UID = "synthetic-job-uid"
				job.TypeMeta = meta.TypeMeta{APIVersion: "batch/v1", Kind: "Job"}
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(job)
				return
			}
			if r.Method == "POST" && (strings.HasSuffix(r.URL.Path, "/configmaps") || strings.HasSuffix(r.URL.Path, "/secrets")) {
				if strings.HasSuffix(r.URL.Path, "/"+deniedResource) && deny {
					w.WriteHeader(403)
					json.NewEncoder(w).Encode(meta.Status{TypeMeta: meta.TypeMeta{Kind: "Status", APIVersion: "v1"}, Status: "Failure", Reason: meta.StatusReasonForbidden, Code: 403, Message: "synthetic namespaced role denies dependent creation"})
					return
				}
				v := map[string]interface{}{}
				json.NewDecoder(r.Body).Decode(&v)
				refs := v["metadata"].(map[string]interface{})["ownerReferences"].([]interface{})
				if len(refs) != 1 || refs[0].(map[string]interface{})["uid"] != "synthetic-job-uid" {
					panic("owner reference mismatch")
				}
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(v)
				return
			}
			if r.Method == "PATCH" && strings.HasSuffix(r.URL.Path, "/jobs/fixture-job") {
				no := false
				job.Spec.Suspend = &no
				json.NewEncoder(w).Encode(job)
				return
			}
			http.Error(w, "unexpected API request", 500)
		}))
		logger := zap.NewNop().Sugar()
		util, err := k8s.NewK8sUtil(logger, &k8s.RuntimeConfig{})
		if err != nil {
			panic(err)
		}
		ex := executors.NewSystemWorkflowExecutorImpl(logger, util)
		template := bean.WorkflowTemplate{Namespace: "executor-fixture", WorkflowNamePrefix: "fixture", WorkflowType: bean.CI_WORKFLOW_NAME,
			ClusterConfig: &rest.Config{Host: server.URL, BearerToken: "synthetic-fixture-only", ContentConfig: rest.ContentConfig{ContentType: "application/json", AcceptContentTypes: "application/json"}}, TerminationGracePeriod: 30,
			PodSpec:    core.PodSpec{RestartPolicy: core.RestartPolicyNever, Containers: []core.Container{{Name: "main", Image: "fixture.invalid/not-executed", Command: []string{"true"}}}},
			ConfigMaps: []api.ConfigSecretMap{{Name: "fixture-config", Data: json.RawMessage(`{"setting":"synthetic"}`)}},
			Secrets:    []api.ConfigSecretMap{{Name: "fixture-secret", Data: json.RawMessage(`{"value":"c3ludGhldGlj"}`)}}}
		objects, runErr := ex.ExecuteWorkflow(template)
		server.Close()
		patch := false
		for _, c := range calls {
			if c == "PATCH /apis/batch/v1/namespaces/executor-fixture/jobs/fixture-job" {
				patch = true
			}
		}
		count := 0
		if objects != nil {
			count = len(objects.Items)
		}
		row := map[string]interface{}{"denyDependentCreates": deny, "deniedResource": deniedResource, "errorReturned": runErr != nil, "unsuspendRequested": patch, "returnedObjects": count, "calls": calls, "scope": "official executor against local controlled HTTP API, no Kubernetes workload or external connection"}
		if runErr != nil {
			row["error"] = runErr.Error()
		}
		if !deny && (runErr != nil || !patch || count != 3) {
			t.Fatalf("success control failed: %v", row)
		}
		if deny && (runErr == nil || patch) {
			t.Fatalf("dependent creation failure must not unsuspend: %v", row)
		}
	}
}
