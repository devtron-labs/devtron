/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package dockerRegistry

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/common-lib/utils/k8s"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	util2 "github.com/devtron-labs/devtron/internal/util"
	ciConfig "github.com/devtron-labs/devtron/pkg/build/pipeline/read"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"strconv"
	"strings"
)

type DockerRegistryIpsConfigService interface {
	IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int, isVirtualEnv bool) (bool, error)
	HandleImagePullSecretOnApplicationDeployment(ctx context.Context, environment *repository2.Environment, artifact *repository3.CiArtifact, ciPipelineId int, valuesFileContent []byte) ([]byte, error)
}

type DockerRegistryIpsConfigServiceImpl struct {
	logger                            *zap.SugaredLogger
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
	k8sUtil                           *k8s.K8sServiceImpl
	dockerArtifactStoreRepository     repository.DockerArtifactStoreRepository
	clusterReadService                read.ClusterReadService
	ciPipelineConfigReadService       ciConfig.CiPipelineConfigReadService
}

func NewDockerRegistryIpsConfigServiceImpl(logger *zap.SugaredLogger, dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository,
	k8sUtil *k8s.K8sServiceImpl,
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository,
	clusterReadService read.ClusterReadService,
	ciPipelineConfigReadService ciConfig.CiPipelineConfigReadService) *DockerRegistryIpsConfigServiceImpl {
	return &DockerRegistryIpsConfigServiceImpl{
		logger:                            logger,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
		k8sUtil:                           k8sUtil,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
		clusterReadService:                clusterReadService,
		ciPipelineConfigReadService:       ciPipelineConfigReadService,
	}
}

func (impl DockerRegistryIpsConfigServiceImpl) IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int, isVirtualEnv bool) (bool, error) {
	impl.logger.Infow("checking if Ips access provided", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId)
	ipsConfig, err := impl.dockerRegistryIpsConfigRepository.FindByDockerRegistryId(dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("Error while getting docker registry ips config", "dockerRegistryId", dockerRegistryId, "err", err)
		if err == pg.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}
	isAccessProvided := CheckIfImagePullSecretAccessProvided(ipsConfig.AppliedClusterIdsCsv, ipsConfig.IgnoredClusterIdsCsv, clusterId, isVirtualEnv)
	return isAccessProvided, nil
}

func (impl DockerRegistryIpsConfigServiceImpl) HandleImagePullSecretOnApplicationDeployment(ctx context.Context, environment *repository2.Environment, artifact *repository3.CiArtifact, ciPipelineId int, valuesFileContent []byte) ([]byte, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "DockerRegistryIpsConfigServiceImpl.HandleImagePullSecretOnApplicationDeployment")
	defer span.End()
	clusterId := environment.ClusterId
	impl.logger.Infow("handling ips if access given", "ciPipelineId", ciPipelineId, "clusterId", clusterId)

	if ciPipelineId == 0 {
		impl.logger.Warn("returning as ciPipelineId is found 0")
		return valuesFileContent, nil
	}

	dockerRegistryId, err := impl.ciPipelineConfigReadService.GetDockerRegistryIdForCiPipeline(ciPipelineId, artifact)
	if err != nil {
		impl.logger.Errorw("error in getting docker registry", "dockerRegistryId", dockerRegistryId, "error", err)
		return valuesFileContent, err
	} else if dockerRegistryId == nil {
		return valuesFileContent, nil
	}

	dockerRegistryBean, err := impl.dockerArtifactStoreRepository.FindOne(*dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("error in getting docker registry", "dockerRegistryId", dockerRegistryId, "error", err)
		if err == pg.ErrNoRows {
			return valuesFileContent, nil
		} else {
			return nil, err
		}
	}

	// check if access provided, if not - return
	ipsConfig := dockerRegistryBean.IpsConfig
	if ipsConfig == nil {
		impl.logger.Warn("returning as ipsConfig is found nil")
		return valuesFileContent, nil
	}
	isVirtualEnv := environment.IsVirtualEnvironment
	ipsAccessProvided := CheckIfImagePullSecretAccessProvided(ipsConfig.AppliedClusterIdsCsv, ipsConfig.IgnoredClusterIdsCsv, clusterId, isVirtualEnv)
	if !ipsAccessProvided {
		impl.logger.Infow("ips access not given", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId)
		return valuesFileContent, nil
	}

	ipsCredentialType := string(ipsConfig.CredentialType)
	ipsName := BuildIpsName(*dockerRegistryId, ipsCredentialType, ipsConfig.CredentialValue)

	// Create or update secret of credential type is not of NAME type
	if ipsCredentialType != IPS_CREDENTIAL_TYPE_NAME {
		err = impl.createOrUpdateDockerRegistryImagePullSecret(clusterId, environment.Namespace, ipsName, dockerRegistryBean)
		if err != nil {
			return nil, err
		}
	}

	// merge ipsName in values
	impl.logger.Infow("setting ips name in values file", "ipsName", ipsName)
	updatedValuesFileContent, err := SetIpsNameInValues(valuesFileContent, ipsName)
	if err != nil {
		impl.logger.Errorw("error in setting ips name", "ipsName", ipsName, "error", err)
		return nil, err
	}

	return updatedValuesFileContent, nil
}

