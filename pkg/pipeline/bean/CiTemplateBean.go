package bean

import "github.com/devtron-labs/devtron/internals/sql/repository/pipelineConfig"

type CiTemplateBean struct {
	CiTemplate         *pipelineConfig.CiTemplate
	CiTemplateOverride *pipelineConfig.CiTemplateOverride
	CiBuildConfig      *CiBuildConfigBean
	UserId             int32
}
