package deploymentConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type DeploymentConfig struct {
	tableName          struct{} `sql:"deployment_config" pg:",discard_unknown_columns"`
	Id                 int      `sql:"id,pk"`
	AppId              int      `sql:"app_id"`
	EnvironmentId      int      `sql:"environment_id"`
	DeploymentAppType  string   `sql:"deployment_app_type"`
	ConfigType         string   `sql:"config_type"`
	RepoUrl            string   `sql:"repo_url"`
	RepoName           string   `sql:"repo_name"`
	ChartLocation      string   `sql:"chart_location"`
	CredentialType     string   `sql:"credential_type"`
	CredentialIdInt    int      `sql:"credential_id_int"`
	CredentialIdString string   `sql:"credential_id_string"`
	Active             bool     `sql:"active"`
	sql.AuditLog
}

type Repository interface {
	Save(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error)
	SaveAll(tx *pg.Tx, configs []*DeploymentConfig) ([]*DeploymentConfig, error)
	Update(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error)
	GetById(id int) (*DeploymentConfig, error)
	GetByAppIdAndEnvId(appId, envId int) (*DeploymentConfig, error)
	GetAppLevelConfig(appId int) (*DeploymentConfig, error)
	GetAppLevelConfigByAppIds(appIds []int) ([]*DeploymentConfig, error)
	GetAppAndEnvLevelConfigsInBulk(envIdToAppIdMap map[int]int) ([]*DeploymentConfig, error)
}

type RepositoryImpl struct {
	dbConnection *pg.DB
}

func NewRepositoryImpl(dbConnection *pg.DB) *RepositoryImpl {
	return &RepositoryImpl{dbConnection: dbConnection}
}

func (impl RepositoryImpl) Save(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Insert(config)
	} else {
		err = impl.dbConnection.Insert(config)
	}
	return config, err
}

func (impl RepositoryImpl) SaveAll(tx *pg.Tx, configs []*DeploymentConfig) ([]*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Insert(configs)
	} else {
		err = impl.dbConnection.Insert(configs)
	}
	return configs, err
}

func (impl RepositoryImpl) Update(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Update(config)
	} else {
		err = impl.dbConnection.Update(config)
	}
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

func (impl RepositoryImpl) GetAppLevelConfigByAppIds(appIds []int) ([]*DeploymentConfig, error) {
	var result []*DeploymentConfig
	err := impl.dbConnection.Model(&result).
		Where("app_id in (?) ", appIds).
		Where("active = ?", true).
		Select()
	return result, err
}

func (impl RepositoryImpl) GetAppAndEnvLevelConfigsInBulk(envIdToAppIdMap map[int]int) ([]*DeploymentConfig, error) {
	var result []*DeploymentConfig
	err := impl.dbConnection.Model(&result).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			for envId, appId := range envIdToAppIdMap {
				query = query.Where("environment_id = ? and app_id = ? ", envId, appId)
			}
			return query, nil
		}).
		Where("active = ?", true).
		Select()
	return result, err
}
