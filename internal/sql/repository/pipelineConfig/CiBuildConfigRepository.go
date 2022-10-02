package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiBuildConfig struct {
	tableName            struct{} `sql:"ci_build_config" pg:",discard_unknown_columns"`
	Id                   int      `sql:"id"`
	Type                 string   `sql:"type"`
	CiTemplateId         int      `sql:"ci_template_id"`
	CiTemplateOverrideId int      `sql:"ci_template_override_id"`
	BuildMetadata        string   `sql:"build_metadata"`
	sql.AuditLog
}

type CiBuildConfigRepository interface {
	Save(ciBuildConfig *CiBuildConfig) error
	Update(ciBuildConfig *CiBuildConfig) error
	Delete(ciBuildConfigId int) error
}

type CiBuildConfigRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiBuildConfigRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiBuildConfigRepositoryImpl {
	return &CiBuildConfigRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl CiBuildConfigRepositoryImpl) Save(ciBuildConfig *CiBuildConfig) error {

	err := impl.dbConnection.Insert(ciBuildConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while saving ciBuildConfig", "ciBuildConfig", ciBuildConfig, "err", err)
	}
	return err
}

func (impl CiBuildConfigRepositoryImpl) Update(ciBuildConfig *CiBuildConfig) error {
	err := impl.dbConnection.Update(ciBuildConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while updating ciBuildConfig", "err", err)
	}
	return err
}

func (impl CiBuildConfigRepositoryImpl) Delete(ciBuildConfigId int) error {
	err := impl.dbConnection.Delete(ciBuildConfigId)
	if err != nil {
		impl.logger.Errorw("error occurred while deleting ciBuildConfig", "ciBuildConfigId", ciBuildConfigId, "err", err)
	}
	return err
}
