package appStoreRepository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type InstalledAppVersionHistoryRepository interface {
	CreateInstalledAppVersionHistory(model *InstalledAppVersionHistory, tx *pg.Tx) (*InstalledAppVersionHistory, error)
	UpdateInstalledAppVersionHistory(model *InstalledAppVersionHistory, tx *pg.Tx) (*InstalledAppVersionHistory, error)
	GetInstalledAppVersionHistory(id int) (*InstalledAppVersionHistory, error)
	GetInstalledAppVersionHistoryByVersionId(installAppVersionId int) ([]*InstalledAppVersionHistory, error)
	GetLatestInstalledAppVersionHistory(installAppVersionId int) (*InstalledAppVersionHistory, error)
	GetLatestInstalledAppVersionHistoryByGitHash(gitHash string) (*InstalledAppVersionHistory, error)
}

type InstalledAppVersionHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewInstalledAppVersionHistoryRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *InstalledAppVersionHistoryRepositoryImpl {
	return &InstalledAppVersionHistoryRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type InstalledAppVersionHistory struct {
	TableName             struct{} `sql:"installed_app_version_history" pg:",discard_unknown_columns"`
	Id                    int      `sql:"id,pk"`
	InstalledAppVersionId int      `sql:"installed_app_version_id,notnull"`
	ValuesYamlRaw         string   `sql:"values_yaml_raw"`
	Status                string   `sql:"status"`
	GitHash               string   `sql:"git_hash"`
	sql.AuditLog
}

func (impl InstalledAppVersionHistoryRepositoryImpl) CreateInstalledAppVersionHistory(model *InstalledAppVersionHistory, tx *pg.Tx) (*InstalledAppVersionHistory, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl InstalledAppVersionHistoryRepositoryImpl) UpdateInstalledAppVersionHistory(model *InstalledAppVersionHistory, tx *pg.Tx) (*InstalledAppVersionHistory, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl InstalledAppVersionHistoryRepositoryImpl) GetInstalledAppVersionHistory(id int) (*InstalledAppVersionHistory, error) {
	model := &InstalledAppVersionHistory{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_version_history.*").
		Where("installed_app_version_history.id = ?", id).Select()
	return model, err
}
func (impl InstalledAppVersionHistoryRepositoryImpl) GetInstalledAppVersionHistoryByVersionId(installAppVersionId int) ([]*InstalledAppVersionHistory, error) {
	var model []*InstalledAppVersionHistory
	err := impl.dbConnection.Model(&model).
		Column("installed_app_version_history.*").
		Where("installed_app_version_history.installed_app_version_id = ?", installAppVersionId).
		Order("installed_app_version_history.id desc").
		Select()
	return model, err
}

func (impl InstalledAppVersionHistoryRepositoryImpl) GetLatestInstalledAppVersionHistory(installAppVersionId int) (*InstalledAppVersionHistory, error) {
	model := &InstalledAppVersionHistory{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_version_history.*").
		Where("installed_app_version_history.installed_app_version_id = ?", installAppVersionId).
		Order("installed_app_version_history.id desc").Limit(1).
		Select()
	return model, err
}

func (impl InstalledAppVersionHistoryRepositoryImpl) GetLatestInstalledAppVersionHistoryByGitHash(gitHash string) (*InstalledAppVersionHistory, error) {
	model := &InstalledAppVersionHistory{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_version_history.*").
		Where("installed_app_version_history.git_hash = ?", gitHash).
		Order("installed_app_version_history.id desc").Limit(1).
		Select()
	return model, err
}
