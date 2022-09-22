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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RoleGroupRepository interface {
	CreateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error)
	UpdateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error)
	GetRoleGroupById(id int32) (*RoleGroup, error)
	GetRoleGroupByName(name string) (*RoleGroup, error)
	GetRoleGroupListByName(name string) ([]*RoleGroup, error)
	GetAllRoleGroup() ([]*RoleGroup, error)
	GetRoleGroupListByCasbinNames(name []string) ([]*RoleGroup, error)

	CreateRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (*RoleGroupRoleMapping, error)
	GetRoleGroupRoleMapping(model int32) (*RoleGroupRoleMapping, error)
	GetRoleGroupRoleMappingByRoleGroupId(roleGroupId int32) ([]*RoleGroupRoleMapping, error)
	DeleteRoleGroupRoleMappingByRoleId(roleId int, tx *pg.Tx) error
	DeleteRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (bool, error)
	GetConnection() (dbConnection *pg.DB)
	GetRoleGroupListByNames(groupNames []string) ([]*RoleGroup, error)
	GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds []int32) ([]*RoleModel, error)
	GetRolesByGroupCasbinName(groupName string) ([]*RoleModel, error)
	GetRolesByGroupNames(groupNames []string) ([]*RoleModel, error)
}

type RoleGroupRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewRoleGroupRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *RoleGroupRepositoryImpl {
	return &RoleGroupRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type RoleGroup struct {
	TableName   struct{} `sql:"role_group" pg:",discard_unknown_columns"`
	Id          int32    `sql:"id,pk"`
	Name        string   `sql:"name,notnull"`
	CasbinName  string   `sql:"casbin_name,notnull"`
	Description string   `sql:"description"`
	Active      bool     `sql:"active,notnull"`
	sql.AuditLog
}

type RoleGroupRoleMapping struct {
	TableName   struct{} `sql:"role_group_role_mapping"  pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	RoleGroupId int32    `sql:"role_group_id,notnull"`
	RoleId      int      `sql:"role_id,notnull"`
	sql.AuditLog
}

func (impl *RoleGroupRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl RoleGroupRepositoryImpl) CreateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	//TODO - Create Entry In UserRole With Default Role for User
	return model, nil
}
func (impl RoleGroupRepositoryImpl) UpdateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}

	//TODO - Create Entry In UserRole With Default Role for User

	return model, nil
}
func (impl RoleGroupRepositoryImpl) GetRoleGroupById(id int32) (*RoleGroup, error) {
	var model RoleGroup
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Where("active = ?", true).Select()
	return &model, err
}
func (impl RoleGroupRepositoryImpl) GetRoleGroupByName(name string) (*RoleGroup, error) {
	var model RoleGroup
	err := impl.dbConnection.Model(&model).Where("name = ?", name).Where("active = ?", true).Order("updated_on desc").Select()
	return &model, err
}
func (impl RoleGroupRepositoryImpl) GetRoleGroupListByName(name string) ([]*RoleGroup, error) {
	var model []*RoleGroup
	err := impl.dbConnection.Model(&model).Where("name ILIKE ?", "%"+name+"%").Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}
func (impl RoleGroupRepositoryImpl) GetAllRoleGroup() ([]*RoleGroup, error) {
	var model []*RoleGroup
	err := impl.dbConnection.Model(&model).Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}

func (impl RoleGroupRepositoryImpl) GetRoleGroupListByCasbinNames(names []string) ([]*RoleGroup, error) {
	var model []*RoleGroup
	err := impl.dbConnection.Model(&model).Where("casbin_name in (?)", pg.In(names)).Where("active = ?", true).Select()
	return model, err
}

func (impl RoleGroupRepositoryImpl) CreateRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (*RoleGroupRoleMapping, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}

	return model, nil
}
func (impl RoleGroupRepositoryImpl) GetRoleGroupRoleMapping(userRoleModelId int32) (*RoleGroupRoleMapping, error) {
	var model RoleGroupRoleMapping
	err := impl.dbConnection.Model(&model).Where("id = ?", userRoleModelId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}

	return &model, nil
}
func (impl RoleGroupRepositoryImpl) GetRoleGroupRoleMappingByRoleGroupId(roleGroupId int32) ([]*RoleGroupRoleMapping, error) {
	var userRoleModels []*RoleGroupRoleMapping
	err := impl.dbConnection.Model(&userRoleModels).Where("role_group_id = ?", roleGroupId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModels, err
	}
	return userRoleModels, nil
}

func (impl RoleGroupRepositoryImpl) DeleteRoleGroupRoleMappingByRoleId(roleId int, tx *pg.Tx) error {
	var roleGroupRoleMapping *RoleGroupRoleMapping
	_, err := tx.Model(roleGroupRoleMapping).Where("role_id = ?", roleId).Delete()
	if err != nil {
		impl.Logger.Error("err in deleting roleGroupRoleMapping by role id", "err", err, "roleId", roleId)
		return err
	}
	return nil
}

func (impl RoleGroupRepositoryImpl) DeleteRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (bool, error) {
	err := tx.Delete(model)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl RoleGroupRepositoryImpl) GetRoleGroupListByNames(groupNames []string) ([]*RoleGroup, error) {
	var model []*RoleGroup
	err := impl.dbConnection.Model(&model).Where("name in (?)", pg.In(groupNames)).Where("active = ?", true).Order("updated_on desc").Select()
	return model, err
}

func (impl RoleGroupRepositoryImpl) GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds []int32) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" WHERE rgm.role_group_id in (?);"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(roleGroupIds))
	if err != nil {
		return roleModels, err
	}
	return roleModels, nil
}

func (impl RoleGroupRepositoryImpl) GetRolesByGroupCasbinName(groupName string) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" INNER JOIN role_group rg on rg.id=rgm.role_group_id" +
		" WHERE rg.casbin_name = ?;"
	_, err := impl.dbConnection.Query(&roleModels, query, groupName)

	if err != nil {
		return roleModels, err
	}
	return roleModels, nil
}

func (impl RoleGroupRepositoryImpl) GetRolesByGroupNames(groupNames []string) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" INNER JOIN role_group rg on rg.id=rgm.role_group_id" +
		" WHERE rg.name in (?);"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(groupNames))

	if err != nil {
		impl.Logger.Errorw("error in getting roles by group names", "err", err, "groupNames", groupNames)
		return roleModels, err
	}
	return roleModels, nil
}
