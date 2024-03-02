package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
)

func ConvertServerConnectionConfigToProto(dockerBean *types.DockerArtifactStoreBean) *bean2.ServerConnectionConfig {
	var registryConnectionConfig *bean2.ServerConnectionConfig
	if dockerBean.RegistryConnectionConfig != nil {
		connectionMethod := 0
		if dockerBean.RegistryConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodSSH {
			connectionMethod = 1
		}
		registryConnectionConfig = &bean2.ServerConnectionConfig{
			ConnectionMethod: bean2.ServerConnectionMethod(connectionMethod),
		}
		if dockerBean.RegistryConnectionConfig.ProxyConfig != nil && dockerBean.RegistryConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodProxy {
			proxyConfig := dockerBean.RegistryConnectionConfig.ProxyConfig
			registryConnectionConfig.ProxyConfig = &bean2.ProxyConfig{
				ProxyUrl: proxyConfig.ProxyUrl,
			}
		}
		if dockerBean.RegistryConnectionConfig.SSHTunnelConfig != nil && dockerBean.RegistryConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodSSH {
			sshTunnelConfig := dockerBean.RegistryConnectionConfig.SSHTunnelConfig
			registryConnectionConfig.SSHTunnelConfig = &bean2.SSHTunnelConfig{
				SSHServerAddress: sshTunnelConfig.SSHServerAddress,
				SSHUsername:      sshTunnelConfig.SSHUsername,
				SSHPassword:      sshTunnelConfig.SSHPassword,
				SSHAuthKey:       sshTunnelConfig.SSHAuthKey,
			}
		}
	}
	return registryConnectionConfig
}
