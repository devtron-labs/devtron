package devtron_integration_manager

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type IntegrationManagerRepository interface {
	SaveIntegrationModule(module *IntegrationModuleEntity) error
	UpdateIntegrationModuleStatus(userId int32, moduleName string, status string, detailedStatus string) error
	GetIntegrationModule(moduleName string) (*IntegrationModuleEntity, error)
	GetAlIntegrationModules() ([]*IntegrationModuleEntity, error)
}

type IntegrationModuleEntity struct {
	tableName      struct{} `sql:"integration_module" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk"`
	Name           string   `sql:"name,notnull"`
	Status         string   `sql:"status,notnull"`
	DetailedStatus string   `sql:"detailed_status, notnull"`
	sql.AuditLog
}

type IntegrationManagerRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewIntegrationManagerRepositoryImpl(dbConnection *pg.DB) *IntegrationManagerRepositoryImpl {
	return &IntegrationManagerRepositoryImpl{dbConnection: dbConnection}
}

func (impl *IntegrationManagerRepositoryImpl) SaveIntegrationModule(module *IntegrationModuleEntity) error {
	err := impl.dbConnection.Insert(module)
	return err
}

func (impl *IntegrationManagerRepositoryImpl) UpdateIntegrationModuleStatus(userId int32, moduleName string, status string, detailedStatus string) error {
	integrationModule, err := impl.GetIntegrationModule(moduleName)
	if err != nil {
		return nil
	}
	integrationModule.Status = status
	integrationModule.DetailedStatus = detailedStatus
	integrationModule.UpdatedOn = time.Now()
	integrationModule.UpdatedBy = userId
	err = impl.dbConnection.Update(integrationModule)
	return err
}

func (impl *IntegrationManagerRepositoryImpl) GetIntegrationModule(moduleName string) (*IntegrationModuleEntity, error) {
	var model *IntegrationModuleEntity
	err := impl.dbConnection.Model(&model).Where("name = ?", moduleName).
		Select()
	return model, err
}

func (impl *IntegrationManagerRepositoryImpl) GetAlIntegrationModules() ([]*IntegrationModuleEntity, error) {
	var models []*IntegrationModuleEntity
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}
