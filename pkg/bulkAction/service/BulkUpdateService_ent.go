package service

import (
	"context"
	"github.com/devtron-labs/devtron/pkg/bulkAction/bean"
	"net/http"
)

func (impl BulkUpdateServiceImpl) BulkHibernateV1(request *bean.BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*bean.BulkApplicationHibernateUnhibernateForEnvironmentResponse, error) {
	return nil, nil
}
