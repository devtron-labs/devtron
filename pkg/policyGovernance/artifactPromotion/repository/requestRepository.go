package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type ArtifactPromotionApprovalRequest struct {
	tableName               struct{}                                 `sql:"artifact_promotion_approval_request" pg:",discard_unknown_columns"`
	Id                      int                                      `sql:"id"`
	PolicyId                int                                      `sql:"policy_id"`
	PolicyEvaluationAuditId int                                      `sql:"policy_evaluation_audit_id"`
	ArtifactId              int                                      `sql:"artifact_id"`
	SourceType              constants.SourceType                     `sql:"source_type"`
	SourcePipelineId        int                                      `sql:"source_pipeline_id"`
	DestinationPipelineId   int                                      `sql:"destination_pipeline_id"`
	Status                  constants.ArtifactPromotionRequestStatus `sql:"status"`
	sql.AuditLog
}

type RequestRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewRequestRepositoryImpl(dbConnection *pg.DB) *RequestRepositoryImpl {
	return &RequestRepositoryImpl{
		dbConnection: dbConnection,
	}
}

type RequestRepository interface {
	Create(tx *pg.Tx, PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	Update(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	UpdateInBulk(tx *pg.Tx, PromotionRequest []*ArtifactPromotionApprovalRequest) error
	FindById(id int) (*ArtifactPromotionApprovalRequest, error)
	FindByArtifactAndDestinationPipelineIds(artifactId int, destinationPipelineId []int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPendingByDestinationPipelineId(destinationPipelineId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) (*ArtifactPromotionApprovalRequest, error)
	FindByPipelineIdAndArtifactIds(pipelineId int, artifactIds []int, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error)
	FindRequestsByStatusesForDestinationPipelines(pipelineId []int, artifactId int, statuses []constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error)
	FindRequestsByArtifactAndOptionalEnv(artifactId int, environmentName string, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPolicyId(policyId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPendingByDestinationPipelineIds(pipelineIds []int) (PromotionRequest []*ArtifactPromotionApprovalRequest, err error)

	// TODO: combine below func based on status
	MarkStaleByIds(tx *pg.Tx, requestIds []int) error
	MarkStaleByDestinationPipelineId(tx *pg.Tx, pipelineIds []int) error
	MarkStaleByPolicyId(tx *pg.Tx, policyId int) error
	MarkStaleByAppEnvIds(tx *pg.Tx, commaSeperatedAppEnvIds [][]int) error
	MarkPromoted(tx *pg.Tx, requestIds []int, userId int32) error
	MarkCancel(requestId int, userId int32) (rowsAffected int, err error)
	GetRequestsApprovedByUserForPipelines(pipelineIds []int, userId int32) ([]int, error)
	HasUserApprovedRequest(artifactId, pipelineId int, userId int32) (bool, error)
}

func (repo *RequestRepositoryImpl) Create(tx *pg.Tx, PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error) {
	_, err := tx.Model(PromotionRequest).Insert()
	if err != nil {
		return nil, err
	}
	return PromotionRequest, nil
}

func (repo *RequestRepositoryImpl) Update(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error) {
	_, err := repo.dbConnection.Model(PromotionRequest).Update()
	if err != nil {
		return nil, err
	}
	return PromotionRequest, nil
}

func (repo *RequestRepositoryImpl) FindById(id int) (*ArtifactPromotionApprovalRequest, error) {
	model := &ArtifactPromotionApprovalRequest{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).
		Select()
	return model, err
}

func (repo *RequestRepositoryImpl) FindRequestsByStatusesForDestinationPipelines(destinationPipelineIds []int, artifactId int, statuses []constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id IN (?) ", pg.In(destinationPipelineIds)).
		Where("artifact_id = ?", artifactId).
		Where("status IN (?) ", pg.In(statuses)).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindPendingByDestinationPipelineId(destinationPipelineId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id = ? ", destinationPipelineId).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindByArtifactAndDestinationPipelineIds(artifactId int, destinationPipelineIds []int) ([]*ArtifactPromotionApprovalRequest, error) {

	models := make([]*ArtifactPromotionApprovalRequest, 0)
	if len(destinationPipelineIds) == 0 {
		return models, nil
	}
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id IN (?) ", pg.In(destinationPipelineIds)).
		Where("artifact_id = ?", artifactId).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Where("artifact_id = ?", artifactId).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindAwaitedRequestByPolicyId(policyId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Where("policy_id = ?", policyId).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) (*ArtifactPromotionApprovalRequest, error) {
	model := &ArtifactPromotionApprovalRequest{}
	err := repo.dbConnection.Model(model).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("status = ? ", constants.PROMOTED).
		Where("artifact_id = ?", artifactId).
		Select()
	return model, err
}

func (repo *RequestRepositoryImpl) FindByPipelineIdAndArtifactIds(pipelineId int, artifactIds []int, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error) {
	var model []*ArtifactPromotionApprovalRequest
	if len(artifactIds) == 0 {
		return model, nil
	}
	err := repo.dbConnection.Model(&model).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("status = ? ", status).
		Where("artifact_id in (?) ", pg.In(artifactIds)).
		Select()
	return model, err
}

func (repo *RequestRepositoryImpl) FindRequestsByArtifactAndOptionalEnv(artifactId int, environmentName string, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)

	query := fmt.Sprintf("SELECT apar.* "+
		" FROM artifact_promotion_approval_request apar"+
		" inner join pipeline p on apar.destination_pipeline_id=p.id "+
		"inner join environment e on p.environment_id=e.id where apar.status = %d and apar.artifact_id = %d and p.deleted=false ", status, artifactId)

	if len(environmentName) > 0 {
		query = query + fmt.Sprintf("and e.environment_name = '%s'", environmentName)
	}
	_, err := repo.dbConnection.Query(&models, query)
	return models, err
}

func (repo *RequestRepositoryImpl) MarkStaleByIds(tx *pg.Tx, requestIds []int) error {
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", constants.STALE).
		Set("updated_on = ?", time.Now()).
		Where("id IN (?)", pg.In(requestIds)).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Update()
	return err
}

func (repo *RequestRepositoryImpl) MarkStaleByDestinationPipelineId(tx *pg.Tx, pipelineIds []int) error {
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", constants.STALE).
		Set("updated_on = ?", time.Now()).
		Where("destination_pipeline_id IN (?)", pg.In(pipelineIds)).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Update()
	return err
}

func (repo *RequestRepositoryImpl) MarkStaleByPolicyId(tx *pg.Tx, policyId int) error {
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", constants.STALE).
		Set("updated_on = ?", time.Now()).
		Where("policy_id = ?", policyId).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Update()
	return err
}

func (repo *RequestRepositoryImpl) MarkStaleByAppEnvIds(tx *pg.Tx, commaSeperatedAppEnvIds [][]int) error {
	if len(commaSeperatedAppEnvIds) == 0 {
		return nil
	}
	// example
	// update artifact_promotion_approval_request
	// set status = 3
	// from pipeline p
	// where (p.app_id,p.environment_id) IN ((4,2)) and p.id = artifact_promotion_approval_request.destination_pipeline_id

	res, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Table("pipeline").
		Set("status = ?", constants.STALE).
		Set("updated_on = ?", time.Now()).
		Where("(pipeline.app_id,pipeline.environment_id) IN ?", pg.InMulti(commaSeperatedAppEnvIds)).
		Where("pipeline.id = artifact_promotion_approval_request.destination_pipeline_id").
		Where("artifact_promotion_approval_request.status = ? ", constants.AWAITING_APPROVAL).
		Update()
	if res != nil {
		fmt.Println("rows affected : ", res.RowsAffected())
	}
	return err
}

func (repo *RequestRepositoryImpl) MarkPromoted(tx *pg.Tx, requestIds []int, userId int32) error {
	if len(requestIds) == 0 {
		return nil
	}
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", constants.PROMOTED).
		Set("updated_on = ?", time.Now()).
		Where("id IN (?)", pg.In(requestIds)).
		Update()
	return err
}

func (repo *RequestRepositoryImpl) UpdateInBulk(tx *pg.Tx, PromotionRequest []*ArtifactPromotionApprovalRequest) error {
	for _, request := range PromotionRequest {
		err := tx.Update(request)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *RequestRepositoryImpl) FindPendingByDestinationPipelineIds(pipelineIds []int) (PromotionRequest []*ArtifactPromotionApprovalRequest, err error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	if len(pipelineIds) == 0 {
		return models, nil
	}
	err = repo.dbConnection.Model(&models).
		Where("destination_pipeline_id in (?) and status = ? ", pg.In(pipelineIds), constants.AWAITING_APPROVAL).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) MarkCancel(requestId int, userId int32) (rowsAffected int, err error) {
	res, err := repo.dbConnection.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", constants.CANCELED).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).
		Where("id = ? and created_by = ? ", requestId, userId).
		Update()
	return res.RowsAffected(), err
}

func (repo *RequestRepositoryImpl) GetRequestsApprovedByUserForPipelines(pipelineIds []int, userId int32) ([]int, error) {
	var ids []int
	if len(pipelineIds) == 0 {
		return ids, nil
	}
	err := repo.dbConnection.Model(&ArtifactPromotionApprovalRequest{}).
		Column("artifact_promotion_approval_request.id").
		Join("inner join request_approval_user_data on artifact_promotion_approval_request.id = request_approval_user_data.approval_request_id and request_type = ? ", models.ARTIFACT_PROMOTION_APPROVAL).
		Where("request_approval_user_data.user_id = ? and artifact_promotion_approval_request.status = ? and artifact_promotion_approval_request.destination_pipeline_id in (?)", userId, constants.AWAITING_APPROVAL, pg.In(pipelineIds)).
		Select(&ids)
	return ids, err
}

func (repo *RequestRepositoryImpl) HasUserApprovedRequest(artifactId, pipelineId int, userId int32) (bool, error) {
	var Result bool
	err := repo.dbConnection.Model(&ArtifactPromotionApprovalRequest{}).
		ColumnExpr(" count(*) > 0 as Result").
		Join("inner join request_approval_user_data on artifact_promotion_approval_request.id = request_approval_user_data.approval_request_id and request_type = ? ", models.ARTIFACT_PROMOTION_APPROVAL).
		Where("request_approval_user_data.user_id = ?  ", userId).
		Where("artifact_promotion_approval_request.status = ?  ", constants.PROMOTED).
		Where("artifact_promotion_approval_request.destination_pipeline_id = ?", pipelineId).
		Where("artifact_promotion_approval_request.artifact_id = ?", artifactId).
		Select(&Result)
	if err == pg.ErrNoRows {
		return false, nil
	}
	return Result, err
}
