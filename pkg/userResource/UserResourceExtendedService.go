package userResource

import (
	"context"
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/userResource/adapter"
	bean5 "github.com/devtron-labs/devtron/pkg/userResource/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/helper"
	"github.com/devtron-labs/devtron/util/commonEnforcementFunctionsUtil"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
)

type UserResourceExtendedServiceImpl struct {
	logger             *zap.SugaredLogger
	chartGroupService  chartGroup.ChartGroupService
	appListingService  app.AppListingService
	appWorkflowService appWorkflow.AppWorkflowService
	*UserResourceServiceImpl
}

func NewUserResourceExtendedServiceImpl(logger *zap.SugaredLogger, teamService team.TeamService,
	envService environment.EnvironmentService,
	appService app.AppCrudOperationService,
	chartGroupService chartGroup.ChartGroupService,
	appListingService app.AppListingService,
	appWorkflowService appWorkflow.AppWorkflowService,
	k8sApplicationService application2.K8sApplicationService,
	clusterService cluster.ClusterService,
	rbacEnforcementUtil commonEnforcementFunctionsUtil.CommonEnforcementUtil,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *UserResourceExtendedServiceImpl {
	return &UserResourceExtendedServiceImpl{
		logger:                  logger,
		chartGroupService:       chartGroupService,
		appListingService:       appListingService,
		appWorkflowService:      appWorkflowService,
		UserResourceServiceImpl: NewUserResourceServiceImpl(logger, teamService, envService, clusterService, k8sApplicationService, enforcerUtil, rbacEnforcementUtil, enforcer, appService),
	}

}

func (impl *UserResourceExtendedServiceImpl) GetResourceOptions(context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.UserResourceResponseDto, error) {
	err := helper.ValidateResourceOptionReqBean(reqBean)
	if err != nil {
		impl.logger.Errorw("error in GetResourceOptions", "err", err, "reqBean", reqBean)
		return nil, err
	}
	f := getResourceOptionExtendedFunc(bean5.UserResourceKind(params.Kind), bean5.Version(params.Version))
	if f == nil {
		impl.logger.Errorw("error encountered in GetResourceOptions, not supported kind", "params", params)
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean5.RequestInvalidKindVersionErrMessage, bean5.RequestInvalidKindVersionErrMessage)
	}
	data, err := f(impl, context, token, reqBean, params)
	if err != nil {
		impl.logger.Errorw("error in GetResourceOptions", "err", err, "reqBean", reqBean)
		return nil, err
	}
	// rbac function get and enforce at service level
	f2 := getResourceOptionRbacExtendedFunc(bean5.UserResourceKind(params.Kind), bean5.Version(params.Version), reqBean.Entity, reqBean.AccessType)
	if f2 == nil {
		impl.logger.Errorw("error encountered in GetResourceOptions, not supported kind for rbac", "params", params)
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean5.RequestInvalidKindVersionErrMessage, bean5.RequestInvalidKindVersionErrMessage)
	}
	finalData, err := f2(impl, token, params, data)
	if err != nil {
		impl.logger.Errorw("error in GetResourceOptions", "err", err, "reqBean", reqBean)
		return nil, err
	}
	return finalData, nil
}

func (impl *UserResourceExtendedServiceImpl) getEnvResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// get env resource options
	env, err := impl.envService.GetEnvironmentListForAutocomplete(false)
	if err != nil {
		impl.logger.Errorw("error in getEnvResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithEnvResp(env), nil

}

func (impl *UserResourceExtendedServiceImpl) getChartGroupResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// get chart group resource options
	// max is passed as 0 to get all chart groups, default behaviour
	chartGroups, err := impl.chartGroupService.ChartGroupList(0)
	if err != nil {
		impl.logger.Errorw("error in getChartGroupResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithChartGroupResp(chartGroups), nil
}

func (impl *UserResourceExtendedServiceImpl) getJobsResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	fetchJobListingRequest := adapter.BuildFetchAppListingReqForJobFromDto(reqBean)
	jobs, err := impl.appListingService.FetchJobs(fetchJobListingRequest)
	if err != nil {
		impl.logger.Errorw("error in getJobsResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithJobsResp(jobs), nil
}

func (impl *UserResourceExtendedServiceImpl) getAppWfsResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	workflows, err := impl.appWorkflowService.FindAllWorkflowsForApps(*reqBean.WorkflowNamesRequest)
	if err != nil {
		impl.logger.Errorw("error in getAppWfsResourceOptions", "err", err)
		return nil, err
	}

	return bean5.NewResourceOptionsDto().WithAppWfsResp(workflows), nil
}
