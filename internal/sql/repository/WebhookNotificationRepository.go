package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type WebhookNotificationRepository interface {
	FindOne(id int) (*WebhookConfig, error)
	UpdateWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error)
	SaveWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error)
	FindAll() ([]WebhookConfig, error)
	FindByIdsIn(ids []int) ([]*WebhookConfig, error)
	FindByTeamIdOrOwnerId(ownerId int32, teamIds []int) ([]WebhookConfig, error)
	FindByName(value string) ([]WebhookConfig, error)
	FindByIds(ids []*int) ([]*WebhookConfig, error)
	MarkSlackConfigDeleted(webhookConfig *WebhookConfig) error
}

type WebhookNotificationRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewWebhookNotificationRepositoryImpl(dbConnection *pg.DB) *WebhookNotificationRepositoryImpl {
	return &WebhookNotificationRepositoryImpl{dbConnection: dbConnection}
}

type WebhookConfig struct {
	tableName   struct{}          `sql:"webhook_config" pg:",discard_unknown_columns"`
	Id          int               `sql:"id,pk"`
	WebHookUrl  string            `sql:"web_hook_url"`
	ConfigName  string            `sql:"config_name"`
	Header      map[string]string `sql:"header"`
	Payload     map[string]string `sql:"payload"`
	Description string            `sql:"description"`
	OwnerId     int32             `sql:"owner_id"`
	Active      bool              `sql:"active"`
	Deleted     bool              `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl *WebhookNotificationRepositoryImpl) FindByIdsIn(ids []int) ([]*WebhookConfig, error) {
	var configs []*WebhookConfig
	err := impl.dbConnection.Model(&configs).
		Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).
		Select()
	return configs, err
}

func (impl *WebhookNotificationRepositoryImpl) FindOne(id int) (*WebhookConfig, error) {
	details := &WebhookConfig{}
	err := impl.dbConnection.Model(details).Where("id = ?", id).
		Where("deleted = ?", false).Select()

	return details, err
}

func (impl *WebhookNotificationRepositoryImpl) FindAll() ([]WebhookConfig, error) {
	var webhookConfigs []WebhookConfig
	err := impl.dbConnection.Model(&webhookConfigs).
		Where("deleted = ?", false).Select()
	return webhookConfigs, err
}

func (impl *WebhookNotificationRepositoryImpl) FindByTeamIdOrOwnerId(ownerId int32, teamIds []int) ([]WebhookConfig, error) {
	var webhookConfigs []WebhookConfig
	if len(teamIds) == 0 {

		err := impl.dbConnection.Model(&webhookConfigs).Where(`owner_id = ?`, ownerId).
			Where("deleted = ?", false).Select()
		return webhookConfigs, err
	} else {
		err := impl.dbConnection.Model(&webhookConfigs).
			Where(`team_id in (?)`, pg.In(teamIds)).
			Where("deleted = ?", false).Select()
		return webhookConfigs, err
	}
}

func (impl *WebhookNotificationRepositoryImpl) UpdateWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error) {
	return webhookConfig, impl.dbConnection.Update(webhookConfig)
}

func (impl *WebhookNotificationRepositoryImpl) SaveWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error) {
	return webhookConfig, impl.dbConnection.Insert(webhookConfig)
}

func (impl *WebhookNotificationRepositoryImpl) FindByName(value string) ([]WebhookConfig, error) {
	var webhookConfigs []WebhookConfig
	err := impl.dbConnection.Model(&webhookConfigs).Where(`config_name like ?`, "%"+value+"%").
		Where("deleted = ?", false).Select()
	return webhookConfigs, err

}

func (repo *WebhookNotificationRepositoryImpl) FindByIds(ids []*int) ([]*WebhookConfig, error) {
	var objects []*WebhookConfig
	err := repo.dbConnection.Model(&objects).Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	return objects, err
}

func (impl *WebhookNotificationRepositoryImpl) MarkSlackConfigDeleted(webhookConfig *WebhookConfig) error {
	webhookConfig.Deleted = true
	return impl.dbConnection.Update(webhookConfig)
}
