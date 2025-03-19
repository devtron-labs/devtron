package beHelper

import (
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	util2 "github.com/devtron-labs/devtron/util"
)

func GetCIPipelineName(appId int) string {
	return fmt.Sprintf("ci-%d-%s", appId, util2.Generate(4))
}

func GetCDPipelineName(appId int) string {
	return fmt.Sprintf("cd-%d-%s", appId, util2.Generate(4))
}

func GetAppWorkflowName(appId int) string {
	return fmt.Sprintf("wf-%d-%s", appId, util2.Generate(4))
}

func GetPipelineNameByPipelineType(pipelineType string, appId int) string {
	return fmt.Sprintf("%s-%d-%s", pipelineType, appId, util.Generate(4))
}
