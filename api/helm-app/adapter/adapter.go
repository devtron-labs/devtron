package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
)

func ConvertServerConnectionConfigToProto(dockerBean *types.DockerArtifactStoreBean) *bean2.ServerConnectionConfig {
	var registryConnectionConfig *bean2.ServerConnectionConfig
	if dockerBean.ServerConnectionConfig != nil {
		connectionMethod := 0
		if dockerBean.ServerConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodSSH {
			connectionMethod = 1
		}
		registryConnectionConfig = &bean2.ServerConnectionConfig{
			ConnectionMethod: bean2.ServerConnectionMethod(connectionMethod),
		}
		if dockerBean.ServerConnectionConfig.ProxyConfig != nil && dockerBean.ServerConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodProxy {
			proxyConfig := dockerBean.ServerConnectionConfig.ProxyConfig
			registryConnectionConfig.ProxyConfig = &bean2.ProxyConfig{
				ProxyUrl: proxyConfig.ProxyUrl,
			}
		}
		if dockerBean.ServerConnectionConfig.SSHTunnelConfig != nil && dockerBean.ServerConnectionConfig.ConnectionMethod == bean.ServerConnectionMethodSSH {
			sshTunnelConfig := dockerBean.ServerConnectionConfig.SSHTunnelConfig
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
