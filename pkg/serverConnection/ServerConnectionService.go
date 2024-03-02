package serverConnection

import (
	"github.com/devtron-labs/devtron/pkg/serverConnection/adapter"
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/serverConnection/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ServerConnectionService interface {
	// methods
	CreateOrUpdateServerConnectionConfig(reqBean *bean.ServerConnectionConfigBean, userId int32, tx *pg.Tx) error
	GetServerConnectionConfigById(id int) (*bean.ServerConnectionConfigBean, error)
}

type ServerConnectionServiceImpl struct {
	logger                     *zap.SugaredLogger
	serverConnectionRepository repository.ServerConnectionRepository
}

func NewServerConnectionServiceImpl(logger *zap.SugaredLogger,
	serverConnectionRepository repository.ServerConnectionRepository) *ServerConnectionServiceImpl {
	return &ServerConnectionServiceImpl{
		logger:                     logger,
		serverConnectionRepository: serverConnectionRepository,
	}
}

func (impl *ServerConnectionServiceImpl) CreateOrUpdateServerConnectionConfig(reqBean *bean.ServerConnectionConfigBean, userId int32, tx *pg.Tx) error {
	existingConfig, err := impl.serverConnectionRepository.GetById(reqBean.ServerConnectionConfigId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching existing server connection config", "err", err, "id", reqBean.ServerConnectionConfigId)
		return err
	}
	config := adapter.ConvertServerConnectionConfigBeanToServerConnectionConfig(reqBean, userId)
	if existingConfig == nil {
		config.AuditLog.CreatedOn = time.Now()
		config.AuditLog.CreatedBy = userId
		err = impl.serverConnectionRepository.Save(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while saving server connection config", "err", err)
			return err
		}
		reqBean.ServerConnectionConfigId = config.Id
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
	serverConnectionConfig := adapter.GetServerConnectionConfigBean(model)
	return serverConnectionConfig, nil
}
