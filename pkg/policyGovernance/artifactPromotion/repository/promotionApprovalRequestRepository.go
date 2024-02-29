package repository

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type ArtifactPromotionApprovalRequest struct {
	tableName               struct{}                            `sql:"artifact_promotion_approval_request" pg:",discard_unknown_columns"`
	Id                      int                                 `sql:"id"`
	PolicyId                int                                 `sql:"policy_id"`
	PolicyEvaluationAuditId int                                 `sql:"policy_evaluation_audit_id"`
	ArtifactId              int                                 `sql:"artifact_id"`
	SourceType              bean.SourceType                     `sql:"source_type"`
	SourcePipelineId        int                                 `sql:"source_pipeline_id"`
	DestinationPipelineId   int                                 `sql:"destination_pipeline_id"`
	Status                  bean.ArtifactPromotionRequestStatus `sql:"status"`
	Active                  bool                                `sql:"active"`
	sql.AuditLog
}

type ArtifactPromotionApprovalRequestRepoImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewArtifactPromotionApprovalRequestImpl(dbConnection *pg.DB) *ArtifactPromotionApprovalRequestRepoImpl {
	return &ArtifactPromotionApprovalRequestRepoImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

type ArtifactPromotionApprovalRequestRepository interface {
	Create(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	Update(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	FindById(id int) (*ArtifactPromotionApprovalRequest, error)
	FindByDestinationPipelineIds(destinationPipelineId []int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPendingByDestinationPipelineId(destinationPipelineId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) (*ArtifactPromotionApprovalRequest, error)
	FindByPipelineIdAndArtifactIds(pipelineId int, artifactIds []int) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPolicyId(policyId int) ([]*ArtifactPromotionApprovalRequest, error)
	MarkStaleByIds(tx *pg.Tx, requestIds []int) error
	MarkStaleByPolicyId(tx *pg.Tx, policyId int) error
	MarkPromoted(tx *pg.Tx, requestIds []int) error
	sql.TransactionWrapper
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) Create(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error) {
	_, err := repo.dbConnection.Model(PromotionRequest).Insert()
	if err != nil {
		return nil, err
	}
	return PromotionRequest, nil
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) Update(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error) {
	_, err := repo.dbConnection.Model(PromotionRequest).Update()
	if err != nil {
		return nil, err
	}
	return PromotionRequest, nil
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindById(id int) (*ArtifactPromotionApprovalRequest, error) {
	model := &ArtifactPromotionApprovalRequest{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).Where("active = ?", true).Select()
	return model, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindPendingByDestinationPipelineId(destinationPipelineId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id = ? ", destinationPipelineId).
		Where("status = ? ", bean.PROMOTED).
		Where("active = ?", true).
		Select()
	return models, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindByDestinationPipelineIds(destinationPipelineIds []int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id IN (?) ", pg.In(destinationPipelineIds)).
		Where("status = ? ", bean.AWAITING_APPROVAL).
		Where("active = ?", true).
		Select()
	return models, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("status = ? ", bean.AWAITING_APPROVAL).
		Where("active = ?", true).
		Where("artifact_id = ?", artifactId).
		Select()
	return models, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindAwaitedRequestByPolicyId(policyId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(&models).
		Where("status = ? ", bean.AWAITING_APPROVAL).
		Where("active = ?", true).
		Where("policy_id = ?", policyId).
		Select()
	return models, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) (*ArtifactPromotionApprovalRequest, error) {
	model := &ArtifactPromotionApprovalRequest{}
	err := repo.dbConnection.Model(model).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("status = ? ", bean.PROMOTED).
		Where("active = ?", true).
		Where("artifact_id = ?", artifactId).
		Select()
	return model, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindByPipelineIdAndArtifactIds(pipelineId int, artifactIds []int) ([]*ArtifactPromotionApprovalRequest, error) {
	var model []*ArtifactPromotionApprovalRequest
	err := repo.dbConnection.Model(model).
		Where("destination_pipeline_id = ? ", pipelineId).
		Where("active = ?", true).
		Where("artifact_id in (?) ", artifactIds).
		Select()
	return model, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(models).
		Where("status = ? ", bean.AWAITING_APPROVAL).
		Where("active = ?", true).
		Where("artifact_id = ?", artifactId).
		Select()
	return models, err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) MarkStaleByIds(tx *pg.Tx, requestIds []int) error {
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", bean.STALE).
		Set("updated_on = ?", time.Now()).
		Set("active=?", false).
		Where("id IN (?)", pg.In(requestIds)).
		Where("active = ?", true).
		Update()
	return err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) MarkStaleByPolicyId(tx *pg.Tx, policyId int) error {
	_, err := repo.dbConnection.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", bean.STALE).
		Set("updated_on = ?", time.Now()).
		Set("active=?", false).
		Where("policy_id = ?", policyId).
		Where("active = ?", true).
		Update()
	return err
}

func (repo *ArtifactPromotionApprovalRequestRepoImpl) MarkPromoted(tx *pg.Tx, requestIds []int) error {
	_, err := tx.Model(&ArtifactPromotionApprovalRequest{}).
		Set("status = ?", bean.PROMOTED).
		Set("updated_on = ?", time.Now()).
		Set("active=?", false).
		Where("id IN (?)", pg.In(requestIds)).
		Where("active = ?", true).
		Update()
	return err
}
