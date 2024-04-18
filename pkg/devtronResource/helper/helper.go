package helper

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetDefaultReleaseNameIfNotProvided(reqBean *bean.DevtronResourceObjectBean) string {
	// The default value of name for release resource -> {releaseVersion}
	return reqBean.Overview.ReleaseVersion
}
