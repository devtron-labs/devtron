package k8s

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"sync"
)

type ClusterCronService interface {
}

type ClusterCronServiceImpl struct {
	logger                *zap.SugaredLogger
	clusterService        cluster.ClusterService
	k8sApplicationService K8sApplicationService
	clusterRepository     clusterRepository.ClusterRepository
}

type ClusterStatusConfig struct {
	ClusterStatusCronTime int `env:"CLUSTER_STATUS_CRON_TIME" envDefault:"15"`
}

func NewClusterCronServiceImpl(logger *zap.SugaredLogger, clusterService cluster.ClusterService,
	k8sApplicationService K8sApplicationService, clusterRepository clusterRepository.ClusterRepository) (*ClusterCronServiceImpl, error) {
	clusterCronServiceImpl := &ClusterCronServiceImpl{
		logger:                logger,
		clusterService:        clusterService,
		k8sApplicationService: k8sApplicationService,
		clusterRepository:     clusterRepository,
	}
	// initialise cron
	newCron := cron.New(cron.WithChain())
	newCron.Start()
	cfg := &ClusterStatusConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server cluster status config: " + err.Error())
	}
	// add function into cron
	_, err = newCron.AddFunc(fmt.Sprintf("@every %dm", cfg.ClusterStatusCronTime), clusterCronServiceImpl.GetAndUpdateClusterConnectionStatus)
	if err != nil {
		fmt.Println("error in adding cron function into cluster cron service")
		return clusterCronServiceImpl, err
	}
	return clusterCronServiceImpl, nil
}

func (impl *ClusterCronServiceImpl) GetAndUpdateClusterConnectionStatus() {
	impl.logger.Debug("starting cluster connection status fetch thread")
	defer impl.logger.Debug("stopped cluster connection status fetch thread")

	//getting all clusters
	clusters, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(clusters))
	mutex := &sync.Mutex{}
	//map of clusterId and error in its connection check process
	respMap := make(map[int]error)
	for _, cluster := range clusters {
		// getting restConfig and clientSet outside the goroutine because we don't want to call goroutine func with receiver function
		restConfig, err := impl.k8sApplicationService.GetRestConfigByCluster(context.Background(), cluster)
		if err != nil {
			impl.logger.Errorw("error in getting restConfig by cluster", "err", err, "clusterId", cluster.Id)
			mutex.Lock()
			respMap[cluster.Id] = err
			mutex.Unlock()
			continue
		}
		k8sHttpClient, err := util.OverrideK8sHttpClientWithTracer(restConfig)
		if err != nil {
			continue
		}
		k8sClientSet, err := kubernetes.NewForConfigAndClient(restConfig, k8sHttpClient)
		if err != nil {
			impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
			mutex.Lock()
			respMap[cluster.Id] = err
			mutex.Unlock()
			continue
		}
		go impl.GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, cluster.Id, respMap, wg, mutex)
	}
	wg.Wait()
	_ = impl.clusterService.HandleErrorInClusterConnections(respMap)
	return
}

func (impl *ClusterCronServiceImpl) GetAndUpdateConnectionStatusForOneCluster(k8sClientSet *kubernetes.Clientset, clusterId int, respMap map[int]error, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()
	err := impl.k8sApplicationService.FetchConnectionStatusForCluster(k8sClientSet, clusterId)
	mutex.Lock()
	respMap[clusterId] = err
	mutex.Unlock()
	return
}
