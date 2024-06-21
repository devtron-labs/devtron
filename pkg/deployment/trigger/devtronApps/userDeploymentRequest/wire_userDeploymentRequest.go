package userDeploymentRequest

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/service"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	repository.NewUserDeploymentRequestRepositoryImpl,
	wire.Bind(new(repository.UserDeploymentRequestRepository), new(*repository.UserDeploymentRequestRepositoryImpl)),

	service.NewUserDeploymentRequestServiceImpl,
	wire.Bind(new(service.UserDeploymentRequestService), new(*service.UserDeploymentRequestServiceImpl)),
)
