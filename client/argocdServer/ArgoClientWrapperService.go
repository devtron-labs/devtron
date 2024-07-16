/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package argocdServer

import (
	"context"
	"encoding/json"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/devtron/client/argocdServer/adapter"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/util/retryFunc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"time"
)

type ACDConfig struct {
	ArgoCDAutoSyncEnabled     bool `env:"ARGO_AUTO_SYNC_ENABLED" envDefault:"true"` // will gradually switch this flag to false in enterprise
	RegisterRepoMaxRetryCount int  `env:"ARGO_REPO_REGISTER_RETRY_COUNT" envDefault:"3"`
	RegisterRepoMaxRetryDelay int  `env:"ARGO_REPO_REGISTER_RETRY_DELAY" envDefault:"10"`
}

func (config *ACDConfig) IsManualSyncEnabled() bool {
	return config.ArgoCDAutoSyncEnabled == false
}

func (config *ACDConfig) IsAutoSyncEnabled() bool {
	return config.ArgoCDAutoSyncEnabled == true
}

func GetACDDeploymentConfig() (*ACDConfig, error) {
	cfg := &ACDConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

const (
	ErrorOperationAlreadyInProgress = "another operation is already in progress" // this string is returned from argocd
)

type ArgoClientWrapperService interface {

	// GetArgoAppWithNormalRefresh - refresh app at argocd side
	GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error

	// SyncArgoCDApplicationIfNeededAndRefresh - if ARGO_AUTO_SYNC_ENABLED=true, app will be refreshed to initiate refresh at argoCD side or else it will be synced and refreshed
	SyncArgoCDApplicationIfNeededAndRefresh(context context.Context, argoAppName string) error

	// UpdateArgoCDSyncModeIfNeeded - if ARGO_AUTO_SYNC_ENABLED=true and app is in manual sync mode or vice versa update app
	UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error)

	// RegisterGitOpsRepoInArgoWithRetry - register a repository in argo-cd with retry mechanism
	RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl string, userId int32) error

	// GetArgoAppByName fetches an argoCd app by its name
	GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error)

	// PatchArgoCdApp performs a patch operation on an argoCd app
	PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error

	// IsArgoAppPatchRequired decides weather the v1alpha1.ApplicationSource requires to be updated
	IsArgoAppPatchRequired(argoAppSpec *v1alpha1.ApplicationSource, currentGitRepoUrl, currentChartPath string) bool

	// GetGitOpsRepoName returns the GitOps repository name, configured for the argoCd app
	GetGitOpsRepoName(ctx context.Context, appName string) (gitOpsRepoName string, err error)

	GetGitOpsRepoURL(ctx context.Context, appName string) (gitOpsRepoURL string, err error)
}

type ArgoClientWrapperServiceImpl struct {
	logger                  *zap.SugaredLogger
	acdClient               application.ServiceClient
	ACDConfig               *ACDConfig
	repositoryService       repository.ServiceClient
	gitOpsConfigReadService config.GitOpsConfigReadService
	gitOperationService     git.GitOperationService
	asyncRunnable           *async.Runnable
}

