package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"log"
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

	// add function into cron
	//TODO: get cron time from env var
	_, err := newCron.AddFunc(fmt.Sprint("@every 15m"), clusterCronServiceImpl.GetAndUpdateClusterConnectionStatus)
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
		restConfig, err := impl.k8sApplicationService.GetRestConfigByCluster(cluster)
		if err != nil {
			impl.logger.Errorw("error in getting restConfig by cluster", "err", err, "clusterId", cluster.Id)
			mutex.Lock()
			respMap[cluster.Id] = err
			mutex.Unlock()
			continue
		}
		k8sClientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
			mutex.Lock()
			respMap[cluster.Id] = err
			mutex.Unlock()
			continue
		}
		go GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, cluster.Id, respMap, wg, mutex)
	}
	wg.Wait()
	impl.HandleErrorInClusterConnections(respMap)
	return
}

func GetAndUpdateConnectionStatusForOneCluster(k8sClientSet *kubernetes.Clientset, clusterId int, respMap map[int]error, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()
	//using livez path as healthz path is deprecated
	path := "/livez"
	response, err := k8sClientSet.Discovery().RESTClient().Get().AbsPath(path).DoRaw(context.Background())
	log.Println("received response for cluster livez status", "response", string(response), "err", err, "clusterId", clusterId)
	if err == nil && string(response) != "ok" {
		err = fmt.Errorf("ErrorNotOk : response != 'ok' : %s", string(response))
	}
	mutex.Lock()
	respMap[clusterId] = err
	mutex.Unlock()
	return
}

func (impl *ClusterCronServiceImpl) HandleErrorInClusterConnections(respMap map[int]error) {
	for clusterId, err := range respMap {
		errorInConnecting := ""
		if err != nil {
			errorInConnecting = err.Error()
		}
		//updating cluster connection status
		errInUpdating := impl.clusterRepository.UpdateClusterConnectionStatus(clusterId, errorInConnecting)
		if errInUpdating != nil {
			impl.logger.Errorw("error in updating cluster connection status", "err", err, "clusterId", clusterId, "errorInConnecting", errorInConnecting)
		}
	}
}
