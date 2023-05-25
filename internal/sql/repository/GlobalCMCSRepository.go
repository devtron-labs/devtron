package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalCMCSRepository interface {
	Save(model *GlobalCMCS) (*GlobalCMCS, error)
	Update(model *GlobalCMCS) (*GlobalCMCS, error)
	FindAllActive() ([]*GlobalCMCS, error)
	FindById(id int) (*GlobalCMCS, error)
	FindByConfigTypeAndName(configType, name string) (*GlobalCMCS, error)
	FindByMountPath(mountPath string) (*GlobalCMCS, error)
	FindAllActiveByPipelineType(pipelineType string) ([]*GlobalCMCS, error)
	Delete(model *GlobalCMCS) error
}

const (
	CM_TYPE_CONFIG     = "CONFIGMAP"
	CS_TYPE_CONFIG     = "SECRET"
	ENVIRONMENT_CONFIG = "environment"
	VOLUME_CONFIG      = "volume"
)

const (
	PIPELINE_TYPE_CI    = "CI"
	PIPELINE_TYPE_CD    = "CD"
	PIPELINE_TYPE_CI_CD = "CI/CD"
)

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
	ConfigType string   `sql:"config_type"` // [CONFIGMAP, SECRET]
	Name       string   `sql:"name"`
	Type       string   `sql:"type"` // [environment, volume]
	//json string of map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data               string `sql:"data"`
	MountPath          string `sql:"mount_path"`
	Deleted            bool   `sql:"deleted,notnull"`
	SecretIngestionFor string `sql:"secret_ingestion_for,notnull"` // [CI, CD, CI/CD]
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

func (impl *GlobalCMCSRepositoryImpl) FindAllActiveByPipelineType(pipelineType string) ([]*GlobalCMCS, error) {
	var models []*GlobalCMCS
	err := impl.dbConnection.Model(&models).
		Where("deleted = ?", false).
		Where("secret_ingestion_for = ? OR secret_ingestion_for = ?", pipelineType, PIPELINE_TYPE_CI_CD).
		Select()
	if err != nil {
		impl.logger.Errorw("err on getting global cm/cs config to be used by default in ci/cd pipeline", "err", err)
		return nil, err
	}
	return models, nil
}

func (impl *GlobalCMCSRepositoryImpl) FindById(id int) (*GlobalCMCS, error) {
	model := &GlobalCMCS{}
	err := impl.dbConnection.Model(model).Where("id = ? ", id).Select()
	if err != nil {
		impl.logger.Errorw("err on getting global cm/cs config to be used by default in ci/cd pipeline", "err", err)
		return nil, err
	}
	return model, err
}

func (impl *GlobalCMCSRepositoryImpl) Delete(model *GlobalCMCS) error {
	err := impl.dbConnection.Delete(model)
	if err != nil {
		impl.logger.Errorw("error in deleting global cm cs ", "err", err)
		return err
	}
	return nil
}
