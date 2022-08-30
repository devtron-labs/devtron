package devtron_integration_manager

import (
	"errors"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type IntegrationManagerService interface {
	InstallModule(userId int32, moduleName string) error
	GetAllModules() ([]*IntegrationModule, error)
	GetModulesStatus(moduleNames []string) ([]*IntegrationModule, error)
	UpdateModuleStatus(userId int32, moduleName string, status string, detailedStatus string) error
}

type IntegrationManagerServiceImpl struct {
	logger                       *zap.SugaredLogger
	statusChecker                IntegrationModuleStatusChecker
	integrationManagerRepository IntegrationManagerRepository
}

type IntegrationModule struct {
	Name           string `json:"name,omitempty"`
	Status         string `json:"status,omitempty"`
	DetailedStatus string `json:"detailedStatus,omitempty"`
}

func NewIntegrationManagerServiceImpl(logger *zap.SugaredLogger, statusChecker IntegrationModuleStatusChecker, integrationManagerRepository IntegrationManagerRepository) *IntegrationManagerServiceImpl {
	integrationManager := &IntegrationManagerServiceImpl{
		logger:                       logger,
		statusChecker:                statusChecker,
		integrationManagerRepository: integrationManagerRepository,
	}

	return integrationManager
}

func (impl *IntegrationManagerServiceImpl) InstallModule(userId int32, moduleName string) error {

	//TODO make request to kubelink
	module := &IntegrationModuleEntity{
		Name:   moduleName,
		Status: "Installing",
	}
	module.CreatedOn = time.Now()
	module.CreatedBy = userId
	module.UpdatedOn = time.Now()
	module.UpdatedBy = userId
	err := impl.integrationManagerRepository.SaveIntegrationModule(module)
	impl.statusChecker.AddIntegrationModule(moduleName)
	return err
}

func (impl *IntegrationManagerServiceImpl) GetAllModules() ([]*IntegrationModule, error) {
	moduleEntities, err := impl.integrationManagerRepository.GetAlIntegrationModules()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching modules from db ", "reason", err)
		if err != pg.ErrNoRows {
			return []*IntegrationModule{}, nil
		} else {
			return nil, errors.New("error while fetching modules")
		}
	}
	modules := impl.convertToModule(moduleEntities)
	return modules, err
}

func (impl *IntegrationManagerServiceImpl) convertToModule(moduleEntities []*IntegrationModuleEntity) []*IntegrationModule {
	var modules []*IntegrationModule
	for _, moduleEntity := range moduleEntities {
		module := &IntegrationModule{Name: moduleEntity.Name, Status: moduleEntity.Status, DetailedStatus: moduleEntity.DetailedStatus}
		modules = append(modules, module)
	}
	return modules
}

func (impl *IntegrationManagerServiceImpl) GetModulesStatus(moduleNames []string) ([]*IntegrationModule, error) {

	moduleEntities, err := impl.integrationManagerRepository.GetModulesStatus(moduleNames)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching modules status", "modules", moduleNames, "reason", err)
		return nil, errors.New("error while fetching modules statuses")
	}
	finalModules := impl.convertToModule(moduleEntities)
	return finalModules, nil
}

func (impl *IntegrationManagerServiceImpl) UpdateModuleStatus(userId int32, moduleName string, status string, detailedStatus string) error {
	err := impl.integrationManagerRepository.UpdateIntegrationModuleStatus(userId, moduleName, status, detailedStatus)
	if err != nil {
		impl.logger.Errorw("error occurred while updating status for module ", "name", moduleName,
			"status", status, "detailedStatus", detailedStatus, "user", userId, "error", err)
		return errors.New("error while updating modules status")
	}
	return nil
}
