/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

/*
@description: user crud
*/
package repository

import (
	"time"

	"github.com/go-pg/pg"
)

type UserAudit struct {
	TableName struct{}  `sql:"user_audit"`
	Id        int32     `sql:"id,pk"`
	UserId    int32     `sql:"user_id, notnull"`
	ClientIp  string    `sql:"client_ip"`
	CreatedOn time.Time `sql:"created_on,type:timestamptz"`
	UpdatedOn time.Time `sql:"updated_on,type:timestamptz"`
}

type UserAuditRepository interface {
	Save(userAudit *UserAudit) error
	GetLatestByUserId(userId int32) (*UserAudit, error)
	GetLatestUser() (*UserAudit, error)
	Update(userAudit *UserAudit) error
}

type UserAuditRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewUserAuditRepositoryImpl(dbConnection *pg.DB) *UserAuditRepositoryImpl {
	return &UserAuditRepositoryImpl{dbConnection: dbConnection}
}

func (impl UserAuditRepositoryImpl) Update(userAudit *UserAudit) error {
	userAuditPresentInDB, err := impl.GetLatestByUserId(userAudit.UserId)
	userAudit.UpdatedOn = time.Now()
	if err == nil {
		userAudit.Id = userAuditPresentInDB.Id
		userAudit.CreatedOn = userAuditPresentInDB.CreatedOn
		err = impl.dbConnection.Update(userAudit)
	} else if err == pg.ErrNoRows {
		userAudit.CreatedOn = userAudit.UpdatedOn
		err = impl.dbConnection.Insert(userAudit)
	}
	return err
}
func (impl UserAuditRepositoryImpl) Save(userAudit *UserAudit) error {
	userAudit.UpdatedOn = time.Now()
	return impl.dbConnection.Insert(userAudit)
}

func (impl UserAuditRepositoryImpl) GetLatestByUserId(userId int32) (*UserAudit, error) {
	userAudit := &UserAudit{}
	err := impl.dbConnection.Model(userAudit).
		Where("user_id = ?", userId).
		Order("id desc").
		Limit(1).
		Select()
	return userAudit, err
}

func (impl UserAuditRepositoryImpl) GetLatestUser() (*UserAudit, error) {
	userAudit := &UserAudit{}
	err := impl.dbConnection.Model(userAudit).
		Where("updated_on is not null").
		Order("updated_on desc").
		Limit(1).
		Select()
	return userAudit, err
}
