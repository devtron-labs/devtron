package in

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/audit"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"go.uber.org/zap"
	"net/http"
)

type InternalProcessingService interface {
	DeleteObjectAndItsDependency(req *bean.DevtronResourceObjectDescriptorBean) error
}

type InternalProcessingServiceImpl struct {
	logger                          *zap.SugaredLogger
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository
	dtResourceReadService           read.ReadService
	dtResourceObjectAuditService    audit.ObjectAuditService
}

func NewInternalProcessingServiceImpl(logger *zap.SugaredLogger,
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository,
	dtResourceReadService read.ReadService,
	dtResourceObjectAuditService audit.ObjectAuditService) *InternalProcessingServiceImpl {
	return &InternalProcessingServiceImpl{
		logger:                          logger,
		devtronResourceSchemaRepository: devtronResourceSchemaRepository,
		devtronResourceObjectRepository: devtronResourceObjectRepository,
		dtResourceReadService:           dtResourceReadService,
		dtResourceObjectAuditService:    dtResourceObjectAuditService,
	}
}

func (impl *InternalProcessingServiceImpl) DeleteObjectAndItsDependency(req *bean.DevtronResourceObjectDescriptorBean) error {
	tx, err := impl.devtronResourceObjectRepository.StartTx()
	// Rollback tx on error.
	defer impl.devtronResourceObjectRepository.RollbackTx(tx)
	if err != nil {
		impl.logger.Errorw("error in getting transaction, DeleteObject", "err", err)
		return err
	}
	resourceObjectId := req.GetResourceIdByIdType()
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema, DeleteObject", "err", err, "kind", req.Kind, "subKind", req.SubKind, "version", req.Version)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return err
	}
	if resourceObjectId == 0 {
		if len(req.Identifier) == 0 {
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceRequestDescriptorData, bean.InvalidResourceRequestDescriptorData)
		}
		devtronResourceObject, err := impl.devtronResourceObjectRepository.FindByObjectIdentifier(req.Identifier, resourceSchema.Id)
		if err != nil {
			impl.logger.Errorw("error in getting object by identifier", "err", err, "identifier", req.Identifier, "devtronResourceSchemaId", resourceSchema.Id)
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
		}
		resourceObjectId, req.IdType = helper.GetResourceObjectIdAndType(devtronResourceObject)
	}
	exists, err := impl.dtResourceReadService.CheckIfDevtronObjectExistsByIdAndIdType(resourceObjectId, resourceSchema.Id, req.IdType)
	if err != nil {
		return err
	}
	if !exists {
		impl.logger.Errorw("no resource object found to be deleted, DeleteObject", "id", resourceObjectId, "idType", req.IdType, "err", err)
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	var updatedResourceObject *repository.DevtronResourceObject
	if req.IdType == bean.ResourceObjectIdType {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.DeleteObjectById(tx, resourceObjectId, resourceSchema.DevtronResourceId, req.UserId)
		if err != nil {
			impl.logger.Errorw("error, DeleteObject", "err", err, "id", resourceObjectId, "idType", req.IdType, "devtronResourceId", resourceSchema.DevtronResourceId)
			return err
		}
	} else if req.IdType == bean.OldObjectId {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.DeleteObjectByOldObjectId(tx, resourceObjectId, resourceSchema.DevtronResourceId, req.UserId)
		if err != nil {
			impl.logger.Errorw("error, DeleteObject", "err", err, "oldObjectId", resourceObjectId, "idType", req.IdType, "devtronResourceId", resourceSchema.DevtronResourceId)
			return err
		}
	} else {
		err = fmt.Errorf(bean.IdTypeNotSupportedError)
		impl.logger.Errorw("error, DeleteObject", "err", err, "id", resourceObjectId, "idType", req.IdType, "devtronResourceId", resourceSchema.DevtronResourceId)
		return err
	}

	err = impl.devtronResourceObjectRepository.DeleteDependencyInObjectData(tx, resourceObjectId, resourceSchema.DevtronResourceId, req.UserId)
	if err != nil {
		impl.logger.Errorw("error, DeleteObject", "err", err, "id", resourceObjectId, "idType", req.IdType, "devtronResourceId", resourceSchema.DevtronResourceId)
		return err
	}
	if updatedResourceObject != nil {
		impl.dtResourceObjectAuditService.SaveAudit(updatedResourceObject, repository.AuditOperationTypeDeleted, nil)
	}
	err = impl.devtronResourceObjectRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction, DeleteObject", "err", err)
		return err
	}
	return nil
}
