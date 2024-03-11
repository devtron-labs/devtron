package repository

import (
	"fmt"
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
	FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindRequestsByArtifactIdAndEnvName(artifactId int, environmentName string, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPolicyId(policyId int) ([]*ArtifactPromotionApprovalRequest, error)
	MarkStaleByIds(tx *pg.Tx, requestIds []int) error
	MarkStaleByDestinationPipelineId(tx *pg.Tx, pipelineIds []int) error
	MarkStaleByPolicyId(tx *pg.Tx, policyId int) error
	MarkPromoted(tx *pg.Tx, requestIds []int) error
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

func (repo *RequestRepositoryImpl) FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("status = ? ", constants.AWAITING_APPROVAL).
		Where("artifact_id = ?", artifactId).
		Select()
	return models, err
}

func (repo *RequestRepositoryImpl) FindRequestsByArtifactIdAndEnvName(artifactId int, environmentName string, status constants.ArtifactPromotionRequestStatus) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)

	query := fmt.Sprintf("select * from artifact_promotion_approval_request apar"+
		" inner join pipeline p on apar.destination_pipeline_id=p.id "+
		"inner join environment e on p.environment_id=e.id where apar.status = %d and apar.artifact_id = %d ", status, artifactId)

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

func (repo *RequestRepositoryImpl) MarkPromoted(tx *pg.Tx, requestIds []int) error {
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
