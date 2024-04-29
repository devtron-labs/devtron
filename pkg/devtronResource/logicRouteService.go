package devtronResource

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
)

func getFuncForGetApiResourceKindUIComponent(kind, component string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectGetAPIBean) error {
	if f, ok := getApiResourceKindUIComponentFuncMap[getKeyForKindAndUIComponent(kind, component)]; ok {
		return f
	} else {
		return nil
	}
}

var getApiResourceKindUIComponentFuncMap = map[string]func(*DevtronResourceServiceImpl, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectGetAPIBean) error{
	getKeyForKindAndUIComponent(bean.DevtronResourceApplication, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceApplication, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceCluster, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceCluster, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceJob, bean.UIComponentCatalog): (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceJob, bean.UIComponentAll):     (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentCatalog):      (*DevtronResourceServiceImpl).updateCatalogDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentConfigStatus): (*DevtronResourceServiceImpl).updateReleaseConfigStatusForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentNote):         (*DevtronResourceServiceImpl).updateReleaseNoteForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentOverview):     (*DevtronResourceServiceImpl).updateReleaseOverviewDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceRelease, bean.UIComponentAll):          (*DevtronResourceServiceImpl).updateCompleteReleaseDataForGetApiResourceObj,

	getKeyForKindAndUIComponent(bean.DevtronResourceReleaseTrack, bean.UIComponentOverview): (*DevtronResourceServiceImpl).updateReleaseTrackOverviewDataForGetApiResourceObj,
	getKeyForKindAndUIComponent(bean.DevtronResourceReleaseTrack, bean.UIComponentAll):      (*DevtronResourceServiceImpl).updateReleaseTrackOverviewDataForGetApiResourceObj,
}

func getFuncToSetUserProvidedDataInResourceObject(kind, subKind, version string) func(*DevtronResourceServiceImpl, string, *bean.DevtronResourceObjectBean) (string, error) {
	if f, ok := setUserProvidedDataByKindVersionFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var setUserProvidedDataByKindVersionFuncMap = map[string]func(*DevtronResourceServiceImpl, string, *bean.DevtronResourceObjectBean) (string, error){
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      (*DevtronResourceServiceImpl).updateUserProvidedDataInReleaseObj,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).updateUserProvidedDataInReleaseTrackObj,
}

func getFuncToValidateCreateResourceRequest(kind, subKind, version string) func(*bean.DevtronResourceObjectBean) error {
	if f, ok := validateCreateResourceRequestFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var validateCreateResourceRequestFuncMap = map[string]func(*bean.DevtronResourceObjectBean) error{
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      validateCreateReleaseRequest,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): validateCreateReleaseTrackRequest,
}

func getFuncToPopulateDefaultValuesForCreateResourceRequest(kind, subKind, version string) func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean) error {
	if f, ok := populateDefaultValuesForCreateResourceRequestFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var populateDefaultValuesForCreateResourceRequestFuncMap = map[string]func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean) error{
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "", bean.DevtronResourceVersionAlpha1):      (*DevtronResourceServiceImpl).populateDefaultValuesForCreateReleaseRequest,
	getKeyForKindAndVersion(bean.DevtronResourceReleaseTrack, "", bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).populateDefaultValuesForCreateReleaseTrackRequest,
}