func NewArgoClientWrapperServiceImpl(logger *zap.SugaredLogger, acdClient application.ServiceClient,
	ACDConfig *ACDConfig, repositoryService repository.ServiceClient, gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOperationService git.GitOperationService, asyncRunnable *async.Runnable) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		logger:                  logger,
		acdClient:               acdClient,
		ACDConfig:               ACDConfig,
		repositoryService:       repositoryService,
		gitOpsConfigReadService: gitOpsConfigReadService,
		gitOperationService:     gitOperationService,
		asyncRunnable:           asyncRunnable,
	}
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppWithNormalRefresh(ctx context.Context, argoAppName string) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ArgoClientWrapperServiceImpl.GetArgoAppWithNormalRefresh")
	defer span.End()
	refreshType := bean.RefreshTypeNormal
	impl.logger.Debugw("trying to normal refresh application through get ", "argoAppName", argoAppName)
	_, err := impl.acdClient.Get(newCtx, &application2.ApplicationQuery{Name: &argoAppName, Refresh: &refreshType})
	if err != nil {
		internalMsg := fmt.Sprintf("%s, err:- %s", constants.CannotGetAppWithRefreshErrMsg, err.Error())
		clientCode, _ := util.GetClientDetailedError(err)
		httpStatusCode := clientCode.GetHttpStatusCodeForGivenGrpcCode()
		err = &util.ApiError{HttpStatusCode: httpStatusCode, Code: strconv.Itoa(httpStatusCode), InternalMessage: internalMsg, UserMessage: err.Error()}
		impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
		return err
	}
	impl.logger.Debugw("done getting the application with refresh with no error", "argoAppName", argoAppName)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) SyncArgoCDApplicationIfNeededAndRefresh(ctx context.Context, argoAppName string) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ArgoClientWrapperServiceImpl.SyncArgoCDApplicationIfNeededAndRefresh")
	defer span.End()
	impl.logger.Info("ArgoCd manual sync for app started", "argoAppName", argoAppName)
	if impl.ACDConfig.IsManualSyncEnabled() {

		impl.logger.Debugw("syncing ArgoCd app as manual sync is enabled", "argoAppName", argoAppName)
		revision := "master"
		pruneResources := true
		_, syncErr := impl.acdClient.Sync(newCtx, &application2.ApplicationSyncRequest{Name: &argoAppName,
			Revision: &revision,
			Prune:    &pruneResources,
		})
		if syncErr != nil {
			impl.logger.Errorw("error in syncing argoCD app", "app", argoAppName, "err", syncErr)
			statusCode, msg := util.GetClientDetailedError(syncErr)
			if statusCode.IsFailedPreconditionCode() && msg == ErrorOperationAlreadyInProgress {
				impl.logger.Info("terminating ongoing sync operation and retrying manual sync", "argoAppName", argoAppName)
				_, terminationErr := impl.acdClient.TerminateOperation(newCtx, &application2.OperationTerminateRequest{
					Name: &argoAppName,
				})
				if terminationErr != nil {
					impl.logger.Errorw("error in terminating sync operation")
					return fmt.Errorf("error in terminating existing sync, err: %w", terminationErr)
				}
				_, syncErr = impl.acdClient.Sync(newCtx, &application2.ApplicationSyncRequest{Name: &argoAppName,
					Revision: &revision,
					Prune:    &pruneResources,
					RetryStrategy: &v1alpha1.RetryStrategy{
						Limit: 1,
					},
				})
				if syncErr != nil {
					impl.logger.Errorw("error in syncing argoCD app", "app", argoAppName, "err", syncErr)
					return syncErr
				}
			} else {
				return syncErr
			}
		}
		impl.logger.Infow("ArgoCd sync completed", "argoAppName", argoAppName)
	}

	runnableFunc := func() {
		// running ArgoCd app refresh in asynchronous mode
		refreshErr := impl.GetArgoAppWithNormalRefresh(context.Background(), argoAppName)
		if refreshErr != nil {
			impl.logger.Errorw("error in refreshing argo app", "argoAppName", argoAppName, "err", refreshErr)
		}
	}
	impl.asyncRunnable.Execute(runnableFunc)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error) {
	if impl.isArgoAppSyncModeMigrationNeeded(argoApplication) {
		syncModeUpdateRequest := impl.CreateRequestForArgoCDSyncModeUpdateRequest(argoApplication)
		validate := false
		_, err = impl.acdClient.Update(ctx, &application2.ApplicationUpdateRequest{Application: syncModeUpdateRequest, Validate: &validate})
		if err != nil {
			impl.logger.Errorw("error in creating argo pipeline ", "name", argoApplication.Name, "err", err)
			return err
		}
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) isArgoAppSyncModeMigrationNeeded(argoApplication *v1alpha1.Application) bool {
	if impl.ACDConfig.IsManualSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated != nil {
		return true
	} else if impl.ACDConfig.IsAutoSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated == nil {
		return true
	}
	return false
}

func (impl *ArgoClientWrapperServiceImpl) CreateRequestForArgoCDSyncModeUpdateRequest(argoApplication *v1alpha1.Application) *v1alpha1.Application {
	// set automated field in update request
	var automated *v1alpha1.SyncPolicyAutomated
	if impl.ACDConfig.IsAutoSyncEnabled() {
		automated = &v1alpha1.SyncPolicyAutomated{
			Prune: true,
		}
	}
	return &v1alpha1.Application{
		ObjectMeta: v1.ObjectMeta{
			Name:      argoApplication.Name,
			Namespace: DevtronInstalationNs,
		},
		Spec: v1alpha1.ApplicationSpec{
			Destination: argoApplication.Spec.Destination,
			Source:      argoApplication.Spec.Source,
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated:   automated,
				SyncOptions: argoApplication.Spec.SyncPolicy.SyncOptions,
				Retry:       argoApplication.Spec.SyncPolicy.Retry,
			}}}
}

