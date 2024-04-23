package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetDefaultReleaseNameIfNotProvided(reqBean *bean.DevtronResourceObjectBean) string {
	// The default value of name for release resource -> {releaseVersion}
	return reqBean.Overview.ReleaseVersion
}

func GetKeyForADependencyMap(oldObjectId, devtronResourceSchemaId int) string {
	// key can be "oldObjectId-schemaId" or "name-schemaId"
	return fmt.Sprintf("%d-%d", oldObjectId, devtronResourceSchemaId)
}
