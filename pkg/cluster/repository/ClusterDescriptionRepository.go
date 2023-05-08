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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ClusterDescription struct {
	ClusterId        int       `sql:"cluster_id"`
	ClusterName      string    `sql:"cluster_name"`
	ClusterCreatedOn time.Time `sql:"cluster_created_on"`
	ClusterCreatedBy int32     `sql:"cluster_created_by"`
	NoteId           int       `sql:"note_id,pk"`
	Description      string    `sql:"description"`
	sql.AuditLog
}

type ClusterDescriptionRepository interface {
	FindByClusterIdWithClusterDetails(id int) (*ClusterDescription, error)
}

func NewClusterDescriptionRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ClusterDescriptionRepositoryImpl {
	return &ClusterDescriptionRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type ClusterDescriptionRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl ClusterDescriptionRepositoryImpl) FindByClusterIdWithClusterDetails(id int) (*ClusterDescription, error) {
	clusterDescription := &ClusterDescription{}
	query := fmt.Sprintf("select cl.id as cluster_id, cl.cluster_name as cluster_name, cl.created_on as cluster_created_on, cl.created_by as cluster_created_by, cln.id as note_id, cln.description, cln.created_by, cln.created_on, cln.updated_by, cln.updated_on from cluster cl left join cluster_note cln on cl.id=cln.cluster_id where cl.id=%d and cl.active=true limit 1 offset 0;", id)
	_, err := impl.dbConnection.Query(clusterDescription, query)
	return clusterDescription, err
}
