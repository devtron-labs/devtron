package executors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestSystemWorkflowStatusUsesActiveTerminalCondition(t *testing.T) {
	cases := []struct {
		name       string
		conditions []batchv1.JobCondition
		want       string
	}{
		{"completed after resume", []batchv1.JobCondition{{Type: batchv1.JobSuspended, Status: corev1.ConditionFalse}, {Type: batchv1.JobSuccessCriteriaMet, Status: corev1.ConditionTrue}, {Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}, "Complete"},
		{"failed after resume", []batchv1.JobCondition{{Type: batchv1.JobSuspended, Status: corev1.ConditionFalse}, {Type: batchv1.JobFailureTarget, Status: corev1.ConditionTrue}, {Type: batchv1.JobFailed, Status: corev1.ConditionTrue}}, "Failed"},
		{"false terminal condition", []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionFalse}, {Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}, "Complete"},
		{"currently suspended", []batchv1.JobCondition{{Type: batchv1.JobSuspended, Status: corev1.ConditionTrue}}, "Suspended"},
		{"only historical conditions", []batchv1.JobCondition{{Type: batchv1.JobSuspended, Status: corev1.ConditionFalse}}, ""},
		{"no conditions", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/apis/batch/v1/namespaces/fixture/jobs/owned-job" {
					http.Error(w, "unexpected API request", 500)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(batchv1.Job{TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"}, Status: batchv1.JobStatus{Conditions: tc.conditions}})
			}))
			defer server.Close()
			logger := zap.NewNop().Sugar()
			util, err := k8s.NewK8sUtil(logger, &k8s.RuntimeConfig{})
			if err != nil {
				t.Fatal(err)
			}
			executor := executors.NewSystemWorkflowExecutorImpl(logger, util)
			got, err := executor.GetWorkflowStatus("owned-job", "fixture", &rest.Config{Host: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != tc.want {
				t.Fatalf("got status %q, want %q", got.Status, tc.want)
			}
		})
	}
}
