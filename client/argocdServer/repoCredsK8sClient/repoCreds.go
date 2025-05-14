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

package repoCredsK8sClient

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient/bean"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	url2 "net/url"
	"path"
	yaml2 "sigs.k8s.io/yaml"
	"strings"
)

type RepositoryCredsK8sClient interface {
	AddOrUpdateOCIRegistry(argoK8sConfig *argoApplication.ArgoK8sConfig, username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error
	DeleteOCIRegistry(argoK8sConfig *argoApplication.ArgoK8sConfig, registryURL, repo string, ociRegistryId int) error
	AddChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, request bean.ChartRepositoryAddRequest) error
	UpdateChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, request bean.ChartRepositoryUpdateRequest) error
	DeleteChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, name, url string) error
}

type RepositoryCredsK8sClientImpl struct {
	logger     *zap.SugaredLogger
	K8sService k8s.K8sService
}

func NewRepositoryCredsK8sClientImpl(
	logger *zap.SugaredLogger,
	K8sService k8s.K8sService,
) *RepositoryCredsK8sClientImpl {
	return &RepositoryCredsK8sClientImpl{
		logger:     logger,
		K8sService: K8sService,
	}
}

func (impl *RepositoryCredsK8sClientImpl) getK8sClientByArgoK8sConfig(config *argoApplication.ArgoK8sConfig) (*v1.CoreV1Client, error) {
	k8sClient, err := impl.K8sService.GetCoreV1ClientByRestConfig(config.RestConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return nil, err
	}
	return k8sClient, nil
}

func (impl *RepositoryCredsK8sClientImpl) AddOrUpdateOCIRegistry(argoK8sConfig *argoApplication.ArgoK8sConfig, username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error {

	secretData, uniqueSecretName, err := getSecretDataAndName(
		username,
		password,
		uniqueId,
		registryUrl,
		repo,
		isPublic)
	if err != nil {
		impl.logger.Errorw("error in getting secretData and secretName", "err", err)
		return err
	}

	err = impl.createOrUpdateArgoRepoSecret(argoK8sConfig, uniqueSecretName, secretData)
	if err != nil {
		impl.logger.Errorw("error in create/update k8s secret", "registryUrl", registryUrl, "repo", repo, "err", err)
		return err
	}
	return nil

}

func (impl *RepositoryCredsK8sClientImpl) DeleteOCIRegistry(argoK8sConfig *argoApplication.ArgoK8sConfig, registryURL, repo string, ociRegistryId int) error {

	_, _, chartName, err := getHostAndFullRepoPathAndChartName(registryURL, repo)
	if err != nil {
		return err
	}

	uniqueSecretName := fmt.Sprintf("%s-%d", chartName, ociRegistryId)

	err = impl.DeleteChartSecret(argoK8sConfig, uniqueSecretName)
	if err != nil {
		impl.logger.Errorw("error in deleting oci registry secret", "secretName", uniqueSecretName, "err", err)
		return err
	}
	return nil
}

func (impl *RepositoryCredsK8sClientImpl) createOrUpdateArgoRepoSecret(argoK8sConfig *argoApplication.ArgoK8sConfig, uniqueSecretName string, secretData map[string]string) error {

	k8sClient, err := impl.getK8sClientByArgoK8sConfig(argoK8sConfig)
	if err != nil {
		return err
	}

	secretLabel := make(map[string]string)
	secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
	err = impl.K8sService.CreateOrUpdateSecretByName(
		k8sClient,
		argoK8sConfig.AcdNamespace,
		uniqueSecretName,
		secretLabel,
		secretData)
	if err != nil {
		impl.logger.Errorw("error in create/update argocd secret by name", "secretName", uniqueSecretName, "err", err)
		return err
	}

	return nil
}

func getSecretDataAndName(username, password string, ociRegistryId int, registryUrl, repo string, isPublic bool) (map[string]string, string, error) {

	host, fullRepoPath, chartName, err := getHostAndFullRepoPathAndChartName(registryUrl, repo)
	if err != nil {
		return nil, "", nil
	}

	uniqueSecretName := fmt.Sprintf("%s-%d", chartName, ociRegistryId)

	secretData := parseSecretDataForOCI(username, password, fullRepoPath, host, isPublic)

	return secretData, uniqueSecretName, nil
}

func getHostAndFullRepoPathAndChartName(registryUrl string, repo string) (string, string, string, error) {
	host, fullRepoPath, err := GetHostAndFullRepoPath(registryUrl, repo)
	if err != nil {
		return "", "", "", err
	}
	repoSplit := strings.Split(fullRepoPath, "/")
	chartName := repoSplit[len(repoSplit)-1]
	return host, fullRepoPath, chartName, nil
}

func GetHostAndFullRepoPath(url string, repo string) (string, string, error) {
	parsedUrl, err := url2.Parse(url)
	if err != nil || parsedUrl.Scheme == "" {
		url = fmt.Sprintf("//%s", url)
		parsedUrl, err = url2.Parse(url)
		if err != nil {
			return "", "", err
		}
	}
	var repoName string
	if len(parsedUrl.Path) > 0 {
		repoName = path.Join(strings.TrimLeft(parsedUrl.Path, "/"), repo)
	} else {
		repoName = repo
	}
	return parsedUrl.Host, repoName, nil
}

func parseSecretDataForOCI(username, password, repoName, repoHost string, isPublic bool) map[string]string {
	secretData := make(map[string]string)
	secretData[bean.REPOSITORY_SECRET_NAME_KEY] = repoName // argocd will use this for passing repository credentials to application
	secretData[bean.REPOSITORY_SECRET_TYPE_KEY] = bean.REPOSITORY_TYPE_HELM
	secretData[bean.REPOSITORY_SECRET_URL_KEY] = repoHost
	if !isPublic {
		secretData[bean.REPOSITORY_SECRET_USERNAME_KEY] = username
		secretData[bean.REPOSITORY_SECRET_PASSWORD_KEY] = password
	}
	secretData[bean.REPOSITORY_SECRET_ENABLE_OCI_KEY] = "true"
	return secretData
}

func (impl *RepositoryCredsK8sClientImpl) AddChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, request bean.ChartRepositoryAddRequest) error {

	k8sClient, err := impl.getK8sClientByArgoK8sConfig(argoK8sConfig)
	if err != nil {
		return err
	}

	updateSuccess := false
	retryCount := 0

	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		secretLabel := make(map[string]string)
		secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
		secretData := impl.CreateSecretDataForHelmChart(
			request.Name,
			request.Username,
			request.Password,
			request.URL,
			request.AllowInsecureConnection,
			request.IsPrivateChart)
		_, err = impl.K8sService.CreateSecret(argoK8sConfig.AcdNamespace, nil, request.Name, "", k8sClient, secretLabel, secretData)
		if err != nil {
			// TODO refactoring:  Implement the below error handling if secret name already exists
			//if statusError, ok := err.(*k8sErrors.StatusError); ok &&
			//	statusError != nil &&
			//	statusError.Status().Code == http.StatusConflict &&
			//	statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
			//	impl.logger.Errorw("secret already exists", "err", statusError.Error())
			//	return nil, fmt.Errorf(statusError.Error())
			//}
			continue
		}
		if err == nil {
			updateSuccess = true
		}
	}
	if !updateSuccess {
		impl.logger.Errorw("error in creating secret for chart repository", "err", err)
		return fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	return nil
}

