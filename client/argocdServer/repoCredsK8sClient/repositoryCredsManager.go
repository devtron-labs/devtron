package repoCredsK8sClient

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient/bean"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/pkg/cluster"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	url2 "net/url"
	"path"
	yaml2 "sigs.k8s.io/yaml"
	"strings"
)

type RepositoryCreds interface {
	AddOrUpdateOCIRegistry(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error
	DeleteOCIRegistry(registryURL, repo string, ociRegistryId int) error
	AddChartRepository(request bean.ChartRepositoryAddRequest) error
	UpdateChartRepository(request bean.ChartRepositoryUpdateRequest) error
	DeleteChartRepository(name, url string) error
}

type RepositorySecretImpl struct {
	logger         *zap.SugaredLogger
	K8sService     k8s.K8sService
	clusterService cluster.ClusterService
	acdAuthConfig  *util2.ACDAuthConfig
}

func NewRepositorySecret(
	logger *zap.SugaredLogger,
	K8sService k8s.K8sService,
	clusterService cluster.ClusterService,
	acdAuthConfig *util2.ACDAuthConfig,
) *RepositorySecretImpl {
	return &RepositorySecretImpl{
		logger:         logger,
		K8sService:     K8sService,
		clusterService: clusterService,
		acdAuthConfig:  acdAuthConfig,
	}
}

func (impl *RepositorySecretImpl) AddOrUpdateOCIRegistry(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error {

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

	err = impl.createOrUpdateArgoRepoSecret(uniqueSecretName, secretData)
	if err != nil {
		impl.logger.Errorw("error in create/update k8s secret", "registryUrl", registryUrl, "repo", repo, "err", err)
		return err
	}
	return nil

}

func (impl *RepositorySecretImpl) DeleteOCIRegistry(registryURL, repo string, ociRegistryId int) error {

	_, _, chartName, err := getHostAndFullRepoPathAndChartName(registryURL, repo)
	if err != nil {
		return err
	}

	uniqueSecretName := fmt.Sprintf("%s-%d", chartName, ociRegistryId)

	err = impl.DeleteChartSecret(uniqueSecretName)
	if err != nil {
		impl.logger.Errorw("error in deleting oci registry secret", "secretName", uniqueSecretName, "err", err)
		return err
	}
	return nil
}

func (impl *RepositorySecretImpl) createOrUpdateArgoRepoSecret(uniqueSecretName string, secretData map[string]string) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster bean from db", "err", err)
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "err", err)
		return err
	}
	client, err := impl.K8sService.GetCoreV1Client(cfg)
	if err != nil {
		impl.logger.Errorw("error in creating kubernetes client", "err", err)
		return err
	}

	secretLabel := make(map[string]string)
	secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE

	err = impl.K8sService.CreateOrUpdateSecretByName(
		client,
		impl.acdAuthConfig.ACDConfigMapNamespace,
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

func (impl *RepositorySecretImpl) AddChartRepository(request bean.ChartRepositoryAddRequest) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster bean from db", "err", err)
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	client, err := impl.K8sService.GetCoreV1Client(cfg)
	if err != nil {
		impl.logger.Errorw("error in creating kubernetes client", "err", err)
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
		_, err = impl.K8sService.CreateSecret(impl.acdAuthConfig.ACDConfigMapNamespace, nil, request.Name, "", client, secretLabel, secretData)
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

func (impl *RepositorySecretImpl) UpdateChartRepository(request bean.ChartRepositoryUpdateRequest) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	client, err := impl.K8sService.GetCoreV1Client(cfg)
	if err != nil {
		return err
	}

	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sService.GetConfigMap(impl.acdAuthConfig.ACDConfigMapNamespace, impl.acdAuthConfig.ACDConfigMapName, client)
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
			_, err = impl.K8sService.UpdateConfigMap(impl.acdAuthConfig.ACDConfigMapNamespace, cm, client)
		} else {
			secretData := impl.CreateSecretDataForHelmChart(request.Name,
				request.Username,
				request.Password,
				request.URL,
				request.AllowInsecureConnection,
				request.IsPrivateChart)
			secret, err := impl.K8sService.GetSecret(impl.acdAuthConfig.ACDConfigMapNamespace, request.PreviousName, client)
			statusError, ok := err.(*errors2.StatusError)
			if err != nil && (ok && statusError != nil && statusError.Status().Code != http.StatusNotFound) {
				impl.logger.Errorw("error in fetching secret", "err", err)
				continue
			}

			if ok && statusError != nil && statusError.Status().Code == http.StatusNotFound {
				secretLabel := make(map[string]string)
				secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
				_, err = impl.K8sService.CreateSecret(impl.acdAuthConfig.ACDConfigMapNamespace, nil, request.PreviousName, "", client, secretLabel, secretData)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", request.PreviousName, "err", err)
					continue
				}
				updateSuccess = true
				break
			}

			if request.PreviousName != request.Name {
				err = impl.DeleteChartSecret(request.PreviousName)
				if err != nil {
					impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", request.Name, "err", err)
					continue
				}
				secretLabel := make(map[string]string)
				secretLabel[bean.ARGOCD_REPOSITORY_SECRET_KEY] = bean.ARGOCD_REPOSITORY_SECRET_VALUE
				_, err = impl.K8sService.CreateSecret(impl.acdAuthConfig.ACDConfigMapNamespace, nil, request.Name, "", client, secretLabel, secretData)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", request.Name, "err", err)
				}
			} else {
				secret.StringData = secretData
				_, err = impl.K8sService.UpdateSecret(impl.acdAuthConfig.ACDConfigMapNamespace, secret, client)
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

func (impl *RepositorySecretImpl) DeleteChartSecret(secretName string) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	client, err := impl.K8sService.GetCoreV1Client(cfg)
	if err != nil {
		return err
	}
	err = impl.K8sService.DeleteSecret(impl.acdAuthConfig.ACDConfigMapNamespace, secretName, client)
	return err
}

func (impl *RepositorySecretImpl) removeRepoData(data map[string]string, name string) (map[string]string, error) {
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
func (impl *RepositorySecretImpl) updateRepoData(data map[string]string, name, authMode, username, password, sshKey, url string) (map[string]string, error) {
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

func (impl *RepositorySecretImpl) createRepoElement(authmode, username, password, sshKey, url, name string) *bean.AcdConfigMapRepositoriesDto {
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
func (impl *RepositorySecretImpl) CreateSecretDataForHelmChart(name, username, password, repoURL string, allowInsecureConnection, isPrivateChart bool) (secretData map[string]string) {
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

func (impl *RepositorySecretImpl) DeleteChartRepository(name, url string) error {

	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return err
	}
	cfg := clusterBean.GetClusterConfig()
	client, err := impl.K8sService.GetCoreV1Client(cfg)
	if err != nil {
		return err
	}
	updateSuccess := false
	retryCount := 0
	//request.RedirectionUrl = ""

	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sService.GetConfigMap(impl.acdAuthConfig.ACDConfigMapNamespace, impl.acdAuthConfig.ACDConfigMapName, client)
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
			_, err = impl.K8sService.UpdateConfigMap(impl.acdAuthConfig.ACDConfigMapNamespace, cm, client)
		} else {
			err = impl.DeleteChartSecret(name)
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
