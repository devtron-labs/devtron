/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package resourceGroup

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ResourceGroup struct {
	tableName   struct{} `sql:"resource_group" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name,notnull"`
	Description string   `sql:"description,notnull"`
	Active      bool     `sql:"active,notnull"`
	ResourceId  int      `sql:"resource_id,notnull"`
	ResourceKey int      `sql:"resource_key,notnull"`
	sql.AuditLog
}

type ResourceGroupRepository interface {
	Save(model *ResourceGroup, tx *pg.Tx) (*ResourceGroup, error)
	Update(model *ResourceGroup, tx *pg.Tx) error
	FindById(id int) (*ResourceGroup, error)
	FindByNameAndParentResource(name string, resourceId int, resourceKey int) (*ResourceGroup, error)
	FindActiveListByParentResource(resourceId int, resourceKey int) ([]*ResourceGroup, error)
	GetConnection() (dbConnection *pg.DB)
}

type ResourceGroupRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewResourceGroupRepositoryImpl(dbConnection *pg.DB) *ResourceGroupRepositoryImpl {
	return &ResourceGroupRepositoryImpl{dbConnection: dbConnection}
}

func (repo *ResourceGroupRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}

func (repo ResourceGroupRepositoryImpl) Save(model *ResourceGroup, tx *pg.Tx) (*ResourceGroup, error) {
	err := tx.Insert(model)
	return model, err
}

func (repo ResourceGroupRepositoryImpl) Update(model *ResourceGroup, tx *pg.Tx) error {
	err := tx.Update(model)
	return err
}

func (repo ResourceGroupRepositoryImpl) FindById(id int) (*ResourceGroup, error) {
	model := &ResourceGroup{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).Where("active = ?", true).
		Select()
	return model, err
}

func (repo ResourceGroupRepositoryImpl) FindByNameAndParentResource(name string, resourceId int, resourceKey int) (*ResourceGroup, error) {
	model := &ResourceGroup{}
	err := repo.dbConnection.Model(model).
		Where("name = ?", name).
		Where("resource_id=?", resourceId).
		Where("resource_key=?", resourceKey).
		Where("active = ?", true).
		Select()
	return model, err
}

func (repo ResourceGroupRepositoryImpl) FindActiveListByParentResource(resourceId int, resourceKey int) ([]*ResourceGroup, error) {
	var models []*ResourceGroup
	err := repo.dbConnection.Model(&models).Where("active=?", true).
		Where("resource_id=?", resourceId).
		Where("resource_key=?", resourceKey).
		Select()
	return models, err
}
