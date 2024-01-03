/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
@description: user crud
*/
package repository

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/go-pg/pg"
	"time"
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
	GetAllNonApiTokenUsersData() ([]*UserAudit, error)
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

func (impl UserAuditRepositoryImpl) GetAllNonApiTokenUsersData() ([]*UserAudit, error) {
	var userAudits []*UserAudit
	err := impl.dbConnection.Model(&userAudits).
		Join("inner join users on users.id=user_audit.user_id").
		Where("users.user_type is NULL or users.user_type != ?", bean.USER_TYPE_API_TOKEN).
		Where("users.active = ?", true).
		Select()
	return userAudits, err
}
