package configDiff

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"net/http"
)

func (impl *DeploymentConfigurationServiceImpl) getSecretDataForDraftOnly(ctx context.Context, appEnvAndClusterMetadata *bean2.AppEnvAndClusterMetadata, userId int32) (*bean2.SecretConfigMetadata, error) {
	return nil, util.GetApiError(http.StatusNotFound, "implementation for draft kind not found", "implementation for draft kind not found")
}

func (impl *DeploymentConfigurationServiceImpl) getSecretDataForPublishedWithDraft(ctx context.Context, appEnvAndClusterMetadata *bean2.AppEnvAndClusterMetadata,
	systemMetadata *resourceQualifiers.SystemMetadata, userId int32) (*bean2.SecretConfigMetadata, error) {
	return nil, util.GetApiError(http.StatusNotFound, "implementation for published with draft kind not found", "implementation for published with draft kind not found")
}
