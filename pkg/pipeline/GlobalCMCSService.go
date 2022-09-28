package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GlobalCMCSService interface {
	Create(model *GlobalCMCSDto) (*GlobalCMCSDto, error)
	FindAllDefaultInCiPipeline() ([]*GlobalCMCSDto, error)
}

type GlobalCMCSServiceImpl struct {
	logger               *zap.SugaredLogger
	globalCMCSRepository repository.GlobalCMCSRepository
}

func NewGlobalCMCSServiceImpl(logger *zap.SugaredLogger,
	globalCMCSRepository repository.GlobalCMCSRepository) *GlobalCMCSServiceImpl {
	return &GlobalCMCSServiceImpl{
		logger:               logger,
		globalCMCSRepository: globalCMCSRepository,
	}
}

const (
	CM_TYPE_CONFIG = "CONFIGMAP"
	CS_TYPE_CONFIG = "SECRET"
)

type GlobalCMCSDto struct {
	Id         int    `json:"id"`
	ConfigType string `json:"configType" validate:"oneof=CONFIGMAP SECRET"`
	Name       string `json:"name"  validate:"required"`
	//map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data                     map[string]string `json:"data"  validate:"required"`
	MountPath                string            `json:"mountPath"`
	UseByDefaultInCiPipeline bool              `json:"useByDefaultInCiPipeline"`
	Deleted                  bool              `json:"deleted"`
	UserId                   int32             `json:"-"`
}

func (impl *GlobalCMCSServiceImpl) Create(config *GlobalCMCSDto) (*GlobalCMCSDto, error) {
	//checking if same name config is present for this type
	sameNameConfig, err := impl.globalCMCSRepository.FindByConfigTypeAndName(config.ConfigType, config.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global cm/cs config by name and configType", "err", err, "configType", config.ConfigType, "name", config.Name)
		return nil, err
	}
	if sameNameConfig != nil && sameNameConfig.Id > 0 {
		impl.logger.Errorw("found global cm/cs config with same name", "configName", config.Name)
		return nil, fmt.Errorf(fmt.Sprintf("found %s with same name, please update the name and try again", config.ConfigType))
	}

	//checking if same name & same mountPath config is present for other type
	otherConfigType := ""
	if config.ConfigType == CM_TYPE_CONFIG {
		otherConfigType = CS_TYPE_CONFIG
	} else if config.ConfigType == CS_TYPE_CONFIG {
		otherConfigType = CM_TYPE_CONFIG
	}
	sameNameMountPathConfig, err := impl.globalCMCSRepository.FindByNameMountPathAndConfigType(otherConfigType, config.Name, config.MountPath)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global cm/cs config by name, mountPath and configType", "err", err, "configType", config.ConfigType, "name", config.Name, "mountPath", config.MountPath)
		return nil, err
	}
	if sameNameMountPathConfig != nil && sameNameMountPathConfig.Id > 0 {
		impl.logger.Errorw("found global cm/cs config with same name and mountPath", "configName", config.Name)
		return nil, fmt.Errorf(fmt.Sprintf("found %s with same name and mount path, please either update the name or mount path and try again", otherConfigType))
	}
	dataByte, err := json.Marshal(config.Data)
	if err != nil {
		impl.logger.Errorw("error in marshaling cm/cs data", "err", err)
		return nil, err
	}
	model := &repository.GlobalCMCS{
		ConfigType:               config.ConfigType,
		Data:                     string(dataByte),
		Name:                     config.Name,
		MountPath:                config.MountPath,
		UseByDefaultInCiPipeline: true,
		Deleted:                  false,
		AuditLog: sql.AuditLog{
			CreatedBy: config.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: config.UserId,
			UpdatedOn: time.Now(),
		},
	}
	model, err = impl.globalCMCSRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("err on creating global cm/cs config ", "err", err)
		return nil, err
	}
	config.Id = model.Id
	return config, nil
}

func (impl *GlobalCMCSServiceImpl) FindAllDefaultInCiPipeline() ([]*GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllDefaultInCiPipeline()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting global cm/cs config which are to be used as default in ci pipelines", "err", err)
		return nil, err
	}
	var configDtos []*GlobalCMCSDto
	for _, model := range models {
		data := make(map[string]string)
		err = json.Unmarshal([]byte(model.Data), &data)
		if err != nil {
			impl.logger.Errorw("error in un-marshaling cm/cs data", "err", err)
		}
		configDto := &GlobalCMCSDto{
			Id:         model.Id,
			ConfigType: model.ConfigType,
			Data:       data,
			Name:       model.Name,
			MountPath:  model.MountPath,
		}
		configDtos = append(configDtos, configDto)
	}

	return configDtos, nil
}
