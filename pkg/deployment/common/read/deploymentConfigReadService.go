package read

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"go.uber.org/zap"
)

type DeploymentConfigReadService interface {
	GetByAppIdAndEnvId(appId, envId int) (*bean.DeploymentConfig, error)
}

type DeploymentConfigReadServiceImpl struct {
	deploymentConfigRepository deploymentConfig.Repository
	logger                     *zap.SugaredLogger
}
