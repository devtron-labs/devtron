package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/errors"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	errorsutil "github.com/argoproj/argo-workflows/v3/util/errors"
	"github.com/argoproj/argo-workflows/v3/util/retry"
	waitutil "github.com/argoproj/argo-workflows/v3/util/wait"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TerminateWorkflow terminates a workflow by setting its spec.shutdown to ShutdownStrategyTerminate
func TerminateWorkflow(ctx context.Context, wfClient v1alpha1.WorkflowInterface, name string) error {
	return patchShutdownStrategy(ctx, wfClient, name, wfv1.ShutdownStrategyTerminate)
}

type AlreadyShutdownError struct {
	workflowName string
	namespace    string
}

func (e AlreadyShutdownError) Error() string {
	return fmt.Sprintf("cannot shutdown a completed workflow: workflow: %q, namespace: %q", e.workflowName, e.namespace)
}

// patchShutdownStrategy patches the shutdown strategy to a workflow.
func patchShutdownStrategy(ctx context.Context, wfClient v1alpha1.WorkflowInterface, name string, strategy wfv1.ShutdownStrategy) error {
	patchObj := map[string]interface{}{
		"spec": map[string]interface{}{
			"shutdown": strategy,
		},
	}
	var err error
	patch, err := json.Marshal(patchObj)
	if err != nil {
		return errors.InternalWrapError(err)
	}
	err = waitutil.Backoff(retry.DefaultRetry, func() (bool, error) {
		wf, err := wfClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return !errorsutil.IsTransientErr(err), err
		}
		if wf.Status.Fulfilled() {
			return true, AlreadyShutdownError{wf.Name, wf.Namespace}
		}
		_, err = wfClient.Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
		if apierr.IsConflict(err) {
			return false, nil
		}
		return !errorsutil.IsTransientErr(err), err
	})
	return err
}
