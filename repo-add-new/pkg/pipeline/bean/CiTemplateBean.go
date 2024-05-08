package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
)

// todo move to proper place
type CiTemplateBean struct {
	CiTemplate         *pipelineConfig.CiTemplate
	CiTemplateOverride *pipelineConfig.CiTemplateOverride
	CiBuildConfig      *CiPipeline.CiBuildConfigBean
	UserId             int32
}
