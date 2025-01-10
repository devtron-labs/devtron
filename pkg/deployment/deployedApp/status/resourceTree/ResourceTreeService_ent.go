package resourceTree

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func (impl *ServiceImpl) FetchResourceTreeWithDrift(ctx context.Context, appId int, envId int, cdPipeline *pipelineConfig.Pipeline,
	deploymentConfig *commonBean.DeploymentConfig) (map[string]interface{}, error) {
	return nil, nil
}
