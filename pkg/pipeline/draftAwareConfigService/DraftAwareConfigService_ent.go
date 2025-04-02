package draftAwareConfigService

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"net/http"
)

func (impl *DraftAwareResourceServiceImpl) performExpressEditActionsOnCmCsForExceptionUser(ctx context.Context, name string, configMapRequest *bean.ConfigDataRequest) error {
	return util.NewApiError(http.StatusNotImplemented, "operations not supported in oss", "operations not supported in oss")
}

func (impl *DraftAwareResourceServiceImpl) performExpressEditActionsOnDeplTemplateForExceptionUser(ctx context.Context, appId, envId int, resourceName string) error {
	return util.NewApiError(http.StatusNotImplemented, "operation not supported in oss", "operation not supported in oss")
}
