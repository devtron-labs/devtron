package notifier

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

const WEBHOOK_CONFIG_TYPE = "webhook"

type WebhookVariable string

const (
	// these fields will be configurable in future
	DevtronContainerImageTag WebhookVariable = "{{devtronContainerImageTag}}"
	DevtronAppName           WebhookVariable = "{{devtronAppName}}"
	DevtronAppId             WebhookVariable = "{{devtronAppId}}"
	DevtronEnvName           WebhookVariable = "{{devtronEnvName}}"
	DevtronEnvId             WebhookVariable = "{{devtronEnvId}}"
	DevtronCiPipelineId      WebhookVariable = "{{devtronCiPipelineId}}"
	DevtronCdPipelineId      WebhookVariable = "{{devtronCdPipelineId}}"
	DevtronTriggeredByEmail  WebhookVariable = "{{devtronTriggeredByEmail}}"
	EventType                WebhookVariable = "{{eventType}}"
)

type WebhookNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []WebhookConfigDto, userId int32) ([]int, error)
	FetchWebhookNotificationConfigById(id int) (*WebhookConfigDto, error)
	GetWebhookVariables() (map[string]WebhookVariable, error)
	FetchAllWebhookNotificationConfig() ([]*WebhookConfigDto, error)
	FetchAllWebhookNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error)
	DeleteNotificationConfig(deleteReq *WebhookConfigDto, userId int32) error
}

type WebhookNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	webhookRepository              repository.WebhookNotificationRepository
	teamService                    team.TeamService
	userRepository                 repository2.UserRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}
type WebhookChannelConfig struct {
	Channel           util2.Channel       `json:"channel" validate:"required"`
	WebhookConfigDtos *[]WebhookConfigDto `json:"configs"`
}

type WebhookConfigDto struct {
	OwnerId     int32                  `json:"userId" validate:"number"`
	WebhookUrl  string                 `json:"webhookUrl" validate:"required"`
	ConfigName  string                 `json:"configName" validate:"required"`
	Header      map[string]interface{} `json:"header"`
	Payload     map[string]interface{} `json:"payload"`
	Description string                 `json:"description"`
	Id          int                    `json:"id" validate:"number"`
}

func NewWebhookNotificationServiceImpl(logger *zap.SugaredLogger, webhookRepository repository.WebhookNotificationRepository, teamService team.TeamService,
	userRepository repository2.UserRepository, notificationSettingsRepository repository.NotificationSettingsRepository) *WebhookNotificationServiceImpl {
	return &WebhookNotificationServiceImpl{
		logger:                         logger,
		webhookRepository:              webhookRepository,
		teamService:                    teamService,
		userRepository:                 userRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

func (impl *WebhookNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []WebhookConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	webhookConfigs := buildWebhookNewConfigs(channelReq, userId)
	for _, config := range webhookConfigs {
		if config.Id != 0 {
			model, err := impl.webhookRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching webhook config", "err", err)
				return []int{}, err
			}
			impl.buildConfigUpdateModel(config, model, userId)
			model, uErr := impl.webhookRepository.UpdateWebhookConfig(model)
			if uErr != nil {
				impl.logger.Errorw("err while updating webhook config", "err", err)
				return []int{}, uErr
			}
		} else {
			_, iErr := impl.webhookRepository.SaveWebhookConfig(config)
			if iErr != nil {
				impl.logger.Errorw("err while inserting webhook config", "err", iErr)
				return []int{}, iErr
			}

		}
		responseIds = append(responseIds, config.Id)
	}
	return responseIds, nil
}

func (impl *WebhookNotificationServiceImpl) FetchWebhookNotificationConfigById(id int) (*WebhookConfigDto, error) {
	webhookConfig, err := impl.webhookRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all webhoook config", "err", err)
		return nil, err
	}
	webhoookConfigDto := impl.adaptWebhookConfig(*webhookConfig)
	return &webhoookConfigDto, nil
}

