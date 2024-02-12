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
	"time"
)

type RoleGroupRepository interface {
	StartATransaction() (*pg.Tx, error)
	CommitATransaction(tx *pg.Tx) error
	CreateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error)
	UpdateRoleGroup(model *RoleGroup, tx *pg.Tx) (*RoleGroup, error)
	UpdateToInactiveByIds(ids []int32, tx *pg.Tx, loggedInUserId int32) error
	GetRoleGroupById(id int32) (*RoleGroup, error)
	GetRoleGroupByName(name string) (*RoleGroup, error)
	GetRoleGroupListByName(name string) ([]*RoleGroup, error)
	GetAllRoleGroup() ([]*RoleGroup, error)
	GetAllExecutingQuery(query string) ([]*RoleGroup, error)
	GetRoleGroupListByCasbinNames(name []string) ([]*RoleGroup, error)
	CheckRoleGroupExistByCasbinName(name string) (bool, error)
	CreateRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (*RoleGroupRoleMapping, error)
	GetRoleGroupRoleMapping(model int32) (*RoleGroupRoleMapping, error)
	GetRoleGroupRoleMappingByRoleGroupId(roleGroupId int32) ([]*RoleGroupRoleMapping, error)
	GetRoleGroupRoleMappingIdsByRoleGroupId(roleGroupId int32) ([]int, error)
	DeleteRoleGroupRoleMappingByRoleId(roleId int, tx *pg.Tx) error
	DeleteRoleGroupRoleMappingByRoleIds(roleId []int, tx *pg.Tx) error
	DeleteRoleGroupRoleMapping(model *RoleGroupRoleMapping, tx *pg.Tx) (bool, error)
	GetConnection() (dbConnection *pg.DB)
	GetRoleGroupListByNames(groupNames []string) ([]*RoleGroup, error)
	GetRolesByRoleGroupIds(roleGroupIds []int32) ([]*RoleModel, error)
	GetRolesByGroupCasbinName(groupName string) ([]*RoleModel, error)
	GetRolesByGroupNames(groupNames []string) ([]*RoleModel, error)
	GetRolesByGroupCasbinNames(groupCasbinNames []string) ([]*RoleModel, error)
	GetRolesByGroupNamesAndEntity(groupNames []string, entity string) ([]*RoleModel, error)
	UpdateRoleGroupIdForRoleGroupMappings(roleId int, newRoleId int) (*RoleGroupRoleMapping, error)
	GetCasbinNamesById(ids []int32) ([]string, error)
	GetRoleGroupRoleMappingIdsByGroupIds(groupIds []int32) ([]int, error)
	DeleteRoleGroupRoleMappingByIds(ids []int, tx *pg.Tx) error
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

