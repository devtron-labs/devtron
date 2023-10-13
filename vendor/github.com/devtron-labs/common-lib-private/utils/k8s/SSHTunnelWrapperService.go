package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib-private/sshTunnel"
	"go.uber.org/zap"
	"net/url"
	"strconv"
	"sync"
	"time"
)

//This service communicates with our sshTunnel library present at devtron-labs/common-lib-private/sshTunnel

type SSHTunnelWrapperService interface {
	StartUpdateConnectionForCluster(cluster *ClusterConfig) (int, error)
	GetPortUsedForACluster(clusterConfig *ClusterConfig) (int, error)
	CleanupForVerificationCluster(clusterName string)
}

type SSHTunnelWrapperServiceImpl struct {
	logger                           *zap.SugaredLogger
	portMapMutex                     *sync.Mutex
	portMap                          map[int]bool //map of port being used or not
	clusterMapMutex                  *sync.Mutex
	clusterConnectionMap             map[int]*ConnectionDetail    //map of clusterId and connection detail
	verificationClusterConnectionMap map[string]*ConnectionDetail //map of cluster name and connection detail
}

func NewSSHTunnelWrapperServiceImpl(logger *zap.SugaredLogger) (*SSHTunnelWrapperServiceImpl, error) {
	impl := &SSHTunnelWrapperServiceImpl{
		logger:                           logger,
		portMapMutex:                     &sync.Mutex{},
		portMap:                          getPortMap(),
		clusterMapMutex:                  &sync.Mutex{},
		clusterConnectionMap:             make(map[int]*ConnectionDetail),
		verificationClusterConnectionMap: make(map[string]*ConnectionDetail),
	}
	return impl, nil
}

type ConnectionDetail struct {
	portUsed   int
	connection *sshTunnel.SSHTunnel
}

// StartUpdateConnectionForCluster takes clusterId and returns the port being used for connection and error if an
func (impl *SSHTunnelWrapperServiceImpl) StartUpdateConnectionForCluster(cluster *ClusterConfig) (int, error) {
	portUsed := 0
	availablePort, err := impl.getAvailablePort()
	if err != nil {
		impl.logger.Errorw("error in getting port for SSH Tunnel connection", "err", err, "clusterId", cluster.ClusterId)
		return portUsed, err
	}
	//using port so that it gets blocked for us
	blockSuccess := impl.usePort(availablePort)
	if !blockSuccess {
		return portUsed, fmt.Errorf("error in getting port for connecting tunnel")
	} else {
		portUsed = availablePort
	}
	//our server url is actually the remote we are trying to connect, splitting it to get host and port
	remoteAddress, remotePort, err := impl.extractHostAndPostFromUrl(cluster.Host)
	if err != nil {
		impl.logger.Errorw("error in extracting host and port from cluster host address", "err", err, "clusterHost", cluster.Host)
		return portUsed, err
	}
	serverAddress, _, err := impl.extractHostAndPostFromUrl(cluster.SSHTunnelServerAddress)
	if err != nil {
		impl.logger.Errorw("error in extracting host and port from cluster host address", "err", err, "clusterHost", cluster.Host)
		return portUsed, err
	}
	//blocking is successful now we can initialise connection
	tunnel := sshTunnel.NewSSHTunnel(cluster.SSHTunnelUser, cluster.SSHTunnelPassword, cluster.SSHTunnelAuthKey,
		serverAddress, remoteAddress, remotePort, portUsed, DefaultTimeoutDuration)

	connectionDetail := &ConnectionDetail{
		portUsed:   portUsed,
		connection: tunnel,
	}
	impl.clusterMapMutex.Lock()
	if !cluster.ToConnectForClusterVerification {
		if oldConnectionDetail := impl.clusterConnectionMap[cluster.ClusterId]; oldConnectionDetail != nil {
			previousPort := oldConnectionDetail.portUsed
			oldConnectionDetail.connection.Stop()
			impl.freePort(previousPort)
		}
		impl.clusterConnectionMap[cluster.ClusterId] = connectionDetail
	} else {
		if oldConnectionDetail := impl.verificationClusterConnectionMap[cluster.ClusterName]; oldConnectionDetail != nil {
			previousPort := oldConnectionDetail.portUsed
			oldConnectionDetail.connection.Stop()
			impl.freePort(previousPort)
		}
		impl.verificationClusterConnectionMap[cluster.ClusterName] = connectionDetail
	}
	defer impl.deferForStartUpdateConnectionForCluster()

	//starting the tunnel
	go func() {
		err = tunnel.Start(context.Background())
		if err != nil {
			impl.logger.Errorw("error in starting tunnel", "err", err, "clusterId", cluster.ClusterId)
			impl.clusterMapMutex.Lock()
			if !cluster.ToConnectForClusterVerification {
				//need to free port
				if currentConnection := impl.clusterConnectionMap[cluster.ClusterId]; currentConnection != nil {
					portToBeFreed := currentConnection.portUsed
					if portToBeFreed == portUsed {
						currentConnection.connection.Stop()
						impl.freePort(portToBeFreed)
						impl.clusterConnectionMap[cluster.ClusterId] = nil
					}
				}
			} else {
				if currentConnection := impl.verificationClusterConnectionMap[cluster.ClusterName]; currentConnection != nil {
					portToBeFreed := currentConnection.portUsed
					if portToBeFreed == portUsed {
						currentConnection.connection.Stop()
						impl.freePort(portToBeFreed)
						impl.verificationClusterConnectionMap[cluster.ClusterName] = nil
					}
				}
			}
			impl.clusterMapMutex.Unlock()
		}
	}()
	time.Sleep(10 * time.Second)
	return portUsed, nil
}

