package remoteConnection

import (
	dockerArtifactStoreRegistry "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/adapter"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	"github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RemoteConnectionService interface {
	// methods
	CreateOrUpdateRemoteConnectionConfig(reqBean *bean.RemoteConnectionConfigBean, userId int32, tx *pg.Tx) error
	GetRemoteConnectionConfigById(id int) (*bean.RemoteConnectionConfigBean, error)
}

type RemoteConnectionServiceImpl struct {
	logger                        *zap.SugaredLogger
	remoteConnectionRepository    repository.RemoteConnectionRepository
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository
}

func NewRemoteConnectionServiceImpl(logger *zap.SugaredLogger,
	remoteConnectionRepository repository.RemoteConnectionRepository,
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository) *RemoteConnectionServiceImpl {
	return &RemoteConnectionServiceImpl{
		logger:                        logger,
		remoteConnectionRepository:    remoteConnectionRepository,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
	}
}

func (impl *RemoteConnectionServiceImpl) CreateOrUpdateRemoteConnectionConfig(reqBean *bean.RemoteConnectionConfigBean, userId int32, tx *pg.Tx) error {
	existingConfig, err := impl.remoteConnectionRepository.GetById(reqBean.RemoteConnectionConfigId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error occurred while fetching existing server connection config", "err", err, "id", reqBean.RemoteConnectionConfigId)
		return err
	}
	config := adapter.ConvertRemoteConnectionConfigBeanToRemoteConnectionConfig(reqBean, userId)
	if existingConfig == nil || existingConfig.Id == 0 {
		err = impl.remoteConnectionRepository.Save(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while saving server connection config", "err", err)
			return err
		}
		reqBean.RemoteConnectionConfigId = config.Id
	} else {
		config.Id = existingConfig.Id
		if config.SSHPassword == bean.SecretDataObfuscatePlaceholder {
			config.SSHPassword = existingConfig.SSHPassword
		}
		if config.SSHAuthKey == bean.SecretDataObfuscatePlaceholder {
			config.SSHAuthKey = existingConfig.SSHAuthKey
		}
		config.CreatedBy = existingConfig.CreatedBy
		config.CreatedOn = existingConfig.CreatedOn
		err = impl.remoteConnectionRepository.Update(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while updating server connection config", "err", err)
			return err
		}
	}
	return nil
}

func (impl *RemoteConnectionServiceImpl) GetRemoteConnectionConfigById(id int) (*bean.RemoteConnectionConfigBean, error) {
	model, err := impl.remoteConnectionRepository.GetById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching remote connection config", "err", err, "remoteConnectionConfigId", id)
		return nil, err
	}
	remoteConnectionConfig := adapter.GetRemoteConnectionConfigBean(model)
	return remoteConnectionConfig, nil
}
