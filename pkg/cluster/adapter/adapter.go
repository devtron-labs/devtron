package adapter

import (
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	remoteConnectionBean "github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	remoteConnectionRepository "github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"time"
)

func ConvertClusterToNewCluster(model *repository.Cluster) *repository.Cluster {
	if len(model.ProxyUrl) > 0 || model.ToConnectWithSSHTunnel {
		// converting old to new
		connectionConfig := &remoteConnectionRepository.RemoteConnectionConfig{
			Id:               model.RemoteConnectionConfigId,
			ProxyUrl:         model.ProxyUrl,
			SSHServerAddress: model.SSHTunnelServerAddress,
			SSHUsername:      model.SSHTunnelUser,
			SSHPassword:      model.SSHTunnelPassword,
			SSHAuthKey:       model.SSHTunnelAuthKey,
			Deleted:          false,
		}
		if len(model.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = remoteConnectionBean.RemoteConnectionMethodProxy
		} else if model.ToConnectWithSSHTunnel {
			connectionConfig.ConnectionMethod = remoteConnectionBean.RemoteConnectionMethodSSH
		}
		model.RemoteConnectionConfig = connectionConfig
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
		connectionConfig := &remoteConnectionBean.RemoteConnectionConfigBean{}
		if len(clusterBean.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = remoteConnectionBean.RemoteConnectionMethodProxy
			connectionConfig.ProxyConfig = &remoteConnectionBean.ProxyConfig{
				ProxyUrl: clusterBean.ProxyUrl,
			}
		}
		if clusterBean.ToConnectWithSSHTunnel && clusterBean.SSHTunnelConfig != nil {
			connectionConfig.ConnectionMethod = remoteConnectionBean.RemoteConnectionMethodSSH
			connectionConfig.SSHTunnelConfig = &remoteConnectionBean.SSHTunnelConfig{
				SSHServerAddress: clusterBean.SSHTunnelConfig.SSHServerAddress,
				SSHUsername:      clusterBean.SSHTunnelConfig.User,
				SSHPassword:      clusterBean.SSHTunnelConfig.Password,
				SSHAuthKey:       clusterBean.SSHTunnelConfig.AuthKey,
			}
		}
		clusterBean.RemoteConnectionConfig = connectionConfig
	}
	return clusterBean
}

func ConvertNewClusterBeanToOldClusterBean(clusterBean *bean.ClusterBean) *bean.ClusterBean {
	if clusterBean.RemoteConnectionConfig != nil {
		if clusterBean.RemoteConnectionConfig.ConnectionMethod == remoteConnectionBean.RemoteConnectionMethodProxy &&
			clusterBean.RemoteConnectionConfig.ProxyConfig != nil {
			clusterBean.ProxyUrl = clusterBean.RemoteConnectionConfig.ProxyConfig.ProxyUrl
		}
		if clusterBean.RemoteConnectionConfig.ConnectionMethod == remoteConnectionBean.RemoteConnectionMethodSSH &&
			clusterBean.RemoteConnectionConfig.SSHTunnelConfig != nil {
			clusterBean.ToConnectWithSSHTunnel = true
			clusterBean.SSHTunnelConfig.SSHServerAddress = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHServerAddress
			clusterBean.SSHTunnelConfig.User = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHUsername
			clusterBean.SSHTunnelConfig.Password = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHPassword
			clusterBean.SSHTunnelConfig.AuthKey = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHAuthKey
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

	var connectionMethod remoteConnectionBean.RemoteConnectionMethod
	var connectionConfig *remoteConnectionRepository.RemoteConnectionConfig
	if clusterBean.RemoteConnectionConfig != nil {
		// if FE provided new bean
		connectionMethod = clusterBean.RemoteConnectionConfig.ConnectionMethod
		connectionConfig = &remoteConnectionRepository.RemoteConnectionConfig{
			ConnectionMethod: connectionMethod,
		}
		if clusterBean.RemoteConnectionConfig.ConnectionMethod == remoteConnectionBean.RemoteConnectionMethodProxy &&
			clusterBean.RemoteConnectionConfig.ProxyConfig != nil {
			connectionConfig.ProxyUrl = clusterBean.RemoteConnectionConfig.ProxyConfig.ProxyUrl
		}
		if clusterBean.RemoteConnectionConfig.ConnectionMethod == remoteConnectionBean.RemoteConnectionMethodSSH &&
			clusterBean.RemoteConnectionConfig.SSHTunnelConfig != nil {
			connectionConfig.SSHServerAddress = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHServerAddress
			connectionConfig.SSHUsername = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHUsername
			connectionConfig.SSHPassword = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHPassword
			connectionConfig.SSHAuthKey = clusterBean.RemoteConnectionConfig.SSHTunnelConfig.SSHAuthKey
		}
	} else if len(clusterBean.ProxyUrl) > 0 || clusterBean.ToConnectWithSSHTunnel {
		// if FE provided old bean
		if len(clusterBean.ProxyUrl) > 0 {
			connectionMethod = remoteConnectionBean.RemoteConnectionMethodProxy
		} else if clusterBean.ToConnectWithSSHTunnel {
			connectionMethod = remoteConnectionBean.RemoteConnectionMethodSSH
		}
		connectionConfig = &remoteConnectionRepository.RemoteConnectionConfig{
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
	model.RemoteConnectionConfig = connectionConfig
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
	if model.RemoteConnectionConfig != nil {
		clusterBean.RemoteConnectionConfig = &remoteConnectionBean.RemoteConnectionConfigBean{
			RemoteConnectionConfigId: model.RemoteConnectionConfigId,
			ConnectionMethod:         model.RemoteConnectionConfig.ConnectionMethod,
			ProxyConfig: &remoteConnectionBean.ProxyConfig{
				ProxyUrl: model.RemoteConnectionConfig.ProxyUrl,
			},
			SSHTunnelConfig: &remoteConnectionBean.SSHTunnelConfig{
				SSHServerAddress: model.RemoteConnectionConfig.SSHServerAddress,
				SSHUsername:      model.RemoteConnectionConfig.SSHUsername,
				SSHPassword:      model.RemoteConnectionConfig.SSHPassword,
				SSHAuthKey:       model.RemoteConnectionConfig.SSHAuthKey,
			},
		}
	}
	return clusterBean
}