func (impl *ArgoClientWrapperServiceImpl) RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl string, userId int32) error {
	callback := func() error {
		return impl.createRepoInArgoCd(ctx, gitOpsRepoUrl)
	}
	argoCdErr := retryFunc.Retry(callback,
		impl.isRetryableArgoRepoCreationError,
		impl.ACDConfig.RegisterRepoMaxRetryCount,
		time.Duration(impl.ACDConfig.RegisterRepoMaxRetryDelay)*time.Second,
		impl.logger)
	if argoCdErr != nil {
		impl.logger.Errorw("error in registering GitOps repository", "repoName", gitOpsRepoUrl, "err", argoCdErr)
		return impl.handleArgoRepoCreationError(argoCdErr, ctx, gitOpsRepoUrl, userId)
	}
	impl.logger.Infow("gitOps repo registered in argo", "repoName", gitOpsRepoUrl)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error) {
	argoApplication, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("err in getting argo app by name", "app", appName)
		return nil, err
	}
	return argoApplication, nil
}

func (impl *ArgoClientWrapperServiceImpl) IsArgoAppPatchRequired(argoAppSpec *v1alpha1.ApplicationSource, currentGitRepoUrl, currentChartPath string) bool {
	return (len(currentGitRepoUrl) != 0 && argoAppSpec.RepoURL != currentGitRepoUrl) ||
		argoAppSpec.Path != currentChartPath ||
		argoAppSpec.TargetRevision != bean.TargetRevisionMaster
}

func (impl *ArgoClientWrapperServiceImpl) PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error {
	patchReq := adapter.GetArgoCdPatchReqFromDto(dto)
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.logger.Errorw("error in creating patch", "err", err)
		return err
	}
	reqString := string(reqbyte)
	_, err = impl.acdClient.Patch(ctx, &application2.ApplicationPatchRequest{Patch: &reqString, Name: &dto.ArgoAppName, PatchType: &dto.PatchType})
	if err != nil {
		impl.logger.Errorw("error in patching argo pipeline ", "name", dto.ArgoAppName, "patch", reqString, "err", err)
		return err
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) GetGitOpsRepoName(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	acdApplication, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "acdAppName", appName, "err", err)
		return gitOpsRepoName, err
	}
	// safety checks nil pointers
	if acdApplication != nil && acdApplication.Spec.Source != nil {
		gitOpsRepoUrl := acdApplication.Spec.Source.RepoURL
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
		return gitOpsRepoName, nil
	}
	return gitOpsRepoName, fmt.Errorf("unable to get any ArgoCd application '%s'", appName)
}

func (impl *ArgoClientWrapperServiceImpl) GetGitOpsRepoURL(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	acdApplication, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "acdAppName", appName, "err", err)
		return gitOpsRepoName, err
	}
	// safety checks nil pointers
	if acdApplication != nil && acdApplication.Spec.Source != nil {
		gitOpsRepoUrl := acdApplication.Spec.Source.RepoURL
		return gitOpsRepoUrl, nil
	}
	return "", fmt.Errorf("unable to get any ArgoCd application '%s'", appName)
}

// createRepoInArgoCd is the wrapper function to Create Repository in ArgoCd
func (impl *ArgoClientWrapperServiceImpl) createRepoInArgoCd(ctx context.Context, gitOpsRepoUrl string) error {
	repo := &v1alpha1.Repository{
		Repo: gitOpsRepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository", "err", err)
		return err
	}
	return nil
}

// isRetryableArgoRepoCreationError returns whether to retry or not, based on the error returned from callback func
// In createRepoInArgoCd, we will retry only if the error matches to bean.ArgoRepoSyncDelayErr
func (impl *ArgoClientWrapperServiceImpl) isRetryableArgoRepoCreationError(argoCdErr error) bool {
	return strings.Contains(argoCdErr.Error(), bean.ArgoRepoSyncDelayErr)
}

// handleArgoRepoCreationError - manages the error thrown while performing createRepoInArgoCd
func (impl *ArgoClientWrapperServiceImpl) handleArgoRepoCreationError(argoCdErr error, ctx context.Context, gitOpsRepoUrl string, userId int32) error {
	emptyRepoErrorMessages := bean.EmptyRepoErrorList
	isEmptyRepoError := false
	for _, errMsg := range emptyRepoErrorMessages {
		if strings.Contains(argoCdErr.Error(), errMsg) {
			isEmptyRepoError = true
		}
	}
	if isEmptyRepoError {
		// - found empty repository, create some file in repository
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
		err := impl.gitOperationService.CreateReadmeInGitRepo(ctx, gitOpsRepoName, userId)
		if err != nil {
			impl.logger.Errorw("error in creating file in git repo", "err", err)
			return err
		}
	}
	// try to register with after creating readme file
	return impl.createRepoInArgoCd(ctx, gitOpsRepoUrl)
}
