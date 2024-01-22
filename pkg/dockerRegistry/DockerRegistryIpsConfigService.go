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
	"github.com/devtron-labs/common-lib/utils/k8s"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"strings"
)

type DockerRegistryIpsConfigService interface {
	IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int, isVirtualEnv bool) (bool, error)
	HandleImagePullSecretOnApplicationDeployment(environment *repository2.Environment, artifact *repository3.CiArtifact, ciPipelineId int, valuesFileContent []byte) ([]byte, error)
}

type DockerRegistryIpsConfigServiceImpl struct {
	logger                            *zap.SugaredLogger
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
	k8sUtil                           *k8s.K8sServiceImpl
	clusterService                    cluster.ClusterService
	ciPipelineRepository              pipelineConfig.CiPipelineRepository
	dockerArtifactStoreRepository     repository.DockerArtifactStoreRepository
	ciTemplateOverrideRepository      pipelineConfig.CiTemplateOverrideRepository
}

func NewDockerRegistryIpsConfigServiceImpl(logger *zap.SugaredLogger, dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository,
	k8sUtil *k8s.K8sServiceImpl, clusterService cluster.ClusterService, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository, ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository) *DockerRegistryIpsConfigServiceImpl {
	return &DockerRegistryIpsConfigServiceImpl{
		logger:                            logger,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
		k8sUtil:                           k8sUtil,
		clusterService:                    clusterService,
		ciPipelineRepository:              ciPipelineRepository,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
		ciTemplateOverrideRepository:      ciTemplateOverrideRepository,
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

func (impl DockerRegistryIpsConfigServiceImpl) HandleImagePullSecretOnApplicationDeployment(environment *repository2.Environment, artifact *repository3.CiArtifact, ciPipelineId int, valuesFileContent []byte) ([]byte, error) {
	clusterId := environment.ClusterId
	impl.logger.Infow("handling ips if access given", "ciPipelineId", ciPipelineId, "clusterId", clusterId)

	if ciPipelineId == 0 {
		impl.logger.Warn("returning as ciPipelineId is found 0")
		return valuesFileContent, nil
	}

	dockerRegistryId, err := impl.getDockerRegistryIdForCiPipeline(ciPipelineId, artifact)
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

func (impl DockerRegistryIpsConfigServiceImpl) getDockerRegistryIdForCiPipeline(ciPipelineId int, artifact *repository3.CiArtifact) (*string, error) {
	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipelineId, "error", err)
		return nil, err
	}

	if ciPipeline.IsExternal && ciPipeline.ParentCiPipeline == 0 {
		impl.logger.Warn("Ignoring for external ci")
		return nil, nil
	}

	if ciPipeline.CiTemplate == nil {
		impl.logger.Warn("returning as ciPipeline.CiTemplate is found nil")
		return nil, nil
	}
	var dockerRegistryId string
	if artifact.DataSource == repository3.POST_CI || artifact.DataSource == repository3.PRE_CD || artifact.DataSource == repository3.POST_CD {
		// if image is generated by plugin at these stages
		if artifact.CredentialsSourceType == repository3.GLOBAL_CONTAINER_REGISTRY {
			dockerRegistryId = artifact.CredentialSourceValue
		}
	} else if artifact.DataSource == repository3.CI_RUNNER {
		// if image is created by ci build
		dockerRegistryId = *ciPipeline.CiTemplate.DockerRegistryId
		if len(dockerRegistryId) == 0 {
			impl.logger.Warn("returning as dockerRegistryId is found empty")
			return nil, nil
		}

		if ciPipeline.IsDockerConfigOverridden {
			//set dockerRegistryId value with the DockerRegistryId of the overridden dockerRegistry
			ciPipId := ciPipelineId
			if ciPipeline.ParentCiPipeline != 0 {
				ciPipId = ciPipeline.ParentCiPipeline
			}
			ciTemplateOverride, err := impl.ciTemplateOverrideRepository.FindByCiPipelineId(ciPipId)
			if err != nil {
				impl.logger.Errorw("error in getting ciTemplateOverride by ciPipelineId", "ciPipelineId", ciPipelineId, "error", err)
				return nil, err
			}
			dockerRegistryId = ciTemplateOverride.DockerRegistryId
		}
	}
	return &dockerRegistryId, nil
}
func (impl DockerRegistryIpsConfigServiceImpl) createOrUpdateDockerRegistryImagePullSecret(clusterId int, namespace string, ipsName string, dockerRegistryBean *repository.DockerArtifactStore) error {
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
	cfg, err := clusterBean.GetClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "clusterId", clusterId, "error", err)
		return err
	}
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
