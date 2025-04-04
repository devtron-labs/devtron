package service

import (
	"context"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bulkAction/bean"
)

func (impl BulkUpdateServiceImpl) BulkHibernateV1(ctx context.Context, request *bean.BulkApplicationForEnvironmentPayload, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool,
	userMetadata *bean2.UserMetadata) (*bean.BulkApplicationHibernateUnhibernateForEnvironmentResponse, error) {
	return nil, nil
}
