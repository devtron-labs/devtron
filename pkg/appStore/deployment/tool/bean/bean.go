package bean

import (
	"github.com/devtron-labs/devtron/internal/util"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
)

type AppStoreManifestResponse struct {
	ChartResponse      *util.ChartCreateResponse
	ValuesConfig       *util.ChartConfig
	RequirementsConfig *util.ChartConfig
}

type AppStoreGitOpsResponse struct {
	ChartGitAttribute *commonBean.ChartGitAttribute
	GitHash           string
}
