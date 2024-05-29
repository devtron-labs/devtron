/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package moduleRepo

import (
	"github.com/go-pg/pg"
	"time"
)

type ModuleResourceStatus struct {
	tableName     struct{}  `sql:"module_resource_status"`
	Id            int       `sql:"id,pk"`
	ModuleId      int       `sql:"module_id,notnull"`
	Group         string    `sql:"group, notnull"`
	Version       string    `sql:"version, notnull"`
	Kind          string    `sql:"kind, notnull"`
	Name          string    `sql:"name, notnull"`
	HealthStatus  string    `sql:"health_status"`
	HealthMessage string    `sql:"health_message"`
	Active        bool      `sql:"active"`
	CreatedOn     time.Time `sql:"created_on, notnull"`
	UpdatedOn     time.Time `sql:"updated_on"`
}

type ModuleResourceStatusRepository interface {
	GetConnection() *pg.DB
	FindAllActiveByModuleId(moduleId int) ([]*ModuleResourceStatus, error)
	Update(status *ModuleResourceStatus, tx *pg.Tx) error
	Save(statuses []*ModuleResourceStatus, tx *pg.Tx) error
}

type ModuleResourceStatusRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewModuleResourceStatusRepositoryImpl(dbConnection *pg.DB) *ModuleResourceStatusRepositoryImpl {
	return &ModuleResourceStatusRepositoryImpl{dbConnection: dbConnection}
}

func (impl ModuleResourceStatusRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl ModuleResourceStatusRepositoryImpl) FindAllActiveByModuleId(moduleId int) ([]*ModuleResourceStatus, error) {
	var moduleResourcesStatus []*ModuleResourceStatus
	err := impl.dbConnection.Model(&moduleResourcesStatus).
		Where("module_id = ?", moduleId).
		Where("active = ?", true).
		Select()
	return moduleResourcesStatus, err
}

func (impl ModuleResourceStatusRepositoryImpl) Update(status *ModuleResourceStatus, tx *pg.Tx) error {
	return tx.Update(status)
}

func (impl ModuleResourceStatusRepositoryImpl) Save(statuses []*ModuleResourceStatus, tx *pg.Tx) error {
	return tx.Insert(&statuses)
}
