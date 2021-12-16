package chartConfig

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

type ConfigmapAndSecretGlobalHistory struct {
	TableName           struct{}   `sql:"config_map_global_history" pg:",discard_unknown_columns"`
	Id                  int        `sql:"id,pk"`
	ConfigMapAppLevelId int        `sql:"config_map_app_level_id,notnull"`
	DataType            ConfigType `sql:"data_type"`
	Data                string     `sql:"data"`
	Deployed            bool       `sql:"deployed"`
	DeployedOn          time.Time  `sql:"deployed_on"`
	DeployedBy          int32      `sql:"deployed_by"`
	Latest              bool       `sql:"latest,notnull"`
	sql.AuditLog
}

type ConfigMapHistoryRepository interface {
	CreateGlobalHistory(model *ConfigmapAndSecretGlobalHistory) (*ConfigmapAndSecretGlobalHistory, error)
	UpdateGlobalHistory(model *ConfigmapAndSecretGlobalHistory) (*ConfigmapAndSecretGlobalHistory, error)
	GetLatestGlobalHistoryByAppLevelIdAndConfigType(appLevelId int, configType ConfigType) (*ConfigmapAndSecretGlobalHistory, error)
	GetLatestGlobalHistoryByAppId(appId int) ([]*ConfigmapAndSecretGlobalHistory, error)

	CreateEnvHistory(model *ConfigmapAndSecretEnvHistory) (*ConfigmapAndSecretEnvHistory, error)
	UpdateEnvHistory(model *ConfigmapAndSecretEnvHistory) (*ConfigmapAndSecretEnvHistory, error)
	GetLatestEnvHistoryByEnvLevelIdAndConfigType(envLevelId int, configType ConfigType) (*ConfigmapAndSecretEnvHistory, error)
	GetLatestEnvHistoryByAppIdAndEnvId(appId int, envId int) ([]*ConfigmapAndSecretEnvHistory, error)
}

type ConfigMapHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewConfigMapHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *ConfigMapHistoryRepositoryImpl {
	return &ConfigMapHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl ConfigMapHistoryRepositoryImpl) CreateGlobalHistory(model *ConfigmapAndSecretGlobalHistory) (*ConfigmapAndSecretGlobalHistory, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating global config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) UpdateGlobalHistory(model *ConfigmapAndSecretGlobalHistory) (*ConfigmapAndSecretGlobalHistory, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("err in updating global config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetLatestGlobalHistoryByAppId(appId int) ([]*ConfigmapAndSecretGlobalHistory, error) {
	var model []*ConfigmapAndSecretGlobalHistory
	err := impl.dbConnection.Model(&model).
		Join("INNER JOIN config_map_app_level cmapp on config_map_global_history.config_map_app_level_id = cmapp.id").
		Where("cmapp.app_id = ?", appId).
		Where("latest = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for global CM/CS history", "err", err, "appId", appId)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetLatestGlobalHistoryByAppLevelIdAndConfigType(appLevelId int, configType ConfigType) (*ConfigmapAndSecretGlobalHistory, error) {
	var model *ConfigmapAndSecretGlobalHistory
	err := impl.dbConnection.Model(&model).Where("config_map_app_level_id = ?", appLevelId).
		Where("latest = ?", true).Where("data_type = ?", configType).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for global CM/CS history", "err", err, "appLevelId", appLevelId, "configType", configType)
		return model, err
	}
	return model, nil
}

//----------------------------------------------

type ConfigmapAndSecretEnvHistory struct {
	TableName           struct{}   `sql:"config_map_env_history" pg:",discard_unknown_columns"`
	Id                  int        `sql:"id,pk"`
	ConfigMapEnvLevelId int        `sql:"config_map_env_level_id,notnull"`
	DataType            ConfigType `sql:"data_type"`
	Data                string     `sql:"data"`
	Deployed            bool       `sql:"deployed"`
	DeployedOn          time.Time  `sql:"deployed_on"`
	DeployedBy          int32      `sql:"deployed_by"`
	Latest              bool       `sql:"latest, notnull"`
	sql.AuditLog
}

func (impl ConfigMapHistoryRepositoryImpl) CreateEnvHistory(model *ConfigmapAndSecretEnvHistory) (*ConfigmapAndSecretEnvHistory, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating env config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) UpdateEnvHistory(model *ConfigmapAndSecretEnvHistory) (*ConfigmapAndSecretEnvHistory, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("err in updating env config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetLatestEnvHistoryByEnvLevelIdAndConfigType(envLevelId int, configType ConfigType) (*ConfigmapAndSecretEnvHistory, error) {
	var model *ConfigmapAndSecretEnvHistory
	err := impl.dbConnection.Model(&model).Where("config_map_env_level_id = ?", envLevelId).
		Where("latest = ?", true).Where("data_type = ?", configType).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for env CM/CS history", "err", err, "envLevelId", envLevelId, "configType", configType)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetLatestEnvHistoryByAppIdAndEnvId(appId int, envId int) ([]*ConfigmapAndSecretEnvHistory, error) {
	var model []*ConfigmapAndSecretEnvHistory
	err := impl.dbConnection.Model(&model).
		Join("INNER JOIN config_map_env_level cmenv on config_map_env_history.config_map_env_level_id = cmenv.id").
		Where("cmenv.app_id = ?", appId).Where("cmenv.environment_id = ?", envId).
		Where("latest = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("err in getting latest entry for env CM/CS history", "err", err, "envId", envId, "appId", appId)
		return model, err
	}
	return model, nil
}
