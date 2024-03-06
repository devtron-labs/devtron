package serverConnection

import (
	"fmt"
	dockerArtifactStoreRegistry "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/serverConnection/adapter"
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/serverConnection/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ServerConnectionService interface {
	// methods
	CreateOrUpdateServerConnectionConfig(reqBean *bean.ServerConnectionConfigBean, userId int32, tx *pg.Tx) error
	GetServerConnectionConfigById(id int) (*bean.ServerConnectionConfigBean, error)
	GetServerConnectionConfigByDockerId(dockerId string) (*bean.ServerConnectionConfigBean, error)
}

type ServerConnectionServiceImpl struct {
	logger                        *zap.SugaredLogger
	serverConnectionRepository    repository.ServerConnectionRepository
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository
}

func NewServerConnectionServiceImpl(logger *zap.SugaredLogger,
	serverConnectionRepository repository.ServerConnectionRepository,
	dockerArtifactStoreRepository dockerArtifactStoreRegistry.DockerArtifactStoreRepository) *ServerConnectionServiceImpl {
	return &ServerConnectionServiceImpl{
		logger:                        logger,
		serverConnectionRepository:    serverConnectionRepository,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
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
		err = impl.serverConnectionRepository.Save(config, tx)
		if err != nil {
			impl.logger.Errorw("error occurred while saving server connection config", "err", err)
			return err
		}
		reqBean.ServerConnectionConfigId = config.Id
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

func (impl *ServerConnectionServiceImpl) GetServerConnectionConfigById(id int) (*bean.ServerConnectionConfigBean, error) {
	model, err := impl.serverConnectionRepository.GetById(id)
	if err != nil {
		return nil, err
	}
	serverConnectionConfig := adapter.GetServerConnectionConfigBean(model)
	return serverConnectionConfig, nil
}

func (impl *ServerConnectionServiceImpl) GetServerConnectionConfigByDockerId(dockerId string) (*bean.ServerConnectionConfigBean, error) {
	fmt.Println("hi")
	dockerRegistry, err := impl.dockerArtifactStoreRepository.FindOne(dockerId)
	if err != nil {
		return nil, err
	}
	serverConnectionConfig, err := impl.GetServerConnectionConfigById(dockerRegistry.ServerConnectionConfigId)
	if err != nil {
		return nil, err
	}
	return serverConnectionConfig, nil
}
