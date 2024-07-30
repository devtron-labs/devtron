/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
