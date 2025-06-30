package service

import (
	"context"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bulkAction/bean"
)

type BulkUpdateServiceEnt interface {
}

func (impl BulkUpdateServiceImpl) isEnvDTConfigUpdateAllowed(ctx context.Context, appId, envId int, userMetadata *userBean.UserMetadata) error {
	return nil
}

func (impl BulkUpdateServiceImpl) isBaseDTConfigUpdateAllowed(ctx context.Context, appId int, userMetadata *userBean.UserMetadata) error {
	return nil
}

func (impl BulkUpdateServiceImpl) BulkHibernateV1(ctx context.Context, request *bean.BulkApplicationForEnvironmentPayload, checkAuthForBulkActions func(token string, appObject string, envObject string) bool,
	userMetadata *userBean.UserMetadata) (*bean.BulkApplicationHibernateUnhibernateForEnvironmentResponse, error) {
	return nil, nil
}
