package timelineStatus

import (
	"golang.org/x/exp/slices"
)

type TimelineStatus string

func ContainsTerminalTimelineStatus(statuses []TimelineStatus) bool {
	for _, status := range statuses {
		if slices.Contains(TerminalTimelineStatusList, status) {
			return true
		}
	}
	return false
}

func (status TimelineStatus) ToString() string {
	return string(status)
}

var TerminalTimelineStatusList = []TimelineStatus{
	TIMELINE_STATUS_APP_HEALTHY,
	TIMELINE_STATUS_DEPLOYMENT_FAILED,
	TIMELINE_STATUS_GIT_COMMIT_FAILED,
	TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
}

var InternalTimelineStatusList = []TimelineStatus{
	TIMELINE_STATUS_DEPLOYMENT_REQUEST_VALIDATED,
	TIMELINE_STATUS_ARGOCD_SYNC_INITIATED,
	TIMELINE_STATUS_DEPLOYMENT_TRIGGERED,
}

const (
	TIMELINE_STATUS_DEPLOYMENT_INITIATED         TimelineStatus = "DEPLOYMENT_INITIATED"
	TIMELINE_STATUS_DEPLOYMENT_REQUEST_VALIDATED TimelineStatus = "DEPLOYMENT_REQUEST_VALIDATED"
	TIMELINE_STATUS_GIT_COMMIT                   TimelineStatus = "GIT_COMMIT"
	TIMELINE_STATUS_GIT_COMMIT_FAILED            TimelineStatus = "GIT_COMMIT_FAILED"
	TIMELINE_STATUS_ARGOCD_SYNC_INITIATED        TimelineStatus = "ARGOCD_SYNC_INITIATED"
	TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED        TimelineStatus = "ARGOCD_SYNC_COMPLETED"
	// TIMELINE_STATUS_DEPLOYMENT_TRIGGERED - is not a terminal status.
	// It indicates that the deployment request has been served to Kubernetes CD agents (helm/ ArgoCD).
	TIMELINE_STATUS_DEPLOYMENT_TRIGGERED TimelineStatus = "DEPLOYMENT_TRIGGERED"

	TIMELINE_STATUS_KUBECTL_APPLY_STARTED  TimelineStatus = "KUBECTL_APPLY_STARTED"
	TIMELINE_STATUS_KUBECTL_APPLY_SYNCED   TimelineStatus = "KUBECTL_APPLY_SYNCED"
	TIMELINE_STATUS_APP_HEALTHY            TimelineStatus = "HEALTHY"
	TIMELINE_STATUS_DEPLOYMENT_FAILED      TimelineStatus = "FAILED"
	TIMELINE_STATUS_FETCH_TIMED_OUT        TimelineStatus = "TIMED_OUT"
	TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS TimelineStatus = "UNABLE_TO_FETCH_STATUS"
	TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED  TimelineStatus = "DEPLOYMENT_SUPERSEDED"
	TIMELINE_STATUS_MANIFEST_GENERATED     TimelineStatus = "HELM_PACKAGE_GENERATED" // TODO: remove as this deployment type is not supported
)

const (
	TIMELINE_DESCRIPTION_DEPLOYMENT_INITIATED         string = "Deployment initiated successfully."
	TIMELINE_DESCRIPTION_VULNERABLE_IMAGE             string = "Deployment failed: Vulnerability policy violated."
	TIMELINE_DESCRIPTION_DEPLOYMENT_REQUEST_VALIDATED string = "Deployment trigger request has been validated successfully."
	TIMELINE_DESCRIPTION_ARGOCD_GIT_COMMIT            string = "Git commit done successfully."
	TIMELINE_DESCRIPTION_ARGOCD_SYNC_INITIATED        string = "ArgoCD sync initiated."
	TIMELINE_DESCRIPTION_ARGOCD_SYNC_COMPLETED        string = "ArgoCD sync completed."
	TIMELINE_DESCRIPTION_DEPLOYMENT_COMPLETED         string = "Deployment has been performed successfully. Waiting for application to be healthy..."
	TIMELINE_DESCRIPTION_DEPLOYMENT_SUPERSEDED        string = "This deployment is superseded."
)
