package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
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
	GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType ConfigType) (*ConfigmapAndSecretHistory, error)
	GetDeployedHistoryList(pipelineId, baseConfigId int, configType ConfigType, componentName string) ([]*ConfigmapAndSecretHistory, error)
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
	CdWorkflowRunner *pipelineConfig.CdWorkflowRunner
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

func (impl ConfigMapHistoryRepositoryImpl) GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType ConfigType) (*ConfigmapAndSecretHistory, error) {
	var history ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&history).Join("INNER JOIN cd_workflow_runner cwr ON cwr.started_on = config_map_history.deployed_on").
		Where("config_map_history.pipeline_id = ?", pipelineId).
		Where("config_map_history.deployed = ?", true).
		Where("config_map_history.data_type = ?", configType).
		Where("cwr.id = ?", wfrId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting configmap/secret history by pipelineId & wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return &history, err
	}
	return &history, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetDeployedHistoryList(pipelineId, baseConfigId int, configType ConfigType, componentName string) ([]*ConfigmapAndSecretHistory, error) {
	var histories []*ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&histories).
		Column("config_map_history.*", "CdWorkflowRunner").
		Join("INNER JOIN cd_workflow_runner cwr ON cwr.started_on = config_map_history.deployed_on").
		Where("config_map_history.pipeline_id = ?", pipelineId).
		Where("config_map_history.deployed = ?", true).
		Where("config_map_history.id <= ?", baseConfigId).
		Where("config_map_history.data_type = ?", configType).
		Where("config_map_history.data LIKE ?", "%"+fmt.Sprintf("\"name\":\"%s\"", componentName)+"%").
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting configmap/secret history list by pipelineId", "err", err, "pipelineId", pipelineId)
		return histories, err
	}
	return histories, nil
}
