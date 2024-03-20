package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type TimeoutWindowResourceMappingRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewTimeoutWindowResourceMappingRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *TimeoutWindowResourceMappingRepositoryImpl {
	return &TimeoutWindowResourceMappingRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type TimeoutWindowResourceMapping struct {
	TableName       struct{}     `sql:"timeout_window_resource_mappings" pg:",discard_unknown_columns"`
	Id              int          `sql:"id,pk"`
	TimeoutWindowId int          `sql:"timeout_window_configuration_id,pk"`
	ResourceId      int          `sql:"resource_id"`
	ResourceType    ResourceType `sql:"resource_type"`
}

type ResourceType int

const (
	DeploymentWindowProfile ResourceType = 1
)

type TimeoutWindowResourceMappingRepository interface {
	Create(tx *pg.Tx, mappings []*TimeoutWindowResourceMapping) ([]*TimeoutWindowResourceMapping, error)
	GetWindowsForResources(resourceIds []int, resourceType ResourceType) ([]*TimeoutWindowResourceMapping, error)
	DeleteAllForResource(tx *pg.Tx, resourceId int, resourceType ResourceType) error
}

func (impl TimeoutWindowResourceMappingRepositoryImpl) Create(tx *pg.Tx, mappings []*TimeoutWindowResourceMapping) ([]*TimeoutWindowResourceMapping, error) {
	return mappings, tx.Insert(&mappings)
}

func (impl TimeoutWindowResourceMappingRepositoryImpl) GetWindowsForResources(resourceIds []int, resourceType ResourceType) ([]*TimeoutWindowResourceMapping, error) {
	var mappings []*TimeoutWindowResourceMapping

	if len(resourceIds) == 0 {
		return mappings, nil
	}
	err := impl.dbConnection.Model(&mappings).
		Where("resource_id IN (?)", pg.In(resourceIds)).
		Where("resource_type = ?", resourceType).
		Select()
	if err == pg.ErrNoRows {
		return []*TimeoutWindowResourceMapping{}, nil
	}
	return mappings, err
}

func (impl TimeoutWindowResourceMappingRepositoryImpl) DeleteAllForResource(tx *pg.Tx, resourceId int, resourceType ResourceType) error {
	_, err := tx.Model(&TimeoutWindowResourceMapping{}).
		Where("resource_id = ?", resourceId).
		Where("resource_type = ?", resourceType).
		Delete()
	return err
}
