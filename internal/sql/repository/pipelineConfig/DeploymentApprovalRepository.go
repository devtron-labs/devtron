package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DeploymentApprovalRepository interface {
	FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int) ([]*DeploymentApprovalRequest, error)
}

type DeploymentApprovalRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDeploymentApprovalRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DeploymentApprovalRepositoryImpl {
	return &DeploymentApprovalRepositoryImpl{dbConnection: dbConnection, logger: logger}
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
	ArtifactId              int      `sql:"artifact_id"`           // keep in mind foreign key constraint
	Active                  bool     `sql:"active,notnull"`        // user can cancel request anytime
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

func (request *DeploymentApprovalRequest) ConvertToApprovalMetadata() *UserApprovalMetadata {
	approvalMetadata := &UserApprovalMetadata{ApprovalRequestId: request.Id}
	requestedUserData := UserApprovalData{}
	requestedUserData.UserId = request.CreatedBy
	requestedUserData.UserActionTime = request.CreatedOn
	approvalMetadata.RequestedUserData = requestedUserData
	var userApprovalData []UserApprovalData
	for _, approvalUser := range request.DeploymentApprovalUsers {
		userApprovalData = append(userApprovalData, UserApprovalData{DataId: approvalUser.Id, UserId: approvalUser.UserId, UserEmail: approvalUser.User.EmailId, UserResponse: approvalUser.UserResponse, UserActionTime: approvalUser.CreatedOn})
	}
	approvalMetadata.ApprovalUsersData = userApprovalData
	return approvalMetadata
}

func (request *DeploymentApprovalRequest) GetApprovedCount() int {
	count := 0
	for _, approvalUser := range request.DeploymentApprovalUsers {
		if approvalUser.UserResponse == APPROVED {
			count++
		}
	}
	return count
}

func (impl *DeploymentApprovalRepositoryImpl) FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int) ([]*DeploymentApprovalRequest, error) {
	impl.logger.Debugw("fetching approval data for artifacts", "ids", artifactIds, "pipelineId", pipelineId)
	var requests []*DeploymentApprovalRequest
	err := impl.dbConnection.
		Model(&requests).
		Column("deployment_approval_request.*", "DeploymentApprovalUsers", "DeploymentApprovalUsers.User").
		Where("artifact_id in (?) ", pg.In(artifactIds)).
		Where("pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner_id is null ").
		Where("active = ?", true).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	return requests, nil
}
