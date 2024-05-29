/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package resourceGroup

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ResourceGroupMapping struct {
	tableName       struct{} `sql:"resource_group_mapping" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	ResourceGroupId int      `sql:"resource_group_id,notnull"`
	ResourceId      int      `sql:"resource_id,notnull"`
	ResourceKey     int      `sql:"resource_key,notnull"`
	ResourceGroup   *ResourceGroup
	sql.AuditLog
}

type ResourceGroupMappingRepository interface {
	Save(model *ResourceGroupMapping, tx *pg.Tx) (*ResourceGroupMapping, error)
	//Update(model *ResourceGroupMapping, tx *pg.Tx) error
	Delete(model *ResourceGroupMapping, tx *pg.Tx) error
	//FindById(id int) (*ResourceGroupMapping, error)
	FindByResourceGroupId(appGroupId int) ([]*ResourceGroupMapping, error)
	//FindAll() ([]*ResourceGroupMapping, error)
	FindByResourceGroupIds(appGroupIds []int) ([]*ResourceGroupMapping, error)
	//GetConnection() (dbConnection *pg.DB)
}

type ResourceGroupMappingRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewResourceGroupMappingRepositoryImpl(dbConnection *pg.DB) *ResourceGroupMappingRepositoryImpl {
	return &ResourceGroupMappingRepositoryImpl{dbConnection: dbConnection}
}

func (repo ResourceGroupMappingRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return repo.dbConnection
}

func (repo ResourceGroupMappingRepositoryImpl) Save(model *ResourceGroupMapping, tx *pg.Tx) (*ResourceGroupMapping, error) {
	err := tx.Insert(model)
	return model, err
}

//func (repo ResourceGroupMappingRepositoryImpl) Update(model *ResourceGroupMapping, tx *pg.Tx) error {
//	err := tx.Update(model)
//	return err
//}

func (repo ResourceGroupMappingRepositoryImpl) Delete(model *ResourceGroupMapping, tx *pg.Tx) error {
	err := tx.Delete(model)
	return err
}

//func (repo ResourceGroupMappingRepositoryImpl) FindById(id int) (*ResourceGroupMapping, error) {
//	model := &ResourceGroupMapping{}
//	err := repo.dbConnection.Model(model).Where("id = ?", id).
//		Select()
//	return model, err
//}

func (repo ResourceGroupMappingRepositoryImpl) FindByResourceGroupId(resourceGroupId int) ([]*ResourceGroupMapping, error) {
	var models []*ResourceGroupMapping
	err := repo.dbConnection.Model(&models).
		Column("resource_group_mapping.*", "ResourceGroup").
		Where("resource_group_mapping.resource_group_id = ?", resourceGroupId).
		Select()
	return models, err
}

//func (repo ResourceGroupMappingRepositoryImpl) FindAll() ([]*ResourceGroupMapping, error) {
//	var models []*ResourceGroupMapping
//	err := repo.dbConnection.Model(&models).Select()
//	return models, err
//}

func (repo ResourceGroupMappingRepositoryImpl) FindByResourceGroupIds(resourceGroupIds []int) ([]*ResourceGroupMapping, error) {
	var models []*ResourceGroupMapping
	err := repo.dbConnection.Model(&models).
		Where("resource_group_id in (?)", pg.In(resourceGroupIds)).
		Select()
	return models, err
}
