package remoteConnection

import (
	dockerArtifactStoreRegistry "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/adapter"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ServerConnectionService interface {
	// methods
	CreateOrUpdateServerConnectionConfig(reqBean *bean.RemoteConnectionConfigBean, userId int32, tx *pg.Tx) error
	GetServerConnectionConfigById(id int) (*bean.RemoteConnectionConfigBean, error)
}

type ServerConnectionServiceImpl struct {
	logger                        *zap.SugaredLogger
	serverConnectionRepository    repository.RemoteConnectionRepository
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository
}

func NewServerConnectionServiceImpl(logger *zap.SugaredLogger,
	serverConnectionRepository repository.RemoteConnectionRepository,
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository) *ServerConnectionServiceImpl {
	return &ServerConnectionServiceImpl{
		logger:                        logger,
		serverConnectionRepository:    serverConnectionRepository,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
	}
}

func (impl *ServerConnectionServiceImpl) CreateOrUpdateServerConnectionConfig(reqBean *bean.RemoteConnectionConfigBean, userId int32, tx *pg.Tx) error {
	existingConfig, err := impl.serverConnectionRepository.GetById(reqBean.RemoteConnectionConfigId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching existing server connection config", "err", err, "id", reqBean.RemoteConnectionConfigId)
		return err
	}
	config := adapter.ConvertServerConnectionConfigBeanToServerConnectionConfig(reqBean, userId)
	if existingConfig == nil {
		err = impl.serverConnectionRepository.Save(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while saving server connection config", "err", err)
			return err
		}
		reqBean.RemoteConnectionConfigId = config.Id
	} else {
		config.Id = existingConfig.Id
		config.CreatedBy = existingConfig.CreatedBy
		config.CreatedOn = existingConfig.CreatedOn
		err = impl.serverConnectionRepository.Update(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while updating server connection config", "err", err)
			return err
		}
	}
	return nil
}

func (impl *ServerConnectionServiceImpl) GetServerConnectionConfigById(id int) (*bean.RemoteConnectionConfigBean, error) {
	model, err := impl.serverConnectionRepository.GetById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching server connection config", "err", err, "serverConnectionConfigId", id)
		return nil, err
	}
	serverConnectionConfig := adapter.GetServerConnectionConfigBean(model)
	return serverConnectionConfig, nil
}
