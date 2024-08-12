package argoRepositoryCreds

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/argoRepositoryCreds/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
	url2 "net/url"
	"path"
	"strings"
)

type RepositorySecret interface {
	CreateArgoRepositorySecret(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error
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

func (impl *RepositorySecretImpl) CreateArgoRepositorySecret(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error {

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

	url := registryUrl

	host, fullRepoPath, err := GetHostAndFullRepoPath(url, repo)
	if err != nil {
		return nil, "", err
	}

	repoSplit := strings.Split(fullRepoPath, "/")
	chartName := repoSplit[len(repoSplit)-1]

	uniqueSecretName := fmt.Sprintf("%s-%d", chartName, ociRegistryId)

	secretData := parseSecretData(username, password, fullRepoPath, host, isPublic)

	return secretData, uniqueSecretName, nil
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

func parseSecretData(username, password, repoName, repoHost string, isPublic bool) map[string]string {
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
