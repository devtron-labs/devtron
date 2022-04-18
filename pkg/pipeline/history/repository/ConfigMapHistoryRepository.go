package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigType string

const (
	CONFIGMAP_TYPE ConfigType = "CONFIGMAP"
	SECRET_TYPE    ConfigType = "SECRET"
)

type ConfigMapHistoryRepository interface {
	CreateHistory(model *ConfigmapAndSecretHistory) (*ConfigmapAndSecretHistory, error)
	GetHistoryForDeployedCMCSById(id, pipelineId int, configType ConfigType) (*ConfigmapAndSecretHistory, error)
	GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType ConfigType) ([]*ConfigmapAndSecretHistory, error)
}

type ConfigMapHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewConfigMapHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *ConfigMapHistoryRepositoryImpl {
	return &ConfigMapHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type ConfigmapAndSecretHistory struct {
	TableName  struct{}   `sql:"config_map_history" pg:",discard_unknown_columns"`
	Id         int        `sql:"id,pk"`
	PipelineId int        `sql:"pipeline_id"`
	AppId      int        `sql:"app_id"`
	DataType   ConfigType `sql:"data_type"`
	Data       string     `sql:"data"`
	Deployed   bool       `sql:"deployed"`
	DeployedOn time.Time  `sql:"deployed_on"`
	DeployedBy int32      `sql:"deployed_by"`
	sql.AuditLog
}

func (impl ConfigMapHistoryRepositoryImpl) CreateHistory(model *ConfigmapAndSecretHistory) (*ConfigmapAndSecretHistory, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating env config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetHistoryForDeployedCMCSById(id, pipelineId int, configType ConfigType) (*ConfigmapAndSecretHistory, error) {
	var history ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&history).Where("id = ?", id).
		Where("pipeline_id = ?", pipelineId).
		Where("data_type = ?", configType).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting CM/CS history", "err", err)
		return &history, err
	}
	return &history, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType ConfigType) ([]*ConfigmapAndSecretHistory, error) {
	var histories []*ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("data_type = ?", configType).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting deployed CM/CS history", "err", err)
		return histories, err
	}
	return histories, nil
}
