package bean

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

// todo move to proper place
type CiTemplateBean struct {
	CiTemplate         *pipelineConfig.CiTemplate
	CiTemplateOverride *pipelineConfig.CiTemplateOverride
	CiBuildConfig      *CiBuildConfigBean
	UserId             int32
}
