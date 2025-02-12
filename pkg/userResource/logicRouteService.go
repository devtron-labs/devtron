package userResource

import (
	"context"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func getUserResourceKindWithEntityAccessKey(kind bean.UserResourceKind, entity, accessType string) string {
	return fmt.Sprintf("%s_%s_%s", kind, entity, accessType)
}

var mapOfUserResourceKindToResourceOptionFunc = map[bean.UserResourceKind]func(impl *UserResourceServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error){
	bean.KindTeam:            (*UserResourceServiceImpl).getTeamResourceOptions,
	bean.KindHelmEnvironment: (*UserResourceServiceImpl).getHelmEnvResourceOptions,
	bean.KindHelmApplication: (*UserResourceServiceImpl).getHelmAppResourceOptions,
	bean.KindCluster:         (*UserResourceServiceImpl).getClusterResourceOptions,
	bean.ClusterApiResources: (*UserResourceServiceImpl).getClusterApiResourceOptions,
	bean.ClusterNamespaces:   (*UserResourceServiceImpl).getClusterNamespacesResourceOptions,
	bean.ClusterResources:    (*UserResourceServiceImpl).getClusterResourceListOptions,
}

func getResourceOptionFunc(kind bean.UserResourceKind) func(impl *UserResourceServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error) {
	if f, ok := mapOfUserResourceKindToResourceOptionFunc[kind]; ok {
		return f
	}
	return nil
}

var mapOfUserResourceKindToResourceOptionExtendedFunc = map[bean.UserResourceKind]func(impl *UserResourceExtendedServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error){
	bean.KindTeam:               (*UserResourceExtendedServiceImpl).getTeamResourceOptions,
	bean.KindHelmEnvironment:    (*UserResourceExtendedServiceImpl).getHelmEnvResourceOptions,
	bean.KindHelmApplication:    (*UserResourceExtendedServiceImpl).getHelmAppResourceOptions,
	bean.KindCluster:            (*UserResourceExtendedServiceImpl).getClusterResourceOptions,
	bean.ClusterApiResources:    (*UserResourceExtendedServiceImpl).getClusterApiResourceOptions,
	bean.ClusterNamespaces:      (*UserResourceExtendedServiceImpl).getClusterNamespacesResourceOptions,
	bean.ClusterResources:       (*UserResourceExtendedServiceImpl).getClusterResourceListOptions,
	bean.KindEnvironment:        (*UserResourceExtendedServiceImpl).getEnvResourceOptions,
	bean.KindDevtronApplication: (*UserResourceExtendedServiceImpl).getAppResourceOptions,
	bean.KindChartGroup:         (*UserResourceExtendedServiceImpl).getChartGroupResourceOptions,
	bean.KindJobs:               (*UserResourceExtendedServiceImpl).getJobsResourceOptions,
	bean.KindWorkflow:           (*UserResourceExtendedServiceImpl).getAppWfsResourceOptions,
}

func getResourceOptionExtendedFunc(kind bean.UserResourceKind) func(impl *UserResourceExtendedServiceImpl, context context.Context, token string, reqBean *apiBean.ResourceOptionsReqDto, params *apiBean.PathParams) (*bean.ResourceOptionsDto, error) {
	if f, ok := mapOfUserResourceKindToResourceOptionExtendedFunc[kind]; ok {
		return f
	}
	return nil
}

var mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacFunc = map[string]func(impl *UserResourceServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error){
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM):            (*UserResourceServiceImpl).enforceRbacForTeamForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmApplication, bean2.ENTITY_APPS, bean2.DEVTRON_APP):          (*UserResourceServiceImpl).enforceRbacForHelmApps,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmEnvironment, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceServiceImpl).enforceRbacForEnvForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindCluster, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):          (*UserResourceServiceImpl).enforceRbacForClusterList,
	getUserResourceKindWithEntityAccessKey(bean.ClusterApiResources, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):  (*UserResourceServiceImpl).enforceRbacForClusterApiResource,
	getUserResourceKindWithEntityAccessKey(bean.ClusterNamespaces, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):    (*UserResourceServiceImpl).enforceRbacForClusterNamespaces,
	getUserResourceKindWithEntityAccessKey(bean.ClusterResources, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):     (*UserResourceServiceImpl).enforceRbacForClusterResourceList,
}