func (impl RoleGroupRepositoryImpl) UpdateToInactiveByIds(ids []int32, tx *pg.Tx, loggedInUserId int32) error {
	var model []*RoleGroup
	_, err := tx.Model(&model).
		Set("active = ?", false).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", loggedInUserId).
		Where("id IN (?)", pg.In(ids)).Update()
	if err != nil {
		impl.Logger.Error("error in UpdateToInactiveByIds", "err", err, "roleGroupIds", ids)
		return err
	}
	return nil

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

func (impl RoleGroupRepositoryImpl) GetAllExecutingQuery(query string) ([]*RoleGroup, error) {
	var model []*RoleGroup
	_, err := impl.dbConnection.Query(&model, query)
	if err != nil {
		impl.Logger.Error("error in GetAllExecutingQuery", "err", err)
		return nil, err
	}
	return model, err
}

func (impl RoleGroupRepositoryImpl) GetRoleGroupListByCasbinNames(names []string) ([]*RoleGroup, error) {
	var model []*RoleGroup
	err := impl.dbConnection.Model(&model).Where("casbin_name in (?)", pg.In(names)).Where("active = ?", true).Select()
	return model, err
}

func (impl RoleGroupRepositoryImpl) CheckRoleGroupExistByCasbinName(name string) (bool, error) {
	var model RoleGroup
	return impl.dbConnection.Model(&model).Where("casbin_name = ?", name).Where("active = ?", true).Exists()

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

func (impl RoleGroupRepositoryImpl) GetRoleGroupRoleMappingIdsByRoleGroupId(roleGroupId int32) ([]int, error) {
	var Id []int
	err := impl.dbConnection.Model().
		Table("role_group_role_mapping").
		Column("role_group_role_mapping.id").
		Where("role_group_id = ?", roleGroupId).Select(&Id)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}
	return Id, nil
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

func (impl RoleGroupRepositoryImpl) DeleteRoleGroupRoleMappingByRoleIds(roleIds []int, tx *pg.Tx) error {
	var roleGroupRoleMapping *RoleGroupRoleMapping
	_, err := tx.Model(roleGroupRoleMapping).Where("role_id in (?)", pg.In(roleIds)).Delete()
	if err != nil {
		impl.Logger.Error("err in deleting roleGroupRoleMapping by role id", "err", err, "roleIds", roleIds)
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

func (impl RoleGroupRepositoryImpl) GetRolesByRoleGroupIds(roleGroupIds []int32) ([]*RoleModel, error) {
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

func (impl RoleGroupRepositoryImpl) GetRolesByGroupCasbinNames(groupCasbinNames []string) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" INNER JOIN role_group rg on rg.id=rgm.role_group_id" +
		" WHERE rg.casbin_name in (?);"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(groupCasbinNames))

	if err != nil {
		impl.Logger.Errorw("error in getting roles by group names", "err", err, "groupCasbinNames", groupCasbinNames)
		return roleModels, err
	}
	return roleModels, nil
}

func (impl RoleGroupRepositoryImpl) GetRolesByGroupNamesAndEntity(groupNames []string, entity string) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "SELECT r.* from roles r" +
		" INNER JOIN role_group_role_mapping rgm on rgm.role_id=r.id" +
		" INNER JOIN role_group rg on rg.id=rgm.role_group_id" +
		" WHERE rg.casbin_name in (?) and r.entity=?;"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(groupNames), entity)
	if err != nil {
		impl.Logger.Errorw("error in getting roles by group names", "err", err, "groupNames", groupNames)
		return roleModels, err
	}
	return roleModels, nil
}
func (impl RoleGroupRepositoryImpl) UpdateRoleGroupIdForRoleGroupMappings(roleId int, newRoleId int) (*RoleGroupRoleMapping, error) {
	var model RoleGroupRoleMapping
	_, err := impl.dbConnection.Model(&model).Set("role_id = ?", newRoleId).
		Where("role_id = ?", roleId).Update()

	return &model, err

}

func (impl RoleGroupRepositoryImpl) StartATransaction() (*pg.Tx, error) {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		impl.Logger.Errorw("error in beginning a transaction", "err", err)
		return nil, err
	}
	return tx, nil
}

func (impl RoleGroupRepositoryImpl) CommitATransaction(tx *pg.Tx) error {
	err := tx.Commit()
	if err != nil {
		impl.Logger.Errorw("error in commiting a transaction", "err", err)
		return err
	}
	return nil
}
func (impl RoleGroupRepositoryImpl) GetCasbinNamesById(ids []int32) ([]string, error) {
	type RoleGroup struct {
		TableName  struct{} `sql:"role_group"`
		CasbinName string   `json:"casbin_name"`
	}
	var models []RoleGroup
	err := impl.dbConnection.Model(&models).Where("id in (?)", pg.In(ids)).
		Where("active = ?", true).Select()
	if err != nil {
		impl.Logger.Errorw("error in GetCasbinNamesById", "ids", ids, "error", err)
		return nil, err
	}
	casbinNames := make([]string, 0, len(models))
	for _, mdl := range models {
		casbinNames = append(casbinNames, mdl.CasbinName)
	}
	return casbinNames, nil
}

func (impl RoleGroupRepositoryImpl) GetRoleGroupRoleMappingIdsByGroupIds(groupIds []int32) ([]int, error) {
	var Id []int
	err := impl.dbConnection.Model().
		Table("role_group_role_mapping").
		Column("role_group_role_mapping.id").
		Where("role_group_id in (?)", pg.In(groupIds)).Select(&Id)
	if err != nil {
		impl.Logger.Errorw("error in GetRoleGroupRoleMappingIdsByGroupIds", "groupIds", groupIds, "error", err)
		return nil, err
	}
	return Id, nil
}

func (impl RoleGroupRepositoryImpl) DeleteRoleGroupRoleMappingByIds(ids []int, tx *pg.Tx) error {
	var userRoleModel *RoleGroupRoleMapping
	_, err := tx.Model(userRoleModel).
		Where("id in (?)", pg.In(ids)).Delete()
	if err != nil {
		impl.Logger.Error("err encountered in DeleteRoleGroupRoleMappingByIds", "ids", ids, "err", err)
		return err
	}
	return nil
}
