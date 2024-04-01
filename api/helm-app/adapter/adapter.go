package adapter

import (
	grpcBean "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
)

func ConvertServerConnectionConfigToProto(dockerBean *types.DockerArtifactStoreBean) *grpcBean.ServerConnectionConfig {
	var registryConnectionConfig *grpcBean.ServerConnectionConfig
	if dockerBean.ServerConnectionConfig != nil {
		connectionMethod := 0
		if dockerBean.ServerConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			connectionMethod = 1
		}
		registryConnectionConfig = &grpcBean.ServerConnectionConfig{
			ConnectionMethod: grpcBean.ServerConnectionMethod(connectionMethod),
		}
		if dockerBean.ServerConnectionConfig.ProxyConfig != nil && dockerBean.ServerConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodProxy {
			proxyConfig := dockerBean.ServerConnectionConfig.ProxyConfig
			registryConnectionConfig.ProxyConfig = &grpcBean.ProxyConfig{
				ProxyUrl: proxyConfig.ProxyUrl,
			}
		}
		if dockerBean.ServerConnectionConfig.SSHTunnelConfig != nil && dockerBean.ServerConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			sshTunnelConfig := dockerBean.ServerConnectionConfig.SSHTunnelConfig
			registryConnectionConfig.SSHTunnelConfig = &grpcBean.SSHTunnelConfig{
				SSHServerAddress: sshTunnelConfig.SSHServerAddress,
				SSHUsername:      sshTunnelConfig.SSHUsername,
				SSHPassword:      sshTunnelConfig.SSHPassword,
				SSHAuthKey:       sshTunnelConfig.SSHAuthKey,
			}
		}
	}
	return registryConnectionConfig
}
