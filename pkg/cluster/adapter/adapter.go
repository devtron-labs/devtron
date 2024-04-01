package adapter

import (
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	serverConnectionBean "github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	serverConnectionRepository "github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"time"
)

func ConvertClusterToNewCluster(model *repository.Cluster) *repository.Cluster {
	if len(model.ProxyUrl) > 0 || model.ToConnectWithSSHTunnel {
		// converting old to new
		connectionConfig := &serverConnectionRepository.RemoteConnectionConfig{
			Id:               model.ServerConnectionConfigId,
			ProxyUrl:         model.ProxyUrl,
			SSHServerAddress: model.SSHTunnelServerAddress,
			SSHUsername:      model.SSHTunnelUser,
			SSHPassword:      model.SSHTunnelPassword,
			SSHAuthKey:       model.SSHTunnelAuthKey,
			Deleted:          false,
		}
		if len(model.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = serverConnectionBean.RemoteConnectionMethodProxy
		} else if model.ToConnectWithSSHTunnel {
			connectionConfig.ConnectionMethod = serverConnectionBean.RemoteConnectionMethodSSH
		}
		model.ServerConnectionConfig = connectionConfig
		// reset old config
		model.ProxyUrl = ""
		model.SSHTunnelUser = ""
		model.SSHTunnelPassword = ""
		model.SSHTunnelServerAddress = ""
		model.SSHTunnelAuthKey = ""
		model.ToConnectWithSSHTunnel = false
	}
	return model
}

func ConvertClusterBeanToNewClusterBean(clusterBean *bean.ClusterBean) *bean.ClusterBean {
	if len(clusterBean.ProxyUrl) > 0 || clusterBean.ToConnectWithSSHTunnel {
		// converting old bean to new bean
		connectionConfig := &serverConnectionBean.RemoteConnectionConfigBean{}
		if len(clusterBean.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = serverConnectionBean.RemoteConnectionMethodProxy
			connectionConfig.ProxyConfig = &serverConnectionBean.ProxyConfig{
				ProxyUrl: clusterBean.ProxyUrl,
			}
		}
		if clusterBean.ToConnectWithSSHTunnel && clusterBean.SSHTunnelConfig != nil {
			connectionConfig.ConnectionMethod = serverConnectionBean.RemoteConnectionMethodSSH
			connectionConfig.SSHTunnelConfig = &serverConnectionBean.SSHTunnelConfig{
				SSHServerAddress: clusterBean.SSHTunnelConfig.SSHServerAddress,
				SSHUsername:      clusterBean.SSHTunnelConfig.User,
				SSHPassword:      clusterBean.SSHTunnelConfig.Password,
				SSHAuthKey:       clusterBean.SSHTunnelConfig.AuthKey,
			}
		}
		clusterBean.ServerConnectionConfig = connectionConfig
	}
	return clusterBean
}

func ConvertNewClusterBeanToOldClusterBean(clusterBean *bean.ClusterBean) *bean.ClusterBean {
	if clusterBean.ServerConnectionConfig != nil {
		if clusterBean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodProxy &&
			clusterBean.ServerConnectionConfig.ProxyConfig != nil {
			clusterBean.ProxyUrl = clusterBean.ServerConnectionConfig.ProxyConfig.ProxyUrl
		}
		if clusterBean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH &&
			clusterBean.ServerConnectionConfig.SSHTunnelConfig != nil {
			clusterBean.ToConnectWithSSHTunnel = true
			clusterBean.SSHTunnelConfig.SSHServerAddress = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHServerAddress
			clusterBean.SSHTunnelConfig.User = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHUsername
			clusterBean.SSHTunnelConfig.Password = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword
			clusterBean.SSHTunnelConfig.AuthKey = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey
		}
	}
	return clusterBean
}

