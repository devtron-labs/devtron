package userResource

import (
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean3 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/adapter"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
	http2 "net/http"
)

func (impl *UserResourceServiceImpl) enforceRbacForTeamForDevtronApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	result := make([]bean2.TeamRequest, 0, len(resourceOptions.TeamsResp))
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if isSuperAdmin {
		result = resourceOptions.TeamsResp
	} else {
		for _, item := range resourceOptions.TeamsResp {
			if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionGet, item.Name); ok {
				result = append(result, item)
			}
		}
	}
	return adapter.BuildUserResourceResponseDto(result), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForTeamForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamsResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForEnvForDevtronApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	result := make([]bean3.EnvironmentBean, 0, len(resourceOptions.EnvResp))
	isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if isSuperAdmin {
		result = resourceOptions.EnvResp
	} else {
		result = impl.rbacEnforcementUtil.CheckAuthorisationForEnvAutocomplete(token, resourceOptions.EnvResp)
	}

	return adapter.BuildUserResourceResponseDto(result), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForEnvForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.EnvResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForDevtronApps(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	result := impl.rbacEnforcementUtil.CheckAuthorisationOnApp(token, resourceOptions.TeamAppResp)
	return adapter.BuildUserResourceResponseDto(result), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForHelmApps(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamAppResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.JobsResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForChartGroup(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*") ||
		impl.enforcer.Enforce(token, casbin.ResourceChartGroup, casbin.ActionGet, "") // doing this for manager check as chart group is by default shown to every user
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.ChartGroupResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForJobsWfs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
	if !isAuthorised {
		return nil, util.GetApiErrorAdapter(http2.StatusForbidden, "403", bean.UnAuthorizedAccess, bean.UnAuthorizedAccess)
	}
	return adapter.BuildUserResourceResponseDto(resourceOptions.JobsResp), nil
}
