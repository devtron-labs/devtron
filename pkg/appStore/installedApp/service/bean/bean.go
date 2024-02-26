package bean

import (
	"github.com/devtron-labs/devtron/internal/util"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
)

type AppStoreManifestResponse struct {
	ChartResponse      *util.ChartCreateResponse
	ValuesConfig       *git.ChartConfig
	RequirementsConfig *git.ChartConfig
}

type AppStoreGitOpsResponse struct {
	ChartGitAttribute *commonBean.ChartGitAttribute
	GitHash           string
}
