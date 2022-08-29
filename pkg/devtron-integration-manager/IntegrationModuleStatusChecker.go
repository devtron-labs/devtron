package devtron_integration_manager

import "go.uber.org/zap"

type IntegrationModuleStatusChecker interface {
	AddIntegrationModule(moduleName string)
}

type IntegrationModuleStatusCheckerImpl struct {
	logger                *zap.SugaredLogger
	taskExecutorService   *TaskExecutorService
	integrationRepository *IntegrationManagerRepository
	modulesToSync         []string
}

func NewIntegrationModuleStatusCheckerImpl(logger *zap.SugaredLogger, taskExecutorService *TaskExecutorService,
	integrationRepository *IntegrationManagerRepository) *IntegrationModuleStatusCheckerImpl {
	checkerImpl := &IntegrationModuleStatusCheckerImpl{
		logger:                logger,
		taskExecutorService:   taskExecutorService,
		integrationRepository: integrationRepository,
	}

	//TODO run cron task for modules state check
	return checkerImpl
}

func (impl *IntegrationModuleStatusCheckerImpl) AddIntegrationModule(moduleName string) {
	for _, module := range impl.modulesToSync {
		if module == moduleName {
			return
		}
	}
	impl.modulesToSync = append(impl.modulesToSync, moduleName)
}