func getFuncToBuildIdentifierForResourceObj(kind, subKind, version string) func(*DevtronResourceServiceImpl, *repository.DevtronResourceObject) (string, error) {
	if f, ok := buildIdentifierForResourceObjFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var buildIdentifierForResourceObjFuncMap = map[string]func(*DevtronResourceServiceImpl, *repository.DevtronResourceObject) (string, error){
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceDevtronApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForDevtronAppResourceObj,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).buildIdentifierForHelmAppResourceObj,
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

func getFuncToUpdateResourceDependenciesDataInResponseObj(kind string, subKind string, version string) func(*DevtronResourceServiceImpl,
	*bean.DevtronResourceObjectDescriptorBean, *apiBean.GetDependencyQueryParams, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectBean) (*bean.DevtronResourceObjectBean, error) {
	if f, ok := updateResourceDependenciesDataInResponseObjFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var updateResourceDependenciesDataInResponseObjFuncMap = map[string]func(*DevtronResourceServiceImpl,
	*bean.DevtronResourceObjectDescriptorBean, *apiBean.GetDependencyQueryParams, *repository.DevtronResourceSchema,
	*repository.DevtronResourceObject, *bean.DevtronResourceObjectBean) (*bean.DevtronResourceObjectBean, error){
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceDevtronApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).updateV1ResourceDataForGetDependenciesApi,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).updateV1ResourceDataForGetDependenciesApi,
	getKeyForKindAndVersion(bean.DevtronResourceCluster, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).updateV1ResourceDataForGetDependenciesApi,
	getKeyForKindAndVersion(bean.DevtronResourceJob, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).updateV1ResourceDataForGetDependenciesApi,

	getKeyForKindAndVersion(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).updateReleaseDataForGetDependenciesApi,
}

func getFuncToExtractConditionsFromFilterCriteria(kind, subKind, version string, resource bean.DevtronResourceKind) func(impl *DevtronResourceServiceImpl, filterCriteria *bean.FilterCriteriaDecoder) ([]int, error) {
	if f, ok := extractConditionsFromFilterCriteriaFuncMap[getKeyForKindSubKindVersionResource(kind, subKind, version, resource)]; ok {
		return f
	} else {
		return nil
	}
}

var extractConditionsFromFilterCriteriaFuncMap = map[string]func(impl *DevtronResourceServiceImpl, filterCriteria *bean.FilterCriteriaDecoder) ([]int, error){
	getKeyForKindSubKindVersionResource(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1, bean.DevtronResourceReleaseTrack): (*DevtronResourceServiceImpl).getReleaseTrackIdsFromFilterValueBasedOnType,
}

func getFuncForProcessingFiltersOnResourceObjects(kind, subKind, version string, resource bean.DevtronResourceKind) func(impl *DevtronResourceServiceImpl, resourceObjects []*repository.DevtronResourceObject, releaseTrackIds []int) ([]*repository.DevtronResourceObject, error) {
	if f, ok := getProcessingFiltersFuncMap[getKeyForKindSubKindVersionResource(kind, subKind, version, resource)]; ok {
		return f
	} else {
		return nil
	}
}

var getProcessingFiltersFuncMap = map[string]func(impl *DevtronResourceServiceImpl, resourceObjects []*repository.DevtronResourceObject, releaseTrackIds []int) ([]*repository.DevtronResourceObject, error){
	getKeyForKindSubKindVersionResource(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1, bean.DevtronResourceReleaseTrack): (*DevtronResourceServiceImpl).getFilteredReleaseObjectsForReleaseTrackIds,
}

func getFuncToApplyFilterResourceKind(kind, subKind, version string) func(impl *DevtronResourceServiceImpl, kind, subKind, version string, resourceObjects []*repository.DevtronResourceObject, filterCriteria []string) ([]*repository.DevtronResourceObject, error) {
	if f, ok := applyFilterResourceKindFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var applyFilterResourceKindFuncMap = map[string]func(impl *DevtronResourceServiceImpl, kind, subKind, version string, resourceObjects []*repository.DevtronResourceObject, filterCriteria []string) ([]*repository.DevtronResourceObject, error){
	getKeyForKindAndVersion(bean.DevtronResourceRelease.ToString(), "", bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).applyFilterCriteriaOnReleaseResourceObjects,
}

func getFuncToHandleResourceObjectUpdateRequest(kind, subKind, version, objectUpdatePath string) func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean,
	*repository.DevtronResourceObject) {
	if f, ok := handleResourceObjectUpdateReqFuncMap[getKeyForKindVersionAndObjectUpdatePath(kind, subKind, version, objectUpdatePath)]; ok {
		return f
	} else {
		return nil
	}
}

