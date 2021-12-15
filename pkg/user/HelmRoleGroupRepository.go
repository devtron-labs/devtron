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

package user

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type HelmRoleGroupRepository interface {
	CreateRoleGroup(model *HelmRoleGroup, tx *pg.Tx) (*HelmRoleGroup, error)
	UpdateRoleGroup(model *HelmRoleGroup, tx *pg.Tx) (*HelmRoleGroup, error)
	GetRoleGroupById(id int32) (*HelmRoleGroup, error)
	GetRoleGroupByName(name string) (*HelmRoleGroup, error)
	GetRoleGroupListByName(name string) ([]*HelmRoleGroup, error)
	GetAllRoleGroup() ([]*HelmRoleGroup, error)
	GetRoleGroupListByCasbinNames(name []string) ([]*HelmRoleGroup, error)

	CreateRoleGroupRoleMapping(model *HelmRoleGroupRoleMapping, tx *pg.Tx) (*HelmRoleGroupRoleMapping, error)
	GetRoleGroupRoleMapping(model int32) (*HelmRoleGroupRoleMapping, error)
	GetRoleGroupRoleMappingByRoleGroupId(roleGroupId int32) ([]*HelmRoleGroupRoleMapping, error)
	DeleteRoleGroupRoleMapping(model *HelmRoleGroupRoleMapping, tx *pg.Tx) (bool, error)
	GetConnection() (dbConnection *pg.DB)
	GetRoleGroupListByNames(groupNames []string) ([]*HelmRoleGroup, error)
	GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds []int32) ([]*HelmRoleModel, error)
}

type HelmRoleGroupRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewHelmRoleGroupRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *HelmRoleGroupRepositoryImpl {
	return &HelmRoleGroupRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type HelmRoleGroup struct {
	TableName   struct{} `sql:"role_group" pg:",discard_unknown_columns"`
	Id          int32    `sql:"id,pk"`
	Name        string   `sql:"name,notnull"`
	CasbinName  string   `sql:"casbin_name,notnull"`
	Description string   `sql:"description"`
	Active      bool     `sql:"active,notnull"`
	sql.AuditLog
}

type HelmRoleGroupRoleMapping struct {
	TableName   struct{} `sql:"role_group_role_mapping"  pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	RoleGroupId int32    `sql:"role_group_id,notnull"`
	RoleId      int      `sql:"role_id,notnull"`
	sql.AuditLog
}

func (impl *HelmRoleGroupRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl HelmRoleGroupRepositoryImpl) CreateRoleGroup(model *HelmRoleGroup, tx *pg.Tx) (*HelmRoleGroup, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	//TODO - Create Entry In UserRole With Default Role for User
	return model, nil
}
func (impl HelmRoleGroupRepositoryImpl) UpdateRoleGroup(model *HelmRoleGroup, tx *pg.Tx) (*HelmRoleGroup, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}

	//TODO - Create Entry In UserRole With Default Role for User

	return model, nil
}
func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupById(id int32) (*HelmRoleGroup, error) {
	var model HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Where("active = ?", true).Select()
	return &model, err
}
func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupByName(name string) (*HelmRoleGroup, error) {
	var model HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("name = ?", name).Where("active = ?", true).Order("updated_on desc").Select()
	return &model, err
}
func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupListByName(name string) ([]*HelmRoleGroup, error) {
	var model []*HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("name ILIKE ?", "%"+name+"%").Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}
func (impl HelmRoleGroupRepositoryImpl) GetAllRoleGroup() ([]*HelmRoleGroup, error) {
	var model []*HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}

func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupListByCasbinNames(names []string) ([]*HelmRoleGroup, error) {
	var model []*HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("casbin_name in (?)", pg.In(names)).Where("active = ?", true).Select()
	return model, err
}

func (impl HelmRoleGroupRepositoryImpl) CreateRoleGroupRoleMapping(model *HelmRoleGroupRoleMapping, tx *pg.Tx) (*HelmRoleGroupRoleMapping, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}

	return model, nil
}
func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupRoleMapping(userRoleModelId int32) (*HelmRoleGroupRoleMapping, error) {
	var model HelmRoleGroupRoleMapping
	err := impl.dbConnection.Model(&model).Where("id = ?", userRoleModelId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}

	return &model, nil
}
func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupRoleMappingByRoleGroupId(roleGroupId int32) ([]*HelmRoleGroupRoleMapping, error) {
	var userRoleModels []*HelmRoleGroupRoleMapping
	err := impl.dbConnection.Model(&userRoleModels).Where("role_group_id = ?", roleGroupId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModels, err
	}
	return userRoleModels, nil
}
func (impl HelmRoleGroupRepositoryImpl) DeleteRoleGroupRoleMapping(model *HelmRoleGroupRoleMapping, tx *pg.Tx) (bool, error) {
	err := tx.Delete(model)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupListByNames(groupNames []string) ([]*HelmRoleGroup, error) {
	var model []*HelmRoleGroup
	err := impl.dbConnection.Model(&model).Where("name in (?)", pg.In(groupNames)).Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}

func (impl HelmRoleGroupRepositoryImpl) GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds []int32) ([]*HelmRoleModel, error) {
	var roleModels []*HelmRoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" WHERE rgm.role_group_id in (?);"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(roleGroupIds))
	if err != nil {
		return roleModels, err
	}
	return roleModels, nil
}
