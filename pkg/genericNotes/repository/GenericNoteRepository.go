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

type NoteType int

const ClusterType NoteType = 0
const AppType NoteType = 1

type ClusterNote struct {
	tableName      struct{} `sql:"cluster_note" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk"`
	Identifer      int      `sql:"identifier"`
	IdentifierType NoteType `sql:"identifier_type"`
	Description    string   `sql:"description"`
	sql.AuditLog
}

type GenericNoteRepository interface {
	Save(model *ClusterNote) error
	FindByClusterId(id int) (*ClusterNote, error)
	FindByAppId(id int) (*ClusterNote, error)
	Update(model *ClusterNote) error
}

func NewGenericNoteRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *GenericNoteRepositoryImpl {
	return &GenericNoteRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type GenericNoteRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl GenericNoteRepositoryImpl) Save(model *ClusterNote) error {
	return impl.dbConnection.Insert(model)
}

func (impl GenericNoteRepositoryImpl) FindByClusterId(id int) (*ClusterNote, error) {
	clusterNote := &ClusterNote{}
	err := impl.dbConnection.
		Model(clusterNote).
		Where("identifier =?", id).
		Where("identifier_type =?", ClusterType).
		Limit(1).
		Select()
	return clusterNote, err
}

func (impl GenericNoteRepositoryImpl) Update(model *ClusterNote) error {
	return impl.dbConnection.Update(model)
}

func (impl GenericNoteRepositoryImpl) FindByAppId(id int) (*ClusterNote, error) {
	clusterNote := &ClusterNote{}
	err := impl.dbConnection.
		Model(clusterNote).
		Where("identifier =?", id).
		Where("identifier_type =?", AppType).
		Limit(1).
		Select()
	return clusterNote, err
}
