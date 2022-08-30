package devtron_integration_manager

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type IntegrationModuleStatusChecker interface {
	AddIntegrationModule(moduleName string)
}

type IntegrationModuleStatusCheckerImpl struct {
	logger                *zap.SugaredLogger
	taskExecutorService   *TaskExecutorService
	integrationRepository IntegrationManagerRepository
	modulesToSync         map[string]bool
	cron                  *cron.Cron
	integrationConfig     *IntegrationManagerConfig
}

func NewIntegrationModuleStatusCheckerImpl(logger *zap.SugaredLogger, taskExecutorService *TaskExecutorService,
	integrationRepository IntegrationManagerRepository, integrationConfig *IntegrationManagerConfig) (*IntegrationModuleStatusCheckerImpl, error) {
	checkerImpl := &IntegrationModuleStatusCheckerImpl{
		logger:                logger,
		taskExecutorService:   taskExecutorService,
		integrationRepository: integrationRepository,
		integrationConfig:     integrationConfig,
	}

	cron := cron.New(
		cron.WithChain())
	cron.Start()

	// add function into cron
	_, err := cron.AddFunc(fmt.Sprintf("@every %dm", integrationConfig.ModuleStatusCheckIntervalInMins), checkerImpl.SyncModuleState)
	if err != nil {
		fmt.Println("error in adding module status check function", err)
		return nil, err
	}

	return checkerImpl, nil
}

func (impl *IntegrationModuleStatusCheckerImpl) AddIntegrationModule(moduleName string) {
	impl.modulesToSync[moduleName] = true
}

func (impl *IntegrationModuleStatusCheckerImpl) SyncModuleState() {

	if len(impl.modulesToSync) < 1 {
		impl.logger.Infow("not checking as there is no module to sync!!")
		return
	}

	// check by hitting kubelink API
	// update all installing status if success comes from kubelink
	status := "Installed" // can be Failed too
	detailedStatus := "detailed-status"

	var modules []string
	for module := range impl.modulesToSync {
		modules = append(modules, module)
	}
	moduleEntities, err := impl.integrationRepository.GetModulesStatus(modules)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching modules from db", "modules", impl.modulesToSync, "error", err)
		return
	}

	for _, moduleEntity := range moduleEntities {
		moduleName := moduleEntity.Name
		moduleStatus := moduleEntity.Status
		if moduleStatus != "Installing" {
			impl.logger.Warnw("fetched module status is not installing", "module", moduleName, "status", moduleStatus)
		} else {
			err := impl.integrationRepository.UpdateIntegrationModuleStatus(1, moduleName, status, detailedStatus)
			if err != nil {
				impl.logger.Errorw("error occurred while updating status of module in db", "status", status, "module", moduleName, "err", err)
				continue
			}
		}
		delete(impl.modulesToSync, moduleName)
	}
}
