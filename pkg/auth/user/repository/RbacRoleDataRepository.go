/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RbacRoleDataRepository interface {
	GetRoleDataForAllRoles() ([]*RbacRoleData, error)
	CreateNewRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error)
	UpdateRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error)
}

type RbacRoleDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacRoleDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *RbacRoleDataRepositoryImpl {
	return &RbacRoleDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacRoleData struct {
	TableName       struct{} `sql:"rbac_role_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id"`
	Entity          string   `sql:"entity"`
	AccessType      string   `sql:"access_type"`
	Role            string   `sql:"role"`
	RoleData        string   `sql:"role_data"`
	RoleDisplayName string   `sql:"role_display_name"`
	RoleDescription string   `sql:"role_description"`
	IsPresetRole    bool     `sql:"is_preset_role,notnull"`
	Deleted         bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacRoleDataRepositoryImpl) GetRoleDataForAllRoles() ([]*RbacRoleData, error) {
	var models []*RbacRoleData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting role data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacRoleDataRepositoryImpl) CreateNewRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error) {
	_, err := tx.Model(model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating role data for a role", "err", err)
		return nil, err
	}
	return model, nil
}

func (repo *RbacRoleDataRepositoryImpl) UpdateRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error) {
	_, err := tx.Model(model).UpdateNotNull()
	if err != nil {
		repo.logger.Errorw("error in updating role data for a role", "err", err)
		return nil, err
	}
	return model, nil
}
