package user

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/user/bean"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"go.uber.org/zap"
)

type DefaultRbacRoleService interface {
	SyncRbacRoleData() error
}

type DefaultRbacRoleServiceImpl struct {
	logger                        *zap.SugaredLogger
	defaultRbacRoleDataRepository repository.DefaultRbacRoleDataRepository
	rbacRoleService               RbacRoleService
}

func NewDefaultRbacRoleServiceImpl(logger *zap.SugaredLogger, defaultRbacRoleDataRepository repository.DefaultRbacRoleDataRepository,
	rbacRoleService RbacRoleService) *DefaultRbacRoleServiceImpl {
	serviceImpl := &DefaultRbacRoleServiceImpl{
		logger:                        logger,
		defaultRbacRoleDataRepository: defaultRbacRoleDataRepository,
		rbacRoleService:               rbacRoleService,
	}
	go serviceImpl.SyncRbacRoleData()
	return serviceImpl
}

func (impl *DefaultRbacRoleServiceImpl) SyncRbacRoleData() error {
	impl.logger.Infow("going to sync rbac role data")
	defaultRbacRoles, err := impl.defaultRbacRoleDataRepository.GetAllDefaultRbacRole()
	if err != nil {
		return err
	}
	impl.logger.Debugw("total configured default roles", "size", len(defaultRbacRoles))
	defaultConfiguredRoles, err := impl.getConfiguredRbacRole()
	if err != nil {
		return err
	}
	for _, defaultRbacRole := range defaultRbacRoles {
		impl.handleDefaultRbacRole(defaultRbacRole, defaultConfiguredRoles)
	}
	return nil
}

func (impl *DefaultRbacRoleServiceImpl) getConfiguredRbacRole() (map[string]*bean.RbacRoleDto, error) {
	defaultConfiguredRoles, err := impl.rbacRoleService.GetAllDefaultRoles()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching default roles, ignoring default rbac sync", "err", err)
		return nil, err
	}
	configuredRoleVsData := make(map[string]*bean.RbacRoleDto)
	for _, configuredRole := range defaultConfiguredRoles {
		roleName := configuredRole.RoleName
		configuredRoleVsData[roleName] = configuredRole
	}
	return configuredRoleVsData, nil
}

func (impl *DefaultRbacRoleServiceImpl) handleDefaultRbacRole(defaultRbacRole *repository.DefaultRbacRoleDto, configuredCustomRoles map[string]*bean.RbacRoleDto) {
	roleName := defaultRbacRole.Role
	if configuredRbacRole, ok := configuredCustomRoles[roleName]; ok {
		enabled := defaultRbacRole.Enabled
		if !enabled {
			//TODO need to disable this entry in db
		} else {
			configuredRbacRoleId := configuredRbacRole.Id
			impl.createOrUpdateDefaultRbacRole(defaultRbacRole, roleName, configuredRbacRoleId)
		}
	} else {
		impl.createOrUpdateDefaultRbacRole(defaultRbacRole, roleName, 0)
	}
}

func (impl *DefaultRbacRoleServiceImpl) createOrUpdateDefaultRbacRole(defaultRbacRole *repository.DefaultRbacRoleDto, roleName string, configuredRbacRoleId int) {
	rbacRoleDto, err := impl.getDefaultRbacRoleDto(defaultRbacRole, roleName)
	if err != nil {
		return
	}
	if configuredRbacRoleId == 0 {
		err = impl.rbacRoleService.CreateDefaultRole(rbacRoleDto, 1)
	} else {
		rbacRoleDto.Id = configuredRbacRoleId
		err = impl.rbacRoleService.UpdateDefaultRole(rbacRoleDto, 1)
	}
	if err != nil {
		impl.logger.Errorw("error occurred while creating/updating default role", "roleName", roleName,
			"configuredRbacRoleId", configuredRbacRoleId, "err", err)
	}
}

func (impl *DefaultRbacRoleServiceImpl) getDefaultRbacRoleDto(defaultRbacRole *repository.DefaultRbacRoleDto, roleName string) (*bean.RbacRoleDto, error) {
	defaultRoleData := defaultRbacRole.DefaultRoleData
	rbacRoleDto := &bean.RbacRoleDto{}
	err := json.Unmarshal([]byte(defaultRoleData), rbacRoleDto)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling rbacRoleDto", "roleName", roleName, "err", err)
	}
	return rbacRoleDto, err
}
