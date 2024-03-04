package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type SMTPNotificationRepository interface {
	FindOne(id int) (*SMTPConfig, error)
	UpdateSMTPConfig(smtpConfig *SMTPConfig) (*SMTPConfig, error)
	SaveSMTPConfig(smtpConfig *SMTPConfig) (*SMTPConfig, error)
	FindAll() ([]*SMTPConfig, error)
	FindByIdsIn(ids []int) ([]*SMTPConfig, error)
	FindByTeamIdOrOwnerId(ownerId int32) ([]*SMTPConfig, error)
	UpdateSMTPConfigDefault() (bool, error)
	FindByIds(ids []*int) ([]*SMTPConfig, error)
	FindDefault() (*SMTPConfig, error)
	MarkSMTPConfigDeleted(smtpConfig *SMTPConfig) error
}

type SMTPNotificationRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewSMTPNotificationRepositoryImpl(dbConnection *pg.DB) *SMTPNotificationRepositoryImpl {
	return &SMTPNotificationRepositoryImpl{dbConnection: dbConnection}
}

type SMTPConfig struct {
	tableName    struct{} `sql:"smtp_config" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	Port         string   `sql:"port"`
	Host         string   `sql:"host"`
	AuthType     string   `sql:"auth_type"`
	AuthUser     string   `sql:"auth_user"`
	AuthPassword string   `sql:"auth_password"`
	FromEmail    string   `sql:"from_email"`
	ConfigName   string   `sql:"config_name"`
	Description  string   `sql:"description"`
	OwnerId      int32    `sql:"owner_id"`
	Default      bool     `sql:"default,notnull"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl *SMTPNotificationRepositoryImpl) FindByIdsIn(ids []int) ([]*SMTPConfig, error) {
	var configs []*SMTPConfig
	err := impl.dbConnection.Model(&configs).
		Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).
		Select()
	return configs, err
}

func (impl *SMTPNotificationRepositoryImpl) FindOne(id int) (*SMTPConfig, error) {
	details := &SMTPConfig{}
	err := impl.dbConnection.Model(details).Where("id = ?", id).
		Where("deleted = ?", false).Select()
	return details, err
}

func (impl *SMTPNotificationRepositoryImpl) FindAll() ([]*SMTPConfig, error) {
	var smtpConfigs []*SMTPConfig
	err := impl.dbConnection.Model(&smtpConfigs).
		Where("deleted = ?", false).Select()
	return smtpConfigs, err
}

func (impl *SMTPNotificationRepositoryImpl) FindByTeamIdOrOwnerId(ownerId int32) ([]*SMTPConfig, error) {
	var smtpConfigs []*SMTPConfig
	err := impl.dbConnection.Model(&smtpConfigs).Where(`owner_id = ?`, ownerId).
		Where("deleted = ?", false).Select()
	return smtpConfigs, err
}

func (impl *SMTPNotificationRepositoryImpl) UpdateSMTPConfig(smtpConfig *SMTPConfig) (*SMTPConfig, error) {
	return smtpConfig, impl.dbConnection.Update(smtpConfig)
}

func (impl *SMTPNotificationRepositoryImpl) SaveSMTPConfig(smtpConfig *SMTPConfig) (*SMTPConfig, error) {
	return smtpConfig, impl.dbConnection.Insert(smtpConfig)
}

func (impl *SMTPNotificationRepositoryImpl) UpdateSMTPConfigDefault() (bool, error) {
	SMTPConfigs, err := impl.FindAll()
	for _, SMTPConfig := range SMTPConfigs {
		SMTPConfig.Default = false
		err = impl.dbConnection.Update(SMTPConfig)
	}
	return true, err
}

func (impl *SMTPNotificationRepositoryImpl) FindByIds(ids []*int) ([]*SMTPConfig, error) {
	var objects []*SMTPConfig
	err := impl.dbConnection.Model(&objects).Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	return objects, err
}

func (impl *SMTPNotificationRepositoryImpl) FindDefault() (*SMTPConfig, error) {
	details := &SMTPConfig{}
	err := impl.dbConnection.Model(details).Where("smtp_config.default = ?", true).
		Where("deleted = ?", false).Select()
	return details, err
}
func (impl *SMTPNotificationRepositoryImpl) MarkSMTPConfigDeleted(smtpConfig *SMTPConfig) error {
	smtpConfig.Deleted = true
	return impl.dbConnection.Update(smtpConfig)
}
