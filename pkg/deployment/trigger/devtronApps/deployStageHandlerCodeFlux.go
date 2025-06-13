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
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"go.opentelemetry.io/otel"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
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
	gitOpsSecret, err := impl.upsertGitRepoSecret(newCtx,
		valuesOverrideResponse.DeploymentConfig.ReleaseConfiguration.FluxCDSpec.GitOpsSecretName,
		valuesOverrideResponse.ManifestPushTemplate.RepoUrl,
		overrideRequest.Namespace,
		clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in creating git repo secret", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	apiClient, err := getClient(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating k8s client", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	//create/update gitOps secret
	if valuesOverrideResponse.Pipeline == nil || !valuesOverrideResponse.Pipeline.DeploymentAppCreated {
		err := impl.createFluxCdApp(newCtx, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in creating flux-cd application", "err", err)
			return err
		}
	} else {
		err := impl.updateFluxCdApp(newCtx, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in updating flux-cd application", "err", err)
			return err
		}
	}
	return nil
}

func getClient(config *rest.Config) (client.Client, error) {
	scheme := runtime.NewScheme()
	// Register core Kubernetes types
	_ = v1.AddToScheme(scheme)
	// Register Flux types
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	return client.New(config, client.Options{Scheme: scheme})
}

func (impl *HandlerServiceImpl) upsertGitRepoSecret(ctx context.Context, secretName, repoUrl, namespace string, clusterConfig *k8s.ClusterConfig) (*v1.Secret, error) {
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

func (impl *HandlerServiceImpl) createFluxCdApp(ctx context.Context, valuesOverrideResponse *app.ValuesOverrideResponse,
	gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.createFluxCdApp")
	defer span.End()
	manifestPushTemplate := valuesOverrideResponse.ManifestPushTemplate
	fluxCdSpec := valuesOverrideResponse.DeploymentConfig.ReleaseConfiguration.FluxCDSpec
	_, err := impl.CreateGitRepository(newCtx, fluxCdSpec, gitOpsSecretName, manifestPushTemplate, apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating git repository", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}

	_, err = impl.CreateHelmRelease(newCtx, fluxCdSpec, manifestPushTemplate, apiClient)
	if err != nil {
		impl.logger.Errorw("error in creating helm release", "fluxCdSpec", fluxCdSpec, "err", err)
		return err
	}

	return nil
}

func (impl *HandlerServiceImpl) updateFluxCdApp(ctx context.Context, valuesOverrideResponse *app.ValuesOverrideResponse,
	gitOpsSecretName string, apiClient client.Client) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.updateFluxCdApp")
	defer span.End()
	manifestPushTemplate := valuesOverrideResponse.ManifestPushTemplate
	fluxCdSpec := valuesOverrideResponse.DeploymentConfig.ReleaseConfiguration.FluxCDSpec
	_, err := impl.UpdateGitRepository(newCtx, fluxCdSpec, manifestPushTemplate, gitOpsSecretName, apiClient)
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

func (impl *HandlerServiceImpl) CreateGitRepository(ctx context.Context, fluxCdSpec bean.FluxCDSpec, secretName string, manifestPushTemplate *bean2.ManifestPushTemplate, apiClient client.Client) (*sourcev1.GitRepository, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.Namespace
	gitRepo := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{Duration: gitRepositoryReconcileInterval},
			URL:      manifestPushTemplate.RepoUrl,
			SecretRef: &meta.LocalObjectReference{
				Name: secretName,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: manifestPushTemplate.TargetRevision,
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
	secretName string, apiClient client.Client) (*sourcev1.GitRepository, error) {
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
	existing.Spec.Reference = &sourcev1.GitRepositoryRef{Branch: manifestPushTemplate.TargetRevision}

	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating git repository", "name", name, "namespace", namespace, "err", err)
		return nil, err
	}
	return existing, nil
}

func (impl *HandlerServiceImpl) CreateHelmRelease(ctx context.Context, fluxCdSpec bean.FluxCDSpec,
	manifestPushTemplate *bean2.ManifestPushTemplate, apiClient client.Client) (*helmv2.HelmRelease, error) {
	name, namespace := fluxCdSpec.GitRepositoryName, fluxCdSpec.Namespace
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
					ValuesFiles: fluxCdSpec.ValuesFiles,
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
			ValuesFiles: fluxCdSpec.ValuesFiles,
		},
	}
	err = apiClient.Update(ctx, existing)
	if err != nil {
		impl.logger.Errorw("error in updating helm release", "name", name, "namespace", namespace, "err", err)
		return nil, fmt.Errorf("failed to update HelmRelease: %w", err)
	}
	return existing, nil
}
