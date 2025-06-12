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

package devtronApps

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"go.opentelemetry.io/otel"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	gitRepositoryReconcileInterval = 1 * time.Minute
	helmReleaseReconcileInterval   = 1 * time.Minute
)

func (impl *HandlerServiceImpl) deployFluxCdApp(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.deployFluxCdApp")
	defer span.End()
	clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(overrideRequest.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", overrideRequest.ClusterId, "error", err)
		return err
	}
	gitOpsSecret, err := impl.upsertGitRepoSecret(newCtx, valuesOverrideResponse.ManifestPushTemplate.RepoUrl, overrideRequest.Namespace, clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in creating git repo secret", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	apiClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		impl.logger.Errorw("error in creating k8s client", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	//create/update gitOps secret
	if valuesOverrideResponse.Pipeline == nil || !valuesOverrideResponse.Pipeline.DeploymentAppCreated {
		err := impl.createFluxCdApp(newCtx, overrideRequest, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in creating flux-cd application", "err", err)
			return err
		}
	} else {
		err := impl.updateFluxCdApp(newCtx, overrideRequest, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in updating flux-cd application", "err", err)
			return err
		}
	}
	return nil
}

func (impl *HandlerServiceImpl) upsertGitRepoSecret(ctx context.Context, repoUrl, namespace string, clusterConfig *k8s.ClusterConfig) (*v1.Secret, error) {
	gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsProviderByRepoURL(repoUrl)
	if err != nil {
		impl.logger.Errorw("error fetching gitops config by repo url", "repoUrl", repoUrl, "err", err)
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

	secretName := fmt.Sprintf("devtron-flux-secret-%d", gitOpsConfig.Id)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
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
		impl.logger.Errorw("error creating secret", "namespace", namespace, "name", secretName, "err", err)
		return nil, err
	}

	// Secret already exists, get and update
	existingSecret, err := coreV1Client.Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error getting existing secret", "namespace", namespace, "name", secretName, "err", err)
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
		impl.logger.Errorw("error updating secret", "namespace", namespace, "name", secretName, "err", err)
		return nil, err
	}

	return updatedSecret, nil
}

