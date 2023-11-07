package devtronResource

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type DevtronResourceService interface {
	GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error)
	UpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error
}

type DevtronResourceServiceImpl struct {
	logger                               *zap.SugaredLogger
	devtronResourceRepository            repository.DevtronResourceRepository
	devtronResourceSchemaRepository      repository.DevtronResourceSchemaRepository
	devtronResourceObjectRepository      repository.DevtronResourceObjectRepository
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository
	userRepository                       repository2.UserRepository
}

func NewDevtronResourceServiceImpl(logger *zap.SugaredLogger,
	devtronResourceRepository repository.DevtronResourceRepository,
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository,
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository,
	userRepository repository2.UserRepository) *DevtronResourceServiceImpl {
	return &DevtronResourceServiceImpl{
		logger:                               logger,
		devtronResourceRepository:            devtronResourceRepository,
		devtronResourceSchemaRepository:      devtronResourceSchemaRepository,
		devtronResourceObjectRepository:      devtronResourceObjectRepository,
		devtronResourceObjectAuditRepository: devtronResourceObjectAuditRepository,
		userRepository:                       userRepository,
	}
}

const RefType = "refType"

func getRefTypeInJson(schemaJsonMap map[string]interface{}, referencedObjects map[string]bool) {
	for key, value := range schemaJsonMap {
		valStr, ok := value.(string)
		if ok {
			if key == RefType {
				referencedObjects[valStr] = true
			}
		} else {
			resolveValForIteration(key, value, referencedObjects)
		}
	}
}

func resolveValForIteration(key string, value interface{}, referencedObjects map[string]bool) {
	if valNew, ok := value.(map[string]interface{}); ok {
		getRefTypeInJson(valNew, referencedObjects)
	} else if valArr, ok := value.([]interface{}); ok {
		for _, val := range valArr {
			resolveValForIteration(key, val, referencedObjects)
		}
	}
}

func (impl DevtronResourceServiceImpl) GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}

	schemaJsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(resourceSchema.Schema), &schemaJsonMap)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling schema", "err", err, "schema", resourceSchema.Schema)
		return nil, err
	}

	referencedPaths := make(map[string]bool)
	getRefTypeInJson(schemaJsonMap, referencedPaths)

	var referencedObjects []interface{}

	for refPath := range referencedPaths {
		refPathSplit := strings.Split(refPath, "/")
		resourceKind := refPathSplit[2]

		referencedItem := make(map[string]interface{})

		if string(bean.DEVTRON_RESOURCE_USER) == resourceKind {
			referencedItem[bean.RefKind] = bean.DEVTRON_RESOURCE_USER
			userModel, err := impl.userRepository.GetAllActiveUsers()
			if err != nil {
				impl.logger.Errorw("error while fetching all users", "err", err, "resource kind", resourceKind)
				return nil, err
			}

			var referencedValues []interface{}
			for _, user := range userModel {
				userList := make(map[string]string)
				userId := strconv.Itoa(int(user.Id))
				userList[bean.RefUserId] = userId
				userList[bean.RefUserName] = user.EmailId
				referencedValues = append(referencedValues, userList)
			}
			referencedItem[bean.RefValues] = referencedValues
			referencedObjects = append(referencedObjects, referencedItem)
		} else {
			impl.logger.Errorw("error while extracting kind of resource; kind not supported as of now", "resource kind", resourceKind)
			return nil, errors.New(fmt.Sprintf("%s kind is not supported", resourceKind))
		}
	}

	referenceObjectsJson, err := json.Marshal(referencedObjects)
	if err != nil {
		impl.logger.Errorw("error in marshalling referencedObjects", "err", err, "referenced objects", referencedObjects)
		return nil, err
	}

	var existingResourceObject *repository.DevtronResourceObject
	if req.OldObjectId > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByOldObjectId(req.OldObjectId, resourceSchema.DevtronResourceId, resourceSchema.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
			return nil, err
		}
	} else if len(req.Name) > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByObjectName(req.Name, resourceSchema.DevtronResourceId, resourceSchema.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
			return nil, err
		}
	}
	resourceObject := &bean.DevtronResourceObjectBean{
		DevtronResourceObjectDescriptorBean: req,
		Schema:                              resourceSchema.Schema,
		ReferencedObjects:                   string(referenceObjectsJson),
	}
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		resourceObject.ObjectData = existingResourceObject.ObjectData
	}
	return resourceObject, nil
}

