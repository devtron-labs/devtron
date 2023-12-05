package devtronResource

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type DevtronResourceService interface {
	GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error)
	CreateOrUpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error
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

func getRefTypeInJsonAndAddRefKey(schemaJsonMap map[string]interface{}, referencedObjects map[string]bool) {
	for key, value := range schemaJsonMap {
		if key == bean.RefTypeKey {
			valStr, ok := value.(string)
			if ok && strings.HasPrefix(valStr, bean.ReferencesPrefix) {
				schemaJsonMap[bean.RefKey] = valStr //adding $ref key for FE schema parsing
				delete(schemaJsonMap, bean.TypeKey) //deleting type because FE will be using $ref and thus type will be invalid
				referencedObjects[valStr] = true
			}
		} else {
			schemaUpdatedWithRef := resolveValForIteration(value, referencedObjects)
			schemaJsonMap[key] = schemaUpdatedWithRef
		}
	}
}

func resolveValForIteration(value interface{}, referencedObjects map[string]bool) interface{} {
	schemaUpdatedWithRef := value
	if valNew, ok := value.(map[string]interface{}); ok {
		getRefTypeInJsonAndAddRefKey(valNew, referencedObjects)
		schemaUpdatedWithRef = valNew
	} else if valArr, ok := value.([]interface{}); ok {
		for index, val := range valArr {
			schemaUpdatedWithRefNew := resolveValForIteration(val, referencedObjects)
			valArr[index] = schemaUpdatedWithRefNew
		}
		schemaUpdatedWithRef = valArr
	}
	return schemaUpdatedWithRef
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
	getRefTypeInJsonAndAddRefKey(schemaJsonMap, referencedPaths)

	//marshaling new schema with $ref keys
	responseSchemaByte, err := json.Marshal(schemaJsonMap)
	if err != nil {
		impl.logger.Errorw("error in json marshaling update schema with ref keys", "err", err)
		return nil, err
	}

	//we need to get metadata from the resource schema because it is the only part which is being used at UI.
	//In future iterations, this should be removed and made generic for the user to work on the whole object.
	responseSchemaResult := gjson.Get(string(responseSchemaByte), bean.ResourceSchemaMetadataPath)
	responseSchema := responseSchemaResult.String()
	for refPath := range referencedPaths {
		refPathSplit := strings.Split(refPath, "/")
		if len(refPathSplit) < 3 {
			return nil, fmt.Errorf("invalid schema found, references not mentioned correctly")
		}
		resourceKind := refPathSplit[2]
		//referencedItem := make(map[string]interface{})
		if resourceKind == string(bean.DEVTRON_RESOURCE_USER) {
			userModel, err := impl.userRepository.GetAllExcludingApiTokenUser()
			if err != nil {
				impl.logger.Errorw("error while fetching all users", "err", err, "resource kind", resourceKind)
				return nil, err
			}
			//creating enums and enumNames
			enums := make([]map[string]interface{}, 0, len(userModel))
			enumNames := make([]interface{}, 0, len(userModel))
			for _, user := range userModel {
				enum := make(map[string]interface{})
				enum[bean.IdKey] = user.Id
				enum[bean.NameKey] = user.EmailId
				enum[bean.IconKey] = true // to get image from user profile when it is done
				enums = append(enums, enum)

				//currently we are referring only object, in future if reference is some field(got from refPathSplit) then we will pass its data as it is
				enumNames = append(enumNames, user.EmailId)
			}

			//updating schema with enum and enumNames
			referencesUpdatePathCommon := fmt.Sprintf("%s.%s", bean.ReferencesKey, resourceKind)
			referenceUpdatePathEnum := fmt.Sprintf("%s.%s", referencesUpdatePathCommon, bean.EnumKey)
			referenceUpdatePathEnumNames := fmt.Sprintf("%s.%s", referencesUpdatePathCommon, bean.EnumNamesKey)
			responseSchema, err = sjson.Set(responseSchema, referenceUpdatePathEnum, enums)
			if err != nil {
				impl.logger.Errorw("error in setting references enum in resourceSchema", "err", err)
				return nil, err
			}
			responseSchema, err = sjson.Set(responseSchema, referenceUpdatePathEnumNames, enumNames)
			if err != nil {
				impl.logger.Errorw("error in setting references enumNames in resourceSchema", "err", err)
				return nil, err
			}
		} else {
			impl.logger.Errorw("error while extracting kind of resource; kind not supported as of now", "resource kind", resourceKind)
			return nil, errors.New(fmt.Sprintf("%s kind is not supported", resourceKind))
		}
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
		Schema:                              responseSchema,
	}
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		metadataObject := gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectMetadataPath)
		resourceObject.ObjectData = metadataObject.String()
	}
	return resourceObject, nil
}

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error {
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		return err
	}
	schema := devtronResourceSchema.Schema

	//we need to put the object got from UI at overview.metadata path since only this part is controlled from UI currently
	objectDataGeneral, err := sjson.Set("", bean.ResourceObjectMetadataPath, json.RawMessage(reqBean.ObjectData))
	if err != nil {
		impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
		return err
	}
	objectDataGeneral, err = sjson.Set(objectDataGeneral, bean.VersionKey, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
		return err
	}

	kindForSchema := reqBean.Kind
	if len(reqBean.SubKind) > 0 {
		kindForSchema += fmt.Sprintf("/%s", reqBean.SubKind)
	}

	objectDataGeneral, err = sjson.Set(objectDataGeneral, bean.KindKey, kindForSchema)
	if err != nil {
		impl.logger.Errorw("error in setting kind in schema", "err", err, "request", reqBean)
		return err
	}

	objectDataGeneral, err = sjson.Set(objectDataGeneral, bean.ResourceObjectIdPath, reqBean.OldObjectId)
	if err != nil {
		impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
		return err
	}

	//validate user provided json with the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(objectDataGeneral)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		impl.logger.Errorw("error in validating resource object json against schema", "err", err, "request", reqBean, "schema", schema)
		return &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: err.Error(),
			UserMessage:     bean.SchemaValidationFailedErrorUserMessage,
		}
	} else if !result.Valid() {
		impl.logger.Errorw("error in validating resource object json against schema", "result", result, "request", reqBean, "schema", schema)
		errStr := ""
		for _, errResult := range result.Errors() {
			errStr += fmt.Sprintln(errResult.String())
		}
		return &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: errStr,
			UserMessage:     bean.SchemaValidationFailedErrorUserMessage,
		}
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
		devtronResourceObject.ObjectData = objectDataGeneral
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
			ObjectData:              objectDataGeneral,
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
