package pipelineConfig

import "time"

type DeploymentApprovalResponse int

const (
	APPROVED DeploymentApprovalResponse = iota
	REJECTED
)

type UserApprovalData struct {
	DataId         int                        `json:"dataId"`
	UserId         int32                      `json:"userId"`
	UserEmail      string                     `json:"userEmail"`
	UserResponse   DeploymentApprovalResponse `json:"userResponse"`
	UserActionTime time.Time                  `json:"userActionTime"`
	UserComment    string                     `json:"userComment"`
}

type ApprovalRequestState int

const (
	InitApprovalState ApprovalRequestState = iota
	RequestedApprovalState
	ApprovedApprovalState
	ConsumedApprovalState
)

type UserApprovalMetadata struct {
	ApprovalRequestId    int                  `json:"approvalRequestId"`
	ApprovalRuntimeState ApprovalRequestState `json:"approvalRuntimeState"`
	RequestedUserData    UserApprovalData     `json:"requestedUserData"`
	ApprovalUsersData    []UserApprovalData   `json:"approvedUsersData"`
}
