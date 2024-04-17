package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetIdentifierForRelease(req *bean.DevtronResourceObjectBean) string {
	return fmt.Sprintf("%s-%s", req.ParentConfig.Data.Name, req.Overview.ReleaseVersion)

}

func GetIdentifierForReleaseTrack(req *bean.DevtronResourceObjectBean) string {
	return req.Name

}

func GetDefaultReleaseNameIfNotProvided(reqBean *bean.DevtronResourceObjectBean) string {
	// The default value of name for release resource -> {releaseVersion}
	return reqBean.Overview.ReleaseVersion
}
