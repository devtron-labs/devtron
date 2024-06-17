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
	"time"
)

type ContainerAction int

const ActionCreate ContainerAction = 0
const ActionAccessed ContainerAction = 1
const ActionTerminate ContainerAction = 2

type EphemeralContainerBean struct {
	tableName           struct{} `sql:"ephemeral_container" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	Name                string   `sql:"name"`
	ClusterId           int      `sql:"cluster_id"`
	Namespace           string   `sql:"namespace"`
	PodName             string   `sql:"pod_name"`
	TargetContainer     string   `sql:"target_container"`
	Config              string   `sql:"config"`
	IsExternallyCreated bool     `sql:"is_externally_created"`
}

type EphemeralContainerAction struct {
	tableName            struct{}        `sql:"ephemeral_container_actions" pg:",discard_unknown_columns"`
	Id                   int             `sql:"id,pk"`
	EphemeralContainerId int             `sql:"ephemeral_container_id"`
	ActionType           ContainerAction `sql:"action_type"`
	PerformedBy          int32           `sql:"performed_by"`
	PerformedAt          time.Time       `sql:"performed_at"`
}

type EphemeralContainersRepository interface {
	sql.TransactionWrapper
	SaveEphemeralContainerData(tx *pg.Tx, model *EphemeralContainerBean) error
	SaveEphemeralContainerActionAudit(tx *pg.Tx, model *EphemeralContainerAction) error
	FindContainerByName(clusterID int, namespace, podName, name string) (*EphemeralContainerBean, error)
}

func NewEphemeralContainersRepositoryImpl(db *pg.DB, transactionUtilImpl *sql.TransactionUtilImpl) *EphemeralContainersRepositoryImpl {
	return &EphemeralContainersRepositoryImpl{
		dbConnection:        db,
		TransactionUtilImpl: transactionUtilImpl,
	}
}

type EphemeralContainersRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func (impl EphemeralContainersRepositoryImpl) SaveEphemeralContainerData(tx *pg.Tx, model *EphemeralContainerBean) error {
	return tx.Insert(model)
}

func (impl EphemeralContainersRepositoryImpl) SaveEphemeralContainerActionAudit(tx *pg.Tx, model *EphemeralContainerAction) error {
	return tx.Insert(model)
}

func (impl EphemeralContainersRepositoryImpl) FindContainerByName(clusterID int, namespace, podName, name string) (*EphemeralContainerBean, error) {
	container := &EphemeralContainerBean{}
	err := impl.dbConnection.Model(container).
		Where("cluster_id = ?", clusterID).
		Where("namespace = ?", namespace).
		Where("pod_name = ?", podName).
		Where("name = ?", name).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if err == pg.ErrNoRows {
		container = nil
	}
	return container, nil
}
