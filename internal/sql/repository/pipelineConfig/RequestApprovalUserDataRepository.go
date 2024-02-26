package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type RequestApprovalRepository interface {
	SaveDeploymentUserData(userData *RequestApprovalUserData) error
	FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*RequestApprovalUserData, error)
	FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*RequestApprovalUserData, error)
}

type RequestApprovalRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewResourceApprovalRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *RequestApprovalRepositoryImpl {
	return &RequestApprovalRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *RequestApprovalRepositoryImpl) FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*RequestApprovalUserData, error) {
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

func (impl *RequestApprovalRepositoryImpl) FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*RequestApprovalUserData, error) {
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

func (impl *RequestApprovalRepositoryImpl) SaveDeploymentUserData(userData *RequestApprovalUserData) error {
	currentTime := time.Now()
	userData.CreatedOn = currentTime
	userData.UpdatedOn = currentTime
	return impl.dbConnection.Insert(userData)
}
