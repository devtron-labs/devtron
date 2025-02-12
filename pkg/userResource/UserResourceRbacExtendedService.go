package userResource

import (
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/adapter"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func (impl *UserResourceServiceImpl) enforceRbacForTeamForDevtronApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamsResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForTeamForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamsResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForEnvForDevtronApp(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.EnvResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForEnvForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.EnvResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForDevtronApps(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamAppResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForHelmApps(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.TeamAppResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForJobs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.JobsResp), nil
}

func (impl *UserResourceServiceImpl) enforceRbacForChartGroup(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.ChartGroupResp), nil
}
func (impl *UserResourceServiceImpl) enforceRbacForJobsWfs(token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	return adapter.BuildUserResourceResponseDto(resourceOptions.JobsResp), nil
}
