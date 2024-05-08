package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/proxy"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"go.uber.org/zap"
	"net/http"
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
	handleK8sApiProxyError(clusterId int)
}

type InterClusterServiceCommunicationHandlerImpl struct {
	logger              *zap.SugaredLogger
	portForwardManager  PortForwardManager
	clusterServiceCache map[ClusterServiceKey]*ProxyServerMetadata
	k8sApiProxyCache    map[int]*ProxyServerMetadata
	lock                *sync.Mutex
	cfg                 *bean.InterClusterCommunicationConfig
}

type ProxyServerMetadata struct {
	forwardedPort         int
	lastActivityTimestamp time.Time
	proxyServer           *httputil.ReverseProxy
}

func NewInterClusterServiceCommunicationHandlerImpl(logger *zap.SugaredLogger, portForwardManager PortForwardManager) (*InterClusterServiceCommunicationHandlerImpl, error) {
	cfg := &bean.InterClusterCommunicationConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error occurred while parsing config ", "err", err)
	}

	clusterServiceCache := map[ClusterServiceKey]*ProxyServerMetadata{}
	k8sApiProxyCache := map[int]*ProxyServerMetadata{}
	return &InterClusterServiceCommunicationHandlerImpl{logger: logger, portForwardManager: portForwardManager, clusterServiceCache: clusterServiceCache, lock: &sync.Mutex{}, cfg: cfg, k8sApiProxyCache: k8sApiProxyCache}, nil
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

	proxyServerMetadata, ok := impl.k8sApiProxyCache[clusterId]
	if ok {
		return proxyServerMetadata.proxyServer, nil
	}
	k8sProxyPort, err := impl.portForwardManager.StartK8sProxy(ctx, clusterId, impl.handleK8sApiProxyError)
	if err != nil {
		return nil, err
	}
	proxyTransport := proxy.NewProxyTransport()
	serverAddr := fmt.Sprintf("http://localhost:%d", k8sProxyPort)
	proxyServer := proxy.GetProxyServerWithPathTrimFunc(serverAddr, proxyTransport, "", "", NewClusterServiceActivityLogger(NewClusterServiceKey(clusterId, "", "", ""), impl.updateLastActivity), func(urlPath string) string {
		return urlPath
	})
	proxyServer.ErrorHandler = impl.handleErrorBeforeResponse

	reverseProxyMetadata := &ProxyServerMetadata{forwardedPort: k8sProxyPort, proxyServer: proxyServer}
	impl.k8sApiProxyCache[clusterId] = reverseProxyMetadata
	go func() {
		// Inactivity Handling
		for {
			if proxyServerMetadata, ok := impl.k8sApiProxyCache[clusterId]; ok {
				lastActivityTimestamp := proxyServerMetadata.lastActivityTimestamp
				if !lastActivityTimestamp.IsZero() && (time.Since(lastActivityTimestamp) > (time.Duration(impl.cfg.ProxyUpTime) * time.Second)) {
					impl.logger.Infow("Stopping K8sProxy because of inactivity", "k8sProxyPort", k8sProxyPort)
					forwardedPort := proxyServerMetadata.forwardedPort
					impl.portForwardManager.StopPortForwarding(context.Background(), forwardedPort)
					return
				}
			} else {
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return reverseProxyMetadata.proxyServer, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) handleErrorBeforeResponse(w http.ResponseWriter, r *http.Request, err error) {
	clusterId, _ := strconv.Atoi(r.Header.Get("Cluster-Id"))
	impl.handleK8sApiProxyError(clusterId)
	errorResponse := bean2.ErrorResponse{
		Kind:    "Status",
		Code:    500,
		Message: "An error occurred. Please try again.",
		Reason:  "Internal Server Error",
	}
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(errorResponse)
}

func (impl *InterClusterServiceCommunicationHandlerImpl) handleK8sApiProxyError(clusterId int) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	if _, ok := impl.k8sApiProxyCache[clusterId]; ok {
		delete(impl.k8sApiProxyCache, clusterId)
	}
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
		proxyServer := proxy.GetProxyServer(serverAddr, proxyTransport, "orchestrator", "", NewClusterServiceActivityLogger(clusterServiceKey, impl.callback)) //TODO Fix this
		reverseProxy = &ProxyServerMetadata{forwardedPort: forwardedPort, proxyServer: proxyServer}
		impl.clusterServiceCache[clusterServiceKey] = reverseProxy
		go func() {
			// inactivity handling
			for {
				proxyServerMetadata := impl.clusterServiceCache[clusterServiceKey]
				lastActivityTimestamp := proxyServerMetadata.lastActivityTimestamp
				if !lastActivityTimestamp.IsZero() && (time.Since(lastActivityTimestamp) > 60*time.Second) {
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
	if proxyServerMetadata, ok := impl.clusterServiceCache[clusterServiceKey]; ok {
		proxyServerMetadata.lastActivityTimestamp = time.Now()
	}
}

func (impl *InterClusterServiceCommunicationHandlerImpl) updateLastActivity(clusterIdString ClusterServiceKey) {
	if proxyServerMetadata, ok := impl.k8sApiProxyCache[clusterIdString.GetClusterId()]; ok {
		proxyServerMetadata.lastActivityTimestamp = time.Now()
	}
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
