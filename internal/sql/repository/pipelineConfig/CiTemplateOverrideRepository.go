package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiTemplateOverride struct {
	tableName        struct{} `sql:"ci_template_override" pg:",discard_unknown_columns"`
	Id               int      `sql:"id"`
	CiPipelineId     int      `sql:"ci_pipeline_id"`
	DockerRegistryId string   `sql:"docker_registry_id"`
	DockerRepository string   `sql:"docker_repository"`
	DockerfilePath   string   `sql:"dockerfile_path"`
	GitMaterialId    int      `sql:"git_material_id"`
	Active           bool     `sql:"active,notnull"`
	sql.AuditLog
	GitMaterial    *GitMaterial
	DockerRegistry *repository.DockerArtifactStore
}

type CiTemplateOverrideRepository interface {
	FindByAppId(appId int) ([]*CiTemplateOverride, error)
	FindByCiPipelineId(ciPipelineId int) (*CiTemplateOverride, error)
}

type CiTemplateOverrideRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiTemplateOverrideRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *CiTemplateOverrideRepositoryImpl {
	return &CiTemplateOverrideRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo *CiTemplateOverrideRepositoryImpl) FindByAppId(appId int) ([]*CiTemplateOverride, error) {
	var ciTemplateOverrides []*CiTemplateOverride
	err := repo.dbConnection.Model(ciTemplateOverrides).
		Join("INNER JOIN ci_pipeline cp on cp.id=ci_template_override.ci_pipeline_id").
		Where("app_id = ?", appId).
		Where("is_docker_config_override = ?", true).
		Where("ci_template_override.active = ?", true).
		Where("cp.deleted = ?", false).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by appId", "err", err, "appId", appId)
		return nil, err
	}
	return ciTemplateOverrides, nil
}

func (repo *CiTemplateOverrideRepositoryImpl) FindByCiPipelineId(ciPipelineId int) (*CiTemplateOverride, error) {
	var ciTemplateOverride CiTemplateOverride
	err := repo.dbConnection.Model(&ciTemplateOverride).
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	return &ciTemplateOverride, nil
}
