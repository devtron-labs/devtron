package devtron_integration_manager

import "go.uber.org/zap"

type IntegrationManagerService interface {
	InstallModule(moduleName string) error
	GetAllModules() []*IntegrationModule
	GetModulesStatus() []*IntegrationModule
	UpdateModuleStatus() *IntegrationModule
}

type IntegrationManagerServiceImpl struct {
	logger                       *zap.SugaredLogger
	executorService              *TaskExecutorService
	integrationManagerRepository *IntegrationManagerRepository
}

type IntegrationModule struct {
	name           string
	status         string
	detailedStatus string
}

func NewIntegrationManagerServiceImpl(logger *zap.SugaredLogger, executorService *TaskExecutorService, integrationManagerRepository *IntegrationManagerRepository) *IntegrationManagerServiceImpl {
	integrationManager := &IntegrationManagerServiceImpl{
		logger:                       logger,
		executorService:              executorService,
		integrationManagerRepository: integrationManagerRepository,
	}

	return integrationManager
}

func (impl *IntegrationManagerServiceImpl) InstallModule(moduleName string) error {
	return nil
}

func (impl *IntegrationManagerServiceImpl) GetAllModules() []*IntegrationModule {
	return nil
}

func (impl *IntegrationManagerServiceImpl) GetModulesStatus() []*IntegrationModule {
	return nil
}
func (impl *IntegrationManagerServiceImpl) UpdateModuleStatus() *IntegrationModule {
	return nil
}