func (impl *RepositoryCredsK8sClientImpl) UpdateChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, request bean.ChartRepositoryUpdateRequest) error {

	k8sClient, err := impl.getK8sClientByArgoK8sConfig(argoK8sConfig)
	if err != nil {
		return err
	}

	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sService.GetConfigMap(argoK8sConfig.AcdNamespace, argoK8sConfig.AcdConfigMapName, k8sClient)
		if err != nil {
			return err
		}
		var repositories []*bean.AcdConfigMapRepositoriesDto
		if cm != nil && cm.Data != nil {
			repoStr := cm.Data["repositories"]
			repoByte, err := yaml2.YAMLToJSON([]byte(repoStr))
			if err != nil {
				impl.logger.Errorw("error in json patch", "err", err)
				return err
			}
			err = json.Unmarshal(repoByte, &repositories)
			if err != nil {
				impl.logger.Errorw("error in unmarshal", "err", err)
				return err
			}
			for _, repo := range repositories {
				if repo.Name == request.PreviousName && repo.Url == request.PreviousURL {
					//chart repo is present in argocd-cm
					isFoundInArgoCdCm = true
					break
				}
			}
		}

		if isFoundInArgoCdCm {
			var data map[string]string
			// if the repo name has been updated then, create a new repo
			if cm != nil && cm.Data != nil {
				data, err = impl.updateRepoData(cm.Data,
					request.Name,
					request.AuthMode,
					request.Username,
					request.Password,
					request.SSHKey,
					request.URL)
				// if the repo name has been updated then, delete the previous repo
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
				if request.PreviousName != request.Name {
					data, err = impl.removeRepoData(cm.Data, request.PreviousName)
				}
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
			}
			cm.Data = data
			_, err = impl.K8sService.UpdateConfigMap(argoK8sConfig.AcdNamespace, cm, k8sClient)
		} else {
			secretData := impl.CreateSecretDataForHelmChart(request.Name,
				request.Username,
				request.Password,
				request.URL,
				request.AllowInsecureConnection,
				request.IsPrivateChart)
			secret, err := impl.K8sService.GetSecret(argoK8sConfig.AcdNamespace, request.PreviousName, k8sClient)
			statusError, ok := err.(*errors2.StatusError)
			if err != nil && (ok && statusError != nil && statusError.Status().Code != http.StatusNotFound) {
				impl.logger.Errorw("error in fetching secret", "err", err)
				continue
			}

			if ok && statusError != nil && statusError.Status().Code == http.StatusNotFound {
				secretLabel := make(map[string]string)
				secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
				_, err = impl.K8sService.CreateSecret(argoK8sConfig.AcdNamespace, nil, request.PreviousName, "", k8sClient, secretLabel, secretData)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", request.PreviousName, "err", err)
					continue
				}
				updateSuccess = true
				break
			}

			if request.PreviousName != request.Name {
				err = impl.DeleteChartSecret(argoK8sConfig, request.PreviousName)
				if err != nil {
					impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", request.Name, "err", err)
					continue
				}
				secretLabel := make(map[string]string)
				secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
				_, err = impl.K8sService.CreateSecret(argoK8sConfig.AcdNamespace, nil, request.Name, "", k8sClient, secretLabel, secretData)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", request.Name, "err", err)
				}
			} else {
				secret.StringData = secretData
				_, err = impl.K8sService.UpdateSecret(argoK8sConfig.AcdNamespace, secret, k8sClient)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", request.Name, "err", err)
				}
			}
			if err != nil {
				impl.logger.Warnw("secret update for chart repo failed", "err", err)
				continue
			}
		}
		if err != nil {
			impl.logger.Warnw(" config map failed", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Warnw(" config map apply succeeded", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	return nil
}

func (impl *RepositoryCredsK8sClientImpl) DeleteChartSecret(argoK8sConfig *argoApplication.ArgoK8sConfig, secretName string) error {
	k8sClient, err := impl.getK8sClientByArgoK8sConfig(argoK8sConfig)
	if err != nil {
		return err
	}
	err = impl.K8sService.DeleteSecret(argoK8sConfig.AcdNamespace, secretName, k8sClient)
	return err
}

func (impl *RepositoryCredsK8sClientImpl) removeRepoData(data map[string]string, name string) (map[string]string, error) {
	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	var helmRepositories []*bean.AcdConfigMapRepositoriesDto
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	//SETUP for repositories
	var repositories []*bean.AcdConfigMapRepositoriesDto
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	found := false
	for index, item := range repositories {
		//if request chart repo found, then delete its values
		if item.Name == name {
			repositories = append(repositories[:index], repositories[index+1:]...)
			found = true
			break
		}
	}

	// if request chart repo not found, add new one
	if !found {
		impl.logger.Errorw("Repo not found", "err", err)
		return nil, fmt.Errorf("Repo not found in config-map")
	}

	rb, err = json.Marshal(repositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	if len(helmRepositoriesYamlByte) > 0 {
		data["helm.repositories"] = string(helmRepositoriesYamlByte)
	}
	if len(repositoriesYamlByte) > 0 {
		data["repositories"] = string(repositoriesYamlByte)
	}
	//dex config copy as it is
	dexConfigStr := data["dex.config"]
	data["dex.config"] = string([]byte(dexConfigStr))
	return data, nil
}

// updateRepoData update the request field in the argo-cm
func (impl *RepositoryCredsK8sClientImpl) updateRepoData(data map[string]string, name, authMode, username, password, sshKey, url string) (map[string]string, error) {
	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	var helmRepositories []*bean.AcdConfigMapRepositoriesDto
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	//SETUP for repositories
	var repositories []*bean.AcdConfigMapRepositoriesDto
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	found := false
	for _, item := range repositories {
		//if request chart repo found, then update its values
		if item.Name == name {
			if authMode == string(constants.AUTH_MODE_USERNAME_PASSWORD) {
				usernameSecret := &bean.KeyDto{Name: username, Key: "username"}
				passwordSecret := &bean.KeyDto{Name: password, Key: "password"}
				item.PasswordSecret = passwordSecret
				item.UsernameSecret = usernameSecret
			} else if authMode == string(constants.AUTH_MODE_ACCESS_TOKEN) {
				// TODO - is it access token or ca cert nd secret
			} else if authMode == string(constants.AUTH_MODE_SSH) {
				keySecret := &bean.KeyDto{Name: sshKey, Key: "key"}
				item.KeySecret = keySecret
			}
			item.Url = url
			found = true
		}
	}

	// if request chart repo not found, add new one
	if !found {
		repoData := impl.createRepoElement(authMode, username, password, sshKey, url, name)
		repositories = append(repositories, repoData)
	}

	rb, err = json.Marshal(repositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	if len(helmRepositoriesYamlByte) > 0 {
		data["helm.repositories"] = string(helmRepositoriesYamlByte)
	}
	if len(repositoriesYamlByte) > 0 {
		data["repositories"] = string(repositoriesYamlByte)
	}
	//dex config copy as it is
	dexConfigStr := data["dex.config"]
	data["dex.config"] = string([]byte(dexConfigStr))
	return data, nil
}

func (impl *RepositoryCredsK8sClientImpl) createRepoElement(authmode, username, password, sshKey, url, name string) *bean.AcdConfigMapRepositoriesDto {
	repoData := &bean.AcdConfigMapRepositoriesDto{}
	if authmode == string(constants.AUTH_MODE_USERNAME_PASSWORD) {
		usernameSecret := &bean.KeyDto{Name: username, Key: "username"}
		passwordSecret := &bean.KeyDto{Name: password, Key: "password"}
		repoData.PasswordSecret = passwordSecret
		repoData.UsernameSecret = usernameSecret
	} else if authmode == string(constants.AUTH_MODE_ACCESS_TOKEN) {
		// TODO - is it access token or ca cert nd secret
	} else if (authmode) == string(constants.AUTH_MODE_SSH) {
		keySecret := &bean.KeyDto{Name: sshKey, Key: "key"}
		repoData.KeySecret = keySecret
	}
	repoData.Url = url
	repoData.Name = name
	repoData.Type = "helm"

	return repoData
}

// Private helm charts credentials are saved as secrets
func (impl *RepositoryCredsK8sClientImpl) CreateSecretDataForHelmChart(name, username, password, repoURL string, allowInsecureConnection, isPrivateChart bool) (secretData map[string]string) {
	secretData = make(map[string]string)
	secretData[bean.REPOSITORY_SECRET_NAME_KEY] = fmt.Sprintf("%s-%s", name, uuid.New().String()) // making repo name unique so that "helm repo add" command in argo-repo-server doesn't give error
	secretData[bean.REPOSITORY_SECRET_TYPE_KEY] = bean.REPOSITORY_TYPE_HELM
	secretData[bean.REPOSITORY_SECRET_URL_KEY] = repoURL
	if isPrivateChart {
		secretData[bean.REPOSITORY_SECRET_USERNAME_KEY] = username
		secretData[bean.REPOSITORY_SECRET_PASSWORD_KEY] = password
	}
	isInsecureConnection := "true"
	if !allowInsecureConnection {
		isInsecureConnection = "false"
	}
	secretData[bean.REPOSITORY_SECRET_INSECURE_KEY] = isInsecureConnection
	return secretData
}

func (impl *RepositoryCredsK8sClientImpl) DeleteChartRepository(argoK8sConfig *argoApplication.ArgoK8sConfig, name, url string) error {

	k8sClient, err := impl.getK8sClientByArgoK8sConfig(argoK8sConfig)
	if err != nil {
		return err
	}

	updateSuccess := false
	retryCount := 0
	//request.RedirectionUrl = ""

	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sService.GetConfigMap(argoK8sConfig.AcdNamespace, argoK8sConfig.AcdConfigMapName, k8sClient)
		if err != nil {
			return err
		}
		var repositories []*bean.AcdConfigMapRepositoriesDto
		if cm != nil && cm.Data != nil {
			repoStr := cm.Data["repositories"]
			repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
			if err != nil {
				impl.logger.Errorw("error in json patch", "err", err)
				return err
			}
			err = json.Unmarshal(repoByte, &repositories)
			if err != nil {
				impl.logger.Errorw("error in unmarshal", "err", err)
				return err
			}
			for _, repo := range repositories {
				if repo.Name == name && repo.Url == url {
					//chart repo is present in argocd-cm
					isFoundInArgoCdCm = true
					break
				}
			}
		}

		if isFoundInArgoCdCm {
			var data map[string]string

			if cm != nil && cm.Data != nil {
				data, err = impl.removeRepoData(cm.Data, name)
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
			}
			cm.Data = data
			_, err = impl.K8sService.UpdateConfigMap(argoK8sConfig.AcdConfigMapName, cm, k8sClient)
		} else {
			err = impl.DeleteChartSecret(argoK8sConfig, name)
			if err != nil {
				impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", name, "err", err)
			}
		}
		if err != nil {
			impl.logger.Warnw(" error in deleting config/secret failed", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Warnw(" config map apply succeeded", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	return nil
}
