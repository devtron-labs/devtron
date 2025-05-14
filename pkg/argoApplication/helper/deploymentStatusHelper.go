package helper

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// GetSyncStartTime assumes that it is always called for calculating start time of latest git hash
func GetSyncStartTime(app *v1alpha1.Application, defaultStartTime time.Time) time.Time {
	startTime := metav1.NewTime(defaultStartTime)
	gitHash := app.Status.Sync.Revision
	if app.Status.OperationState != nil &&
		app.Status.OperationState.Operation.Sync != nil &&
		app.Status.OperationState.Operation.Sync.Revision == gitHash {
		startTime = app.Status.OperationState.StartedAt
	} else if len(app.Status.History) != 0 {
		if app.Status.History.LastRevisionHistory().Revision == gitHash &&
			app.Status.History.LastRevisionHistory().DeployStartedAt != nil {
			startTime = *app.Status.History.LastRevisionHistory().DeployStartedAt
		}
	}
	return startTime.Time
}

// GetSyncFinishTime assumes that it is always called for calculating finish time of latest git hash
func GetSyncFinishTime(app *v1alpha1.Application, defaultEndTime time.Time) time.Time {
	finishTime := metav1.NewTime(defaultEndTime)
	gitHash := app.Status.Sync.Revision
	if app.Status.OperationState != nil &&
		app.Status.OperationState.Operation.Sync != nil &&
		app.Status.OperationState.Operation.Sync.Revision == gitHash &&
		app.Status.OperationState.FinishedAt != nil {
		finishTime = *app.Status.OperationState.FinishedAt
	} else if app.Status.History != nil {
		for _, history := range app.Status.History {
			if history.Revision == gitHash {
				finishTime = history.DeployedAt
			}
		}
	}
	return finishTime.Time
}
