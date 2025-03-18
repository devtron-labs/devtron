package userResource

import (
	"context"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func getUserResourceKindWithEntityAccessKey(kind bean.UserResourceKind, version bean.Version, entity, accessType string) string {
	return fmt.Sprintf("%s_%s_%s", getUserResourceKindWithVersionKey(kind, version), entity, accessType)
}
func getUserResourceKindWithVersionKey(kind bean.UserResourceKind, version bean.Version) string {
	return fmt.Sprintf("%s_%s", kind, version)
}

var mapOfUserResourceKindToAllResourceOptionsFunc = map[string]func(impl *UserResourceServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error){
	getUserResourceKindWithVersionKey(bean.KindTeam, bean.Alpha1Version):            (*UserResourceServiceImpl).getTeamResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindHelmEnvironment, bean.Alpha1Version): (*UserResourceServiceImpl).getHelmEnvResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindHelmApplication, bean.Alpha1Version): (*UserResourceServiceImpl).getHelmAppResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindCluster, bean.Alpha1Version):         (*UserResourceServiceImpl).getClusterResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterApiResources, bean.Alpha1Version): (*UserResourceServiceImpl).getClusterApiResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterNamespaces, bean.Alpha1Version):   (*UserResourceServiceImpl).getClusterNamespacesResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterResources, bean.Alpha1Version):    (*UserResourceServiceImpl).getClusterResourceListOptions,
}

func getAllResourceOptionsFunc(kind bean.UserResourceKind, version bean.Version) func(impl *UserResourceServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error) {
	if f, ok := mapOfUserResourceKindToAllResourceOptionsFunc[getUserResourceKindWithVersionKey(kind, version)]; ok {
		return f
	}
	return nil
}

var mapOfUserResourceKindToAllResourceOptionsExtendedFunc = map[string]func(impl *UserResourceExtendedServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error){
	getUserResourceKindWithVersionKey(bean.KindTeam, bean.Alpha1Version):               (*UserResourceExtendedServiceImpl).getTeamResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindHelmEnvironment, bean.Alpha1Version):    (*UserResourceExtendedServiceImpl).getHelmEnvResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindHelmApplication, bean.Alpha1Version):    (*UserResourceExtendedServiceImpl).getHelmAppResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindCluster, bean.Alpha1Version):            (*UserResourceExtendedServiceImpl).getClusterResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterApiResources, bean.Alpha1Version):    (*UserResourceExtendedServiceImpl).getClusterApiResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterNamespaces, bean.Alpha1Version):      (*UserResourceExtendedServiceImpl).getClusterNamespacesResourceOptions,
	getUserResourceKindWithVersionKey(bean.ClusterResources, bean.Alpha1Version):       (*UserResourceExtendedServiceImpl).getClusterResourceListOptions,
	getUserResourceKindWithVersionKey(bean.KindEnvironment, bean.Alpha1Version):        (*UserResourceExtendedServiceImpl).getEnvResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindDevtronApplication, bean.Alpha1Version): (*UserResourceExtendedServiceImpl).getAppResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindChartGroup, bean.Alpha1Version):         (*UserResourceExtendedServiceImpl).getChartGroupResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindJobs, bean.Alpha1Version):               (*UserResourceExtendedServiceImpl).getJobsResourceOptions,
	getUserResourceKindWithVersionKey(bean.KindWorkflow, bean.Alpha1Version):           (*UserResourceExtendedServiceImpl).getAppWfsResourceOptions,
}

func getAllResourceOptionsExtendedFunc(kind bean.UserResourceKind, version bean.Version) func(impl *UserResourceExtendedServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error) {
	if f, ok := mapOfUserResourceKindToAllResourceOptionsExtendedFunc[getUserResourceKindWithVersionKey(kind, version)]; ok {
		return f
	}
	return nil
}

var mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacFunc = map[string]func(impl *UserResourceServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error){
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM):            (*UserResourceServiceImpl).enforceRbacForTeamForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmApplication, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceServiceImpl).enforceRbacForHelmAppsListing,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmEnvironment, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceServiceImpl).enforceRbacForEnvForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindCluster, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):          (*UserResourceServiceImpl).enforceRbacForClusterList,
	getUserResourceKindWithEntityAccessKey(bean.ClusterApiResources, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):  (*UserResourceServiceImpl).enforceRbacForClusterApiResource,
	getUserResourceKindWithEntityAccessKey(bean.ClusterNamespaces, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):    (*UserResourceServiceImpl).enforceRbacForClusterNamespaces,
	getUserResourceKindWithEntityAccessKey(bean.ClusterResources, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):     (*UserResourceServiceImpl).enforceRbacForClusterResourceList,
}

func getResourceOptionRbacFunc(kind bean.UserResourceKind, version bean.Version, entity string, accessType string) func(impl *UserResourceServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	if f, ok := mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacFunc[getUserResourceKindWithEntityAccessKey(kind, version, entity, accessType)]; ok {
		return f
	}
	return nil
}

var mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacExtendedFunc = map[string]func(impl *UserResourceExtendedServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error){
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM):            (*UserResourceExtendedServiceImpl).enforceRbacForTeamForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmApplication, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceExtendedServiceImpl).enforceRbacForHelmAppsListing,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmEnvironment, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceExtendedServiceImpl).enforceRbacForEnvForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindCluster, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):          (*UserResourceExtendedServiceImpl).enforceRbacForClusterList,
	getUserResourceKindWithEntityAccessKey(bean.ClusterApiResources, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):  (*UserResourceExtendedServiceImpl).enforceRbacForClusterApiResource,
	getUserResourceKindWithEntityAccessKey(bean.ClusterNamespaces, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):    (*UserResourceExtendedServiceImpl).enforceRbacForClusterNamespaces,
	getUserResourceKindWithEntityAccessKey(bean.ClusterResources, bean.Alpha1Version, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):     (*UserResourceExtendedServiceImpl).enforceRbacForClusterResourceList,
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.DEVTRON_APP):                     (*UserResourceExtendedServiceImpl).enforceRbacForTeamForDevtronApp,
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean.Alpha1Version, bean2.EntityJobs, bean2.EmptyAccessType):                  (*UserResourceExtendedServiceImpl).enforceRbacForTeamForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindEnvironment, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.DEVTRON_APP):              (*UserResourceExtendedServiceImpl).enforceRbacForEnvForDevtronApp,
	getUserResourceKindWithEntityAccessKey(bean.KindEnvironment, bean.Alpha1Version, bean2.EntityJobs, bean2.EmptyAccessType):           (*UserResourceExtendedServiceImpl).enforceRbacForEnvForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindDevtronApplication, bean.Alpha1Version, bean2.ENTITY_APPS, bean2.DEVTRON_APP):       (*UserResourceExtendedServiceImpl).enforceRbacForDevtronApps,
	getUserResourceKindWithEntityAccessKey(bean.KindChartGroup, bean.Alpha1Version, bean2.CHART_GROUP_ENTITY, bean2.EmptyAccessType):    (*UserResourceExtendedServiceImpl).enforceRbacForChartGroup,
	getUserResourceKindWithEntityAccessKey(bean.KindJobs, bean.Alpha1Version, bean2.EntityJobs, bean2.EmptyAccessType):                  (*UserResourceExtendedServiceImpl).enforceRbacForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindWorkflow, bean.Alpha1Version, bean2.EntityJobs, bean2.EmptyAccessType):              (*UserResourceExtendedServiceImpl).enforceRbacForJobsWfs,
}

func getResourceOptionRbacExtendedFunc(kind bean.UserResourceKind, version bean.Version, entity string, accessType string) func(impl *UserResourceExtendedServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	if f, ok := mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacExtendedFunc[getUserResourceKindWithEntityAccessKey(kind, version, entity, accessType)]; ok {
		return f
	}
	return nil
}
