package bean

type UserDeploymentRequestStatus string

func (status UserDeploymentRequestStatus) ToString() string {
	return string(status)
}

func (status UserDeploymentRequestStatus) IsTerminalTimelineStatus() bool {
	switch status {
	case DeploymentRequestCompleted, DeploymentRequestSuperseded,
		DeploymentRequestFailed, DeploymentRequestTerminated:
		return true
	}
	return false
}

func (status UserDeploymentRequestStatus) IsTriggered() bool {
	switch status {
	case DeploymentRequestTriggered:
		return true
	}
	return false
}

func (status UserDeploymentRequestStatus) IsTriggerHistorySaved() bool {
	switch status {
	case DeploymentRequestTriggerAuditCompleted:
		return true
	}
	return false
}

func (status UserDeploymentRequestStatus) IsCompleted() bool {
	switch status {
	case DeploymentRequestCompleted:
		return true
	}
	return false
}

const (
	DeploymentRequestPending               UserDeploymentRequestStatus = "PENDING"
	DeploymentRequestTriggerAuditCompleted UserDeploymentRequestStatus = "TRIGGER_AUDIT_COMPLETED"
	DeploymentRequestTriggered             UserDeploymentRequestStatus = "TRIGGERED"
	DeploymentRequestCompleted             UserDeploymentRequestStatus = "COMPLETED"
	DeploymentRequestSuperseded            UserDeploymentRequestStatus = "SUPERSEDED"
	DeploymentRequestFailed                UserDeploymentRequestStatus = "FAILED"
	DeploymentRequestTerminated            UserDeploymentRequestStatus = "TERMINATED"
)
