package adapter

import (
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/serverConnection/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func GetServerConnectionConfigBean(model *repository.ServerConnectionConfig) *bean.ServerConnectionConfigBean {
	var configBean *bean.ServerConnectionConfigBean
	if model != nil {
		configBean = &bean.ServerConnectionConfigBean{
			ServerConnectionConfigId: model.Id,
			ConnectionMethod:         model.ConnectionMethod,
		}
		if model.ConnectionMethod == bean.ServerConnectionMethodProxy {
			configBean.ProxyConfig = &bean.ProxyConfig{
				ProxyUrl: model.ProxyUrl,
			}
		}
		if model.ConnectionMethod == bean.ServerConnectionMethodSSH {
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

func ConvertServerConnectionConfigBeanToServerConnectionConfig(configBean *bean.ServerConnectionConfigBean, userId int32) *repository.ServerConnectionConfig {
	var model repository.ServerConnectionConfig
	if configBean != nil {
		model = repository.ServerConnectionConfig{
			Id:               configBean.ServerConnectionConfigId,
			ConnectionMethod: configBean.ConnectionMethod,
			Deleted:          false,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedBy: userId,
				UpdatedOn: time.Now(),
			},
		}
		if configBean.ProxyConfig != nil && configBean.ConnectionMethod == bean.ServerConnectionMethodProxy {
			proxyConfig := configBean.ProxyConfig
			model.ProxyUrl = proxyConfig.ProxyUrl
		}
		if configBean.SSHTunnelConfig != nil && configBean.ConnectionMethod == bean.ServerConnectionMethodSSH {
			sshTunnelConfig := configBean.SSHTunnelConfig
			model.SSHServerAddress = sshTunnelConfig.SSHServerAddress
			model.SSHUsername = sshTunnelConfig.SSHUsername
			model.SSHPassword = sshTunnelConfig.SSHPassword
			model.SSHAuthKey = sshTunnelConfig.SSHAuthKey
		}
	}
	return &model
}
