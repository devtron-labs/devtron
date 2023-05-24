package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"
)

type GlobalCMCSService interface {
	Create(model *GlobalCMCSDto) (*GlobalCMCSDto, error)
	UpdateDataById(config *GlobalCMCSDataUpdateDto) (*GlobalCMCSDataUpdateDto, error)
	GetGlobalCMCSDataByConfigTypeAndName(configName string, configType string) (*GlobalCMCSDto, error)
	FindAllActiveByPipelineType(pipelineType string) ([]*GlobalCMCSDto, error)
	FindAllActive() ([]*GlobalCMCSDto, error)
	AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs []*GlobalCMCSDto, steps *[]v1alpha1.ParallelSteps, volumes *[]v12.Volume, templates *[]v1alpha1.Template) error
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

type GlobalCMCSDto struct {
	Id         int    `json:"id"`
	ConfigType string `json:"configType" validate:"oneof=CONFIGMAP SECRET"`
	Name       string `json:"name"  validate:"required"`
	Type       string `json:"type" validate:"oneof=environment volume"`
	//map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data               map[string]string `json:"data"  validate:"required"`
	MountPath          string            `json:"mountPath"`
	Deleted            bool              `json:"deleted"`
	UserId             int32             `json:"-"`
	SecretIngestionFor string            `json:"SecretIngestionFor"` // value can be one of [ci, cd, ci/cd]
}

type GlobalCMCSDataUpdateDto struct {
	Id                 int               `json:"id"`
	Data               map[string]string `json:"data"  validate:"required"`
	SecretIngestionFor string            `json:"SecretIngestionFor"` // value can be one of [ci, cd, ci/cd]
	UserId             int32             `json:"-"`
}

func (impl *GlobalCMCSServiceImpl) validateGlobalCMCSData(config *GlobalCMCSDto) error {
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

func (impl *GlobalCMCSServiceImpl) Create(config *GlobalCMCSDto) (*GlobalCMCSDto, error) {

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

func (impl *GlobalCMCSServiceImpl) ConvertGlobalCmcsDbObjectToGlobalCmcsDto(GlobalCMCSDBObject []*repository.GlobalCMCS) []*GlobalCMCSDto {
	var configDtos []*GlobalCMCSDto
	for _, model := range GlobalCMCSDBObject {
		data := make(map[string]string)
		err := json.Unmarshal([]byte(model.Data), &data)
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
	return configDtos
}

func (impl *GlobalCMCSServiceImpl) FindAllActiveByPipelineType(pipelineType string) ([]*GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllActiveByPipelineType(pipelineType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting all global cm/cs configs", "err", err)
		return nil, err
	}
	configDtos := impl.ConvertGlobalCmcsDbObjectToGlobalCmcsDto(models)
	return configDtos, nil
}

func (impl *GlobalCMCSServiceImpl) FindAllActive() ([]*GlobalCMCSDto, error) {
	models, err := impl.globalCMCSRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting all global cm/cs configs", "err", err)
		return nil, err
	}
	configDtos := impl.ConvertGlobalCmcsDbObjectToGlobalCmcsDto(models)
	return configDtos, nil
}

func (impl *GlobalCMCSServiceImpl) GetGlobalCMCSDataByConfigTypeAndName(configName string, configType string) (*GlobalCMCSDto, error) {

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
	GlobalCMCSDto := &GlobalCMCSDto{
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

func (impl *GlobalCMCSServiceImpl) AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs []*GlobalCMCSDto, steps *[]v1alpha1.ParallelSteps, volumes *[]v12.Volume, templates *[]v1alpha1.Template) error {

	cmIndex := 0
	csIndex := 0

	for _, config := range globalCmCsConfigs {
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			ownerDelete := true
			cmBody := v12.ConfigMap{
				TypeMeta: v1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: config.Name,
					OwnerReferences: []v1.OwnerReference{{
						APIVersion:         "argoproj.io/v1alpha1",
						Kind:               "Workflow",
						Name:               "{{workflow.name}}",
						UID:                "{{workflow.uid}}",
						BlockOwnerDeletion: &ownerDelete,
					}},
				},
				Data: config.Data,
			}
			cmJson, err := json.Marshal(cmBody)
			if err != nil {
				impl.logger.Errorw("error in building json", "err", err)
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						ConfigMap: &v12.ConfigMapVolumeSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: config.Name,
							},
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-cm-gb-" + strconv.Itoa(cmIndex),
						Template: "cm-gb-" + strconv.Itoa(cmIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "cm-gb-" + strconv.Itoa(cmIndex),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(cmJson),
				},
			})
			cmIndex++
		} else if config.ConfigType == repository.CS_TYPE_CONFIG {
			secretDataMap := make(map[string][]byte)
			for key, value := range config.Data {
				secretDataMap[key] = []byte(value)
			}
			ownerDelete := true
			secretObject := v12.Secret{
				TypeMeta: v1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: config.Name,
					OwnerReferences: []v1.OwnerReference{{
						APIVersion:         "argoproj.io/v1alpha1",
						Kind:               "Workflow",
						Name:               "{{workflow.name}}",
						UID:                "{{workflow.uid}}",
						BlockOwnerDeletion: &ownerDelete,
					}},
				},
				Data: secretDataMap,
				Type: "Opaque",
			}
			secretJson, err := json.Marshal(secretObject)
			if err != nil {
				impl.logger.Errorw("error in building json", "err", err)
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						Secret: &v12.SecretVolumeSource{
							SecretName: config.Name,
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-sec-gb-" + strconv.Itoa(csIndex),
						Template: "sec-gb-" + strconv.Itoa(csIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "sec-gb-" + strconv.Itoa(csIndex),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(secretJson),
				},
			})
			csIndex++
		}
	}

	return nil
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
