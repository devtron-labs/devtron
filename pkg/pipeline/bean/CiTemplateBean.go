package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
)

type CiTemplateBean struct {
	CiTemplate         *pipelineConfig.CiTemplate
	CiTemplateOverride *pipelineConfig.CiTemplateOverride
	CiBuildConfig      *CiBuildConfigBean
	UserId             int32
}

type CiTemplateMetadata struct {
	DockerfilePath   string
	DockerRepository string
	DockerRegistry   *types.DockerArtifactStoreBean
	CheckoutPath     string
	CiBuildConfig    *CiBuildConfigBean
}
