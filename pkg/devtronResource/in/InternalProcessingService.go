package in

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/audit"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
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
		impl.logger.Errorw("error in checking existing resource object, DeleteObject", "id", resourceObjectId, "idType", req.IdType, "err", err)
		return err
	}
	// for some devtron resources; we use it directly in dependencies.
	// So even though resource object is not preset, there might be usage of them in dependencies.
	if exists {
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
		if updatedResourceObject != nil {
			impl.dtResourceObjectAuditService.SaveAudit(updatedResourceObject, repository.AuditOperationTypeDeleted, nil)
		}
	}
	err = impl.checkDeleteObjImpactAndPerformInfoPatch(tx, resourceObjectId, resourceSchema.Id, req)
	if err != nil {
		impl.logger.Errorw("error, checkDeleteObjImpactAndPerformInfoPatch", "err", err, "req", req)
		return err
	}
	err = impl.devtronResourceObjectRepository.DeleteDependencyInObjectData(tx, resourceObjectId, resourceSchema.DevtronResourceId, req.UserId)
	if err != nil {
		impl.logger.Errorw("error in DeleteDependencyInObjectData, DeleteObject", "err", err, "id", resourceObjectId, "idType", req.IdType, "devtronResourceId", resourceSchema.DevtronResourceId)
		return err
	}
	err = impl.devtronResourceObjectRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction, DeleteObject", "err", err)
		return err
	}
	return nil
}

// TODO: pick this in v2 part to refactor
func (impl *InternalProcessingServiceImpl) checkDeleteObjImpactAndPerformInfoPatch(tx *pg.Tx, objectId, schemaId int,
	req *bean.DevtronResourceObjectDescriptorBean) error {
	if req.Kind == bean.DevtronResourceApplication.ToString() &&
		req.SubKind == bean.DevtronResourceDevtronApplication.ToString() &&
		req.Version == bean.DevtronResourceVersion1.ToString() {
		//check release objects having this app as upstream
		//getting release schemaId
		devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean.DevtronResourceRelease.ToString(),
			"", bean.DevtronResourceVersionAlpha1.ToString())
		if err != nil {
			impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "kind", bean.DevtronResourceRelease)
			return err
		}
		releaseObjs, err := impl.devtronResourceObjectRepository.GetDownstreamObjectsByOwnSchemaIdAndUpstreamId(devtronResourceSchema.Id, objectId, schemaId)
		if err != nil {
			impl.logger.Errorw("error, GetDownstreamObjectsByOwnSchemaIdAndUpstreamId", "err", err)
			return err
		}
		objectsToBePatched := make([]*repository.DevtronResourceObject, 0, len(releaseObjs))
		for i := range releaseObjs {
			releaseObj := releaseObjs[i]
			var statusToBeUpdated bean.ReleaseConfigStatus

			//get rollout status of this releaseObj
			rollOutStatus := gjson.Get(releaseObj.ObjectData, bean.ReleaseResourceRolloutStatusPath).String()
			if rollOutStatus == bean.PartiallyDeployedReleaseRolloutStatus.ToString() ||
				rollOutStatus == bean.CompletelyDeployedReleaseRolloutStatus.ToString() {
				statusToBeUpdated = bean.CorruptedReleaseConfigStatus
			} else {
				statusToBeUpdated = bean.DraftReleaseConfigStatus
				//also update lock in this case, set to unlock
				releaseObj.ObjectData, err = helper.PatchResourceObjectDataAtAPath(releaseObj.ObjectData, bean.ReleaseResourceConfigStatusIsLockedPath, false)
				if err != nil {
					impl.logger.Errorw("error, PatchResourceObjectData", "err", err, "releaseObj", releaseObj)
					continue
				}
			}
			//patch config status as corrupted
			releaseObj.ObjectData, err = helper.PatchResourceObjectDataAtAPath(releaseObj.ObjectData, bean.ReleaseResourceConfigStatusStatusPath, statusToBeUpdated)
			if err != nil {
				impl.logger.Errorw("error, PatchResourceObjectData", "err", err, "releaseObj", releaseObj)
				continue
			} else {
				objectsToBePatched = append(objectsToBePatched, releaseObj)
			}
		}
		if len(objectsToBePatched) > 0 {
			err = impl.devtronResourceObjectRepository.UpdateInBulk(tx, objectsToBePatched)
			if err != nil {
				impl.logger.Errorw("error, UpdateInBulk", "err", err, "objects", objectsToBePatched)
				return err
			}
		}
	}
	return nil
}