func (impl DockerRegistryIpsConfigServiceImpl) createOrUpdateDockerRegistryImagePullSecret(clusterId int, namespace string, ipsName string, dockerRegistryBean *repository.DockerArtifactStore) error {
	impl.logger.Infow("creating/updating ips", "ipsName", ipsName, "clusterId", clusterId)

	username := dockerRegistryBean.Username
	password := dockerRegistryBean.Password.String()
	registryURL := dockerRegistryBean.RegistryURL
	var email string

	// fetch from custom credentials
	if dockerRegistryBean.IpsConfig.CredentialType == IPS_CREDENTIAL_TYPE_CUSTOM_CREDENTIAL {
		var dockerIpsCustomCredential DockerIpsCustomCredential
		credentialValue := dockerRegistryBean.IpsConfig.CredentialValue
		err := json.Unmarshal([]byte(credentialValue), &dockerIpsCustomCredential)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling custom credentials", "credentialValue", credentialValue, "error", err)
			return err
		}
		if len(dockerIpsCustomCredential.Server) > 0 {
			registryURL = dockerIpsCustomCredential.Server
		}
		if len(dockerIpsCustomCredential.Username) > 0 {
			username = dockerIpsCustomCredential.Username
		}
		if len(dockerIpsCustomCredential.Password) > 0 {
			password = dockerIpsCustomCredential.Password
		}
		if len(dockerIpsCustomCredential.Email) > 0 {
			email = dockerIpsCustomCredential.Email
		}
	}

	registryType := dockerRegistryBean.RegistryType

	// ignore for ecr ec2_iam role
	if registryType == repository.REGISTRYTYPE_ECR {
		awsAccessKeyId := dockerRegistryBean.AWSAccessKeyId
		awsSecretAccessKey := dockerRegistryBean.AWSSecretAccessKey
		if len(awsAccessKeyId) == 0 || len(awsSecretAccessKey) == 0 {
			impl.logger.Info("ignoring for ecr ec2_iam role")
			return nil
		}
		// create credential for ecr
		impl.logger.Info("creating ecr credential")
		ecrUsername, ecrPassword, err := CreateCredentialForEcr(dockerRegistryBean.AWSRegion, awsAccessKeyId, awsSecretAccessKey.String())
		if err != nil {
			impl.logger.Errorw("error in creating ecr credential", "clusterId", clusterId, "error", err)
			return err
		}
		username = ecrUsername
		password = ecrPassword
	}

	// for gcr and artifact-registry, remove single quote from start and end, with this secret does not work
	if (registryType == repository.REGISTRYTYPE_GCR || registryType == repository.REGISTRYTYPE_ARTIFACT_REGISTRY) && username == repository.JSON_KEY_USERNAME {
		if strings.HasPrefix(password, "'") {
			password = password[1:]
		}
		if strings.HasSuffix(password, "'") {
			password = password[:len(password)-1]
		}
	}

	clusterBean, err := impl.clusterReadService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", clusterId, "error", err)
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	k8sClient, err := impl.k8sUtil.GetCoreV1Client(cfg)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "clusterId", clusterId, "error", err)
		return err
	}
	secret, err := impl.k8sUtil.GetSecret(namespace, ipsName, k8sClient)
	if err != nil {
		statusError, ok := err.(*k8sErrors.StatusError)
		if !ok || (statusError != nil && statusError.Status().Code != http.StatusNotFound) {
			impl.logger.Errorw("error in getting secret", "clusterId", clusterId, "namespace", namespace, "ipsName", ipsName, "error", err)
			return err
		}
		// create secret
		impl.logger.Infow("creating ips", "ipsName", ipsName, "clusterId", clusterId)
		ipsData := BuildIpsData(registryURL, username, password, email)
		_, err = impl.k8sUtil.CreateSecret(namespace, ipsData, ipsName, v1.SecretTypeDockerConfigJson, k8sClient, nil, nil)
		if err != nil {
			if statusError, ok = err.(*k8sErrors.StatusError); ok {
				errorCode := int(statusError.ErrStatus.Code)
				err = &util2.ApiError{Code: strconv.Itoa(errorCode), HttpStatusCode: errorCode, UserMessage: statusError.Error(), InternalMessage: statusError.Error()}
			}
			impl.logger.Errorw("error in creating secret", "clusterId", clusterId, "namespace", namespace, "ipsName", ipsName, "error", err)
			return err
		}
	} else {
		// update secret if username or password changed
		secretUsername, secretPassword := GetUsernamePasswordFromIpsSecret(registryURL, secret.Data)
		if username != secretUsername || password != secretPassword {
			impl.logger.Infow("updating ips", "ipsName", ipsName, "clusterId", clusterId)
			ipsData := BuildIpsData(registryURL, username, password, email)
			secret.Data = ipsData
			_, err = impl.k8sUtil.UpdateSecret(namespace, secret, k8sClient)
			if err != nil {
				impl.logger.Errorw("error in updating secret", "clusterId", clusterId, "namespace", namespace, "ipsName", ipsName, "error", err)
				return err
			}
		}
	}
	return nil
}
