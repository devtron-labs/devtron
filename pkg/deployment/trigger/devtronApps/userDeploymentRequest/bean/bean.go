package bean

type UserDeploymentRequestStatus string

func (status UserDeploymentRequestStatus) ToString() string {
	return string(status)
}

func (status UserDeploymentRequestStatus) IsTerminalTimelineStatus() bool {
	switch status {
	case DeploymentRequestCompleted, DeploymentRequestSuperseded:
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

const (
	DeploymentRequestPending    UserDeploymentRequestStatus = "PENDING"
	DeploymentRequestTriggered  UserDeploymentRequestStatus = "TRIGGERED"
	DeploymentRequestCompleted  UserDeploymentRequestStatus = "COMPLETED"
	DeploymentRequestSuperseded UserDeploymentRequestStatus = "SUPERSEDED"
)
