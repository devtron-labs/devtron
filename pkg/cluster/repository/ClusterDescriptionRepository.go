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
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository"
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
	FindByClusterIdWithClusterDetails(clusterId int) (*ClusterDescription, error)
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

func (impl ClusterDescriptionRepositoryImpl) FindByClusterIdWithClusterDetails(clusterId int) (*ClusterDescription, error) {
	clusterDescription := &ClusterDescription{}
	query := "SELECT cl.id AS cluster_id, cl.cluster_name AS cluster_name, cl.created_on AS cluster_created_on, cl.created_by AS cluster_created_by, gn.id AS note_id, gn.description, gn.created_by, gn.created_on, gn.updated_by, gn.updated_on FROM" +
		" cluster cl LEFT JOIN" +
		" generic_note gn " +
		" ON cl.id=gn.identifier AND (gn.identifier_type = %d OR gn.identifier_type IS NULL)" +
		" WHERE cl.id=%d AND cl.active=true " +
		" LIMIT 1 OFFSET 0;"
	query = fmt.Sprintf(query, clusterId, repository.ClusterType)
	_, err := impl.dbConnection.Query(clusterDescription, query)
	return clusterDescription, err
}
