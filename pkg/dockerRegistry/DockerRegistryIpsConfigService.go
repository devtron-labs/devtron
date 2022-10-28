/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package dockerRegistry

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"strings"
)

type DockerRegistryIpsConfigService interface {
	IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int) (bool, error)
	CreateOrUpdateDockerRegistryImagePullSecret(clusterId int, namespace string, ipsName string, dockerRegistryBean *repository.DockerArtifactStore) error
	HandleImagePullSecretOnApplicationDeployment(environment *repository2.Environment, ciPipelineId int, valuesFileContent []byte) ([]byte, error)
}

type DockerRegistryIpsConfigServiceImpl struct {
	logger                            *zap.SugaredLogger
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
	k8sUtil                           *util.K8sUtil
	clusterService                    cluster.ClusterService
	ciPipelineRepository              pipelineConfig.CiPipelineRepository
	dockerArtifactStoreRepository     repository.DockerArtifactStoreRepository
}

func NewDockerRegistryIpsConfigServiceImpl(logger *zap.SugaredLogger, dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository,
	k8sUtil *util.K8sUtil, clusterService cluster.ClusterService, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository) *DockerRegistryIpsConfigServiceImpl {
	return &DockerRegistryIpsConfigServiceImpl{
		logger:                            logger,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
		k8sUtil:                           k8sUtil,
		clusterService:                    clusterService,
		ciPipelineRepository:              ciPipelineRepository,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
	}
}

func (impl DockerRegistryIpsConfigServiceImpl) IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int) (bool, error) {
	impl.logger.Infow("checking if Ips access provided", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId)

	ipsConfig, err := impl.dockerRegistryIpsConfigRepository.FindByDockerRegistryId(dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("Error while getting docker registry ips config", "dockerRegistryId", dockerRegistryId, "err", err)
		return false, err
	}
	isAccessProvided := CheckIfImagePullSecretAccessProvided(ipsConfig.AppliedClusterIdsCsv, ipsConfig.IgnoredClusterIdsCsv, clusterId)
	return isAccessProvided, nil
}

func (impl DockerRegistryIpsConfigServiceImpl) HandleImagePullSecretOnApplicationDeployment(environment *repository2.Environment, ciPipelineId int, valuesFileContent []byte) ([]byte, error) {
	clusterId := environment.ClusterId
	impl.logger.Infow("handling ips if access given", "ciPipelineId", ciPipelineId, "clusterId", clusterId)

	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipelineId, "error", err)
		return nil, err
	}

	dockerRegistryId := ciPipeline.CiTemplate.DockerRegistryId
	dockerRegistryBean, err := impl.dockerArtifactStoreRepository.FindOne(dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("error in getting docker registry", "dockerRegistryId", dockerRegistryId, "error", err)
		return nil, err
	}

	// check if access provided, if not - return
	ipsConfig := dockerRegistryBean.IpsConfig
	ipsAccessProvided := CheckIfImagePullSecretAccessProvided(ipsConfig.AppliedClusterIdsCsv, ipsConfig.IgnoredClusterIdsCsv, clusterId)
	if !ipsAccessProvided {
		impl.logger.Infow("ips access not given", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId)
		return valuesFileContent, nil
	}

	ipsCredentialType := string(ipsConfig.CredentialType)
	ipsName := BuildIpsName(dockerRegistryId, ipsCredentialType, ipsConfig.CredentialValue)

	// Create or update secret of credential type is not of NAME type
	if ipsCredentialType != IPS_CREDENTIAL_TYPE_NAME {
		err = impl.CreateOrUpdateDockerRegistryImagePullSecret(clusterId, environment.Namespace, ipsName, dockerRegistryBean)
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

func (impl DockerRegistryIpsConfigServiceImpl) CreateOrUpdateDockerRegistryImagePullSecret(clusterId int, namespace string, ipsName string, dockerRegistryBean *repository.DockerArtifactStore) error {
	impl.logger.Infow("creating/updating ips", "ipsName", ipsName, "clusterId", clusterId)

	username := dockerRegistryBean.Username
	password := dockerRegistryBean.Password
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
		ecrUsername, ecrPassword, err := CreateCredentialForEcr(dockerRegistryBean.AWSRegion, awsAccessKeyId, awsSecretAccessKey)
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

	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", clusterId, "error", err)
		return err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "clusterId", clusterId, "error", err)
		return err
	}
	k8sClient, err := impl.k8sUtil.GetClient(cfg)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "clusterId", clusterId, "error", err)
		return err
	}
	secret, err := impl.k8sUtil.GetSecret(namespace, ipsName, k8sClient)
	if err != nil {
		statusError, _ := err.(*k8sErrors.StatusError)
		if statusError.Status().Code != http.StatusNotFound {
			impl.logger.Errorw("error in getting secret", "clusterId", clusterId, "namespace", namespace, "ipsName", ipsName, "error", err)
			return err
		}
		// create secret
		impl.logger.Infow("creating ips", "ipsName", ipsName, "clusterId", clusterId)
		ipsData := BuildIpsData(registryURL, username, password, email)
		_, err = impl.k8sUtil.CreateSecret(namespace, ipsData, ipsName, v1.SecretTypeDockerConfigJson, k8sClient)
		if err != nil {
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
