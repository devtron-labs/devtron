package pipelineConfig

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DeploymentApprovalRepository interface {
	FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int) ([]*DeploymentApprovalRequest, error)
	FetchById(requestId int) (*DeploymentApprovalRequest, error)
	Save(deploymentApprovalRequest *DeploymentApprovalRequest) error
	Update(deploymentApprovalRequest *DeploymentApprovalRequest) error
	SaveDeploymentUserData(userData *DeploymentApprovalUserData) error
	ConsumeApprovalRequest(requestId int) error
}

type DeploymentApprovalRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDeploymentApprovalRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DeploymentApprovalRepositoryImpl {
	return &DeploymentApprovalRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type DeploymentApprovalRequest struct {
	tableName                   struct{} `sql:"deployment_approval_request" pg:",discard_unknown_columns"`
	Id                          int      `sql:"id,pk"`
	PipelineId                  int      `sql:"pipeline_id"`    // keep in mind foreign key constraint
	ArtifactId                  int      `sql:"artifact_id"`    // keep in mind foreign key constraint
	Active                      bool     `sql:"active,notnull"` // user can cancel request anytime
	ArtifactDeploymentTriggered bool     `sql:"artifact_deployment_triggered"`
	DeploymentApprovalUsers     []*DeploymentApprovalUserData
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

func (impl *DeploymentApprovalRepositoryImpl) FetchApprovalDataForArtifacts(artifactIds []int, pipelineId int) ([]*DeploymentApprovalRequest, error) {
	impl.logger.Debugw("fetching approval data for artifacts", "ids", artifactIds, "pipelineId", pipelineId)
	var requests []*DeploymentApprovalRequest
	err := impl.dbConnection.
		Model(&requests).
		Column("deployment_approval_request.*", /*"DeploymentApprovalUsers", "DeploymentApprovalUsers.User"*/).
		Where("artifact_id in (?) ", pg.In(artifactIds)).
		Where("pipeline_id = ?", pipelineId).
		Where("artifact_deployment_triggered = ?", false).
		Where("active = ?", true).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	var requestIdMap map[int]*DeploymentApprovalRequest
	for _, request := range requests {
		requestIdMap[request.Id] = request
	}
	var usersData []*DeploymentApprovalUserData
	err = impl.dbConnection.
		Model(&usersData).
		Column("deployment_approval_user_data.*", "User").
		Where("approval_request_id in (?) ", pg.In(artifactIds)).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	for _, userData := range usersData {
		approvalRequestId := userData.ApprovalRequestId
		deploymentApprovalRequest := requestIdMap[approvalRequestId]
		approvalUsers := deploymentApprovalRequest.DeploymentApprovalUsers
		approvalUsers = append(approvalUsers, userData)
		deploymentApprovalRequest.DeploymentApprovalUsers = approvalUsers
	}
	return requests, nil
}

func (impl *DeploymentApprovalRepositoryImpl) FetchById(requestId int) (*DeploymentApprovalRequest, error) {
	var request *DeploymentApprovalRequest
	err := impl.dbConnection.
		Model(request).Where("id = ?", requestId).Where("active = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching request data", "id", requestId, "err", err)
		return nil, err
	}
	return request, nil
}

func (impl *DeploymentApprovalRepositoryImpl) ConsumeApprovalRequest(requestId int) error {
	request, err := impl.FetchById(requestId)
	if err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching approval request", "requestId", requestId, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		return errors.New("approval request not raised for this artifact")
	}
	request.ArtifactDeploymentTriggered = true
	return impl.Update(request)
}

func (impl *DeploymentApprovalRepositoryImpl) Save(deploymentApprovalRequest *DeploymentApprovalRequest) error {
	currentTime := time.Now()
	deploymentApprovalRequest.CreatedOn = currentTime
	deploymentApprovalRequest.UpdatedOn = currentTime
	return impl.dbConnection.Insert(deploymentApprovalRequest)
}

func (impl *DeploymentApprovalRepositoryImpl) Update(deploymentApprovalRequest *DeploymentApprovalRequest) error {
	deploymentApprovalRequest.UpdatedOn = time.Now()
	return impl.dbConnection.Update(deploymentApprovalRequest)
}

func (impl *DeploymentApprovalRepositoryImpl) SaveDeploymentUserData(userData *DeploymentApprovalUserData) error {
	currentTime := time.Now()
	userData.CreatedOn = currentTime
	userData.UpdatedOn = currentTime
	return impl.dbConnection.Insert(userData)
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
