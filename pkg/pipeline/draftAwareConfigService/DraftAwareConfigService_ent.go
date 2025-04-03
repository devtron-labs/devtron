package draftAwareConfigService

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"net/http"
)

func (impl *DraftAwareConfigServiceImpl) performExpressEditActionsOnCmCsForExceptionUser(ctx context.Context, configProperty *bean.ConfigNameAndType, configMapRequest *bean.ConfigDataRequest, isSuperAdmin bool, userEmail string) error {
	return util.NewApiError(http.StatusNotImplemented, "operations not supported in oss", "operations not supported in oss")
}

func (impl *DraftAwareConfigServiceImpl) performExpressEditActionsOnDeplTemplateForExceptionUser(ctx context.Context, deplConfigMetadata *bean.DeploymentConfigMetadata, isExpressEdit bool, isSuperAdmin bool, userEmail string) error {
	return util.NewApiError(http.StatusNotImplemented, "operation not supported in oss", "operation not supported in oss")
}
