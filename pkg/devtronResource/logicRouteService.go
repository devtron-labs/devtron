package devtronResource

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
)

func getApiResourceKindUIComponentFunc(kind, component string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectGetAPIBean) error {
	if f, ok := getApiResourceKindUIComponentFuncMap[getKeyForKindAndUIComponent(kind, component)]; ok {
		return f
	} else {
		return nil
	}
}

func getFuncToSetUserProvidedDataInResourceObject(kind, subKind, version string) func(*DevtronResourceServiceImpl, string, *bean.DevtronResourceObjectBean) (string, error) {
	if f, ok := setUserProvidedDataByKindVersionFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

func getFuncToValidateCreateResourceRequest(kind, subKind, version string) func(*bean.DevtronResourceObjectBean) error {
	if f, ok := validateCreateResourceRequestFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

func getFuncToPopulateDefaultValuesForCreateResourceRequest(kind, subKind, version string) func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean) error {
	if f, ok := populateDefaultValuesForCreateResourceRequestFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

func getFuncToBuildIdentifierForResourceObj(kind, subKind, version string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceObject) (string, error) {
	if f, ok := buildIdentifierForResourceObjFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

func getFuncToHandleResourceObjectUpdateRequest(kind, subKind, version, objectUpdatePath string) func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean,
	*repository.DevtronResourceObject) {
	if f, ok := handleResourceObjectUpdateReqFuncMap[getKeyForKindVersionAndObjectUpdatePath(kind, subKind, version, objectUpdatePath)]; ok {
		return f
	} else {
		return nil
	}
}

var getApiResourceKindUIComponentFuncMap = map[string]func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectGetAPIBean) error{
	getKeyForKindAndUIComponent(bean.DevtronResourceApplication, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceApplication, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceCluster, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceCluster, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceJob, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceJob, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentCatalog):      (*DevtronResourceServiceImpl).updateCatalogDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentConfigStatus): (*DevtronResourceServiceImpl).updateReleaseConfigStatusInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentNote):         (*DevtronResourceServiceImpl).updateReleaseNoteInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentOverview):     (*DevtronResourceServiceImpl).updateReleaseOverviewDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentAll):          (*DevtronResourceServiceImpl).updateCompleteReleaseDataInResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceReleaseTrack, bean.UIComponentOverview): (*DevtronResourceServiceImpl).updateReleaseTrackOverviewDataInResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceReleaseTrack, bean.UIComponentAll):      (*DevtronResourceServiceImpl).updateReleaseTrackOverviewDataInResourceObj,
}

var setUserProvidedDataByKindVersionFuncMap = map[string]func(*DevtronResourceServiceImpl, string, *bean.DevtronResourceObjectBean) (string, error){
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      (*DevtronResourceServiceImpl).updateUserProvidedDataInReleaseObj,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).updateUserProvidedDataInReleaseTrackObj,
}

var validateCreateResourceRequestFuncMap = map[string]func(*bean.DevtronResourceObjectBean) error{
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      validateCreateReleaseRequest,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): validateCreateReleaseTrackRequest,
}

var populateDefaultValuesForCreateResourceRequestFuncMap = map[string]func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean) error{
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      (*DevtronResourceServiceImpl).populateDefaultValuesForCreateReleaseRequest,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).populateDefaultValuesForCreateReleaseTrackRequest,
}

var buildIdentifierForResourceObjFuncMap = map[string]func(*DevtronResourceServiceImpl, *repository.DevtronResourceObject) (string, error){
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceDevtronApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForDevtronAppResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierFormHelmAppResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceCluster, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForClusterResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceJob, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForDevtronJobResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceCdPipeline, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForCdPipelineResourceObj,

	getKeyForKindAndVersion(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForReleaseResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForReleaseTrackResourceObj,
}

var handleResourceObjectUpdateReqFuncMap = map[string]func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean, *repository.DevtronResourceObject){
	getKeyForKindVersionAndObjectUpdatePath(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1, bean.ResourceObjectDependenciesPath): (*DevtronResourceServiceImpl).handleReleaseDependencyUpdateRequest,
}

func getKeyForKindAndUIComponent[K, C any](kind K, component C) string {
	return fmt.Sprintf("%s-%s", kind, component)
}

func getKeyForKindAndVersion[K, S, V ~string](kind K, subKind S, version V) string {
	return fmt.Sprintf("%s-%s-%s", kind, subKind, version)
}

func getKeyForKindVersionAndObjectUpdatePath[K, S, V, P ~string](kind K, subKind S, version V, objectUpdatePath P) string {
	return fmt.Sprintf("%s-%s-%s-%s", kind, subKind, version, objectUpdatePath)
}

func listApiResourceKindFunc(kind string) func(*DevtronResourceServiceImpl, []*repository.DevtronResourceObject,
	[]*repository.DevtronResourceObject, map[int][]int, bool) ([]*bean.DevtronResourceObjectGetAPIBean, error) {
	if f, ok := listApiResourceKindFuncMap[kind]; ok {
		return f
	} else {
		return nil
	}
}

var listApiResourceKindFuncMap = map[string]func(impl *DevtronResourceServiceImpl, objects []*repository.DevtronResourceObject,
	childObjects []*repository.DevtronResourceObject, resourceObjectIndexChildMap map[int][]int, isLite bool) ([]*bean.DevtronResourceObjectGetAPIBean, error){
	bean.DevtronResourceReleaseTrack.ToString(): (*DevtronResourceServiceImpl).listReleaseTracks,
}
