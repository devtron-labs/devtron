package chartRepo

import (
	"context"
	repository3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ChartRepositoryServiceExtendedImpl struct {
	argoRepositoryService repository2.ServiceClient
	argoUserService       argo.ArgoUserService
	*ChartRepositoryServiceImpl
}

func NewChartRepositoryServiceExtendedImpl(repositoryService repository2.ServiceClient, logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, K8sUtil *util.K8sUtil, clusterService cluster.ClusterService,
	aCDAuthConfig *util2.ACDAuthConfig, client *http.Client, serverEnvConfig *serverEnvConfig.ServerEnvConfig, argoUserService argo.ArgoUserService) *ChartRepositoryServiceExtendedImpl {

	return &ChartRepositoryServiceExtendedImpl{
		ChartRepositoryServiceImpl: &ChartRepositoryServiceImpl{
			logger:          logger,
			repoRepository:  repoRepository,
			K8sUtil:         K8sUtil,
			clusterService:  clusterService,
			aCDAuthConfig:   aCDAuthConfig,
			client:          client,
			serverEnvConfig: serverEnvConfig,
		},
		argoRepositoryService: repositoryService,
		argoUserService:       argoUserService,
	}

}

func (impl *ChartRepositoryServiceExtendedImpl) ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation) {

	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}

	chartRepo, err := impl.CreateChartRepo(request)
	if err != nil {
		return nil, err, nil
	}

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually ", "err", err)
	}
	return chartRepo, err, nil
}

func (impl *ChartRepositoryServiceExtendedImpl) ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}
	//validationResult := &DetailedErrorHelmRepoValidation{}
	chartRepo, err := impl.UpdateData(request)
	if err != nil {
		return nil, err, validationResult
	}
	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}
	return chartRepo, nil, validationResult
}

func (impl *ChartRepositoryServiceExtendedImpl) CreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {
	chartRepo, err := impl.ChartRepositoryServiceImpl.CreateChartRepo(request)
	if err != nil {
		impl.logger.Errorw("error in saving chart repo", "err", err)
	}
	if len(request.UserName) >= 0 && len(request.Password) >= 0 {
		err = impl.RegisterHelmRepoInArgo(request)
	}
	if err != nil {
		impl.logger.Errorw("error in registering helm repo in argocd", "err", err)
		return nil, err
	}
	return chartRepo, nil
}

func (impl *ChartRepositoryServiceExtendedImpl) UpdateData(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {

	chartRepo, err := impl.repoRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in finding repository by id", "err", err)
		return nil, err
	}

	repoURL := chartRepo.Url

	chartRepo, err = impl.ChartRepositoryServiceImpl.UpdateData(request)
	if err != nil {
		impl.logger.Errorw("Error in updating chart repo in database")
		return nil, err
	}
	if len(request.UserName) >= 0 && len(request.Password) >= 0 {
		err = impl.UpdateHelmRepoInArgo(repoURL, request)
	}
	if err != nil {
		impl.logger.Errorw("Error in updating repo in argo", "err", err)
		return nil, err
	}
	return chartRepo, err
}

func (impl *ChartRepositoryServiceExtendedImpl) RegisterHelmRepoInArgo(request *ChartRepoDto) error {
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return err
	}
	ctx := context.WithValue(context.Background(), "token", acdToken)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	repoAddRequestDTO := &v1alpha1.Repository{
		Name:     request.Name,
		Repo:     request.Url,
		Username: request.UserName,
		Password: request.Password,
		Insecure: request.AllowInsecureConnection,
		Type:     HELM,
	}
	_, err = impl.argoRepositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: repoAddRequestDTO, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
		return err
	}
	impl.ChartRepositoryServiceImpl.logger.Debugw("repo registered in argo", "name", request.Url)
	return err
}

func (impl *ChartRepositoryServiceExtendedImpl) UpdateHelmRepoInArgo(repoURL string, request *ChartRepoDto) error {
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return err
	}

	ctx := context.WithValue(context.Background(), "token", acdToken)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	repoDeleteRequestDTO := repository3.RepoQuery{Repo: repoURL}

	_, err = impl.argoRepositoryService.Delete(ctx, &repoDeleteRequestDTO)
	if err != nil {
		impl.logger.Errorw("error in updating repo ")
	}

	_, err = impl.CreateChartRepo(request)
	impl.logger.Debugw("Helm Repo updated in argo")
	return nil
}

func (impl *ChartRepositoryServiceExtendedImpl) DeleteChartRepo(request *ChartRepoDto) error {

	chartRepo, err := impl.repoRepository.FindById(request.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in finding chart repo by id", "err", err, "id", request.Id)
		return err
	}

	err = impl.ChartRepositoryServiceImpl.DeleteChartRepo(request)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repo from db")
		return err
	}
	repoDeleteRequestDTO := repository3.RepoQuery{Repo: chartRepo.Url}
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return err
	}
	ctx := context.WithValue(context.Background(), "token", acdToken)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	_, err = impl.argoRepositoryService.Delete(ctx, &repoDeleteRequestDTO)
	if err != nil {
		impl.logger.Errorw("error in updating repo ")
	}
	return nil
}
