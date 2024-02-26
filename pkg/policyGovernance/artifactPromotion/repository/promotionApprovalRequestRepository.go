package repository

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ArtifactPromotionApprovalRequest struct {
	tableName               struct{}        `sql:"artifact_promotion_approval_request" pg:",discard_unknown_columns"`
	Id                      int             `sql:"id"`
	PolicyId                int             `sql:"policy_id"`
	PolicyEvaluationAuditId int             `sql:"policy_evaluation_audit_id"`
	ArtifactId              int             `sql:"artifact_id"`
	SourceType              bean.SourceType `sql:"source_type"`
	SourcePipelineId        int             `sql:"source_pipeline_id"`
	DestinationPipelineId   int             `sql:"destination_pipeline_id"`
	Status                  int             `sql:"status"`
	Active                  bool            `sql:"active"`
	sql.AuditLog
}

type ArtifactPromotionApprovalRequestRepoImpl struct {
	dbConnection *pg.DB
}

func NewArtifactPromotionApprovalRequestImpl(dbConnection *pg.DB) *ArtifactPromotionApprovalRequestRepoImpl {
	return &ArtifactPromotionApprovalRequestRepoImpl{
		dbConnection: dbConnection,
	}
}

type ArtifactPromotionApprovalRequestRepository interface {
	Create(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	Update(PromotionRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalRequest, error)
	FindById(id int) (*ArtifactPromotionApprovalRequest, error)
	FindPendingByDestinationPipelineId(destinationPipelineId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
	FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, artifactId int) (*ArtifactPromotionApprovalRequest, error)
	FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error)
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

func (repo *ArtifactPromotionApprovalRequestRepoImpl) FindAwaitedRequestsByArtifactId(artifactId int) ([]*ArtifactPromotionApprovalRequest, error) {
	models := make([]*ArtifactPromotionApprovalRequest, 0)
	err := repo.dbConnection.Model(models).
		Where("status = ? ", bean.AWAITING_APPROVAL).
		Where("active = ?", true).
		Where("artifact_id = ?", artifactId).
		Select()
	return models, err
}
