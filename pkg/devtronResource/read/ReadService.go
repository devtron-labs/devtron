package read

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"go.uber.org/zap"
)

type ReadService interface {
	CheckIfExistingDevtronObject(id, devtronResourceSchemaId int, idType bean.IdType, identifier string) (exists bool, err error)
	CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId int, idType bean.IdType) (exists bool, err error)
}
type ReadServiceImpl struct {
	logger                          *zap.SugaredLogger
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository
}

func NewReadServiceImpl(logger *zap.SugaredLogger,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository) *ReadServiceImpl {
	return &ReadServiceImpl{
		logger:                          logger,
		devtronResourceObjectRepository: devtronResourceObjectRepository,
	}
}

// CheckIfExistingDevtronObject : this method check if it is existing object in the db.
func (impl *ReadServiceImpl) CheckIfExistingDevtronObject(id, devtronResourceSchemaId int, idType bean.IdType, identifier string) (exists bool, err error) {
	if id > 0 {
		exists, err = impl.CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId, idType)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by id", "err", err, "id", id, "idType", idType, "devtronResourceSchemaId", devtronResourceSchemaId)
			return exists, err
		}
	} else if len(identifier) > 0 {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistByIdentifier(identifier, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by identifier", "err", err, "identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
			return exists, err
		}
	}
	return exists, err
}

// CheckIfDevtronObjectExistsByIdAndIdType : this method check if it is existing object in the db by resource id or oldObjectId.
func (impl *ReadServiceImpl) CheckIfDevtronObjectExistsByIdAndIdType(id, devtronResourceSchemaId int, idType bean.IdType) (exists bool, err error) {
	if idType == bean.ResourceObjectIdType {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistById(id, devtronResourceSchemaId)
		if err != nil {
			return exists, err
		}
	} else if idType == bean.OldObjectId {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistById(id, devtronResourceSchemaId)
		if err != nil {
			return exists, err
		}
	} else {
		return exists, fmt.Errorf(bean.IdTypeNotSupportedError)
	}
	return exists, err
}
