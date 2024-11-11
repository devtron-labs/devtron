package adapter

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func BuildConfigUpdateModelForWebhook(webhookConfig *repository.WebhookConfig, model *repository.WebhookConfig, userId int32) {
	model.WebHookUrl = webhookConfig.WebHookUrl
	model.ConfigName = webhookConfig.ConfigName
	model.Description = webhookConfig.Description
	model.Payload = webhookConfig.Payload
	model.Header = webhookConfig.Header
	model.OwnerId = webhookConfig.OwnerId
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}

func AdaptWebhookConfig(webhookConfig repository.WebhookConfig) beans.WebhookConfigDto {
	webhookConfigDto := beans.WebhookConfigDto{
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
func AdaptSMTPConfig(smtpConfig *repository.SMTPConfig) *beans.SMTPConfigDto {
	smtpConfigDto := &beans.SMTPConfigDto{
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
func BuildWebhookNewConfigs(webhookReq []beans.WebhookConfigDto, userId int32) []*repository.WebhookConfig {
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
func BuildSMTPNewConfigs(smtpReq []*beans.SMTPConfigDto, userId int32) []*repository.SMTPConfig {
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
func BuildConfigUpdateModelSMTP(smtpConfig *repository.SMTPConfig, model *repository.SMTPConfig, userId int32) {
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
func BuildSlackNewConfigs(slackReq []beans.SlackConfigDto, userId int32) []*repository.SlackConfig {
	var slackConfigs []*repository.SlackConfig
	for _, c := range slackReq {
		slackConfig := &repository.SlackConfig{
			Id:          c.Id,
			ConfigName:  c.ConfigName,
			WebHookUrl:  c.WebhookUrl,
			Description: c.Description,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		if c.TeamId != 0 {
			slackConfig.TeamId = c.TeamId
		} else {
			slackConfig.OwnerId = userId
		}
		slackConfigs = append(slackConfigs, slackConfig)
	}
	return slackConfigs
}
func AdaptSlackConfig(slackConfig repository.SlackConfig) beans.SlackConfigDto {
	slackConfigDto := beans.SlackConfigDto{
		OwnerId:     slackConfig.OwnerId,
		TeamId:      slackConfig.TeamId,
		WebhookUrl:  slackConfig.WebHookUrl,
		ConfigName:  slackConfig.ConfigName,
		Description: slackConfig.Description,
		Id:          slackConfig.Id,
	}
	return slackConfigDto
}
func BuildConfigUpdateModelForSlack(slackConfig *repository.SlackConfig, model *repository.SlackConfig, userId int32) {
	model.WebHookUrl = slackConfig.WebHookUrl
	model.ConfigName = slackConfig.ConfigName
	model.Description = slackConfig.Description
	if slackConfig.TeamId != 0 {
		model.TeamId = slackConfig.TeamId
	} else {
		model.OwnerId = slackConfig.OwnerId
	}
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}
func BuildConfigUpdateModelForSES(sesConfig *repository.SESConfig, model *repository.SESConfig, userId int32) {
	model.Id = sesConfig.Id
	model.OwnerId = sesConfig.OwnerId
	model.Region = sesConfig.Region
	model.AccessKey = sesConfig.AccessKey
	model.SecretKey = sesConfig.SecretKey
	model.FromEmail = sesConfig.FromEmail
	model.SessionToken = sesConfig.SessionToken
	model.ConfigName = sesConfig.ConfigName
	model.Description = sesConfig.Description
	model.Default = sesConfig.Default
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}
func BuildSESNewConfigs(sesReq []*beans.SESConfigDto, userId int32) []*repository.SESConfig {
	var sesConfigs []*repository.SESConfig
	for _, c := range sesReq {
		sesConfig := &repository.SESConfig{
			Id:           c.Id,
			Region:       c.Region,
			AccessKey:    c.AccessKey,
			SecretKey:    c.SecretKey,
			ConfigName:   c.ConfigName,
			FromEmail:    c.FromEmail,
			SessionToken: c.SessionToken,
			Description:  c.Description,
			Default:      c.Default,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}

		sesConfig.OwnerId = userId
		sesConfigs = append(sesConfigs, sesConfig)
	}
	return sesConfigs
}
func AdaptSESConfig(sesConfig *repository.SESConfig) *beans.SESConfigDto {
	sesConfigDto := &beans.SESConfigDto{
		OwnerId:      sesConfig.OwnerId,
		Region:       sesConfig.Region,
		AccessKey:    sesConfig.AccessKey,
		SecretKey:    sesConfig.SecretKey,
		FromEmail:    sesConfig.FromEmail,
		SessionToken: sesConfig.SessionToken,
		ConfigName:   sesConfig.ConfigName,
		Description:  sesConfig.Description,
		Id:           sesConfig.Id,
		Default:      sesConfig.Default,
	}
	return sesConfigDto
}
