package user

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"go.uber.org/zap"
)

type RbacRoleService interface {
	GetAllDefaultRoles() ([]*bean.RbacRoleDto, error)
}

type RbacRoleServiceImpl struct {
	logger                 *zap.SugaredLogger
	rbacRoleDataRepository repository.RbacRoleDataRepository
}

func NewRbacRoleServiceImpl(logger *zap.SugaredLogger,
	rbacRoleDataRepository repository.RbacRoleDataRepository) *RbacRoleServiceImpl {
	return &RbacRoleServiceImpl{
		logger:                 logger,
		rbacRoleDataRepository: rbacRoleDataRepository,
	}
}
func (impl *RbacRoleServiceImpl) GetAllDefaultRoles() ([]*bean.RbacRoleDto, error) {
	//getting all roles from default data repository
	defaultRoles, err := impl.rbacRoleDataRepository.GetRoleDataForAllRoles()
	if err != nil {
		impl.logger.Errorw("error in getting all default roles data", "err", err)
		return nil, err
	}
	defaultRolesResp := make([]*bean.RbacRoleDto, 0, len(defaultRoles))
	for _, defaultRole := range defaultRoles {
		defaultRoleResp := &bean.RbacRoleDto{
			Id:              defaultRole.Id,
			RoleName:        defaultRole.Role,
			RoleDisplayName: defaultRole.RoleDisplayName,
			RoleDescription: defaultRole.RoleDescription,
			RbacPolicyEntityGroupDto: &bean.RbacPolicyEntityGroupDto{
				Entity:     defaultRole.Entity,
				AccessType: defaultRole.AccessType,
			},
		}
		defaultRolesResp = append(defaultRolesResp, defaultRoleResp)
	}
	return defaultRolesResp, nil
}
