package deployedApp

import (
	"context"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"go.uber.org/zap"
)

type DeployedAppService interface {
	RotatePods(ctx context.Context, podRotateRequest *bean.PodRotateRequest) (*k8s.RotatePodResponse, error)
}

type DeployedAppServiceImpl struct {
	logger           *zap.SugaredLogger
	k8sCommonService k8s.K8sCommonService
	envRepository    repository.EnvironmentRepository
}

func NewDeployedAppServiceImpl(logger *zap.SugaredLogger,
	k8sCommonService k8s.K8sCommonService,
	envRepository repository.EnvironmentRepository) *DeployedAppServiceImpl {
	return &DeployedAppServiceImpl{
		logger:           logger,
		k8sCommonService: k8sCommonService,
		envRepository:    envRepository,
	}
}

func (impl *DeployedAppServiceImpl) RotatePods(ctx context.Context, podRotateRequest *bean.PodRotateRequest) (*k8s.RotatePodResponse, error) {
	impl.logger.Infow("rotate pod request", "payload", podRotateRequest)
	//extract cluster id and namespace from env id
	environmentId := podRotateRequest.EnvironmentId
	environment, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching env details", "envId", environmentId, "err", err)
		return nil, err
	}
	var resourceIdentifiers []util5.ResourceIdentifier
	for _, resourceIdentifier := range podRotateRequest.ResourceIdentifiers {
		resourceIdentifier.Namespace = environment.Namespace
		resourceIdentifiers = append(resourceIdentifiers, resourceIdentifier)
	}
	rotatePodRequest := &k8s.RotatePodRequest{
		ClusterId: environment.ClusterId,
		Resources: resourceIdentifiers,
	}
	response, err := impl.k8sCommonService.RotatePods(ctx, rotatePodRequest)
	if err != nil {
		return nil, err
	}
	//TODO KB: make entry in cd workflow runner
	return response, nil
}