func (impl *WebhookNotificationServiceImpl) GetWebhookVariables() (map[string]WebhookVariable, error) {
	variables := map[string]WebhookVariable{
		"devtronContainerImageTag": DevtronContainerImageTag,
		"devtronAppName":           DevtronAppName,
		"devtronAppId":             DevtronAppId,
		"devtronEnvName":           DevtronEnvName,
		"devtronEnvId":             DevtronEnvId,
		"devtronCiPipelineId":      DevtronCiPipelineId,
		"devtronCdPipelineId":      DevtronCdPipelineId,
		"devtronTriggeredByEmail":  DevtronTriggeredByEmail,
		"eventType":                EventType,
	}

	return variables, nil
}

func (impl *WebhookNotificationServiceImpl) FetchAllWebhookNotificationConfig() ([]*WebhookConfigDto, error) {
	var responseDto []*WebhookConfigDto
	webhookConfigs, err := impl.webhookRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all webhoook config", "err", err)
		return []*WebhookConfigDto{}, err
	}
	for _, webhookConfig := range webhookConfigs {
		webhookConfigDto := impl.adaptWebhookConfig(webhookConfig)
		responseDto = append(responseDto, &webhookConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*WebhookConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *WebhookNotificationServiceImpl) FetchAllWebhookNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error) {
	var responseDto []*NotificationChannelAutoResponse
	webhookConfigs, err := impl.webhookRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all webhoook config", "err", err)
		return []*NotificationChannelAutoResponse{}, err
	}
	for _, webhookConfig := range webhookConfigs {
		webhookConfigDto := &NotificationChannelAutoResponse{
			Id:         webhookConfig.Id,
			ConfigName: webhookConfig.ConfigName,
		}
		responseDto = append(responseDto, webhookConfigDto)
	}
	return responseDto, nil
}

func (impl *WebhookNotificationServiceImpl) adaptWebhookConfig(webhookConfig repository.WebhookConfig) WebhookConfigDto {
	webhookConfigDto := WebhookConfigDto{
		OwnerId:     webhookConfig.OwnerId,
		WebhookUrl:  webhookConfig.WebHookUrl,
		ConfigName:  webhookConfig.ConfigName,
		Header:      webhookConfig.Header,
		Payload:     webhookConfig.Payload,
		Description: webhookConfig.Description,
		Id:          webhookConfig.Id,
	}
	return webhookConfigDto
}

func buildWebhookNewConfigs(webhookReq []WebhookConfigDto, userId int32) []*repository.WebhookConfig {
	var webhookConfigs []*repository.WebhookConfig
	for _, c := range webhookReq {
		webhookConfig := &repository.WebhookConfig{
			Id:          c.Id,
			ConfigName:  c.ConfigName,
			WebHookUrl:  c.WebhookUrl,
			Header:      c.Header,
			Payload:     c.Payload,
			Description: c.Description,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		webhookConfig.OwnerId = userId
		webhookConfigs = append(webhookConfigs, webhookConfig)
	}
	return webhookConfigs
}

func (impl *WebhookNotificationServiceImpl) buildConfigUpdateModel(webhookConfig *repository.WebhookConfig, model *repository.WebhookConfig, userId int32) {
	model.WebHookUrl = webhookConfig.WebHookUrl
	model.ConfigName = webhookConfig.ConfigName
	model.Description = webhookConfig.Description
	model.Payload = webhookConfig.Payload
	model.Header = webhookConfig.Header
	model.OwnerId = webhookConfig.OwnerId
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}

func (impl *WebhookNotificationServiceImpl) DeleteNotificationConfig(deleteReq *WebhookConfigDto, userId int32) error {
	existingConfig, err := impl.webhookRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, WEBHOOK_CONFIG_TYPE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting webhook config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
	}

	existingConfig.UpdatedOn = time.Now()
	existingConfig.UpdatedBy = userId
	//deleting webhook config
	err = impl.webhookRepository.MarkWebhookConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting webhook config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
