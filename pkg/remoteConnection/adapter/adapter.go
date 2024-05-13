package adapter

import (
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func GetRemoteConnectionConfigBean(model *repository.RemoteConnectionConfig) *bean.RemoteConnectionConfigBean {
	var configBean *bean.RemoteConnectionConfigBean
	if model != nil {
		if len(model.SSHPassword) > 0 {
			model.SSHPassword = bean.SecretDataObfuscatePlaceholder
		}
		if len(model.SSHAuthKey) > 0 {
			model.SSHAuthKey = bean.SecretDataObfuscatePlaceholder
		}
		configBean = &bean.RemoteConnectionConfigBean{
			RemoteConnectionConfigId: model.Id,
			ConnectionMethod:         model.ConnectionMethod,
		}
		if model.ConnectionMethod == bean.RemoteConnectionMethodProxy {
			configBean.ProxyConfig = &bean.ProxyConfig{
				ProxyUrl: model.ProxyUrl,
			}
		}
		if model.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			configBean.SSHTunnelConfig = &bean.SSHTunnelConfig{
				SSHServerAddress: model.SSHServerAddress,
				SSHUsername:      model.SSHUsername,
				SSHPassword:      model.SSHPassword,
				SSHAuthKey:       model.SSHAuthKey,
			}
		}
	}
	return configBean
}

func ConvertRemoteConnectionConfigBeanToRemoteConnectionConfig(configBean *bean.RemoteConnectionConfigBean, userId int32) *repository.RemoteConnectionConfig {
	var model repository.RemoteConnectionConfig
	if configBean != nil {
		model = repository.RemoteConnectionConfig{
			Id:               configBean.RemoteConnectionConfigId,
			ConnectionMethod: configBean.ConnectionMethod,
			Deleted:          false,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedBy: userId,
				UpdatedOn: time.Now(),
			},
		}
		if configBean.ProxyConfig != nil && configBean.ConnectionMethod == bean.RemoteConnectionMethodProxy {
			proxyConfig := configBean.ProxyConfig
			model.ProxyUrl = proxyConfig.ProxyUrl
		}
		if configBean.SSHTunnelConfig != nil && configBean.ConnectionMethod == bean.RemoteConnectionMethodSSH {
			sshTunnelConfig := configBean.SSHTunnelConfig
			model.SSHServerAddress = sshTunnelConfig.SSHServerAddress
			model.SSHUsername = sshTunnelConfig.SSHUsername
			model.SSHPassword = sshTunnelConfig.SSHPassword
			model.SSHAuthKey = sshTunnelConfig.SSHAuthKey
		}
	}
	return &model
}
