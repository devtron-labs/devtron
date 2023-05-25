package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiTemplateOverride struct {
	tableName                 struct{} `sql:"ci_template_override" pg:",discard_unknown_columns"`
	Id                        int      `sql:"id"`
	CiPipelineId              int      `sql:"ci_pipeline_id"`
	DockerRegistryId          string   `sql:"docker_registry_id"`
	DockerRepository          string   `sql:"docker_repository"`
	DockerfilePath            string   `sql:"dockerfile_path"`
	GitMaterialId             int      `sql:"git_material_id"`
	BuildContextGitMaterialId int      `sql:"build_context_git_material_id"`
	Active                    bool     `sql:"active,notnull"`
	CiBuildConfigId           int      `sql:"ci_build_config_id"`
	sql.AuditLog
	GitMaterial    *GitMaterial
	DockerRegistry *repository.DockerArtifactStore
	CiBuildConfig  *CiBuildConfig
}

type CiTemplateOverrideRepository interface {
	Save(templateOverrideConfig *CiTemplateOverride) (*CiTemplateOverride, error)
	Update(templateOverrideConfig *CiTemplateOverride) (*CiTemplateOverride, error)
	FindByAppId(appId int) ([]*CiTemplateOverride, error)
	FindByCiPipelineIds(ciPipelineIds []int) ([]*CiTemplateOverride, error)
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

func (repo *CiTemplateOverrideRepositoryImpl) Save(templateOverrideConfig *CiTemplateOverride) (*CiTemplateOverride, error) {
	err := repo.dbConnection.Insert(templateOverrideConfig)
	if err != nil {
		repo.logger.Errorw("error in saving templateOverrideConfig", "err", err)
		return nil, err
	}
	return templateOverrideConfig, nil
}

func (repo *CiTemplateOverrideRepositoryImpl) Update(templateOverrideConfig *CiTemplateOverride) (*CiTemplateOverride, error) {
	err := repo.dbConnection.Update(templateOverrideConfig)
	if err != nil {
		repo.logger.Errorw("error in updating templateOverrideConfig", "err", err)
		return nil, err
	}
	return templateOverrideConfig, nil
}

func (repo *CiTemplateOverrideRepositoryImpl) FindByAppId(appId int) ([]*CiTemplateOverride, error) {
	var ciTemplateOverrides []*CiTemplateOverride
	err := repo.dbConnection.Model(&ciTemplateOverrides).
		Column("ci_template_override.*", "CiBuildConfig").
		Join("INNER JOIN ci_pipeline cp on cp.id=ci_template_override.ci_pipeline_id").
		Join("INNER JOIN ci_build_config cbc on cbc.id=ci_template_override.ci_build_config_id").
		Where("app_id = ?", appId).
		Where("is_docker_config_overridden = ?", true).
		Where("ci_template_override.active = ?", true).
		Where("cp.deleted = ?", false).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by appId", "err", err, "appId", appId)
		return nil, err
	}
	return ciTemplateOverrides, nil
}

func (repo *CiTemplateOverrideRepositoryImpl) FindByCiPipelineIds(ciPipelineIds []int) ([]*CiTemplateOverride, error) {
	var ciTemplateOverrides []*CiTemplateOverride
	err := repo.dbConnection.Model(&ciTemplateOverrides).
		Column("ci_template_override.*", "CiBuildConfig").
		Where("ci_template_override.ci_pipeline_id in (?)", pg.In(ciPipelineIds)).
		Where("ci_template_override.active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by appId", "err", err, "ciPipelineIds", ciPipelineIds)
		return nil, err
	}
	return ciTemplateOverrides, nil
}

func (repo *CiTemplateOverrideRepositoryImpl) FindByCiPipelineId(ciPipelineId int) (*CiTemplateOverride, error) {
	ciTemplateOverride := &CiTemplateOverride{}
	err := repo.dbConnection.Model(ciTemplateOverride).
		Column("ci_template_override.*", "GitMaterial", "DockerRegistry", "CiBuildConfig").
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("ci_template_override.active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return ciTemplateOverride, err
	}
	return ciTemplateOverride, nil
}
