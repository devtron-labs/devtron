package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalAuthorisationConfigRepository interface {
	StartATransaction() (*pg.Tx, error)
	CommitATransaction(tx *pg.Tx) error
	GetAllActiveConfigs() ([]*GlobalAuthorisationConfig, error)
	SetConfigToInactiveByConfigType(configType string) error
	GetByConfigType(configType string) (*GlobalAuthorisationConfig, error)
	GetByConfigTypes(configType []string) ([]*GlobalAuthorisationConfig, error)
	CreateConfig(tx *pg.Tx, model []*GlobalAuthorisationConfig) error
	UpdateConfig(tx *pg.Tx, model []*GlobalAuthorisationConfig) error
	SetConfigsToInactiveExceptGivenConfigs(tx *pg.Tx, configTypes []string) error
}

type GlobalAuthorisationConfigRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

type GlobalAuthorisationConfig struct {
	TableName  struct{} `sql:"global_authorisation_config" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	ConfigType string   `sql:"config_type,notnull"`
	Active     bool     `sql:"active,notnull"`
	sql.AuditLog
}

func NewGlobalAuthorisationConfigRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *GlobalAuthorisationConfigRepositoryImpl {
	globalAuthorisationConfigRepositoryImpl := &GlobalAuthorisationConfigRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
	return globalAuthorisationConfigRepositoryImpl
}
func (repo *GlobalAuthorisationConfigRepositoryImpl) StartATransaction() (*pg.Tx, error) {
	tx, err := repo.dbConnection.Begin()
	if err != nil {
		repo.logger.Errorw("error in beginning a transaction", "err", err)
		return nil, err
	}
	return tx, nil
}
func (repo *GlobalAuthorisationConfigRepositoryImpl) CommitATransaction(tx *pg.Tx) error {
	err := tx.Commit()
	if err != nil {
		repo.logger.Errorw("error in commiting a transaction", "err", err)
		return err
	}
	return nil
}

func (repo *GlobalAuthorisationConfigRepositoryImpl) GetAllActiveConfigs() ([]*GlobalAuthorisationConfig, error) {
	var models []*GlobalAuthorisationConfig
	err := repo.dbConnection.Model(&models).Where("active = ?", true).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all active config from global authorisation config", "err", err)
		return models, err
	}
	return models, err
}

func (repo *GlobalAuthorisationConfigRepositoryImpl) SetConfigToInactiveByConfigType(configType string) error {
	var model *GlobalAuthorisationConfig
	_, err := repo.dbConnection.Model(&model).Set("active = ?", false).
		Where("active = ?", true).
		Where("config_type = ?", configType).
		Update()
	if err != nil {
		repo.logger.Errorw("error in setting config type to inactive", "err", err, "configType", configType)
		return err
	}
	return nil
}
func (repo *GlobalAuthorisationConfigRepositoryImpl) GetByConfigType(configType string) (*GlobalAuthorisationConfig, error) {
	var model *GlobalAuthorisationConfig
	err := repo.dbConnection.Model(&model).
		Where("config_type= ? ", configType).
		Where("active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting config by configType", "err", err, "configType", configType)
		return nil, err
	}
	return model, err
}

func (repo *GlobalAuthorisationConfigRepositoryImpl) GetByConfigTypes(configTypes []string) ([]*GlobalAuthorisationConfig, error) {
	var model []*GlobalAuthorisationConfig
	err := repo.dbConnection.Model(&model).
		Where("config_type in (?) ", pg.In(configTypes)).
		Select()
	if err != nil {
		repo.logger.Errorw("error in getting configs by configTypes", "err", err, "configTypes", configTypes)
		return nil, err
	}
	return model, err
}

func (repo *GlobalAuthorisationConfigRepositoryImpl) CreateConfig(tx *pg.Tx, model []*GlobalAuthorisationConfig) error {
	err := tx.Insert(&model)
	if err != nil {
		repo.logger.Errorw("error in creating global authorisation config", "err", err)
		return err
	}
	return nil
}

func (repo *GlobalAuthorisationConfigRepositoryImpl) UpdateConfig(tx *pg.Tx, model []*GlobalAuthorisationConfig) error {
	_, err := tx.Model(&model).Update()
	if err != nil {
		repo.logger.Errorw("error in updating global authorisation config", "err", err)
		return err
	}
	return nil

}

func (repo *GlobalAuthorisationConfigRepositoryImpl) SetConfigsToInactiveExceptGivenConfigs(tx *pg.Tx, configTypes []string) error {
	var model []*GlobalAuthorisationConfig
	_, err := tx.Model(&model).Set("active = ?", false).Where("config_type not in (?)", pg.In(configTypes)).
		Update()
	if err != nil {
		repo.logger.Errorw("error in updating configs to inactive except configs", "err", err, "configTypes", configTypes)
		return err
	}
	return err
}
