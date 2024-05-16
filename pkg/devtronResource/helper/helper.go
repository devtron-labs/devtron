package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"net/http"
	"strconv"
	"strings"
)

func GetKindAndSubKindFrom(resourceKindVar string) (kind, subKind string, err error) {
	kindSplits := strings.Split(resourceKindVar, "/")
	if len(kindSplits) == 1 {
		kind = kindSplits[0]
	} else if len(kindSplits) == 2 {
		kind = kindSplits[0]
		subKind = kindSplits[1]
	} else {
		return kind, subKind, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
	}
	return kind, subKind, nil
}

func BuildExtendedResourceKindUsingKindAndSubKind(kind, subKind string) bean.DevtronResourceKind {
	if len(kind) != 0 && len(subKind) != 0 {
		return bean.DevtronResourceKind(fmt.Sprintf("%s/%s", kind, subKind))
	}
	return bean.DevtronResourceKind(kind)
}

func GetDefaultReleaseNameIfNotProvided(reqBean *bean.DtResourceObjectCreateReqBean) string {
	// The default value of name for release resource -> {releaseVersion}
	return reqBean.Overview.ReleaseVersion
}

func GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId int,
	devtronResourcesSchemaMapById map[int]*repository.DevtronResourceSchema,
	devtronResourcesMapById map[int]*repository.DevtronResource) (string, string, string, error) {
	devtronResourceSchema := devtronResourcesSchemaMapById[devtronResourceSchemaId]
	if devtronResourceSchema == nil {
		return "", "", "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceSchemaId, bean.InvalidResourceSchemaId)
	}
	kind, subKind := GetKindSubKindOfResourceBySchemaObject(devtronResourceSchema, devtronResourcesMapById)
	return kind, subKind, devtronResourceSchema.Version, nil
}

func GetKindSubKindOfResourceBySchemaObject(devtronResourceSchema *repository.DevtronResourceSchema,
	devtronResourcesMapById map[int]*repository.DevtronResource) (string, string) {
	kind, subKind := "", ""
	if devtronResourceSchema != nil {
		devtronResource := devtronResourceSchema.DevtronResource
		return GetKindSubKindOfResource(&devtronResource, devtronResourcesMapById)
	}
	return kind, subKind
}

func GetKindSubKindOfResource(devtronResource *repository.DevtronResource, devtronResourcesMapById map[int]*repository.DevtronResource) (string, string) {
	kind, subKind := "", ""
	if devtronResource != nil {
		if devtronResource.ParentKindId > 0 {
			devtronParentResource := devtronResourcesMapById[devtronResource.ParentKindId]
			if devtronParentResource != nil {
				kind = devtronParentResource.Kind
				subKind = devtronResource.Kind
			}
		} else {
			kind = devtronResource.Kind
		}
	}
	return kind, subKind
}

func GetFilterKeyObjectFromId(devtronResourceSchemaId int, oldObjectId string) bean.FilterKeyObject {
	// key can be "schemaId/oldObjectId" or "schemaId/*"
	return fmt.Sprintf("%d/%s", devtronResourceSchemaId, oldObjectId)
}

func GetResourceObjectIdAndType(existingResourceObject *repository.DevtronResourceObject) (objectId int, idType bean.IdType) {
	idType = bean.IdType(gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectIdTypePath).String())
	if idType == bean.ResourceObjectIdType {
		objectId = existingResourceObject.Id
	} else if idType == bean.OldObjectId {
		objectId = existingResourceObject.OldObjectId
	}
	return objectId, idType
}

func GetKeyForPipelineIdAndAppId(pipelineId, appId int) string {
	return fmt.Sprintf("%v-%v", pipelineId, appId)
}

func GetAppIdFromPipelineIdAppIdKey(key string) (int, error) {
	objs := strings.Split(key, "-")
	if len(objs) != 2 {
		return 0, util.GetApiErrorAdapter(http.StatusInternalServerError, "500", "Not able to process", "Invalid pipeline and app id key")
	}
	return strconv.Atoi(objs[1])
}
