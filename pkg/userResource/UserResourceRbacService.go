/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package userResource

import (
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/userResource/adapter"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func (impl *UserResourceServiceImpl) enforceRbacForTeamForHelmApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForTeamForHelmApp")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamsResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForEnvForHelmApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForEnvForHelmApp")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.HelmEnvResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterList(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForClusterList")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.ClusterResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterApiResource(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForClusterApiResource")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.ApiResourcesResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterResourceList(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForClusterResourceList")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.ClusterResourcesResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterNamespaces(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		impl.logger.Errorw("user is unauthorized to enforceRbacForClusterNamespaces")
		return adapter.BuildNullDataUserResourceResponseDto(), nil
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.NameSpaces), nil
}