func (impl *DevtronResourceServiceImpl) UpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error {
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		return err
	}
	schema := devtronResourceSchema.Schema
	reqBean.ObjectData, err = sjson.Set(reqBean.ObjectData, "version", reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
		return err
	}

	kindForSchema := reqBean.Kind
	if len(reqBean.SubKind) > 0 {
		kindForSchema += fmt.Sprintf("/%s", reqBean.SubKind)
	}

	reqBean.ObjectData, err = sjson.Set(reqBean.ObjectData, "kind", kindForSchema)
	if err != nil {
		impl.logger.Errorw("error in setting kind in schema", "err", err, "request", reqBean)
		return err
	}

	reqBean.ObjectData, err = sjson.Set(reqBean.ObjectData, "overview.id", reqBean.OldObjectId)
	if err != nil {
		impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
		return err
	}

	//validate user provided json with the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(reqBean.ObjectData)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		impl.logger.Errorw("error in validating resource object json against schema", "err", err, "request", reqBean, "schema", schema)
		return err
	} else if !result.Valid() {
		impl.logger.Errorw("error in validating resource object json against schema", "result", result, "request", reqBean, "schema", schema)
		//not using below errStr currently
		errStr := ""
		for _, errResult := range result.Errors() {
			errStr += fmt.Sprintln(errResult.String())
		}
		return fmt.Errorf("Please provide data for all required fields before saving")
	}
	//schema is validated, getting older objects with given id or name
	var devtronResourceObject *repository.DevtronResourceObject
	if reqBean.OldObjectId > 0 {
		devtronResourceObject, err = impl.devtronResourceObjectRepository.FindByOldObjectId(reqBean.OldObjectId, devtronResourceSchema.DevtronResourceId, devtronResourceSchema.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
			return err
		}
	} else if len(reqBean.Name) > 0 {
		devtronResourceObject, err = impl.devtronResourceObjectRepository.FindByObjectName(reqBean.Name, devtronResourceSchema.DevtronResourceId, devtronResourceSchema.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
			return err
		}
	}

	//TODO: update object data with id(oldObjectId) or name

	var auditAction repository.AuditOperationType
	if devtronResourceObject != nil && devtronResourceObject.Id > 0 {
		auditAction = repository.AuditOperationTypeUpdate
		//object already exists, update the same
		devtronResourceObject.ObjectData = reqBean.ObjectData
		devtronResourceObject.UpdatedBy = reqBean.UserId
		devtronResourceObject.UpdatedOn = time.Now()
		_, err = impl.devtronResourceObjectRepository.Update(devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in updating", "err", err, "req", devtronResourceObject)
			return err
		}
	} else {
		auditAction = repository.AuditOperationTypeCreate
		//object does not exist, create new
		devtronResourceObject = &repository.DevtronResourceObject{
			OldObjectId:             reqBean.OldObjectId,
			Name:                    reqBean.Name,
			DevtronResourceId:       devtronResourceSchema.DevtronResourceId,
			DevtronResourceSchemaId: devtronResourceSchema.Id,
			ObjectData:              reqBean.ObjectData,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: reqBean.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: reqBean.UserId,
			},
		}
		_, err = impl.devtronResourceObjectRepository.Save(devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in saving", "err", err, "req", devtronResourceObject)
			return err
		}
	}

	//save audit
	auditModel := &repository.DevtronResourceObjectAudit{
		DevtronResourceObjectId: devtronResourceObject.Id,
		ObjectData:              devtronResourceObject.ObjectData,
		AuditOperation:          auditAction,
		AuditLog: sql.AuditLog{
			CreatedOn: devtronResourceObject.CreatedOn,
			CreatedBy: devtronResourceObject.CreatedBy,
			UpdatedBy: devtronResourceObject.UpdatedBy,
			UpdatedOn: devtronResourceSchema.UpdatedOn,
		},
	}
	err = impl.devtronResourceObjectAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceObject audit", "err", err, "auditModel", auditModel)
	}
	return nil
}
