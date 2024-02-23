package registry

import (
	"fmt"
	"github.com/devtron-labs/common-lib-private/sshTunnel/bean"
	"github.com/devtron-labs/common-lib-private/utils/ssh"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
)

type RegistryService interface {
	GetHostUrlForSSHTunnelConfiguredRegistry(registry *RegistryConfig) (string, error)
	CreateHttpClientWithProxy(rawProxyUrl string) (*http.Client, error)
	GetProxyEnvForCLI(registry *RegistryConfig) []string
}

type RegistryServiceImpl struct {
	logger                  *zap.SugaredLogger
	sshTunnelWrapperService *ssh.SSHTunnelWrapperServiceImpl
}

func NewRegistryServiceImpl(logger *zap.SugaredLogger, sshTunnelWrapperService *ssh.SSHTunnelWrapperServiceImpl) (*RegistryServiceImpl, error) {
	impl := &RegistryServiceImpl{
		logger:                  logger,
		sshTunnelWrapperService: sshTunnelWrapperService,
	}
	return impl, nil
}

func (impl *RegistryServiceImpl) GetHostUrlForSSHTunnelConfiguredRegistry(registry *RegistryConfig) (string, error) {
	var sshTunnelUrl string
	port, err := impl.sshTunnelWrapperService.GetPortUsedForARegistry(registry)
	if err != nil {
		impl.logger.Errorw("error in getting port of ssh tunnel connected registry", "err", err, "registryId", registry.RegistryId)
		return sshTunnelUrl, err
	}
	sshTunnelUrl = fmt.Sprintf("https://%s:%d", bean.LocalHostAddress, port)
	return sshTunnelUrl, nil
}

func (impl *RegistryServiceImpl) CreateHttpClientWithProxy(rawProxyUrl string) (*http.Client, error) {
	proxyUrl, err := url.Parse(rawProxyUrl)
	if err != nil {
		impl.logger.Errorw("error in parsing proxy url", "err", err, "proxyUrl", rawProxyUrl)
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	return &http.Client{Transport: transport}, err
}

func (impl *RegistryServiceImpl) GetProxyEnvForCLI(registry *RegistryConfig) []string {
	envProxy := append(os.Environ(),
		fmt.Sprintf("HTTP_PROXY=%s", registry.ProxyConfig.ProxyUrl),
		fmt.Sprintf("HTTPS_PROXY=%s", registry.ProxyConfig.ProxyUrl),
		fmt.Sprintf("http_proxy=%s", registry.ProxyConfig.ProxyUrl),
		fmt.Sprintf("https_proxy=%s", registry.ProxyConfig.ProxyUrl),
	)
	return envProxy
}
