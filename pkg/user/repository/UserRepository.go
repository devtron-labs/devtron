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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error)
	UpdateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error)
	GetById(id int32) (*UserModel, error)
	GetByIdIncludeDeleted(id int32) (*UserModel, error)
	GetAllExcludingApiTokenUser() ([]UserModel, error)
	FetchActiveUserByEmail(email string) (bean.UserInfo, error)
	FetchUserDetailByEmail(email string) (bean.UserInfo, error)
	GetByIds(ids []int32) ([]UserModel, error)
	GetConnection() (dbConnection *pg.DB)
	FetchUserMatchesByEmailIdExcludingApiTokenUser(email string) ([]UserModel, error)
	FetchActiveOrDeletedUserByEmail(email string) (*UserModel, error)
}

type UserRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewUserRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *UserRepositoryImpl {
	return &UserRepositoryImpl{dbConnection: dbConnection, Logger: logger}
}

type UserModel struct {
	TableName   struct{} `sql:"users"`
	Id          int32    `sql:"id,pk"`
	EmailId     string   `sql:"email_id,notnull"`
	AccessToken string   `sql:"access_token"`
	Active      bool     `sql:"active,notnull"`
	UserType    string   `sql:"user_type"`
	sql.AuditLog
}

type UserRoleModel struct {
	TableName struct{} `sql:"user_roles"`
	Id        int      `sql:"id,pk"`
	UserId    int32    `sql:"user_id,notnull"`
	RoleId    int      `sql:"role_id,notnull"`
	User      UserModel
	sql.AuditLog
}

func (impl UserRepositoryImpl) CreateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	//TODO - Create Entry In UserRole With Default Role for User
	return userModel, nil
}
func (impl UserRepositoryImpl) UpdateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error) {
	err := tx.Update(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}

	//TODO - Create Entry In UserRole With Default Role for User

	return userModel, nil
}
func (impl UserRepositoryImpl) GetById(id int32) (*UserModel, error) {
	var model UserModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Where("active = ?", true).Select()
	return &model, err
}

func (impl UserRepositoryImpl) GetByIdIncludeDeleted(id int32) (*UserModel, error) {
	var model UserModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}

func (impl UserRepositoryImpl) GetAllExcludingApiTokenUser() ([]UserModel, error) {
	var userModel []UserModel
	err := impl.dbConnection.Model(&userModel).
		Where("active = ?", true).
		Where("user_type is NULL or user_type != ?", bean.USER_TYPE_API_TOKEN).
		Order("updated_on desc").Select()
	return userModel, err
}

func (impl UserRepositoryImpl) FetchActiveUserByEmail(email string) (bean.UserInfo, error) {
	var users bean.UserInfo

	query := "SELECT u.id, u.email_id, u.access_token, u.user_type FROM users u " +
		"WHERE u.active = true and u.email_id ILIKE ? order by u.updated_on desc"
	_, err := impl.dbConnection.Query(&users, query, email)
	if err != nil {
		impl.Logger.Error("Exception caught:", err)
		return users, err
	}

	return users, nil
}

func (impl UserRepositoryImpl) FetchUserDetailByEmail(email string) (bean.UserInfo, error) {
	//impl.Logger.Info("reached at FetchUserDetailByEmail:")
	var users []bean.UserRole
	var userFinal bean.UserInfo

	query := "SELECT u.id, u.email_id, u.user_type, r.role FROM users u" +
		" INNER JOIN user_roles ur ON ur.user_id=u.id" +
		" INNER JOIN roles r ON r.id=ur.role_id" +
		" WHERE u.email_id= ? and u.active = true" +
		" ORDER BY u.updated_on desc;"
	_, err := impl.dbConnection.Query(&users, query, email)
	if err != nil {
		return userFinal, err
	}

	var role []string
	for _, item := range users {
		userFinal.Exist = true
		userFinal.Id = item.Id
		userFinal.EmailId = item.EmailId
		role = append(role, item.Role)
	}
	userFinal.Roles = role
	return userFinal, nil
}
func (impl UserRepositoryImpl) GetByIds(ids []int32) ([]UserModel, error) {
	var model []UserModel
	err := impl.dbConnection.Model(&model).Where("id in (?)", pg.In(ids)).Where("active = ?", true).Select()
	return model, err
}

func (impl *UserRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl UserRepositoryImpl) FetchUserMatchesByEmailIdExcludingApiTokenUser(email string) ([]UserModel, error) {
	var model []UserModel
	err := impl.dbConnection.Model(&model).
		Where("email_id like (?)", "%"+email+"%").
		Where("user_type is NULL or user_type != ?", bean.USER_TYPE_API_TOKEN).
		Where("active = ?", true).Select()
	return model, err
}

func (impl UserRepositoryImpl) FetchActiveOrDeletedUserByEmail(email string) (*UserModel, error) {
	var model UserModel
	err := impl.dbConnection.Model(&model).Where("email_id ILIKE (?)", email).Limit(1).Select()
	return &model, err
}
