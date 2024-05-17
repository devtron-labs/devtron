package application

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/client/proxy"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"go.uber.org/zap"
	"net"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InterClusterServiceCommunicationHandler interface {
	GetClusterServiceProxyHandler(ctx context.Context, clusterServiceKey ClusterServiceKey) (*httputil.ReverseProxy, error)
	GetClusterServiceProxyPort(ctx context.Context, clusterServiceKey ClusterServiceKey) (int, error)
	GetK8sApiProxyHandler(ctx context.Context, clusterId int) (*httputil.ReverseProxy, error)
}

type InterClusterServiceCommunicationHandlerImpl struct {
	logger              *zap.SugaredLogger
	portForwardManager  PortForwardManager
	clusterServiceCache map[ClusterServiceKey]*ProxyServerMetadata
	lock                *sync.Mutex
}

type ProxyServerMetadata struct {
	forwardedPort         int
	lastActivityTimestamp time.Time
	proxyServer           *httputil.ReverseProxy
}

func NewInterClusterServiceCommunicationHandlerImpl(logger *zap.SugaredLogger, portForwardManager PortForwardManager) (*InterClusterServiceCommunicationHandlerImpl, error) {
	clusterServiceCache := map[ClusterServiceKey]*ProxyServerMetadata{}
	return &InterClusterServiceCommunicationHandlerImpl{logger: logger, portForwardManager: portForwardManager, clusterServiceCache: clusterServiceCache, lock: &sync.Mutex{}}, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) GetClusterServiceProxyHandler(ctx context.Context, clusterServiceKey ClusterServiceKey) (*httputil.ReverseProxy, error) {
	reverseProxy, err := impl.getProxyMetadata(ctx, clusterServiceKey)
	if err != nil {
		return nil, err
	}

	return reverseProxy.proxyServer, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) GetClusterServiceProxyPort(ctx context.Context, clusterServiceKey ClusterServiceKey) (int, error) {
	proxyMetadata, err := impl.getProxyMetadata(ctx, clusterServiceKey)
	if err != nil {
		return 0, err
	}
	return proxyMetadata.forwardedPort, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) GetK8sApiProxyHandler(ctx context.Context, clusterId int) (*httputil.ReverseProxy, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	dummyClusterKey := NewClusterServiceKey(clusterId, "dummy", "dummy", "dummy")
	proxyServerMetadata, ok := impl.clusterServiceCache[dummyClusterKey]
	if ok {
		return proxyServerMetadata.proxyServer, nil
	}
	k8sProxyPort, err := impl.portForwardManager.StartK8sProxy(ctx, clusterId)
	if err != nil {
		return nil, err
	}
	proxyTransport := proxy.NewProxyTransport()
	serverAddr := fmt.Sprintf("http://localhost:%d", k8sProxyPort)
	proxyServer := proxy.GetProxyServerWithPathTrimFunc(serverAddr, proxyTransport, "", "", NewClusterServiceActivityLogger(dummyClusterKey, impl.callback), func(urlPath string) string {
		return urlPath
	}) // TODO Fix this

	reverseProxyMetadata := &ProxyServerMetadata{forwardedPort: k8sProxyPort, proxyServer: proxyServer}
	impl.clusterServiceCache[dummyClusterKey] = reverseProxyMetadata
	go func() {
		// inactivity handling
		for {
			proxyServerMetadata := impl.clusterServiceCache[dummyClusterKey]
			lastActivityTimestamp := proxyServerMetadata.lastActivityTimestamp
			if !lastActivityTimestamp.IsZero() && (time.Since(lastActivityTimestamp) > 60*time.Second) {
				impl.logger.Infow("stopping forwarded port because of inactivity", "k8sProxyPort", k8sProxyPort)
				forwardedPort := proxyServerMetadata.forwardedPort
				impl.portForwardManager.StopPortForwarding(context.Background(), forwardedPort)
				delete(impl.clusterServiceCache, dummyClusterKey)
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return reverseProxyMetadata.proxyServer, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) getProxyMetadata(ctx context.Context, clusterServiceKey ClusterServiceKey) (*ProxyServerMetadata, error) {
	var reverseProxy *ProxyServerMetadata
	if proxyMetadata, ok := impl.clusterServiceCache[clusterServiceKey]; ok {
		reverseProxy = proxyMetadata
	} else {
		portForwardRequest := bean.PortForwardRequest{
			ClusterId:   clusterServiceKey.GetClusterId(),
			Namespace:   clusterServiceKey.GetNamespace(),
			ServiceName: clusterServiceKey.GetServiceName(),
			TargetPort:  clusterServiceKey.GetServicePort(),
		}
		forwardedPort, err := impl.portForwardManager.ForwardPort(ctx, portForwardRequest)
		if err != nil {
			return &ProxyServerMetadata{}, err
		}
		proxyTransport := proxy.NewProxyTransport()
		serverAddr := fmt.Sprintf("http://localhost:%d", forwardedPort)
		proxyServer := proxy.GetProxyServer(serverAddr, proxyTransport, "orchestrator", "", NewClusterServiceActivityLogger(clusterServiceKey, impl.callback)) // TODO Fix this
		reverseProxy = &ProxyServerMetadata{forwardedPort: forwardedPort, proxyServer: proxyServer}
		impl.clusterServiceCache[clusterServiceKey] = reverseProxy
		go func() {
			// inactivity handling
			for {
				proxyServerMetadata := impl.clusterServiceCache[clusterServiceKey]
				active := portActive("localhost", proxyServerMetadata.forwardedPort)
				lastActivityTimestamp := proxyServerMetadata.lastActivityTimestamp
				if !active || (!lastActivityTimestamp.IsZero() && (time.Since(lastActivityTimestamp) > 60*time.Second)) {
					impl.logger.Infow("stopping forwarded port because of inactivity", "forwardedPort", forwardedPort)
					forwardedPort := proxyServerMetadata.forwardedPort
					impl.portForwardManager.StopPortForwarding(context.Background(), forwardedPort)
					delete(impl.clusterServiceCache, clusterServiceKey)
					return
				}
				time.Sleep(10 * time.Second)
			}
		}()
	}
	return reverseProxy, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) callback(clusterServiceKey ClusterServiceKey) {
	proxyServerMetadata := impl.clusterServiceCache[clusterServiceKey]
	proxyServerMetadata.lastActivityTimestamp = time.Now()
}

type ClusterServiceKey string

func NewClusterServiceKey(clusterId int, serviceName string, namespace string, servicePort string) ClusterServiceKey {
	return ClusterServiceKey(fmt.Sprintf("%d_$_%s_$_%s_$_%s", clusterId, serviceName, namespace, servicePort))
}

func (key ClusterServiceKey) GetClusterId() int {
	clusterIdStringVal := strings.Split(string(key), "_$_")[0]
	clusterId, _ := strconv.Atoi(clusterIdStringVal)
	return clusterId
}

func (key ClusterServiceKey) GetNamespace() string {
	return strings.Split(string(key), "_$_")[2]
}
func (key ClusterServiceKey) GetServiceName() string {
	return strings.Split(string(key), "_$_")[1]
}
func (key ClusterServiceKey) GetServicePort() string {
	return strings.Split(string(key), "_$_")[3]
}

type ClusterServiceActivityLogger struct {
	proxy.RequestActivityLogger
	clusterServiceKey ClusterServiceKey
	callback          func(clusterServiceKey ClusterServiceKey)
}

func NewClusterServiceActivityLogger(clusterServiceKey ClusterServiceKey, callback func(clusterServiceKey ClusterServiceKey)) ClusterServiceActivityLogger {
	return ClusterServiceActivityLogger{clusterServiceKey: clusterServiceKey, callback: callback}
}

func (csal ClusterServiceActivityLogger) LogActivity() {
	csal.callback(csal.clusterServiceKey)
}

func portActive(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second) // Adjust the timeout as needed
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
