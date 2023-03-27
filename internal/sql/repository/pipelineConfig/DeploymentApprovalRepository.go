package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"time"
)

type DeploymentApprovalRepository interface {
}

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

type DeploymentApprovalRequest struct {
	tableName               struct{} `sql:"deployment_approval_request" pg:",discard_unknown_columns"`
	Id                      int      `sql:"id,pk"`
	PipelineId              int      `sql:"pipeline_id"`           // keep in mind foreign key constraint
	ArtifactId              int      `sql:"artifactId"`            // keep in mind foreign key constraint
	CdWorkflowRunnerId      int      `sql:"cd_workflow_runner_id"` // keep in mind foreign key constraint
	DeploymentApprovalUsers []*DeploymentApprovalUserData
	sql.AuditLog
}

type DeploymentApprovalUserData struct {
	tableName         struct{}                   `sql:"deployment_approval_user_data" pg:",discard_unknown_columns"`
	Id                int                        `sql:"id,pk"`
	ApprovalRequestId int                        `sql:"approval_request_id"` // keep in mind foreign key constraint
	UserId            int32                      `sql:"user_id"`             // keep in mid foreign key constraint
	UserResponse      DeploymentApprovalResponse `sql:"user_response"`
	Comments          string                     `sql:"comments"`
	User              *repository.UserModel
	sql.AuditLog
}

func (impl *DeploymentApprovalRequest) ConvertToApprovalMetadata() *UserApprovalMetadata {
	approvalMetadata := &UserApprovalMetadata{ApprovalRequestId: impl.Id}
	requestedUserData := UserApprovalData{}
	requestedUserData.UserId = impl.CreatedBy
	requestedUserData.UserActionTime = impl.CreatedOn
	approvalMetadata.RequestedUserData = requestedUserData
	var userApprovalData []UserApprovalData
	for _, approvalUser := range impl.DeploymentApprovalUsers {
		userApprovalData = append(userApprovalData, UserApprovalData{DataId: approvalUser.Id, UserId: approvalUser.UserId, UserEmail: approvalUser.User.EmailId, UserResponse: approvalUser.UserResponse, UserActionTime: approvalUser.CreatedOn})
	}
	approvalMetadata.ApprovalUsersData = userApprovalData
	return approvalMetadata
}
