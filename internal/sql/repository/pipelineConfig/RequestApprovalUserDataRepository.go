package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type Constraint string

const UNIQUE_USER_REQUEST_ACTION Constraint = "unique_user_request_action"

type RequestApprovalUserdataRepository interface {
	SaveDeploymentUserData(userData *RequestApprovalUserData) error
	FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*RequestApprovalUserData, error)
	FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*RequestApprovalUserData, error)
}

type RequestApprovalUserDataRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewRequestApprovalUserDataRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *RequestApprovalUserDataRepositoryImpl {
	return &RequestApprovalUserDataRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *RequestApprovalUserDataRepositoryImpl) FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*RequestApprovalUserData, error) {
	var usersData []*RequestApprovalUserData
	if len(requestIds) == 0 {
		return usersData, nil
	}
	err := impl.dbConnection.
		Model(&usersData).
		Column("request_approval_user_data.*", "User").
		Where("approval_request_id in (?) ", pg.In(requestIds)).
		Where("request_type = ?", requestType).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "requestIds", requestIds, "err", err)
		return nil, err
	}
	return usersData, nil
}

func (impl *RequestApprovalUserDataRepositoryImpl) FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*RequestApprovalUserData, error) {
	var results []*RequestApprovalUserData
	err := impl.dbConnection.
		Model(&results).
		Column("request_approval_user_data.*", "User").
		Where("request_approval_user_data.approval_request_id = ? ", approvalRequestId).
		Where("request_approval_user_data.request_type = ?", requestType).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "results", results, "err", err)
		return nil, err
	}
	return results, nil

}

func (impl *RequestApprovalUserDataRepositoryImpl) SaveDeploymentUserData(userData *RequestApprovalUserData) error {
	currentTime := time.Now()
	userData.CreatedOn = currentTime
	userData.UpdatedOn = currentTime
	return impl.dbConnection.Insert(userData)
}
