package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ManifestPushConfig struct {
	tableName         struct{} `sql:"manifest_push_config" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	AppId             int      `sql:"app_id"`
	EnvId             int      `sql:"env_id"`
	CredentialsConfig string   `sql:"credentials_config"`
	ChartName         string   `sql:"chart_name"`
	ChartBaseVersion  string   `sql:"chart_base_version"`
	StorageType       string   `sql:"storage_type"`
	Deleted           bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

type ManifestPushConfigRepository interface {
	SaveConfig(manifestPushConfig *ManifestPushConfig) (*ManifestPushConfig, error)
	GetManifestPushConfigByAppIdAndEnvId(appId, envId int) (*ManifestPushConfig, error)
	UpdateConfig(manifestPushConfig *ManifestPushConfig) error
	GetManifestPushConfigByStoreId(storeId string) (*ManifestPushConfig, error)
}

type ManifestPushConfigRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewManifestPushConfigRepository(logger *zap.SugaredLogger,
	dbConnection *pg.DB,
) *ManifestPushConfigRepositoryImpl {
	return &ManifestPushConfigRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (impl ManifestPushConfigRepositoryImpl) SaveConfig(manifestPushConfig *ManifestPushConfig) (*ManifestPushConfig, error) {
	err := impl.dbConnection.Insert(manifestPushConfig)
	if err != nil {
		return manifestPushConfig, err
	}
	return manifestPushConfig, err
}

func (impl ManifestPushConfigRepositoryImpl) GetManifestPushConfigByAppIdAndEnvId(appId, envId int) (*ManifestPushConfig, error) {
	manifestPushConfig := &ManifestPushConfig{}
	err := impl.dbConnection.Model(manifestPushConfig).
		Where("app_id = ? ", appId).
		Where("env_id = ? ", envId).
		Where("deleted = ? ", false).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return manifestPushConfig, err
	}
	return manifestPushConfig, nil
}

func (impl ManifestPushConfigRepositoryImpl) UpdateConfig(manifestPushConfig *ManifestPushConfig) error {
	_, err := impl.dbConnection.Model(manifestPushConfig).WherePK().Update()
	return err
}

func (impl ManifestPushConfigRepositoryImpl) GetManifestPushConfigByStoreId(storeId string) (*ManifestPushConfig, error) {
	manifestPushConfig := &ManifestPushConfig{}
	err := impl.dbConnection.Model(manifestPushConfig).
		Where("credentials_config LIKE ? ", "%\"ContainerRegistryName\":\""+storeId+"\"}").
		Where("deleted = ? ", false).
		Limit(1).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return manifestPushConfig, err
	}
	return manifestPushConfig, nil
}
