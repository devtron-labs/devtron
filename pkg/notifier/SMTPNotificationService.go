package notifier

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

const SMTP_CONFIG_TYPE = "smtp"

type SMTPNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []*SMTPConfigDto, userId int32) ([]int, error)
	FetchSMTPNotificationConfigById(id int) (*SMTPConfigDto, error)
	FetchAllSMTPNotificationConfig() ([]*SMTPConfigDto, error)
	FetchAllSMTPNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error)
	DeleteNotificationConfig(channelReq *SMTPConfigDto, userId int32) error
}

type SMTPNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	smtpRepository                 repository.SMTPNotificationRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}

func NewSMTPNotificationServiceImpl(logger *zap.SugaredLogger, smtpRepository repository.SMTPNotificationRepository,
	teamService team.TeamService, notificationSettingsRepository repository.NotificationSettingsRepository) *SMTPNotificationServiceImpl {
	return &SMTPNotificationServiceImpl{
		logger:                         logger,
		teamService:                    teamService,
		smtpRepository:                 smtpRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

type SMTPChannelConfig struct {
	Channel        util2.Channel    `json:"channel" validate:"required"`
	SMTPConfigDtos []*SMTPConfigDto `json:"configs"`
}

type SMTPConfigDto struct {
	Id           int    `json:"id"`
	Port         string `json:"port"`
	Host         string `json:"host"`
	AuthType     string `json:"authType"`
	AuthUser     string `json:"authUser"`
	AuthPassword string `json:"authPassword"`
	FromEmail    string `json:"fromEmail"`
	ConfigName   string `json:"configName"`
	Description  string `json:"description"`
	OwnerId      int32  `json:"ownerId"`
	Default      bool   `json:"default"`
	Deleted      bool   `json:"deleted"`
}

func (impl *SMTPNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []*SMTPConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	smtpConfigs := buildSMTPNewConfigs(channelReq, userId)
	for _, config := range smtpConfigs {
		if config.Id != 0 {

			if config.Default {
				_, err := impl.smtpRepository.UpdateSMTPConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				}
			} else {
				_, err := impl.smtpRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			model, err := impl.smtpRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching smtp config", "err", err)
				return []int{}, err
			}
			impl.buildConfigUpdateModel(config, model, userId)
			model, uErr := impl.smtpRepository.UpdateSMTPConfig(model)
			if uErr != nil {
				impl.logger.Errorw("err while updating smtp config", "err", err)
				return []int{}, uErr
			}
		} else {

			if config.Default {
				_, err := impl.smtpRepository.UpdateSMTPConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				}
			} else {
				_, err := impl.smtpRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			_, iErr := impl.smtpRepository.SaveSMTPConfig(config)
			if iErr != nil {
				impl.logger.Errorw("err while inserting smtp config", "err", iErr)
				return []int{}, iErr
			}
		}
		responseIds = append(responseIds, config.Id)
	}
	return responseIds, nil
}

func (impl *SMTPNotificationServiceImpl) FetchSMTPNotificationConfigById(id int) (*SMTPConfigDto, error) {
	smtpConfig, err := impl.smtpRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return nil, err
	}
	smtpConfigDto := impl.adaptSMTPConfig(smtpConfig)
	return smtpConfigDto, nil
}

func (impl *SMTPNotificationServiceImpl) FetchAllSMTPNotificationConfig() ([]*SMTPConfigDto, error) {
	var responseDto []*SMTPConfigDto
	smtpConfigs, err := impl.smtpRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return []*SMTPConfigDto{}, err
	}
	for _, smtpConfig := range smtpConfigs {
		smtpConfigDto := impl.adaptSMTPConfig(smtpConfig)
		smtpConfigDto.AuthPassword = "**********"
		responseDto = append(responseDto, smtpConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*SMTPConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SMTPNotificationServiceImpl) FetchAllSMTPNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error) {
	var responseDto []*NotificationChannelAutoResponse
	smtpConfigs, err := impl.smtpRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return []*NotificationChannelAutoResponse{}, err
	}
	for _, smtpConfig := range smtpConfigs {
		smtpConfigDto := &NotificationChannelAutoResponse{
			Id:         smtpConfig.Id,
			ConfigName: smtpConfig.ConfigName}
		responseDto = append(responseDto, smtpConfigDto)
	}
	return responseDto, nil
}

func (impl *SMTPNotificationServiceImpl) adaptSMTPConfig(smtpConfig *repository.SMTPConfig) *SMTPConfigDto {
	smtpConfigDto := &SMTPConfigDto{
		OwnerId:      smtpConfig.OwnerId,
		Port:         smtpConfig.Port,
		Host:         smtpConfig.Host,
		AuthType:     smtpConfig.AuthType,
		AuthUser:     smtpConfig.AuthUser,
		AuthPassword: smtpConfig.AuthPassword,
		FromEmail:    smtpConfig.FromEmail,
		ConfigName:   smtpConfig.ConfigName,
		Description:  smtpConfig.Description,
		Id:           smtpConfig.Id,
		Default:      smtpConfig.Default,
		Deleted:      false,
	}
	return smtpConfigDto
}

func buildSMTPNewConfigs(smtpReq []*SMTPConfigDto, userId int32) []*repository.SMTPConfig {
	var smtpConfigs []*repository.SMTPConfig
	for _, c := range smtpReq {
		smtpConfig := &repository.SMTPConfig{
			Id:           c.Id,
			Port:         c.Port,
			Host:         c.Host,
			AuthType:     c.AuthType,
			AuthUser:     c.AuthUser,
			AuthPassword: c.AuthPassword,
			ConfigName:   c.ConfigName,
			FromEmail:    c.FromEmail,
			Deleted:      false,
			Description:  c.Description,
			Default:      c.Default,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		smtpConfig.OwnerId = userId
		smtpConfigs = append(smtpConfigs, smtpConfig)
	}
	return smtpConfigs
}

func (impl *SMTPNotificationServiceImpl) buildConfigUpdateModel(smtpConfig *repository.SMTPConfig, model *repository.SMTPConfig, userId int32) {
	model.Id = smtpConfig.Id
	model.OwnerId = smtpConfig.OwnerId
	model.Port = smtpConfig.Port
	model.Host = smtpConfig.Host
	model.AuthUser = smtpConfig.AuthUser
	model.AuthType = smtpConfig.AuthType
	model.AuthPassword = smtpConfig.AuthPassword
	model.FromEmail = smtpConfig.FromEmail
	model.ConfigName = smtpConfig.ConfigName
	model.Description = smtpConfig.Description
	model.Default = smtpConfig.Default
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
	model.Deleted = false
}

func (impl *SMTPNotificationServiceImpl) DeleteNotificationConfig(deleteReq *SMTPConfigDto, userId int32) error {
	existingConfig, err := impl.smtpRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, SMTP_CONFIG_TYPE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting smtp config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
	}
	existingConfig.UpdatedOn = time.Now()
	existingConfig.UpdatedBy = userId
	//deleting smtp config
	err = impl.smtpRepository.MarkSMTPConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting smtp config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