var handleResourceObjectUpdateReqFuncMap = map[string]func(*DevtronResourceServiceImpl, *bean.DevtronResourceObjectBean, *repository.DevtronResourceObject){
	getKeyForKindVersionAndObjectUpdatePath(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1, bean.ResourceObjectDependenciesPath): (*DevtronResourceServiceImpl).handleReleaseDependencyUpdateRequest,
}

func getFuncToListApiResourceKind(kind string) func(*DevtronResourceServiceImpl, []*repository.DevtronResourceObject,
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
	bean.DevtronResourceRelease.ToString():      (*DevtronResourceServiceImpl).listRelease,
}

func getFuncToGetResourceIdAndIdTypeFromIdentifier(kind string, subKind string, version string) func(*DevtronResourceServiceImpl, *bean.ResourceIdentifier) (int, bean.IdType, error) {
	if f, ok := getResourceIdAndIdTypeFromIdentifierFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var getResourceIdAndIdTypeFromIdentifierFuncMap = map[string]func(*DevtronResourceServiceImpl, *bean.ResourceIdentifier) (int, bean.IdType, error){
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceDevtronApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
	getKeyForKindAndVersion(bean.DevtronResourceCluster, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
	getKeyForKindAndVersion(bean.DevtronResourceJob, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
	getKeyForKindAndVersion(bean.DevtronResourceCdPipeline, "",
		bean.DevtronResourceVersion1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1): (*DevtronResourceServiceImpl).getIdAndIdTypeFromIdentifierForDevtronApps,
}

func getFuncToUpdateMetadataInDependency(kind string, subKind string, version string) func(int, *bean.DependencyMetaDataBean) interface{} {
	if f, ok := updateMetadataInDependencyFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var updateMetadataInDependencyFuncMap = map[string]func(int, *bean.DependencyMetaDataBean) interface{}{
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceDevtronApplication,
		bean.DevtronResourceVersion1): updateAppMetaDataInDependencyObj,
	getKeyForKindAndVersion(bean.DevtronResourceApplication, bean.DevtronResourceHelmApplication,
		bean.DevtronResourceVersion1): updateAppMetaDataInDependencyObj,
	getKeyForKindAndVersion(bean.DevtronResourceCdPipeline, "",
		bean.DevtronResourceVersion1): updateCdPipelineMetaDataInDependencyObj,
}

func getFuncToUpdateDependencyConfigData(kind string, subKind string, version string) func(string, *bean.DependencyConfigBean, bool) error {
	if f, ok := updateDependencyConfigDataFuncMap[getKeyForKindAndVersion(kind, subKind, version)]; ok {
		return f
	} else {
		return nil
	}
}

var updateDependencyConfigDataFuncMap = map[string]func(string, *bean.DependencyConfigBean, bool) error{
	getKeyForKindAndVersion(bean.DevtronResourceRelease, "",
		bean.DevtronResourceVersionAlpha1): updateReleaseDependencyConfigDataInObj,
}

func getKeyForKindAndUIComponent[K, C any](kind K, component C) string {
	return fmt.Sprintf("%s-%s", kind, component)
}

func getKeyForKindAndVersion[K, S, V ~string](kind K, subKind S, version V) string {
	return fmt.Sprintf("%s-%s-%s", kind, subKind, version)
}

func getKeyForKindSubKindVersionResource[K, S, C ~string](kind K, subKind S, version C, resource bean.DevtronResourceKind) string {
	return fmt.Sprintf("%s-%s-%s-%s", kind, subKind, version, resource)
}

func getKeyForKindVersionAndObjectUpdatePath[K, S, V, P ~string](kind K, subKind S, version V, objectUpdatePath P) string {
	return fmt.Sprintf("%s-%s-%s-%s", kind, subKind, version, objectUpdatePath)
}