func getResourceOptionRbacFunc(kind bean.UserResourceKind, entity string, accessType string) func(impl *UserResourceServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	if f, ok := mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacFunc[getUserResourceKindWithEntityAccessKey(kind, entity, accessType)]; ok {
		return f
	}
	return nil
}

var mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacExtendedFunc = map[string]func(impl *UserResourceExtendedServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error){
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM):            (*UserResourceExtendedServiceImpl).enforceRbacForTeamForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmApplication, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceExtendedServiceImpl).enforceRbacForHelmApps,
	getUserResourceKindWithEntityAccessKey(bean.KindHelmEnvironment, bean2.ENTITY_APPS, bean2.APP_ACCESS_TYPE_HELM): (*UserResourceExtendedServiceImpl).enforceRbacForEnvForHelmApp,
	getUserResourceKindWithEntityAccessKey(bean.KindCluster, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):          (*UserResourceExtendedServiceImpl).enforceRbacForClusterList,
	getUserResourceKindWithEntityAccessKey(bean.ClusterApiResources, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):  (*UserResourceExtendedServiceImpl).enforceRbacForClusterApiResource,
	getUserResourceKindWithEntityAccessKey(bean.ClusterNamespaces, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):    (*UserResourceExtendedServiceImpl).enforceRbacForClusterNamespaces,
	getUserResourceKindWithEntityAccessKey(bean.ClusterResources, bean2.CLUSTER_ENTITIY, bean2.EmptyAccessType):     (*UserResourceExtendedServiceImpl).enforceRbacForClusterResourceList,
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean2.ENTITY_APPS, bean2.DEVTRON_APP):                     (*UserResourceExtendedServiceImpl).enforceRbacForTeamForDevtronApp,
	getUserResourceKindWithEntityAccessKey(bean.KindTeam, bean2.EntityJobs, bean2.EmptyAccessType):                  (*UserResourceExtendedServiceImpl).enforceRbacForTeamForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindEnvironment, bean2.ENTITY_APPS, bean2.DEVTRON_APP):              (*UserResourceExtendedServiceImpl).enforceRbacForEnvForDevtronApp,
	getUserResourceKindWithEntityAccessKey(bean.KindEnvironment, bean2.EntityJobs, bean2.EmptyAccessType):           (*UserResourceExtendedServiceImpl).enforceRbacForEnvForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindDevtronApplication, bean2.ENTITY_APPS, bean2.DEVTRON_APP):       (*UserResourceExtendedServiceImpl).enforceRbacForDevtronApps,
	getUserResourceKindWithEntityAccessKey(bean.KindChartGroup, bean2.CHART_GROUP_ENTITY, bean2.EmptyAccessType):    (*UserResourceExtendedServiceImpl).enforceRbacForChartGroup,
	getUserResourceKindWithEntityAccessKey(bean.KindJobs, bean2.EntityJobs, bean2.EmptyAccessType):                  (*UserResourceExtendedServiceImpl).enforceRbacForJobs,
	getUserResourceKindWithEntityAccessKey(bean.KindWorkflow, bean2.EntityJobs, bean2.EmptyAccessType):              (*UserResourceExtendedServiceImpl).enforceRbacForJobsWfs,
}

func getResourceOptionRbacExtendedFunc(kind bean.UserResourceKind, entity string, accessType string) func(impl *UserResourceExtendedServiceImpl, token string, params *apiBean.PathParams, resourceOptions *bean.ResourceOptionsDto) (*bean.UserResourceResponseDto, error) {
	if f, ok := mapOfKindWithEntityAccessTypeKeyToResourceOptionRbacExtendedFunc[getUserResourceKindWithEntityAccessKey(kind, entity, accessType)]; ok {
		return f
	}
	return nil
}
