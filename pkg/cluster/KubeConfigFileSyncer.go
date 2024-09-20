package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type KubeConfigFileSyncer interface {
	SyncNow()
}

type KubeConfigFileSyncerImpl struct {
	logger            *zap.SugaredLogger
	clusterService    *ClusterServiceImpl
	syncMutex         *sync.Mutex
	config            *KubeConfigFileSyncerConfig
	syncingInProgress bool
}

type KubeConfigFileSyncerConfig struct {
	KubeconfigSyncTimeInSecs int `env:"KUBE_CONFIG_SYNC_TIME_IN_SECS" envDefault:"600"`
}

func NewKubeConfigFileSyncerImpl(logger *zap.SugaredLogger, clusterService *ClusterServiceImpl) (*KubeConfigFileSyncerImpl, error) {
	syncerConfig := &KubeConfigFileSyncerConfig{}
	err := env.Parse(syncerConfig)
	if err != nil {
		logger.Fatal("error occurred while reading kubeconfig syncer config ", "err", err)
	}
	ConfigFileSyncerCron := cron.New(cron.WithChain())
	ConfigFileSyncerCron.Start()
	syncerImpl := &KubeConfigFileSyncerImpl{logger: logger, clusterService: clusterService, syncMutex: &sync.Mutex{}}
	_, err = ConfigFileSyncerCron.AddFunc(fmt.Sprintf("@every %ds", syncerConfig.KubeconfigSyncTimeInSecs), syncerImpl.SyncNow)
	if err != nil {
		logger.Errorw("error occurred while starting cron job for kubeconfig syncer", "time_in_secs", syncerConfig.KubeconfigSyncTimeInSecs)
		return nil, err
	}
	go syncerImpl.SyncNow()
	return syncerImpl, nil
}

func (impl *KubeConfigFileSyncerImpl) SyncNow() {
	if impl.syncingInProgress {
		impl.logger.Warn("syncing in progress, so ignoring!!")
		return
	}
	impl.syncMutex.Lock()
	defer impl.syncMutex.Unlock()
	impl.setSyncingInProgress(true)
	defer impl.setSyncingInProgress(false)
	impl.syncKubeconfig()
}

func (impl *KubeConfigFileSyncerImpl) setSyncingInProgress(inProgress bool) {
	impl.syncingInProgress = inProgress
}

func (impl *KubeConfigFileSyncerImpl) syncKubeconfig() error {
	kubefolder, dirEntries, err := impl.readKubeConfigFolder()
	if err != nil {
		return errors.New("failed to read kubeconfig folder")
	}
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			fileInfo, err := entry.Info()
			if err != nil {
				impl.logger.Errorw("error occurred while fetching file info", "err", err)
				continue
			}
			fileName := fileInfo.Name()
			impl.processKubeconfigFile(kubefolder, fileName)
		}
	}
	return nil
}

func (impl *KubeConfigFileSyncerImpl) readKubeConfigFolder() (string, []os.DirEntry, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		impl.logger.Errorw("error occurred while reading user home dir", "err", err)
		return "", nil, err
	}
	kubeFolderPath := filepath.Join(userHomeDir, ".kube")
	files, err := os.ReadDir(kubeFolderPath)
	if err != nil {
		impl.logger.Errorw("error occurred while reading dir ", "path", kubeFolderPath, "err", err)
		return "", nil, err
	}
	return kubeFolderPath, files, nil
}

func (impl *KubeConfigFileSyncerImpl) processKubeconfigFile(kubefolder, fileName string) {
	filePath := path.Join(kubefolder, fileName)
	config, err := clientcmd.LoadFromFile(filePath)
	if err != nil {
		impl.logger.Errorw("error occurred while reading file", "filePath", filePath, "error", err)
		return
	}
	contexts := config.Contexts
	for currentContextName := range contexts {
		currentCtx := contexts[currentContextName]
		for name, cluster := range config.Clusters {
			if name == currentCtx.Cluster {
				clusterName := fmt.Sprintf("%s__%s", fileName, name)
				clusterExists, clusterBean, err := impl.checkClusterExists(clusterName)
				if err != nil {
					continue
				}
				impl.saveOrUpdateCluster(clusterExists, clusterName, clusterBean, cluster, config, currentCtx)
			}
		}
	}
}

