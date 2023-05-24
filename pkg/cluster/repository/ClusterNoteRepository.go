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

type ClusterNote struct {
	tableName   struct{} `sql:"cluster_note" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	ClusterId   int      `sql:"cluster_id"`
	Description string   `sql:"description"`
	sql.AuditLog
}

type ClusterNoteRepository interface {
	Save(model *ClusterNote) error
	FindByClusterId(id int) (*ClusterNote, error)
	Update(model *ClusterNote) error
}

func NewClusterNoteRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ClusterNoteRepositoryImpl {
	return &ClusterNoteRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type ClusterNoteRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl ClusterNoteRepositoryImpl) Save(model *ClusterNote) error {
	return impl.dbConnection.Insert(model)
}

func (impl ClusterNoteRepositoryImpl) FindByClusterId(id int) (*ClusterNote, error) {
	clusterNote := &ClusterNote{}
	err := impl.dbConnection.
		Model(clusterNote).
		Where("cluster_id =?", id).
		Limit(1).
		Select()
	return clusterNote, err
}

func (impl ClusterNoteRepositoryImpl) Update(model *ClusterNote) error {
	return impl.dbConnection.Update(model)
}
