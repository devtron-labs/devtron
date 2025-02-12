package userResource

import (
	"context"
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	bean4 "github.com/devtron-labs/devtron/pkg/k8s/bean"
	"github.com/devtron-labs/devtron/pkg/team"
	bean5 "github.com/devtron-labs/devtron/pkg/userResource/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/helper"
	"github.com/devtron-labs/devtron/util/commonEnforcementFunctionsUtil"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
)

type UserResourceService interface {
	GetResourceOptions(context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto,
		params *apiBean.PathParams) (*bean5.UserResourceResponseDto, error)
}
type UserResourceServiceImpl struct {
	logger                *zap.SugaredLogger
	teamService           team.TeamService
	envService            environment.EnvironmentService
	clusterService        cluster.ClusterService
	k8sApplicationService application2.K8sApplicationService
	enforcerUtil          rbac.EnforcerUtil
	rbacEnforcementUtil   commonEnforcementFunctionsUtil.CommonEnforcementUtil
	enforcer              casbin.Enforcer
	appService            app.AppCrudOperationService
}

func NewUserResourceServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	envService environment.EnvironmentService,
	clusterService cluster.ClusterService,
	k8sApplicationService application2.K8sApplicationService,
	enforcerUtil rbac.EnforcerUtil,
	rbacEnforcementUtil commonEnforcementFunctionsUtil.CommonEnforcementUtil,
	enforcer casbin.Enforcer,
	appService app.AppCrudOperationService) *UserResourceServiceImpl {
	return &UserResourceServiceImpl{
		logger:                logger,
		teamService:           teamService,
		envService:            envService,
		clusterService:        clusterService,
		k8sApplicationService: k8sApplicationService,
		enforcerUtil:          enforcerUtil,
		rbacEnforcementUtil:   rbacEnforcementUtil,
		enforcer:              enforcer,
		appService:            appService,
	}
}

func (impl *UserResourceServiceImpl) GetResourceOptions(context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.UserResourceResponseDto, error) {
	err := helper.ValidateResourceOptionReqBean(reqBean)
	if err != nil {
		impl.logger.Errorw("error in GetResourceOptions", "err", err, "reqBean", reqBean)
		return nil, err
	}
	// validation based on kind ,sub kind and entity and access type
	f := getResourceOptionFunc(bean5.UserResourceKind(params.Kind), bean5.Version(params.Version))
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
	f2 := getResourceOptionRbacFunc(bean5.UserResourceKind(params.Kind), bean5.Version(params.Version), reqBean.Entity, reqBean.AccessType)
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

func (impl *UserResourceServiceImpl) getTeamResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// get team resource options
	teams, err := impl.teamService.FetchAllActive()
	if err != nil {
		impl.logger.Errorw("error in getTeamResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithTeamsResp(teams), nil

}
func (impl *UserResourceServiceImpl) getAppResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// empty app type is devtron-app
	apps, err := impl.appService.GetAppListByTeamIds(reqBean.TeamIds, app2.DevtronApp)
	if err != nil {
		impl.logger.Errorw("error encountered in getAppResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithTeamAppResp(apps), nil
}
func (impl *UserResourceServiceImpl) getHelmAppResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	apps, err := impl.appService.GetAppListByTeamIds(reqBean.TeamIds, app2.DevtronChart)
	if err != nil {
		impl.logger.Errorw("error encountered in getAppResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithTeamAppResp(apps), nil
}

func (impl *UserResourceServiceImpl) getHelmEnvResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {

	// get helm env resource options
	env, err := impl.envService.GetCombinedEnvironmentListForDropDown(token, true, impl.rbacEnforcementUtil.CheckAuthorizationByEmailInBatchForGlobalEnvironment)
	if err != nil {
		impl.logger.Errorw("error in getEnvResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithHelmEnvResp(env), nil
}

func (impl *UserResourceServiceImpl) getClusterResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// System User id is passed as 1 as rbac enforcement is handled globally at bottom level and Super-admin is passed as true here
	clusters, err := impl.clusterService.FindAllForClusterByUserId(1, true)
	if err != nil {
		impl.logger.Errorw("error encountered in getClusterResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithClusterResp(clusters), nil
}

func (impl *UserResourceServiceImpl) getClusterNamespacesResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// System User id is passed as 1 as rbac enforcement is handled globally at bottom level and Super-admin is passed as true here
	namespaces, err := impl.clusterService.FindAllNamespacesByUserIdAndClusterId(bean.SystemUserId, reqBean.ClusterId, true)
	if err != nil {
		impl.logger.Errorw("error encountered in getClusterNamespacesResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithNameSpaces(namespaces), nil
}

func (impl *UserResourceServiceImpl) getClusterApiResourceOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	// System User id is passed as 1 as rbac enforcement is handled globally at bottom level and Super-admin is passed as true here
	apiResources, err := impl.k8sApplicationService.GetAllApiResources(context, reqBean.ClusterId, true, bean.SystemUserId)
	if err != nil {
		impl.logger.Errorw("error encountered in getClusterApiResourceOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithApiResourcesResp(apiResources), nil
}

func (impl *UserResourceServiceImpl) getClusterResourceListOptions(context context.Context, token string,
	reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean5.ResourceOptionsDto, error) {
	//  rbac enforcement is handled globally at bottom level and rbac function is passed as true
	clusterRbacFunc := func(token, clusterName string, request bean4.ResourceRequestBean, casbinAction string) bool {
		return true
	}
	resourceList, err := impl.k8sApplicationService.GetResourceList(context, token, reqBean.ResourceRequestBean, clusterRbacFunc)
	if err != nil {
		impl.logger.Errorw("error in getClusterResourceListOptions", "err", err)
		return nil, err
	}
	return bean5.NewResourceOptionsDto().WithClusterResourcesResp(resourceList), nil
}
