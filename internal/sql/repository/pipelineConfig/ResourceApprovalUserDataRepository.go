package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ResourceApprovalRepository interface {
	FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*ResourceApprovalUserData, error)
	FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*ResourceApprovalUserData, error)
}

type ResourceApprovalRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewResourceApprovalRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ResourceApprovalRepositoryImpl {
	return &ResourceApprovalRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *ResourceApprovalRepositoryImpl) FetchApprovalDataForRequests(requestIds []int, requestType repository.RequestType) ([]*ResourceApprovalUserData, error) {
	var usersData []*ResourceApprovalUserData
	if len(requestIds) == 0 {
		return usersData, nil
	}
	err := impl.dbConnection.
		Model(&usersData).
		Column("resource_approval_user_data.*", "User").
		Where("approval_request_id in (?) ", pg.In(requestIds)).
		Where("request_type = ?", requestType).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "requestIds", requestIds, "err", err)
		return nil, err
	}
	return usersData, nil
}

func (impl *ResourceApprovalRepositoryImpl) FetchApprovedDataByApprovalId(approvalRequestId int, requestType repository.RequestType) ([]*ResourceApprovalUserData, error) {
	var results []*ResourceApprovalUserData
	err := impl.dbConnection.
		Model(&results).
		Column("resource_approval_user_data.*", "User").
		Where("resource_approval_user_data.approval_request_id = ? ", approvalRequestId).
		Where("resource_approval_user_data.request_type = ?", requestType).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching artifacts", "results", results, "err", err)
		return nil, err
	}
	return results, nil

}
