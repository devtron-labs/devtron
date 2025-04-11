package pipelineConfig

import "github.com/devtron-labs/devtron/pkg/bean/common"

func (p *CiPipeline) GetWorkflowCacheConfig() common.WorkflowCacheConfigType {
	//in oss, there is no pipeline level workflow cache config, so we pass inherit to get the app level config
	return common.WorkflowCacheConfigInherit
}