// created a method to remove the confusion of defer LIFO nature, if more defer statements are added in future
func (impl *SSHTunnelWrapperServiceImpl) deferForStartUpdateConnectionForCluster() {
	impl.clusterMapMutex.Unlock()
}

func (impl *SSHTunnelWrapperServiceImpl) GetPortUsedForACluster(clusterConfig *ClusterConfig) (int, error) {
	portUsed := 0
	if !clusterConfig.ToConnectForClusterVerification {
		connectionDetail := impl.clusterConnectionMap[clusterConfig.ClusterId]
		if connectionDetail != nil {
			portUsed = connectionDetail.portUsed
		}
	} else {
		//currently we are not handling concurrent request case
		connectionDetail := impl.verificationClusterConnectionMap[clusterConfig.ClusterName]
		if connectionDetail != nil {
			portUsed = connectionDetail.portUsed
		}
	}
	if portUsed == 0 {
		var err error
		portUsed, err = impl.StartUpdateConnectionForCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in connecting with cluster through SSH tunnel", "err", err, "clusterId")
			return portUsed, err
		}
	}
	return portUsed, nil
}

func (impl *SSHTunnelWrapperServiceImpl) CleanupForVerificationCluster(clusterName string) {
	if currentConnection := impl.verificationClusterConnectionMap[clusterName]; currentConnection != nil {
		portToBeFreed := currentConnection.portUsed
		currentConnection.connection.Stop()
		impl.freePort(portToBeFreed)
		impl.verificationClusterConnectionMap[clusterName] = nil
	}
}

func (impl *SSHTunnelWrapperServiceImpl) extractHostAndPostFromUrl(urlStr string) (host string, port int, err error) {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		impl.logger.Errorw("error in parsing url", "err", err, "url", urlStr)
		return "", 0, err
	}
	portStr := u.Port()
	if len(portStr) > 0 {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			impl.logger.Errorw("error in converting port string to int", "err", err, "port", portStr)
			return "", 0, err
		}
	}
	//adding scheme less host path
	host += u.Hostname()
	return host, port, nil
}
func (impl *SSHTunnelWrapperServiceImpl) usePort(port int) bool {
	impl.portMapMutex.Lock()
	defer impl.portMapMutex.Unlock()
	if impl.portMap[port] {
		//port already being used, unsuccessful port update
		return false
	} else {
		impl.portMap[port] = true
		return true
	}
}

func (impl *SSHTunnelWrapperServiceImpl) freePort(port int) {
	impl.portMapMutex.Lock()
	defer impl.portMapMutex.Unlock()
	delete(impl.portMap, port)
}

func getPortMap() map[int]bool {
	m := make(map[int]bool, EndingDynamicPort-StartingDynamicPort+1)
	for i := StartingDynamicPort; i <= EndingDynamicPort; i++ {
		if i == KubelinkDefaultPort { //skipping kubelink port
			continue
		}
		m[i] = false
	}
	return m
}

func (impl *SSHTunnelWrapperServiceImpl) getAvailablePort() (int, error) {
	for key := StartingDynamicPort; key <= EndingDynamicPort; key++ {
		if key == KubelinkDefaultPort { //skipping kubelink port
			continue
		}
		if !impl.portMap[key] {
			return key, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}
