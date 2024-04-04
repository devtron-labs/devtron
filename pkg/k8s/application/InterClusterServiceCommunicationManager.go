package application

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/client/proxy"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"go.uber.org/zap"
	"net/http/httputil"
	"strconv"
	"strings"
)

type InterClusterServiceCommunicationHandler interface {
	GetClusterServiceProxyHandler(ctx context.Context, clusterServiceKey ClusterServiceKey) (*httputil.ReverseProxy, error)
	GetClusterServiceProxyPort(ctx context.Context, clusterServiceKey ClusterServiceKey) (string, error)
}

type InterClusterServiceCommunicationHandlerImpl struct {
	logger              *zap.SugaredLogger
	portForwardManager  PortForwardManager
	clusterServiceCache map[ClusterServiceKey]ProxyServerMetadata
}

type ProxyServerMetadata struct {
	forwardedPort string
	proxyServer   *httputil.ReverseProxy
}

func NewInterClusterServiceCommunicationHandlerImpl(logger *zap.SugaredLogger, portForwardManager PortForwardManager) (InterClusterServiceCommunicationHandler, error) {
	clusterServiceCache := map[ClusterServiceKey]ProxyServerMetadata{}
	return &InterClusterServiceCommunicationHandlerImpl{logger: logger, portForwardManager: portForwardManager, clusterServiceCache: clusterServiceCache}, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) GetClusterServiceProxyHandler(ctx context.Context, clusterServiceKey ClusterServiceKey) (*httputil.ReverseProxy, error) {
	reverseProxy, err := impl.getProxyMetadata(ctx, clusterServiceKey)
	if err != nil {
		return nil, err
	}

	return reverseProxy.proxyServer, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) GetClusterServiceProxyPort(ctx context.Context, clusterServiceKey ClusterServiceKey) (string, error) {
	proxyMetadata, err := impl.getProxyMetadata(ctx, clusterServiceKey)
	if err != nil {
		return "", err
	}
	return proxyMetadata.forwardedPort, nil
}

func (impl *InterClusterServiceCommunicationHandlerImpl) getProxyMetadata(ctx context.Context, clusterServiceKey ClusterServiceKey) (ProxyServerMetadata, error) {
	var reverseProxy ProxyServerMetadata
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
			return ProxyServerMetadata{}, err
		}
		proxyTransport := proxy.NewProxyTransport()
		serverAddr := fmt.Sprintf("http://localhost:%s", forwardedPort)
		proxyServer := proxy.GetProxyServer(serverAddr, proxyTransport, "/orchestrator/k8s", "") //TODO Fix this
		reverseProxy = ProxyServerMetadata{forwardedPort: forwardedPort, proxyServer: proxyServer}
		impl.clusterServiceCache[clusterServiceKey] = reverseProxy
	}
	return reverseProxy, nil
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
