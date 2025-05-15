package helper

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"time"
)

// GetSyncStartTime assumes that it is always called for calculating start time of latest git hash
func GetSyncStartTime(app *v1alpha1.Application) (time.Time, bool) {
	gitHash := app.Status.Sync.Revision
	if app.Status.OperationState != nil &&
		app.Status.OperationState.Operation.Sync != nil &&
		app.Status.OperationState.Operation.Sync.Revision == gitHash {
		return app.Status.OperationState.StartedAt.Time, true
	} else if len(app.Status.History) != 0 {
		if app.Status.History.LastRevisionHistory().Revision == gitHash &&
			app.Status.History.LastRevisionHistory().DeployStartedAt != nil {
			startTime := *app.Status.History.LastRevisionHistory().DeployStartedAt
			return startTime.Time, true
		}
	}
	return time.Time{}, false
}

// GetSyncFinishTime assumes that it is always called for calculating finish time of latest git hash
func GetSyncFinishTime(app *v1alpha1.Application) (time.Time, bool) {
	gitHash := app.Status.Sync.Revision
	if app.Status.OperationState != nil &&
		app.Status.OperationState.Operation.Sync != nil &&
		app.Status.OperationState.Operation.Sync.Revision == gitHash &&
		app.Status.OperationState.FinishedAt != nil {
		finishTime := *app.Status.OperationState.FinishedAt
		return finishTime.Time, true
	} else if len(app.Status.History) != 0 {
		if app.Status.History.LastRevisionHistory().Revision == gitHash &&
			app.Status.History.LastRevisionHistory().DeployStartedAt != nil {
			finishTime := *app.Status.History.LastRevisionHistory().DeployStartedAt
			return finishTime.Time, true
		}
	}
	return time.Time{}, false
}
