package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type UserGroupMapRepository interface {
	GetConnection() *pg.DB
	GetByUserId(userId int32) ([]*UserGroup, error)
	Save(models []*UserGroup, tx *pg.Tx) error
	Update(models []*UserGroup, tx *pg.Tx) error
	GetActiveByUserId(userId int32) ([]*UserGroup, error)
}

type UserGroupMapRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewUserGroupMapRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *UserGroupMapRepositoryImpl {
	return &UserGroupMapRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type UserGroup struct {
	TableName         struct{} `sql:"user_groups"`
	Id                int      `sql:"id,pk"`
	UserId            int32    `sql:"user_id"`
	GroupName         string   `sql:"group_name,notnull"`
	IsGroupClaimsData bool     `sql:"is_group_claims_data,notnull"`
	Active            bool     `sql:"active,notnull"`
	sql.AuditLog
}

func (repo *UserGroupMapRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *UserGroupMapRepositoryImpl) GetByUserId(userId int32) ([]*UserGroup, error) {
	var models []*UserGroup
	err := repo.dbConnection.Model(&models).Where("user_id = ?", userId).Select()
	if err != nil {
		repo.logger.Errorw("error, GetByUserId", "err", err, "userId", userId)
		return nil, err
	}
	return models, nil
}

func (repo *UserGroupMapRepositoryImpl) Save(models []*UserGroup, tx *pg.Tx) error {
	err := tx.Insert(&models)
	if err != nil {
		repo.logger.Errorw("error, Save", "err", err, "models", models)
		return err
	}
	return nil
}

func (repo *UserGroupMapRepositoryImpl) Update(models []*UserGroup, tx *pg.Tx) error {
	_, err := tx.Model(&models).Update()
	if err != nil {
		repo.logger.Errorw("error, UpdateInBatch", "err", err, "models", models)
		return err
	}
	return nil
}

func (repo *UserGroupMapRepositoryImpl) GetActiveByUserId(userId int32) ([]*UserGroup, error) {
	var models []*UserGroup
	err := repo.dbConnection.Model(&models).Where("user_id = ?", userId).
		Where("active = ?", true).Select()
	if err != nil {
		repo.logger.Errorw("error, GetByUserId", "err", err, "userId", userId)
		return nil, err
	}
	return models, nil

}