func ConvertClusterBeanToCluster(clusterBean *bean.ClusterBean, userId int32) *repository.Cluster {

	model := &repository.Cluster{}

	model.ClusterName = clusterBean.ClusterName
	model.Active = true
	model.ServerUrl = clusterBean.ServerUrl
	model.Config = clusterBean.Config
	model.PrometheusEndpoint = clusterBean.PrometheusUrl
	model.InsecureSkipTlsVerify = clusterBean.InsecureSkipTLSVerify

	if clusterBean.PrometheusAuth != nil {
		model.PUserName = clusterBean.PrometheusAuth.UserName
		model.PPassword = clusterBean.PrometheusAuth.Password
		model.PTlsClientCert = clusterBean.PrometheusAuth.TlsClientCert
		model.PTlsClientKey = clusterBean.PrometheusAuth.TlsClientKey
	}

	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()

	var connectionMethod serverConnectionBean.RemoteConnectionMethod
	var connectionConfig *serverConnectionRepository.RemoteConnectionConfig
	if clusterBean.ServerConnectionConfig != nil {
		// if FE provided new bean
		connectionMethod = clusterBean.ServerConnectionConfig.ConnectionMethod
		connectionConfig = &serverConnectionRepository.RemoteConnectionConfig{
			ConnectionMethod: connectionMethod,
		}
		if clusterBean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodProxy &&
			clusterBean.ServerConnectionConfig.ProxyConfig != nil {
			connectionConfig.ProxyUrl = clusterBean.ServerConnectionConfig.ProxyConfig.ProxyUrl
		}
		if clusterBean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH &&
			clusterBean.ServerConnectionConfig.SSHTunnelConfig != nil {
			connectionConfig.SSHServerAddress = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHServerAddress
			connectionConfig.SSHUsername = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHUsername
			connectionConfig.SSHPassword = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword
			connectionConfig.SSHAuthKey = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey
		}
	} else if len(clusterBean.ProxyUrl) > 0 || clusterBean.ToConnectWithSSHTunnel {
		// if FE provided old bean
		if len(clusterBean.ProxyUrl) > 0 {
			connectionMethod = serverConnectionBean.RemoteConnectionMethodProxy
		} else if clusterBean.ToConnectWithSSHTunnel {
			connectionMethod = serverConnectionBean.RemoteConnectionMethodSSH
		}
		connectionConfig = &serverConnectionRepository.RemoteConnectionConfig{
			ConnectionMethod: connectionMethod,
			ProxyUrl:         clusterBean.ProxyUrl,
		}
		if clusterBean.SSHTunnelConfig != nil {
			sshTunnelConfig := clusterBean.SSHTunnelConfig
			connectionConfig.SSHServerAddress = sshTunnelConfig.SSHServerAddress
			connectionConfig.SSHAuthKey = sshTunnelConfig.AuthKey
			connectionConfig.SSHPassword = sshTunnelConfig.Password
			connectionConfig.SSHUsername = sshTunnelConfig.User
		}
		// reset old config
		model.ProxyUrl = ""
		model.SSHTunnelUser = ""
		model.SSHTunnelPassword = ""
		model.SSHTunnelServerAddress = ""
		model.SSHTunnelAuthKey = ""
		model.ToConnectWithSSHTunnel = false
	}
	model.ServerConnectionConfig = connectionConfig
	return model
}

func GetClusterBean(model repository.Cluster) bean.ClusterBean {
	model = *ConvertClusterToNewCluster(&model) // repo model is converted according to new struct
	clusterBean := bean.ClusterBean{}
	clusterBean.Id = model.Id
	clusterBean.ClusterName = model.ClusterName
	clusterBean.ServerUrl = model.ServerUrl
	clusterBean.PrometheusUrl = model.PrometheusEndpoint
	clusterBean.AgentInstallationStage = model.AgentInstallationStage
	clusterBean.Active = model.Active
	clusterBean.Config = model.Config
	clusterBean.K8sVersion = model.K8sVersion
	clusterBean.InsecureSkipTLSVerify = model.InsecureSkipTlsVerify
	clusterBean.IsVirtualCluster = model.IsVirtualCluster
	clusterBean.ErrorInConnecting = model.ErrorInConnecting
	clusterBean.PrometheusAuth = &bean.PrometheusAuth{
		UserName:      model.PUserName,
		Password:      model.PPassword,
		TlsClientCert: model.PTlsClientCert,
		TlsClientKey:  model.PTlsClientKey,
	}
	if model.ServerConnectionConfig != nil {
		clusterBean.ServerConnectionConfig = &serverConnectionBean.RemoteConnectionConfigBean{
			RemoteConnectionConfigId: model.ServerConnectionConfigId,
			ConnectionMethod:         model.ServerConnectionConfig.ConnectionMethod,
			ProxyConfig: &serverConnectionBean.ProxyConfig{
				ProxyUrl: model.ServerConnectionConfig.ProxyUrl,
			},
			SSHTunnelConfig: &serverConnectionBean.SSHTunnelConfig{
				SSHServerAddress: model.ServerConnectionConfig.SSHServerAddress,
				SSHUsername:      model.ServerConnectionConfig.SSHUsername,
				SSHPassword:      model.ServerConnectionConfig.SSHPassword,
				SSHAuthKey:       model.ServerConnectionConfig.SSHAuthKey,
			},
		}
	}
	return clusterBean
}
