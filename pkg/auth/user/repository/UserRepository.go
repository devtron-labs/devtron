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
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error)
	UpdateUser(userModel *UserModel, tx *pg.Tx) (*UserModel, error)
	GetById(id int32) (*UserModel, error)
	GetByIdIncludeDeleted(id int32) (*UserModel, error)
	GetAllExcludingApiTokenUser() ([]UserModel, error)
	GetAllExecutingQuery(query string) ([]UserModel, error)
	GetAllActiveUsers() ([]UserModel, error)
	//GetAllUserRoleMappingsForRoleId(roleId int) ([]UserRoleModel, error)
	FetchActiveUserByEmail(email string) (bean.UserInfo, error)
	FetchUserDetailByEmail(email string) (bean.UserInfo, error)
	GetByIds(ids []int32) ([]UserModel, error)
	GetConnection() (dbConnection *pg.DB)
	FetchUserMatchesByEmailIdExcludingApiTokenUser(email string) ([]UserModel, error)
	FetchActiveOrDeletedUserByEmail(email string) (*UserModel, error)
	UpdateRoleIdForUserRolesMappings(roleId int, newRoleId int) (*UserRoleModel, error)
	GetCountExecutingQuery(query string) (int, error)
	UpdateWindowIdtoNull(userIds []int32) error
	UpdateTimeWindowId(tx *pg.Tx, userid int32, windowId int) error
	StartATransaction() (*pg.Tx, error)
	CommitATransaction(tx *pg.Tx) error
	GetUserWithTimeoutWindowConfiguration(emailId string) (*UserModel, error)
}

type UserRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewUserRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *UserRepositoryImpl {
	return &UserRepositoryImpl{dbConnection: dbConnection, Logger: logger}
}

type UserModel struct {
	TableName                    struct{} `sql:"users" pg:",discard_unknown_columns"`
	Id                           int32    `sql:"id,pk"`
	EmailId                      string   `sql:"email_id,notnull"`
	AccessToken                  string   `sql:"access_token"`
	Active                       bool     `sql:"active,notnull"`
	UserType                     string   `sql:"user_type"`
	TimeoutWindowConfigurationId int      `sql:"timeout_window_configuration_id"`
	TimeoutWindowConfiguration   *repository.TimeoutWindowConfiguration
	UserAudit                    *UserAudit `sql:"-"`
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

func (impl UserRepositoryImpl) GetAllExecutingQuery(query string) ([]UserModel, error) {
	var userModel []UserModel
	_, err := impl.dbConnection.Query(&userModel, query)
	if err != nil {
		impl.Logger.Error("error in GetAllExecutingQuery", "err", err, "query", query)
		return nil, err
	}
	return userModel, err

}

func (impl UserRepositoryImpl) GetAllActiveUsers() ([]UserModel, error) {
	var userModel []UserModel
	err := impl.dbConnection.Model(&userModel).
		Where("active = ?", true).
		Order("email_id").Select()
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

func (impl UserRepositoryImpl) UpdateRoleIdForUserRolesMappings(roleId int, newRoleId int) (*UserRoleModel, error) {
	var model UserRoleModel
	_, err := impl.dbConnection.Model(&model).Set("role_id = ? ", newRoleId).Where("role_id = ? ", roleId).Update()
	return &model, err

}

func (impl UserRepositoryImpl) GetCountExecutingQuery(query string) (int, error) {
	var totalCount int
	_, err := impl.dbConnection.Query(&totalCount, query)
	if err != nil {
		impl.Logger.Error("Exception caught: GetCountExecutingQuery", err)
		return totalCount, err
	}
	return totalCount, err
}

func (impl UserRepositoryImpl) UpdateWindowIdtoNull(userIds []int32) error {
	var model []UserModel
	_, err := impl.dbConnection.Model(&model).Set("timeout_window_configuration_id = null").
		Where("id in (?)", pg.In(userIds)).Update()
	if err != nil {
		impl.Logger.Error("error in UpdateFKtoNull", "err", err, "userIds", userIds)
		return err
	}
	return nil
}

func (impl UserRepositoryImpl) UpdateTimeWindowId(tx *pg.Tx, userid int32, windowId int) error {
	var model []UserModel
	_, err := tx.Model(&model).Set("timeout_window_configuration_id = ? ", windowId).
		Where("id = ? ", userid).Update()
	if err != nil {
		impl.Logger.Error("error in UpdateTimeWindowId", "err", err, "userid", userid, "windowId", windowId)
		return err
	}
	return nil
}

func (impl UserRepositoryImpl) StartATransaction() (*pg.Tx, error) {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		impl.Logger.Errorw("error in beginning a transaction", "err", err)
		return nil, err
	}
	return tx, nil
}

func (impl UserRepositoryImpl) CommitATransaction(tx *pg.Tx) error {
	err := tx.Commit()
	if err != nil {
		impl.Logger.Errorw("error in commiting a transaction", "err", err)
		return err
	}
	return nil
}

func (impl UserRepositoryImpl) GetUserWithTimeoutWindowConfiguration(emailId string) (*UserModel, error) {
	var model UserModel
	err := impl.dbConnection.Model(&model).
		Column("user_model.*", "TimeoutWindowConfiguration").
		Join("left join timeout_window_configuration twc on twc.id = user_model.timeout_window_configuration_id").
		Where("user_model.email_id like (?) ", emailId).
		Where("user_model.active = ? ", true).
		Select()
	if err != nil {
		impl.Logger.Errorw("error in GetUserWithTimeoutWindowConfiguration", "err", err, "emailId", emailId)
		return &model, err
	}
	return &model, nil
}

/*
func (impl UserRepositoryImpl) GetUserWithTimeoutWindowConfiguration_V2(emailId string) (*UserModel, error) {
	var model UserModel
	//formattedTimeForQuery := time.Now().Format(helper.QueryTimeFormat)
	err := impl.dbConnection.Model(&model).
		Column("user_model.*", "TimeoutWindowConfiguration").
		Join("left join timeout_window_configuration twc on twc.id = user_model.timeout_window_configuration_id").
		Where("user_model.email_id like (?) ", emailId).
		Where("user_model.active = ? ", true).
		Where("twc.timeout_window_expression_format = ?", 1).
		Where("twc.timeout_window_expression > ?", time.Now()).
		//Where("TO_TIMESTAMP(twc.timeout_window_expression,'YYYY-MM-DD HH24:MI:SS') < '?' )", formattedTimeForQuery).
		Select()

	if err != nil {
		impl.Logger.Errorw("error in GetUserWithTimeoutWindow", "err", err, "emailId", emailId)
		return &model, err
	}
	return &model, nil
}
*/
