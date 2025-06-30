/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fluxcd

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type DeploymentService interface {
	DeployFluxCdApp(ctx context.Context, request *DeploymentRequest) error
	DeleteFluxDeploymentApp(ctx context.Context, request *DeploymentAppDeleteRequest) error
}

type DeploymentServiceImpl struct {
	logger                  *zap.SugaredLogger
	K8sUtil                 *k8s.K8sServiceImpl
	gitOpsConfigReadService config.GitOpsConfigReadService
}

func NewDeploymentService(logger *zap.SugaredLogger,
	K8sUtil *k8s.K8sServiceImpl,
	gitOpsConfigReadService config.GitOpsConfigReadService) *DeploymentServiceImpl {
	return &DeploymentServiceImpl{
		logger:                  logger,
		K8sUtil:                 K8sUtil,
		gitOpsConfigReadService: gitOpsConfigReadService,
	}
}

const (
	gitRepositoryReconcileInterval = 1 * time.Minute
	helmReleaseReconcileInterval   = 1 * time.Minute
)

type DeploymentRequest struct {
	ClusterConfig    *k8s.ClusterConfig
	DeploymentConfig *bean.DeploymentConfig
	IsAppCreated     bool
}

type DeploymentAppDeleteRequest struct {
	ClusterConfig    *k8s.ClusterConfig
	DeploymentConfig *bean.DeploymentConfig
}

func GetValuesFileArrForDevtronInlineApps(chartLocation string) []string {
	//order matters here, last file will override previous file
	//for external flux apps this array might have some other data and we will add our devtronValueFileName (format: _{envId}-values.yaml) along with this array
	return []string{path.Join(chartLocation, "values.yaml")}
}

func (impl *DeploymentServiceImpl) DeployFluxCdApp(ctx context.Context, request *DeploymentRequest) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.deployFluxCdApp")
	defer span.End()
	fluxCdSpec := request.DeploymentConfig.ReleaseConfiguration.FluxCDSpec
	clusterId := request.ClusterConfig.ClusterId
	gitOpsSecret, err := impl.upsertGitRepoSecret(newCtx, fluxCdSpec, request.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in creating git repo secret", "clusterId", clusterId, "err", err)
		return err
	}
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(request.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "clusterId", clusterId, "err", err)
		return err
	}
	apiClient, err := getClient(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating k8s client", "clusterId", clusterId, "err", err)
		return err
	}
	//create/update gitOps secret
	if !request.IsAppCreated {
		err := impl.createFluxCdApp(newCtx, fluxCdSpec, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in creating flux-cd application", "err", err)
			return err
		}
	} else {
		err := impl.updateFluxCdApp(newCtx, fluxCdSpec, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in updating flux-cd application", "err", err)
			return err
		}
	}
	return nil
}

func (impl *DeploymentServiceImpl) upsertGitRepoSecret(ctx context.Context, fluxCdSpec bean.FluxCDSpec, clusterConfig *k8s.ClusterConfig) (*v1.Secret, error) {
	gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsProviderByRepoURL(fluxCdSpec.RepoUrl)
	if err != nil {
		impl.logger.Errorw("error fetching gitops config by repo url", "repoUrl", fluxCdSpec.RepoUrl, "err", err)
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(gitOpsConfig.Username),
		"password": []byte(gitOpsConfig.Token),
	}

	labels := map[string]string{
		"managed-by": "devtron",
		"providerId": gitOpsConfig.Provider,
	}

	namespace := fluxCdSpec.GitRepositoryNamespace
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fluxCdSpec.GitOpsSecretName,
			Namespace: fluxCdSpec.GitRepositoryNamespace,
			Labels:    labels,
		},
		Type: v1.SecretTypeBasicAuth,
		Data: data,
	}

	coreV1Client, err := impl.K8sUtil.GetCoreV1Client(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting core v1 client", "clusterId", clusterConfig.ClusterId, "err", err)
		return nil, err
	}
	// Try to create the secret first
	createdSecret, err := coreV1Client.Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err == nil {
		return createdSecret, nil
	}
	if !k8serrors.IsAlreadyExists(err) {
		impl.logger.Errorw("error creating secret", "namespace", namespace, "name", secret.GetName(), "err", err)
		return nil, err
	}

	// Secret already exists, get and update
	existingSecret, err := coreV1Client.Secrets(namespace).Get(ctx, secret.GetName(), metav1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error getting existing secret", "namespace", namespace, "name", secret.GetName(), "err", err)
		return nil, err
	}

	// Update data and labels
	existingSecret.Data = data
	if existingSecret.Labels == nil {
		existingSecret.Labels = make(map[string]string)
	}
	for k, v := range labels {
		existingSecret.Labels[k] = v
	}

	updatedSecret, err := coreV1Client.Secrets(namespace).Update(ctx, existingSecret, metav1.UpdateOptions{})
	if err != nil {
		impl.logger.Errorw("error updating secret", "namespace", namespace, "name", secret.GetName(), "err", err)
		return nil, err
	}

	return updatedSecret, nil
}

