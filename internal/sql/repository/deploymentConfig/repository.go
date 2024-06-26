package deploymentConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type DeploymentConfig struct {
	Id                 int    `sql:"id,pk"`
	AppId              int    `sql:"app_id"`
	EnvironmentId      int    `sql:"environment_id"`
	DeploymentAppType  string `sql:"deployment_app_type"`
	ConfigType         string `sql:"config_type"`
	RepoUrl            string `sql:"repo_url"`
	ChartLocation      string `sql:"chart_location"`
	CredentialType     string `sql:"credential_type"`
	CredentialIdInt    int    `sql:"credential_id_int"`
	CredentialIdString string `sql:"credential_id_string"`
	Active             bool   `sql:"active"`
	sql.AuditLog
}

type Repository interface {
	Save(config *DeploymentConfig) (*DeploymentConfig, error)
	Update(config *DeploymentConfig) (*DeploymentConfig, error)
	GetById(id int) (*DeploymentConfig, error)
	GetByAppIdAndEnvId(appId, envId int) (*DeploymentConfig, error)
	GetAppLevelConfig(appId int) (*DeploymentConfig, error)
}

type RepositoryImpl struct {
	dbConnection *pg.DB
}

func NewRepositoryImpl(dbConnection *pg.DB) *RepositoryImpl {
	return &RepositoryImpl{dbConnection: dbConnection}
}

func (impl RepositoryImpl) Save(config *DeploymentConfig) (*DeploymentConfig, error) {
	err := impl.dbConnection.Insert(config)
	return config, err
}

func (impl RepositoryImpl) Update(config *DeploymentConfig) (*DeploymentConfig, error) {
	err := impl.dbConnection.Update(config)
	return config, err
}

func (impl RepositoryImpl) GetById(id int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).Where("id = ?", id).Where("active = ?", true).Select()
	return result, err
}

func (impl RepositoryImpl) GetByAppIdAndEnvId(appId, envId int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).
		Where("app_id = ?", appId).
		Where("environment_id = ? ", envId).
		Where("active = ?", true).
		Select()
	return result, err
}

func (impl RepositoryImpl) GetAppLevelConfig(appId int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).
		Where("app_id = ? ", appId).
		Where("environment_id is NULL").
		Where("active = ?", true).
		Select()
	return result, err
}
