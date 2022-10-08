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
	FindAllActive() ([]*GlobalCMCSDto, error)
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

type GlobalCMCSDto struct {
	Id         int    `json:"id"`
	ConfigType string `json:"configType" validate:"oneof=CONFIGMAP SECRET"`
	Name       string `json:"name"  validate:"required"`
	Type       string `json:"type" validate:"oneof=environment volume"`
	//map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data      map[string]string `json:"data"  validate:"required"`
	MountPath string            `json:"mountPath"`
	Deleted   bool              `json:"deleted"`
	UserId    int32             `json:"-"`
}

func (impl *GlobalCMCSServiceImpl) Create(config *GlobalCMCSDto) (*GlobalCMCSDto, error) {
	//checking if same name config is present for this type
	sameNameConfig, err := impl.globalCMCSRepository.FindByConfigTypeAndName(config.ConfigType, config.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global cm/cs config by name and configType", "err", err, "configType", config.ConfigType, "name", config.Name)
		return nil, err
	}
	if config.Type == repository.VOLUME_CONFIG {
		//checking if same mountPath config is present for any type
		sameMountPathConfig, err := impl.globalCMCSRepository.FindByMountPath(config.MountPath)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting global cm/cs config by mountPath and configType", "err", err, "mountPath", config.MountPath)
			return nil, err
		}
		if (sameMountPathConfig != nil && sameMountPathConfig.Id > 0) && (sameNameConfig != nil && sameNameConfig.Id > 0) {
			impl.logger.Errorw("found global cm/cs config with same name and same mountPath", "configName", config.Name)
			return nil, fmt.Errorf("found configs with same name & mount path, please update the name & mountPath and try again")
		} else if sameMountPathConfig != nil && sameMountPathConfig.Id > 0 {
			impl.logger.Errorw("found global cm/cs config with same mountPath", "configName", config.Name)
			return nil, fmt.Errorf("found configs with same mount path, please update the mount path and try again")
		}
	}
	if sameNameConfig != nil && sameNameConfig.Id > 0 {
		impl.logger.Errorw("found global cm/cs config with same name", "configName", config.Name)
		return nil, fmt.Errorf("found %s with same name, please update the name and try again", config.ConfigType)
	}
	dataByte, err := json.Marshal(config.Data)
	if err != nil {
		impl.logger.Errorw("error in marshaling cm/cs data", "err", err)
		return nil, err
	}
	model := &repository.GlobalCMCS{
		ConfigType: config.ConfigType,
		Data:       json.RawMessage(dataByte),
		Name:       config.Name,
		MountPath:  config.MountPath,
		Type:       config.Type,
		Deleted:    false,
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

func (impl *GlobalCMCSServiceImpl) FindAllActive() ([]*GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting all global cm/cs configs", "err", err)
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
			Type:       model.Type,
			Data:       data,
			Name:       model.Name,
			MountPath:  model.MountPath,
		}
		configDtos = append(configDtos, configDto)
	}

	return configDtos, nil
}
