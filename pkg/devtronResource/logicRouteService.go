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

func getFuncToBuildIdentifierForResourceObj(kind string, subKind string, version string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceObject) (string, error) {
	if f, ok := buildIdentifierForResourceObjFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
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
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForDevtronAppResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierFormHelmAppResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceCluster, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForClusterResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceJob, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForDevtronJobResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceCdPipeline, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForCdPipelineResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForReleaseResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).buildIdentifierForReleaseTrackResourceObj,
}

func getKeyForKindAndUIComponent[K, C any](kind K, component C) string {
	return fmt.Sprintf("%s-%s", kind, component)
}

func getKeyForKindAndVersion[K, S, C ~string](kind K, subKind S, version C) string {
	return fmt.Sprintf("%s-%s-%s", kind, subKind, version)
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