func (impl *DeploymentServiceImpl) createFluxCdApp(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.createFluxCdApp")
	defer span.End()
	_, err := impl.CreateGitRepository(newCtx, fluxCdSpec, gitOpsSecretName, apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating git repository", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}

	_, err = impl.CreateHelmRelease(newCtx, fluxCdSpec, apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating helm release", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}
	return nil
}

func (impl *DeploymentServiceImpl) updateFluxCdApp(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.updateFluxCdApp")
	defer span.End()
	_, err := impl.UpdateGitRepository(newCtx, fluxCdSpec, gitOpsSecretName, apiClient)
	if err != nil {
		impl.logger.Errorw("error in updating git repository", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}

	_, err = impl.UpdateHelmRelease(newCtx, fluxCdSpec, apiClient)
	if err != nil {
		impl.logger.Errorw("error in updating helm release", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}
	return nil
}

func (impl *DeploymentServiceImpl) CreateGitRepository(ctx context.Context, fluxCdSpec bean.FluxCDSpec, secretName string, apiClient client.Client) (*sourcev1.GitRepository, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.GitRepositoryNamespace
	gitRepo := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{Duration: gitRepositoryReconcileInterval},
			URL:      fluxCdSpec.RepoUrl,
			SecretRef: &meta.LocalObjectReference{
				Name: secretName,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: fluxCdSpec.RevisionTarget,
			},
		},
	}
	err := apiClient.Create(ctx, gitRepo)
	if err != nil {
		impl.logger.Errorw("error creating GitRepository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	return gitRepo, nil
}

func (impl *DeploymentServiceImpl) UpdateGitRepository(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	secretName string, apiClient client.Client) (*sourcev1.GitRepository, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.GitRepositoryNamespace
	key := types.NamespacedName{Name: name, Namespace: namespace}
	existing := &sourcev1.GitRepository{}

	err := apiClient.Get(ctx, key, existing)
	if err != nil {
		impl.logger.Errorw("error in getting git repository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	existing.Spec.URL = fluxCdSpec.RepoUrl
	existing.Spec.Interval = metav1.Duration{Duration: gitRepositoryReconcileInterval}
	existing.Spec.SecretRef = &meta.LocalObjectReference{Name: secretName}
	existing.Spec.Reference = &sourcev1.GitRepositoryRef{Branch: fluxCdSpec.RevisionTarget}

	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating git repository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	return existing, nil
}

func (impl *DeploymentServiceImpl) CreateHelmRelease(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	apiClient client.Client) (*helmv2.HelmRelease, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.HelmReleaseNamespace
	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "devtron",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: helmReleaseReconcileInterval}, //TODO
			DriftDetection: &helmv2.DriftDetection{
				Mode: helmv2.DriftDetectionEnabled,
			},
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					ReconcileStrategy: "Revision",
					Chart:             fluxCdSpec.ChartLocation,
					Version:           fluxCdSpec.ChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.GitRepositoryKind,
						Name:      name, //using same name for git repository and helm release which will be = deploymentAppName
						Namespace: namespace,
					},
					ValuesFiles: fluxCdSpec.GetFinalValuesFilePathArray(),
				},
			},
		},
	}
	err := apiClient.Create(ctx, helmRelease)
	if err != nil {
		impl.logger.Errorw("error in creating helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to create HelmRelease: %w", err)
	}
	return helmRelease, nil
}

