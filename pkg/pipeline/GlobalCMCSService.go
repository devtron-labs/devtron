package pipeline

import (
	"encoding/json"
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
	ConfigType string `sql:"config_type" validate:"oneof=CONFIGMAP SECRET"`
	Name       string `sql:"name"  validate:"required"`
	//map of key:value, example: '{ "a" : "b", "c" : "d"}'
	Data                     map[string]string `json:"data"  validate:"required"`
	MountPath                string            `json:"mountPath"`
	UseByDefaultInCiPipeline bool              `json:"useByDefaultInCiPipeline"`
	Deleted                  bool              `json:"deleted"`
	UserId                   int32             `json:"-"`
}

func (impl *GlobalCMCSServiceImpl) Create(config *GlobalCMCSDto) (*GlobalCMCSDto, error) {
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
