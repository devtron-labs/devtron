package userResource

import (
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/userResource/adapter"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
	http2 "net/http"
)

func (impl *UserResourceServiceImpl) enforceRbacForTeamForHelmApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamsResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForEnvForHelmApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.HelmEnvResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterList(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.ClusterResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterApiResource(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.ApiResourcesResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterResourceList(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.ClusterResourcesResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForClusterNamespaces(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.NameSpaces), nil
}
