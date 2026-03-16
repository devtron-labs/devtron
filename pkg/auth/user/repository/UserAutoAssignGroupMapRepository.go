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
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type UserAutoAssignGroupMapRepository interface {
	GetConnection() *pg.DB
	GetByUserId(userId int32) ([]*UserAutoAssignedGroup, error)
	Save(models []*UserAutoAssignedGroup, tx *pg.Tx) error
	Update(models []*UserAutoAssignedGroup, tx *pg.Tx) error
	GetActiveByUserId(userId int32) ([]*UserAutoAssignedGroup, error)
	GetAllActiveWithUser() ([]*UserAutoAssignedGroup, error)
}

type UserAutoAssignGroupMapRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewUserAutoAssignGroupMapRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *UserAutoAssignGroupMapRepositoryImpl {
	return &UserAutoAssignGroupMapRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type UserAutoAssignedGroup struct {
	TableName         struct{}   `sql:"user_auto_assigned_groups"`
	Id                int        `sql:"id,pk"`
	UserId            int32      `sql:"user_id"`
	GroupName         string     `sql:"group_name,notnull"`
	IsGroupClaimsData bool       `sql:"is_group_claims_data,notnull"`
	Active            bool       `sql:"active,notnull"`
	User              *UserModel `pg:"fk:user_id"`
	sql.AuditLog
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) GetByUserId(userId int32) ([]*UserAutoAssignedGroup, error) {
	var models []*UserAutoAssignedGroup
	err := repo.dbConnection.Model(&models).Where("user_id = ?", userId).Select()
	if err != nil {
		repo.logger.Errorw("error, GetByUserId", "err", err, "userId", userId)
		return nil, err
	}
	return models, nil
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) Save(models []*UserAutoAssignedGroup, tx *pg.Tx) error {
	err := tx.Insert(&models)
	if err != nil {
		repo.logger.Errorw("error, Save", "err", err, "models", models)
		return err
	}
	return nil
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) Update(models []*UserAutoAssignedGroup, tx *pg.Tx) error {
	_, err := tx.Model(&models).Update()
	if err != nil {
		repo.logger.Errorw("error, UpdateInBatch", "err", err, "models", models)
		return err
	}
	return nil
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) GetActiveByUserId(userId int32) ([]*UserAutoAssignedGroup, error) {
	var models []*UserAutoAssignedGroup
	err := repo.dbConnection.Model(&models).Where("user_id = ?", userId).
		Where("active = ?", true).Select()
	if err != nil {
		repo.logger.Errorw("error, GetActiveByUserId", "err", err, "userId", userId)
		return nil, err
	}
	return models, nil
}

func (repo *UserAutoAssignGroupMapRepositoryImpl) GetAllActiveWithUser() ([]*UserAutoAssignedGroup, error) {
	var models []*UserAutoAssignedGroup
	err := repo.dbConnection.Model(&models).
		Column("user_auto_assigned_group.*").
		Relation("User", func(q *orm.Query) (*orm.Query, error) {
			return q.Where("\"user\".active = ?", true), nil
		}).
		Where("user_auto_assigned_group.active = ?", true).
		Select()
	if err != nil {
		repo.logger.Errorw("error, GetAllActiveWithUser", "err", err)
		return nil, err
	}
	return models, nil
}
