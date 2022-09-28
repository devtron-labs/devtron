package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalConfigMapRepository interface {
	Create(model *GlobalConfigMap) (*GlobalConfigMap, error)
	Update(model *GlobalConfigMap) (*GlobalConfigMap, error)
}

type GlobalConfigMapRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalConfigMapRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *GlobalConfigMapRepositoryImpl {
	return &GlobalConfigMapRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type GlobalConfigMap struct {
	TableName  struct{} `sql:"global_config_map" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	ConfigType string   `sql:"config_type,notnull"`
	Name       string   `sql:"name"`
	//json string of map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data                     string `sql:"data"`
	MountPath                string `sql:"mount_path"`
	UseByDefaultInCiPipeline bool   `sql:"use_by_default_in_ci_pipeline,notnull"`
	Deleted                  bool   `sql:"default,notnull"`
	sql.AuditLog
}

func (impl *GlobalConfigMapRepositoryImpl) Create(model *GlobalConfigMap) (*GlobalConfigMap, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err on saving global cm/cs config ", "err", err)
		return model, err
	}
	return model, nil
}

func (impl *GlobalConfigMapRepositoryImpl) Update(model *GlobalConfigMap) (*GlobalConfigMap, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("err on updating global cm/cs config ", "err", err)
		return model, err
	}
	return model, nil
}
