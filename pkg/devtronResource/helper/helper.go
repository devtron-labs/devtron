package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
)

func GetDefaultReleaseNameIfNotProvided(reqBean *bean.DevtronResourceObjectBean) string {
	// The default value of name for release resource -> {releaseVersion}
	return reqBean.Overview.ReleaseVersion
}

func GetKeyForADependencyMap(oldObjectId, devtronResourceSchemaId int) string {
	// key can be "oldObjectId-schemaId" or "name-schemaId"
	return fmt.Sprintf("%d-%d", oldObjectId, devtronResourceSchemaId)
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
