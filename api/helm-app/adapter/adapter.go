package adapter

import (
	grpcBean "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
)

func ConvertRemoteConnectionConfigToProto(dockerBean *types.DockerArtifactStoreBean) *grpcBean.RemoteConnectionConfig {
	var registryConnectionConfig *grpcBean.RemoteConnectionConfig
	if dockerBean.RemoteConnectionConfig != nil {
		connectionMethod := 0
		if dockerBean.RemoteConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			connectionMethod = 1
		}
		registryConnectionConfig = &grpcBean.RemoteConnectionConfig{
			ConnectionMethod: grpcBean.RemoteConnectionMethod(connectionMethod),
		}
		if dockerBean.RemoteConnectionConfig.ProxyConfig != nil && dockerBean.RemoteConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodProxy {
			proxyConfig := dockerBean.RemoteConnectionConfig.ProxyConfig
			registryConnectionConfig.ProxyConfig = &grpcBean.ProxyConfig{
				ProxyUrl: proxyConfig.ProxyUrl,
			}
		}
		if dockerBean.RemoteConnectionConfig.SSHTunnelConfig != nil && dockerBean.RemoteConnectionConfig.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			sshTunnelConfig := dockerBean.RemoteConnectionConfig.SSHTunnelConfig
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
