package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalCMCSRepository interface {
	Save(model *GlobalCMCS) (*GlobalCMCS, error)
	Update(model *GlobalCMCS) (*GlobalCMCS, error)
	FindAllActive() ([]*GlobalCMCS, error)
	FindByConfigTypeAndName(configType, name string) (*GlobalCMCS, error)
	FindByMountPath(mountPath string) (*GlobalCMCS, error)
}

type GlobalCMCSRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalCMCSRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *GlobalCMCSRepositoryImpl {
	return &GlobalCMCSRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type GlobalCMCS struct {
	TableName  struct{} `sql:"global_cm_cs" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	ConfigType string   `sql:"config_type"`
	Name       string   `sql:"name"`
	//json string of map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data      json.RawMessage `sql:"data"`
	MountPath string          `sql:"mount_path"`
	Deleted   bool            `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl *GlobalCMCSRepositoryImpl) Save(model *GlobalCMCS) (*GlobalCMCS, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err on saving global cm/cs config ", "err", err)
		return nil, err
	}
	return model, nil
}

func (impl *GlobalCMCSRepositoryImpl) Update(model *GlobalCMCS) (*GlobalCMCS, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Errorw("err on updating global cm/cs config ", "err", err)
		return nil, err
	}
	return model, nil
}

func (impl *GlobalCMCSRepositoryImpl) FindAllActive() ([]*GlobalCMCS, error) {
	var models []*GlobalCMCS
	err := impl.dbConnection.Model(&models).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err on getting global cm/cs config to be used by default in ci pipeline", "err", err)
		return nil, err
	}
	return models, nil
}

func (impl *GlobalCMCSRepositoryImpl) FindByConfigTypeAndName(configType, name string) (*GlobalCMCS, error) {
	model := &GlobalCMCS{}
	err := impl.dbConnection.Model(model).
		Where("config_type = ?", configType).
		Where("name = ?", name).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err on getting global cm/cs config by configType and name", "err", err)
		return nil, err
	}
	return model, nil
}

func (impl *GlobalCMCSRepositoryImpl) FindByMountPath(mountPath string) (*GlobalCMCS, error) {
	model := &GlobalCMCS{}
	err := impl.dbConnection.Model(model).
		Where("mount_path = ?", mountPath).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err on getting global cm/cs config by mountPath", "err", err, "mountPath", mountPath)
		return nil, err
	}
	return model, nil
}