func (impl *DeploymentServiceImpl) UpdateHelmRelease(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	apiClient client.Client) (*helmv2.HelmRelease, error) {
	name, namespace := fluxCdSpec.HelmReleaseName, fluxCdSpec.HelmReleaseNamespace
	key := types.NamespacedName{Name: name, Namespace: namespace}
	existing := &helmv2.HelmRelease{}

	gitRepositoryName, gitRepositoryNamespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.GitRepositoryNamespace

	err := apiClient.Get(ctx, key, existing)
	if err != nil {
		impl.logger.Errorw("error in getting helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to get HelmRelease: %w", err)
	}
	existing.Spec.DriftDetection = &helmv2.DriftDetection{
		Mode: helmv2.DriftDetectionEnabled,
	}
	existing.Spec.Interval = metav1.Duration{Duration: helmReleaseReconcileInterval}
	existing.Spec.Chart = &helmv2.HelmChartTemplate{
		Spec: helmv2.HelmChartTemplateSpec{
			ReconcileStrategy: "Revision",
			Chart:             fluxCdSpec.ChartLocation,
			Version:           fluxCdSpec.ChartVersion,
			SourceRef: helmv2.CrossNamespaceObjectReference{
				Kind:      sourcev1.GitRepositoryKind,
				Name:      gitRepositoryName,
				Namespace: gitRepositoryNamespace,
			},
			ValuesFiles: fluxCdSpec.GetFinalValuesFilePathArray(),
		},
	}
	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to update HelmRelease: %w", err)
	}
	return existing, nil
}

func (impl *DeploymentServiceImpl) DeleteFluxDeploymentApp(ctx context.Context, request *DeploymentAppDeleteRequest) error {
	fluxCdSpec := request.DeploymentConfig.ReleaseConfiguration.FluxCDSpec
	clusterId := fluxCdSpec.ClusterId
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(request.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "clusterId", clusterId, "err", err)
		return err
	}
	apiClient, err := getClient(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating k8s client", "clusterId", clusterId, "err", err)
		return err
	}
	name, namespace := fluxCdSpec.HelmReleaseName, fluxCdSpec.HelmReleaseNamespace
	key := types.NamespacedName{Name: name, Namespace: namespace}

	existingHelmRelease := &helmv2.HelmRelease{}
	err = apiClient.Get(ctx, key, existingHelmRelease)
	if err != nil {
		impl.logger.Errorw("error in getting helm release", "key", key, "err", err)
		return err
	}
	err = apiClient.Delete(ctx, existingHelmRelease)
	if err != nil {
		impl.logger.Errorw("error in deleting helm release", "key", key, "err", err)
		return err
	}

	key = types.NamespacedName{Name: fluxCdSpec.GitRepositoryName, Namespace: fluxCdSpec.GitRepositoryNamespace}
	existingGitRepository := &sourcev1.GitRepository{}
	err = apiClient.Get(ctx, key, existingGitRepository)
	if err != nil {
		impl.logger.Errorw("error in getting git repository", "key", key, "err", err)
		return err
	}
	err = apiClient.Delete(ctx, existingGitRepository)
	if err != nil {
		impl.logger.Errorw("error in deleting git repository", "name", name, "namespace", namespace, "err", err)
		return err
	}
	return nil
}

func getClient(config *rest.Config) (client.Client, error) {
	return client.New(config, client.Options{Scheme: getSchemeWithFluxCRDTypes()})
}

func getSchemeWithFluxCRDTypes() *runtime.Scheme {
	scheme := runtime.NewScheme()
	// Register core Kubernetes types
	_ = v1.AddToScheme(scheme)
	// Register Flux types
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	return scheme
}