func (impl *KubeConfigFileSyncerImpl) checkClusterExists(clusterName string) (bool, *ClusterBean, error) {
	clusterBean, err := impl.clusterService.FindOneActive(clusterName)
	if err != nil {
		impl.logger.Errorw("error occurred while finding cluster", "clusterName", clusterName)
		return false, nil, err
	}
	alreadyExists := clusterBean != nil && clusterBean.Id > 0
	return alreadyExists, clusterBean, nil
}

func (impl *KubeConfigFileSyncerImpl) saveOrUpdateCluster(clusterExists bool, clusterName string, clusterBean *ClusterBean, cluster *api.Cluster, config *api.Config, currentCtx *api.Context) {
	newClusterBean := impl.getClusterBean(clusterName, cluster, config, currentCtx)
	if clusterExists {
		impl.updateCluster(clusterBean, newClusterBean)
	} else {
		_, err := impl.clusterService.Save(context.Background(), newClusterBean, bean.AdminUserId)
		if err != nil {
			impl.logger.Errorw("error occurred while saving cluster data", "err", err)
		}
	}

}

func (impl *KubeConfigFileSyncerImpl) updateCluster(existingClusterBean *ClusterBean, newClusterBean *ClusterBean) {
	overriddenClusterBean := impl.compareClusterBeans(existingClusterBean, newClusterBean)
	if overriddenClusterBean != nil {
		_, err := impl.clusterService.Update(context.Background(), overriddenClusterBean, bean.AdminUserId)
		if err != nil {
			impl.logger.Errorw("error occurred while updating cluster data", "err", err)
		}
	}
}

func (impl *KubeConfigFileSyncerImpl) getClusterBean(name string, cluster *api.Cluster, config *api.Config, currentCtx *api.Context) *ClusterBean {
	authInfos := config.AuthInfos
	info := currentCtx.AuthInfo
	authInfo := authInfos[info]
	if authInfo == nil {
		return nil
	}
	token := authInfo.Token
	clusterConfig := map[string]string{k8s.BearerToken: token}
	insecureSkipTLSVerify := cluster.InsecureSkipTLSVerify
	if !insecureSkipTLSVerify {
		clusterConfig[k8s.TlsKey] = string(authInfo.ClientKeyData)
		clusterConfig[k8s.CertData] = string(authInfo.ClientCertificateData)
		clusterConfig[k8s.CertificateAuthorityData] = string(cluster.CertificateAuthorityData)
	}
	clusterBean := &ClusterBean{
		ClusterName:           name,
		Active:                true,
		ServerUrl:             cluster.Server,
		InsecureSkipTLSVerify: insecureSkipTLSVerify,
		Config:                clusterConfig,
	}
	return clusterBean
}

func (impl *KubeConfigFileSyncerImpl) compareClusterBeans(existingBean *ClusterBean, newClusterBean *ClusterBean) *ClusterBean {
	if existingBean == nil {
		impl.logger.Warnw("existing cluster passed is nul so ignoring bean")
		return nil
	}
	config := existingBean.Config
	newClusterConfig := newClusterBean.Config
	existingToken := config[k8s.BearerToken]
	newToken := newClusterConfig[k8s.BearerToken]
	if existingBean.ServerUrl != newClusterBean.ServerUrl || existingToken != newToken {
		newClusterBean.Id = existingBean.Id
		newClusterBean.ServerUrl = existingBean.ServerUrl
		newClusterBean.Config = existingBean.Config
		newClusterBean.Active = true
		return newClusterBean
	}
	return nil
}
