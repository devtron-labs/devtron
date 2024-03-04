package adapter

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean4 "github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/serverConnection/repository"
	"time"
)

func ConvertClusterToNewCluster(model *repository.Cluster) *repository.Cluster {
	if len(model.ProxyUrl) > 0 || model.ToConnectWithSSHTunnel {
		// converting old to new
		connectionConfig := &repository3.ServerConnectionConfig{
			Id:               model.ServerConnectionConfigId,
			ProxyUrl:         model.ProxyUrl,
			SSHServerAddress: model.SSHTunnelServerAddress,
			SSHUsername:      model.SSHTunnelUser,
			SSHPassword:      model.SSHTunnelPassword,
			SSHAuthKey:       model.SSHTunnelAuthKey,
			Deleted:          false,
		}
		if len(model.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = bean4.ServerConnectionMethodProxy
		} else if model.ToConnectWithSSHTunnel {
			connectionConfig.ConnectionMethod = bean4.ServerConnectionMethodSSH
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

func ConvertClusterBeanToNewClusterBean(clusterBean *cluster.ClusterBean) *cluster.ClusterBean {
	if len(clusterBean.ProxyUrl) > 0 || clusterBean.ToConnectWithSSHTunnel {
		// converting old bean to new bean
		connectionConfig := &bean4.ServerConnectionConfigBean{}
		if len(clusterBean.ProxyUrl) > 0 {
			connectionConfig.ConnectionMethod = bean4.ServerConnectionMethodProxy
			connectionConfig.ProxyConfig = &bean4.ProxyConfig{
				ProxyUrl: clusterBean.ProxyUrl,
			}
		}
		if clusterBean.ToConnectWithSSHTunnel && clusterBean.SSHTunnelConfig != nil {
			connectionConfig.ConnectionMethod = bean4.ServerConnectionMethodSSH
			connectionConfig.SSHTunnelConfig = &bean4.SSHTunnelConfig{
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

func ConvertClusterBeanToCluster(clusterBean *cluster.ClusterBean, userId int32) *repository.Cluster {

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

	var connectionMethod bean4.ServerConnectionMethod
	var connectionConfig *repository3.ServerConnectionConfig
	if clusterBean.ServerConnectionConfig != nil {
		// if FE provided new bean
		connectionMethod = clusterBean.ServerConnectionConfig.ConnectionMethod
		connectionConfig = &repository3.ServerConnectionConfig{
			ConnectionMethod: connectionMethod,
		}
		if clusterBean.ServerConnectionConfig.ConnectionMethod == bean4.ServerConnectionMethodProxy &&
			clusterBean.ServerConnectionConfig.ProxyConfig != nil {
			connectionConfig.ProxyUrl = clusterBean.ServerConnectionConfig.ProxyConfig.ProxyUrl
		}
		if clusterBean.ServerConnectionConfig.ConnectionMethod == bean4.ServerConnectionMethodSSH &&
			clusterBean.ServerConnectionConfig.SSHTunnelConfig != nil {
			connectionConfig.SSHServerAddress = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHServerAddress
			connectionConfig.SSHUsername = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHUsername
			connectionConfig.SSHPassword = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword
			connectionConfig.SSHAuthKey = clusterBean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey
		}
	} else if len(clusterBean.ProxyUrl) > 0 || clusterBean.ToConnectWithSSHTunnel {
		// if FE provided old bean
		if len(clusterBean.ProxyUrl) > 0 {
			connectionMethod = bean4.ServerConnectionMethodProxy
		} else if clusterBean.ToConnectWithSSHTunnel {
			connectionMethod = bean4.ServerConnectionMethodSSH
		}
		connectionConfig = &repository3.ServerConnectionConfig{
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
