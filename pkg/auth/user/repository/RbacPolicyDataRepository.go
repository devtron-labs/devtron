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

type RbacPolicyDataRepository interface {
	GetPolicyDataForAllRoles() ([]*RbacPolicyData, error)
	CreateNewPolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error)
	UpdatePolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error)
}

type RbacPolicyDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacPolicyDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *RbacPolicyDataRepositoryImpl {
	return &RbacPolicyDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacPolicyData struct {
	TableName    struct{} `sql:"rbac_policy_data" pg:",discard_unknown_columns"`
	Id           int      `sql:"id"`
	Entity       string   `sql:"entity"`
	AccessType   string   `sql:"access_type"`
	Role         string   `sql:"role"`
	PolicyData   string   `sql:"policy_data"`
	IsPresetRole bool     `sql:"is_preset_role,notnull"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacPolicyDataRepositoryImpl) GetPolicyDataForAllRoles() ([]*RbacPolicyData, error) {
	var models []*RbacPolicyData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacPolicyDataRepositoryImpl) CreateNewPolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error) {
	_, err := tx.Model(model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}

func (repo *RbacPolicyDataRepositoryImpl) UpdatePolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error) {
	_, err := tx.Model(model).UpdateNotNull()
	if err != nil {
		repo.logger.Errorw("error in updating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}