func (impl *HandlerServiceImpl) createFluxCdApp(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.createFluxCdApp")
	defer span.End()
	deploymentAppName := valuesOverrideResponse.Pipeline.DeploymentAppName
	manifestPushTemplate := valuesOverrideResponse.ManifestPushTemplate
	_, err := impl.CreateGitRepository(newCtx, deploymentAppName, overrideRequest.Namespace, gitOpsSecretName,
		manifestPushTemplate.RepoUrl, '', apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating git repository", "name", deploymentAppName, "namespace", overrideRequest.Namespace, "err", err)
		return err
	}

	_, err = impl.CreateHelmRelease(newCtx, deploymentAppName, overrideRequest.Namespace, manifestPushTemplate, apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating helm release", "name", deploymentAppName, "namespace", overrideRequest.Namespace, "err", err)
		return err
	}

	err = impl.updateReleaseSpecForDeploymentConfiguration(valuesOverrideResponse.DeploymentConfig, deploymentAppName, overrideRequest.Namespace,
		gitOpsSecretName, overrideRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating release spec", "deploymentConfig", valuesOverrideResponse.DeploymentConfig, "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) updateReleaseSpecForDeploymentConfiguration(deploymentConfig *bean.DeploymentConfig,
	name, namespace, gitOpsSecretName string, userId int32) error {
	deploymentConfig.ReleaseConfiguration = adapter.NewFluxSpecReleaseConfig(namespace, name, name, gitOpsSecretName)
	_, err := impl.deploymentConfigService.CreateOrUpdateConfig(nil, deploymentConfig, userId)
	if err != nil {
		impl.logger.Errorw("error in updating deployment configuration", "deploymentConfig", deploymentConfig, "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) updateFluxCdApp(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.updateFluxCdApp")
	defer span.End()
	manifestPushTemplate := valuesOverrideResponse.ManifestPushTemplate
	fluxCdSpec := valuesOverrideResponse.DeploymentConfig.ReleaseConfiguration.FluxCDSpec
	_, err := impl.UpdateGitRepository(newCtx, fluxCdSpec, manifestPushTemplate, gitOpsSecretName, '', apiClient)
	if err != nil {
		impl.logger.Errorw("error in updating git repository", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}

	_, err = impl.UpdateHelmRelease(newCtx, fluxCdSpec, manifestPushTemplate, apiClient)
	if err != nil {
		impl.logger.Errorw("error in updating helm release", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) CreateGitRepository(ctx context.Context, name, namespace, secretName, repoURL, branchName string, apiClient client.Client) (*sourcev1.GitRepository, error) {
	gitRepo := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{Duration: gitRepositoryReconcileInterval},
			URL:      repoURL,
			SecretRef: &meta.LocalObjectReference{
				Name: secretName,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: branchName,
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

func (impl *HandlerServiceImpl) UpdateGitRepository(ctx context.Context, fluxCdSpec bean.FluxCDSpec, manifestPushTemplate *bean2.ManifestPushTemplate,
	secretName, branchName string, apiClient client.Client) (*sourcev1.GitRepository, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.Namespace
	key := types.NamespacedName{Name: name, Namespace: namespace}
	existing := &sourcev1.GitRepository{}

	err := apiClient.Get(ctx, key, existing)
	if err != nil {
		impl.logger.Errorw("error in getting git repository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	existing.Spec.URL = manifestPushTemplate.RepoUrl
	existing.Spec.Interval = metav1.Duration{Duration: gitRepositoryReconcileInterval}
	existing.Spec.SecretRef = &meta.LocalObjectReference{Name: secretName}
	existing.Spec.Reference = &sourcev1.GitRepositoryRef{Branch: branchName}

	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating git repository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	return existing, nil
}

func (impl *HandlerServiceImpl) CreateHelmRelease(ctx context.Context, name, namespace string,
	manifestPushTemplate *bean2.ManifestPushTemplate, apiClient client.Client) (*helmv2.HelmRelease, error) {
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
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   manifestPushTemplate.ChartLocation,
					Version: manifestPushTemplate.ChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.GitRepositoryKind,
						Name:      name, //using same name for git repository and helm release which will be = deploymentAppName
						Namespace: namespace,
					},
					ValuesFiles: getValuesFileArr(manifestPushTemplate.ChartLocation, manifestPushTemplate.ValuesFilePath),
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

func (impl *HandlerServiceImpl) UpdateHelmRelease(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	manifestPushTemplate *bean2.ManifestPushTemplate, apiClient client.Client) (*helmv2.HelmRelease, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.Namespace
	key := types.NamespacedName{Name: name, Namespace: namespace}
	existing := &helmv2.HelmRelease{}

	err := apiClient.Get(ctx, key, existing)
	if err != nil {
		impl.logger.Errorw("error in getting helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to get HelmRelease: %w", err)
	}
	existing.Spec.Interval = metav1.Duration{Duration: helmReleaseReconcileInterval}
	existing.Spec.Chart = &helmv2.HelmChartTemplate{
		Spec: helmv2.HelmChartTemplateSpec{
			Chart:   manifestPushTemplate.ChartLocation,
			Version: manifestPushTemplate.ChartVersion,
			SourceRef: helmv2.CrossNamespaceObjectReference{
				Kind:      sourcev1.GitRepositoryKind,
				Name:      name,
				Namespace: namespace,
			},
			ValuesFiles: getValuesFileArr(manifestPushTemplate.ChartLocation, manifestPushTemplate.ValuesFilePath),
		},
	}
	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to update HelmRelease: %w", err)
	}
	return existing, nil
}

func getValuesFileArr(chartLocation, valuesFilePath string) []string {
	//order matters here, last file will override previous file
	return []string{path.Join(chartLocation, "values.yaml"), path.Join(chartLocation, valuesFilePath)}
}
