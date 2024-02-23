package serverConnection

import (
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/serverConnection/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ServerConnectionService interface {
	// methods
	CreateOrUpdateServerConnectionConfig(reqBean *bean.ServerConnectionConfigBean, userId int32, tx *pg.Tx) error
	GetServerConnectionConfigById(id int) (*bean.ServerConnectionConfigBean, error)
	ConvertServerConnectionConfigBeanToServerConnectionConfig(configBean *bean.ServerConnectionConfigBean, userId int32) *repository.ServerConnectionConfig
	GetServerConnectionConfigBean(model *repository.ServerConnectionConfig) *bean.ServerConnectionConfigBean
}

type ServerConnectionServiceImpl struct {
	logger                     *zap.SugaredLogger
	serverConnectionRepository repository.ServerConnectionRepository
}

func NewServerConnectionServiceImpl(logger *zap.SugaredLogger,
	serverConnectionRepository repository.ServerConnectionRepository) (*ServerConnectionServiceImpl, error) {
	impl := &ServerConnectionServiceImpl{
		logger:                     logger,
		serverConnectionRepository: serverConnectionRepository,
	}
	return impl, nil
}

func (impl *ServerConnectionServiceImpl) GetServerConnectionConfigBean(model *repository.ServerConnectionConfig) *bean.ServerConnectionConfigBean {
	var configBean bean.ServerConnectionConfigBean
	if model != nil {
		configBean = bean.ServerConnectionConfigBean{
			ConnectionMethod: model.ConnectionMethod,
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
	return &configBean
}

func (impl *ServerConnectionServiceImpl) ConvertServerConnectionConfigBeanToServerConnectionConfig(configBean *bean.ServerConnectionConfigBean, userId int32) *repository.ServerConnectionConfig {
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

func (impl *ServerConnectionServiceImpl) CreateOrUpdateServerConnectionConfig(reqBean *bean.ServerConnectionConfigBean, userId int32, tx *pg.Tx) error {
	existingConfig, err := impl.serverConnectionRepository.GetById(reqBean.ServerConnectionConfigId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching existing server connection config", "err", err, "id", reqBean.ServerConnectionConfigId)
		return err
	}
	config := impl.ConvertServerConnectionConfigBeanToServerConnectionConfig(reqBean, userId)
	if existingConfig == nil {
		config.AuditLog.CreatedOn = time.Now()
		config.AuditLog.CreatedBy = userId
		err = impl.serverConnectionRepository.Save(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while saving server connection config", "err", err)
			return err
		}
	} else {
		config.Id = existingConfig.Id
		err = impl.serverConnectionRepository.Update(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while updating server connection config", "err", err)
			return err
		}
	}
	return nil
}

func (impl *ServerConnectionServiceImpl) GetServerConnectionConfigById(id int) (*bean.ServerConnectionConfigBean, error) {
	model, err := impl.serverConnectionRepository.GetById(id)
	if err != nil {
		return nil, err
	}
	serverConnectionConfig := impl.GetServerConnectionConfigBean(model)
	return serverConnectionConfig, nil
}
