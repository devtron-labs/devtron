package pipelineConfig

import (
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
}

type CiTemplateOverrideRepository interface {
	FindByCiPipelineIds(ciPipelineIds []int) ([]*CiTemplateOverride, error)
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

func (repo *CiTemplateOverrideRepositoryImpl) FindByCiPipelineIds(ciPipelineIds []int) ([]*CiTemplateOverride, error) {
	var ciTemplateOverrides []*CiTemplateOverride
	err := repo.dbConnection.Model(ciTemplateOverrides).
		Where("ci_pipeline_id in (?)", pg.In(ciPipelineIds)).
		Where("active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting ciTemplateOverride by ciPipelineIds", "err", err, "ciPipelineIds", ciPipelineIds)
		return nil, err
	}
	return ciTemplateOverrides, nil
}
