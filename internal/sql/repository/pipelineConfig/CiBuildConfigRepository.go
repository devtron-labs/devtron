package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiBuildConfig struct {
	tableName     struct{} `sql:"ci_build_config" pg:",discard_unknown_columns"`
	Id            int      `sql:"id"`
	CiTemplateId  int      `sql:"ci_template_id"`
	CiPipelineId  int      `sql:"ci_pipeline_id"`
	Type          string   `sql:"type"`
	BuildMetadata string   `sql:"build_metadata"`
	sql.AuditLog
}

type CiBuildConfigRepository interface {
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
