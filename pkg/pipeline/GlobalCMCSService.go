package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type GlobalConfigMapService interface {
	Create(model *GlobalConfigMap) (*GlobalConfigMap, error)
	Update(model *GlobalConfigMap) (*GlobalConfigMap, error)
}

type GlobalConfigMapServiceImpl struct {
	logger                    *zap.SugaredLogger
	globalConfigMapRepository repository.GlobalConfigMapRepository
}

func NewGlobalConfigMapServiceImpl(logger *zap.SugaredLogger, globalConfigMapRepository repository.GlobalConfigMapRepository) *GlobalConfigMapServiceImpl {
	return &GlobalConfigMapServiceImpl{
		logger:                    logger,
		globalConfigMapRepository: globalConfigMapRepository,
	}
}

type GlobalConfigMapDto struct {
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
		impl.Logger.Errorw("err on saving global cm/cs config ", "err", err)
		return model, err
	}
	return model, nil
}

func (impl *GlobalConfigMapRepositoryImpl) Update(model *GlobalConfigMap) (*GlobalConfigMap, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Errorw("err on updating global cm/cs config ", "err", err)
		return model, err
	}
	return model, nil
}
