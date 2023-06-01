package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GlobalCMCSService interface {
	Create(model *bean.GlobalCMCSDto) (*bean.GlobalCMCSDto, error)
	UpdateDataById(config *GlobalCMCSDataUpdateDto) (*GlobalCMCSDataUpdateDto, error)
	GetGlobalCMCSDataByConfigTypeAndName(configName string, configType string) (*bean.GlobalCMCSDto, error)
	FindAllActiveByPipelineType(pipelineType string) ([]*bean.GlobalCMCSDto, error)
	FindAllActive() ([]*bean.GlobalCMCSDto, error)
	DeleteById(id int) error
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

type GlobalCMCSDataUpdateDto struct {
	Id                 int               `json:"id"`
	Data               map[string]string `json:"data"  validate:"required"`
	SecretIngestionFor string            `json:"SecretIngestionFor"` // value can be one of [ci, cd, ci/cd]
	UserId             int32             `json:"-"`
}

func (impl *GlobalCMCSServiceImpl) validateGlobalCMCSData(config *bean.GlobalCMCSDto) error {
	sameNameConfig, err := impl.globalCMCSRepository.FindByConfigTypeAndName(config.ConfigType, config.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global cm/cs config by name and configType", "err", err, "configType", config.ConfigType, "name", config.Name)
		return err
	}
	if config.Type == repository.VOLUME_CONFIG {
		//checking if same mountPath config is present for any type
		sameMountPathConfig, err := impl.globalCMCSRepository.FindByMountPath(config.MountPath)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting global cm/cs config by mountPath and configType", "err", err, "mountPath", config.MountPath)
			return err
		}
		if (sameMountPathConfig != nil && sameMountPathConfig.Id > 0) && (sameNameConfig != nil && sameNameConfig.Id > 0) {
			impl.logger.Errorw("found global cm/cs config with same name and same mountPath", "configName", config.Name)
			return fmt.Errorf("found configs with same name & mount path, please update the name & mountPath and try again")
		} else if sameMountPathConfig != nil && sameMountPathConfig.Id > 0 {
			impl.logger.Errorw("found global cm/cs config with same mountPath", "configName", config.Name)
			return fmt.Errorf("found configs with same mount path, please update the mount path and try again")
		}
	}
	if sameNameConfig != nil && sameNameConfig.Id > 0 {
		impl.logger.Errorw("found global cm/cs config with same name", "configName", config.Name)
		return fmt.Errorf("found %s with same name, please update the name and try again", config.ConfigType)
	}
	return nil
}

func (impl *GlobalCMCSServiceImpl) Create(config *bean.GlobalCMCSDto) (*bean.GlobalCMCSDto, error) {

	err := impl.validateGlobalCMCSData(config)
	if err != nil {
		return nil, err
	}
	//checking if same name config is present for this type
	dataByte, err := json.Marshal(config.Data)
	if err != nil {
		impl.logger.Errorw("error in marshaling cm/cs data", "err", err)
		return nil, err
	}
	if config.SecretIngestionFor == "" {
		config.SecretIngestionFor = "CI/CD"
	}
	model := &repository.GlobalCMCS{
		ConfigType: config.ConfigType,
		Data:       string(dataByte),
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
		SecretIngestionFor: config.SecretIngestionFor,
	}
	model, err = impl.globalCMCSRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("err on creating global cm/cs config ", "err", err)
		return nil, err
	}
	config.Id = model.Id
	return config, nil
}

func (impl *GlobalCMCSServiceImpl) UpdateDataById(config *GlobalCMCSDataUpdateDto) (*GlobalCMCSDataUpdateDto, error) {

	model, err := impl.globalCMCSRepository.FindById(config.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching data from global cm cs")
		return nil, err
	}
	//checking if same name config is present for this type
	dataByte, err := json.Marshal(config.Data)
	if err != nil {
		impl.logger.Errorw("error in marshaling cm/cs data", "err", err)
		return nil, err
	}
	model.Data = string(dataByte)
	if config.SecretIngestionFor != "" {
		model.SecretIngestionFor = config.SecretIngestionFor
	}
	model.UpdatedBy = config.UserId
	model.UpdatedOn = time.Now()
	model, err = impl.globalCMCSRepository.Update(model)
	if err != nil {
		impl.logger.Errorw("err on creating global cm/cs config ", "err", err)
		return nil, err
	}
	config.Id = model.Id
	return config, nil
}

func (impl *GlobalCMCSServiceImpl) ConvertGlobalCmcsDbObjectToGlobalCmcsDto(GlobalCMCSDBObject []*repository.GlobalCMCS) []*bean.GlobalCMCSDto {
	var configDtos []*bean.GlobalCMCSDto
	for _, model := range GlobalCMCSDBObject {
		data := make(map[string]string)
		err := json.Unmarshal([]byte(model.Data), &data)
		if err != nil {
			impl.logger.Errorw("error in un-marshaling cm/cs data", "err", err)
		}
		configDto := &bean.GlobalCMCSDto{
			Id:         model.Id,
			ConfigType: model.ConfigType,
			Type:       model.Type,
			Data:       data,
			Name:       model.Name,
			MountPath:  model.MountPath,
		}
		configDtos = append(configDtos, configDto)
	}
	return configDtos
}

func (impl *GlobalCMCSServiceImpl) FindAllActiveByPipelineType(pipelineType string) ([]*bean.GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllActiveByPipelineType(pipelineType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting all global cm/cs configs", "err", err)
		return nil, err
	}
	configDtos := impl.ConvertGlobalCmcsDbObjectToGlobalCmcsDto(models)
	return configDtos, nil
}

func (impl *GlobalCMCSServiceImpl) FindAllActive() ([]*bean.GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting all global cm/cs configs", "err", err)
		return nil, err
	}
	configDtos := impl.ConvertGlobalCmcsDbObjectToGlobalCmcsDto(models)
	return configDtos, nil
}

func (impl *GlobalCMCSServiceImpl) GetGlobalCMCSDataByConfigTypeAndName(configName string, configType string) (*bean.GlobalCMCSDto, error) {

	model, err := impl.globalCMCSRepository.FindByConfigTypeAndName(configType, configName)
	if err != nil {
		impl.logger.Errorw("error in fetching data from ")
		return nil, err
	}
	data := make(map[string]string)
	err = json.Unmarshal([]byte(model.Data), &data)
	if err != nil {
		impl.logger.Errorw("error in un-marshaling cm/cs data", "err", err)
	}
	GlobalCMCSDto := &bean.GlobalCMCSDto{
		Id:                 model.Id,
		ConfigType:         model.ConfigType,
		Name:               model.Name,
		Type:               model.Type,
		Data:               data,
		MountPath:          model.MountPath,
		Deleted:            model.Deleted,
		SecretIngestionFor: model.SecretIngestionFor,
	}
	return GlobalCMCSDto, nil
}

func (impl *GlobalCMCSServiceImpl) DeleteById(id int) error {

	model, err := impl.globalCMCSRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching model by id", "err", err)
		return err
	}
	err = impl.globalCMCSRepository.Delete(model)
	if err != nil {
		impl.logger.Errorw("error in deleting model")
		return err
	}
	return nil
}
