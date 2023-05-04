package user

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/bean"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

type RbacRoleService interface {
	GetDefaultRoleDetail(roleId int) (*bean.RbacRoleDto, error)
	GetAllDefaultRolesByEntityAccessType(entity, accessType string) ([]*bean.RbacRoleDto, error)
	GetAllDefaultRoles() ([]*bean.RbacRoleDto, error)
	GetRbacPolicyResourceListForAllEntityAccessTypes() ([]*bean.RbacPolicyEntityGroupDto, error)
	GetPolicyResourceListByEntityAccessType(entity, accessType string) (*bean.RbacPolicyEntityGroupDto, error)
	CreateDefaultRole(requestDto *bean.RbacRoleDto, userId int32) error
	UpdateDefaultRole(requestDto *bean.RbacRoleDto, userId int32) error
}

type RbacRoleServiceImpl struct {
	logger                             *zap.SugaredLogger
	rbacPolicyResourceDetailRepository repository.RbacPolicyResourceDetailRepository
	rbacRoleResourceDetailRepository   repository.RbacRoleResourceDetailRepository
	rbacRoleDataRepository             repository.RbacRoleDataRepository
	rbacPolicyDataRepository           repository.RbacPolicyDataRepository
	rbacDataCacheFactory               repository.RbacDataCacheFactory
	userAuthRepository                 repository.UserAuthRepository
	userCommonService                  UserCommonService
}

func NewRbacRoleServiceImpl(logger *zap.SugaredLogger,
	rbacPolicyResourceDetailRepository repository.RbacPolicyResourceDetailRepository,
	rbacRoleResourceDetailRepository repository.RbacRoleResourceDetailRepository,
	rbacRoleDataRepository repository.RbacRoleDataRepository,
	rbacPolicyDataRepository repository.RbacPolicyDataRepository,
	rbacDataCacheFactory repository.RbacDataCacheFactory,
	userAuthRepository repository.UserAuthRepository,
	userCommonService UserCommonService,
) *RbacRoleServiceImpl {
	return &RbacRoleServiceImpl{
		logger:                             logger,
		rbacPolicyResourceDetailRepository: rbacPolicyResourceDetailRepository,
		rbacRoleResourceDetailRepository:   rbacRoleResourceDetailRepository,
		rbacRoleDataRepository:             rbacRoleDataRepository,
		rbacPolicyDataRepository:           rbacPolicyDataRepository,
		rbacDataCacheFactory:               rbacDataCacheFactory,
		userAuthRepository:                 userAuthRepository,
		userCommonService:                  userCommonService,
	}
}

func (impl *RbacRoleServiceImpl) GetDefaultRoleDetail(roleId int) (*bean.RbacRoleDto, error) {
	//getting all roles from default data repository
	defaultRole, err := impl.rbacRoleDataRepository.GetById(roleId)
	if err != nil {
		impl.logger.Errorw("error in getting default role by id", "err", err, "id", roleId)
		return nil, err
	}
	//getting default role by entity, accessType and role
	defaultPolicy, err := impl.rbacPolicyDataRepository.GetPolicyByRoleDetails(defaultRole.Entity, defaultRole.AccessType, defaultRole.Role)
	if err != nil {
		impl.logger.Errorw("error in getting default policy by entity, accessType and role", "err", err)
		return nil, err
	}
	entityAccessType := getEntityAccessTypeString(defaultRole.Entity, defaultRole.AccessType)
	//getting policy resource list by entity and accessType
	policyResourceList, err := impl.rbacPolicyResourceDetailRepository.GetPolicyResourceDetailByEntityAccessType(entityAccessType)
	if err != nil {
		impl.logger.Errorw("error in getting all resource detail for default policy by entity & accessType", "err", err)
		return nil, err
	}
	//map of policyResourceValue and resource
	policyResourceMap := make(map[string]string)
	for _, policyResource := range policyResourceList {
		policyResourceMap[policyResource.PolicyResourceValue] = policyResource.Resource
	}
	var defaultPolicyDataObj repository.PolicyCacheDetailObj
	err = json.Unmarshal([]byte(defaultPolicy.PolicyData), &defaultPolicyDataObj)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling default policy data", "err", err)
		return nil, err
	}
	resourceDetailList, err := impl.getPolicyResourceDetailListForRoleFromDefaultPolicy(defaultPolicyDataObj, policyResourceMap)
	if err != nil {
		impl.logger.Errorw("error, getPolicyResourceDetailListForRoleFromDefaultPolicy", "err", err)
		return nil, err
	}
	defaultRoleResp := &bean.RbacRoleDto{
		Id:              roleId,
		RoleName:        defaultRole.Role,
		RoleDisplayName: defaultRole.RoleDisplayName,
		RoleDescription: defaultRole.RoleDescription,
		RbacPolicyEntityGroupDto: &bean.RbacPolicyEntityGroupDto{
			Entity:             defaultRole.Entity,
			AccessType:         defaultRole.AccessType,
			ResourceDetailList: resourceDetailList,
		},
	}
	return defaultRoleResp, nil
}

func (impl *RbacRoleServiceImpl) GetAllDefaultRolesByEntityAccessType(entity, accessType string) ([]*bean.RbacRoleDto, error) {
	//getting all roles from default data repository
	defaultRoles, err := impl.rbacRoleDataRepository.GetRoleDataByEntityAccessType(entity, accessType)
	if err != nil {
		impl.logger.Errorw("error in getting all default roles data by entity and accessType", "entity", entity, "accessType", accessType, "err", err)
		return nil, err
	}
	defaultRolesResp := make([]*bean.RbacRoleDto, 0, len(defaultRoles))
	for _, defaultRole := range defaultRoles {
		defaultRoleResp := &bean.RbacRoleDto{
			Id:              defaultRole.Id,
			RoleName:        defaultRole.Role,
			RoleDisplayName: defaultRole.RoleDisplayName,
			RoleDescription: defaultRole.RoleDescription,
			RbacPolicyEntityGroupDto: &bean.RbacPolicyEntityGroupDto{
				Entity:     defaultRole.Entity,
				AccessType: defaultRole.AccessType,
			},
		}
		defaultRolesResp = append(defaultRolesResp, defaultRoleResp)
	}
	return defaultRolesResp, nil
}

func (impl *RbacRoleServiceImpl) GetAllDefaultRoles() ([]*bean.RbacRoleDto, error) {
	//getting all roles from default data repository
	defaultRoles, err := impl.rbacRoleDataRepository.GetRoleDataForAllRoles()
	if err != nil {
		impl.logger.Errorw("error in getting all default roles data", "err", err)
		return nil, err
	}
	defaultRolesResp := make([]*bean.RbacRoleDto, 0, len(defaultRoles))
	for _, defaultRole := range defaultRoles {
		defaultRoleResp := &bean.RbacRoleDto{
			Id:              defaultRole.Id,
			RoleName:        defaultRole.Role,
			RoleDisplayName: defaultRole.RoleDisplayName,
			RoleDescription: defaultRole.RoleDescription,
			RbacPolicyEntityGroupDto: &bean.RbacPolicyEntityGroupDto{
				Entity:     defaultRole.Entity,
				AccessType: defaultRole.AccessType,
			},
		}
		defaultRolesResp = append(defaultRolesResp, defaultRoleResp)
	}
	return defaultRolesResp, nil
}

func (impl *RbacRoleServiceImpl) GetRbacPolicyResourceListForAllEntityAccessTypes() ([]*bean.RbacPolicyEntityGroupDto, error) {
	//getting all resource detail for policy from db
	rbacPolicyResourceList, err := impl.rbacPolicyResourceDetailRepository.GetAllPolicyResourceDetail()
	if err != nil {
		impl.logger.Errorw("error in getting all rbac policy resource list", "err", err)
		return nil, err
	}
	//map of entity/accessType or entity and their corresponding resource detail list
	rbacPolicyResourceEntityMap := make(map[string][]*bean.RbacPolicyResource)
	for _, rbacPolicyResource := range rbacPolicyResourceList {
		resourceDetailDto := &bean.RbacPolicyResource{
			Resource: rbacPolicyResource.Resource,
			Actions:  rbacPolicyResource.AllowedActions,
		}
		for _, entityAccessType := range rbacPolicyResource.EligibleEntityAccessTypes {
			if list, ok := rbacPolicyResourceEntityMap[entityAccessType]; ok {
				list = append(list, resourceDetailDto)
				rbacPolicyResourceEntityMap[entityAccessType] = list
			} else {
				rbacPolicyResourceEntityMap[entityAccessType] = []*bean.RbacPolicyResource{resourceDetailDto}
			}
		}
	}
	rbacPolicyGroupListDto := make([]*bean.RbacPolicyEntityGroupDto, 0)
	for entityAccessType, policyResourceDetailList := range rbacPolicyResourceEntityMap {
		var entity, accessType string
		entitySplit := strings.Split(entityAccessType, "/")
		switch len(entitySplit) {
		case 1:
			entity = entitySplit[0]
		case 2:
			entity = entitySplit[0]
			accessType = entitySplit[1]
		default:
			continue
		}
		rbacPolicyGroupDto := &bean.RbacPolicyEntityGroupDto{
			Entity:             entity,
			AccessType:         accessType,
			ResourceDetailList: policyResourceDetailList,
		}
		rbacPolicyGroupListDto = append(rbacPolicyGroupListDto, rbacPolicyGroupDto)
	}
	return rbacPolicyGroupListDto, nil
}

func (impl *RbacRoleServiceImpl) CreateDefaultRole(requestDto *bean.RbacRoleDto, userId int32) error {
	entityAccessType := getEntityAccessTypeString(requestDto.Entity, requestDto.AccessType)
	//initiating a transaction
	dbConnection := impl.rbacPolicyResourceDetailRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return err
	}
	//creating new default role entry
	rolePValObj, err := impl.createOrUpdateDefaultRoleDataEntry(nil, requestDto, entityAccessType, userId, tx)
	if err != nil {
		impl.logger.Errorw("service error, createOrUpdateDefaultRoleDataEntry", "err", err, "payload", requestDto)
		return err
	}
	//creating/updating policy data entry
	_, err = impl.createOrUpdateNewDefaultPolicyDataEntry(nil, requestDto, entityAccessType, rolePValObj, userId, tx)
	if err != nil {
		impl.logger.Errorw("service error, createOrUpdateNewDefaultPolicyDataEntry", "err", err, "payload", requestDto)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	// syncing default role cache
	go impl.rbacDataCacheFactory.SyncRoleDataCache()
	// syncing default policy cache
	go impl.rbacDataCacheFactory.SyncPolicyCache()
	return nil
}

func (impl *RbacRoleServiceImpl) UpdateDefaultRole(requestDto *bean.RbacRoleDto, userId int32) error {
	//get default role entry
	oldDefaultRole, err := impl.rbacRoleDataRepository.GetById(requestDto.Id)
	if err != nil {
		impl.logger.Errorw("error in getting default role data by id", "err", err, "id", requestDto.Id)
		return err
	}
	entityAccessType := getEntityAccessTypeString(requestDto.Entity, requestDto.AccessType)
	//initiating a transaction
	dbConnection := impl.rbacPolicyResourceDetailRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return err
	}
	//getting old defaultPolicyData from db
	oldDefaultPolicy, err := impl.rbacPolicyDataRepository.GetPolicyByRoleDetails(requestDto.Entity, requestDto.AccessType, requestDto.RoleName)
	if err != nil {
		impl.logger.Errorw("error in getting default policy by role details", "err", err, "entity", requestDto.Entity, "accessType", requestDto.AccessType, "role", requestDto.RoleName)
		return err
	}
	var oldDefaultPolicyObj repository.PolicyCacheDetailObj
	err = json.Unmarshal([]byte(oldDefaultPolicy.PolicyData), &oldDefaultPolicyObj)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling, getting old default policy data obj", "err", err)
		return err
	}
	//creating new default role entry
	rolePValObj, err := impl.createOrUpdateDefaultRoleDataEntry(oldDefaultRole, requestDto, entityAccessType, userId, tx)
	if err != nil {
		impl.logger.Errorw("service error, createOrUpdateDefaultRoleDataEntry", "err", err, "payload", requestDto)
		return err
	}
	//creating/updating policy data entry
	newDefaultPolicyObj, err := impl.createOrUpdateNewDefaultPolicyDataEntry(oldDefaultPolicy, requestDto, entityAccessType, rolePValObj, userId, tx)
	if err != nil {
		impl.logger.Errorw("service error, createOrUpdateNewDefaultPolicyDataEntry", "err", err, "payload", requestDto)
		return err
	}
	//TODO: check if need to make this async
	if requestDto.UpdatePoliciesForExistingProvidedRoles {
		err = impl.UpdateExistingUserOrGroupPolicies(&oldDefaultPolicyObj, newDefaultPolicyObj, requestDto.Entity, requestDto.AccessType, requestDto.RoleName)
		if err != nil {
			impl.logger.Errorw("error, UpdateExistingUserOrGroupPolicies", "err", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	// syncing default role cache
	go impl.rbacDataCacheFactory.SyncRoleDataCache()
	// syncing default policy cache
	go impl.rbacDataCacheFactory.SyncPolicyCache()
	return nil
}

func (impl *RbacRoleServiceImpl) GetPolicyResourceListByEntityAccessType(entity, accessType string) (*bean.RbacPolicyEntityGroupDto, error) {
	entityAccessType := getEntityAccessTypeString(entity, accessType)
	//getting all resource details needed for creating default policy
	policyResourceList, err := impl.rbacPolicyResourceDetailRepository.GetPolicyResourceDetailByEntityAccessType(entityAccessType)
	if err != nil {
		impl.logger.Errorw("error in getting all resource detail for default policy by entity & accessType", "entityAccessType", entityAccessType, "err", err)
		return nil, err
	}
	resourceDetailList := make([]*bean.RbacPolicyResource, 0, len(policyResourceList))
	for _, rbacPolicyResource := range policyResourceList {
		resourceDetailDto := &bean.RbacPolicyResource{
			Resource: rbacPolicyResource.Resource,
			Actions:  rbacPolicyResource.AllowedActions,
		}
		resourceDetailList = append(resourceDetailList, resourceDetailDto)
	}
	rbacPolicyGroupDto := &bean.RbacPolicyEntityGroupDto{
		Entity:             entity,
		AccessType:         accessType,
		ResourceDetailList: resourceDetailList,
	}
	return rbacPolicyGroupDto, nil
}

func (impl *RbacRoleServiceImpl) getPolicyResourceDetailListForRoleFromDefaultPolicy(defaultPolicyData repository.PolicyCacheDetailObj,
	policyResourceMap map[string]string) ([]*bean.RbacPolicyResource, error) {
	//map of resource and array of its actions
	rbacPolicyResourceMap := make(map[string][]string)
	for _, resActObj := range defaultPolicyData.ResActObjSet {
		for policyResourceValue, resource := range policyResourceMap {
			isEqual, err := util.IsAJSONStringAndAnInterfaceEqual(policyResourceValue, resActObj.Res)
			if err != nil {
				impl.logger.Errorw("error in checking if policyResourceValue and default policy resource is equal or not", "err", err, "policyResourceValue", policyResourceValue, "resource", resActObj.Res)
				return nil, err
			}
			if isEqual {
				if actionArr, ok := rbacPolicyResourceMap[resource]; ok {
					//assuming action will not have indexKeyMap as action cannot be dynamic
					actionArr = append(actionArr, resActObj.Act.Value)
					rbacPolicyResourceMap[resource] = actionArr
				} else {
					actionArr = []string{resActObj.Act.Value}
					rbacPolicyResourceMap[resource] = actionArr
				}
				break
			}
		}
	}
	rbacPolicyResourceArr := make([]*bean.RbacPolicyResource, 0, len(rbacPolicyResourceMap))
	for resource, actions := range rbacPolicyResourceMap {
		rbacPolicyResource := &bean.RbacPolicyResource{
			Resource: resource,
			Actions:  actions,
		}
		rbacPolicyResourceArr = append(rbacPolicyResourceArr, rbacPolicyResource)
	}
	return rbacPolicyResourceArr, nil
}

func (impl *RbacRoleServiceImpl) createOrUpdateDefaultRoleDataEntry(oldDefaultRole *repository.RbacRoleData, requestDto *bean.RbacRoleDto, entityAccessType string, userId int32, tx *pg.Tx) (repository.PValDetailObj, error) {
	rolePValObj := repository.PValDetailObj{}
	//getting all resources needed for creating default role
	roleResourceList, err := impl.rbacRoleResourceDetailRepository.GetRoleResourceDetailByEntityAccessType(entityAccessType)
	if err != nil {
		impl.logger.Errorw("error in getting all resource detail for role by entity & accessType", "err", err)
		return rolePValObj, err
	}
	rolePValObj, defaultRoleData, err := impl.getDefaultRoleData(entityAccessType, requestDto, roleResourceList)
	if err != nil {
		impl.logger.Errorw("error in getting default role data", "err", err)
		return rolePValObj, err
	}
	id := 0
	isPresetRole := false
	if oldDefaultRole != nil {
		id = oldDefaultRole.Id
		isPresetRole = oldDefaultRole.IsPresetRole
	}
	rbacRoleDataModel := &repository.RbacRoleData{
		Id:              id,
		Entity:          requestDto.Entity,
		AccessType:      requestDto.AccessType,
		Role:            requestDto.RoleName,
		RoleDisplayName: requestDto.RoleDisplayName,
		RoleData:        defaultRoleData,
		RoleDescription: requestDto.RoleDescription,
		IsPresetRole:    isPresetRole,
		Deleted:         false,
		AuditLog: sql.AuditLog{
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}

	if id == 0 {
		rbacRoleDataModel.CreatedOn = time.Now()
		rbacRoleDataModel.CreatedBy = userId
		_, err = impl.rbacRoleDataRepository.CreateNewRoleDataForRoleWithTxn(rbacRoleDataModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new default role data entry", "err", err, "model", rbacRoleDataModel)
			return rolePValObj, err
		}
	} else {
		_, err = impl.rbacRoleDataRepository.UpdateRoleDataForRoleWithTxn(rbacRoleDataModel, tx)
		if err != nil {
			impl.logger.Errorw("error in updating default role data entry", "err", err, "model", rbacRoleDataModel)
			return rolePValObj, err
		}
	}
	return rolePValObj, nil
}

func (impl *RbacRoleServiceImpl) createOrUpdateNewDefaultPolicyDataEntry(oldDefaultPolicy *repository.RbacPolicyData, requestDto *bean.RbacRoleDto,
	entityAccessType string, rolePValObj repository.PValDetailObj, userId int32, tx *pg.Tx) (*repository.PolicyCacheDetailObj, error) {
	var defaultPolicy *repository.PolicyCacheDetailObj
	//getting all resource details needed for creating default policy
	policyResourceList, err := impl.rbacPolicyResourceDetailRepository.GetPolicyResourceDetailByEntityAccessType(entityAccessType)
	if err != nil {
		impl.logger.Errorw("error in getting all resource detail for default policy by entity & accessType", "err", err)
		return defaultPolicy, err
	}
	policyResourceListMap := make(map[string]*repository.RbacPolicyResourceDetail)
	for _, policyResource := range policyResourceList {
		policyResourceListMap[policyResource.Resource] = policyResource
	}
	defaultPolicy, err = impl.getDefaultPolicyDataObj(rolePValObj, requestDto.ResourceDetailList, policyResourceListMap)
	if err != nil {
		impl.logger.Errorw("error in getting default policy data", "err", err)
		return defaultPolicy, err
	}
	//marshaling default policy data
	marshalledDefaultPolicy, err := json.Marshal(defaultPolicy)
	if err != nil {
		impl.logger.Errorw("error in marshalling default policy data", "defaultPolicy", defaultPolicy, "err", err)
		return defaultPolicy, err
	}
	id := 0
	isPresetRole := false
	if oldDefaultPolicy != nil {
		id = oldDefaultPolicy.Id
		isPresetRole = oldDefaultPolicy.IsPresetRole
	}
	defaultPolicyData := string(marshalledDefaultPolicy)
	rbacPolicyDataModel := &repository.RbacPolicyData{
		Id:           id,
		Entity:       requestDto.Entity,
		AccessType:   requestDto.AccessType,
		Role:         requestDto.RoleName,
		PolicyData:   defaultPolicyData,
		IsPresetRole: isPresetRole,
		Deleted:      false,
		AuditLog: sql.AuditLog{
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	if id == 0 {
		rbacPolicyDataModel.CreatedOn = time.Now()
		rbacPolicyDataModel.CreatedBy = userId
		_, err = impl.rbacPolicyDataRepository.CreateNewPolicyDataForRoleWithTxn(rbacPolicyDataModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new default policy data entry", "err", err, "model", rbacPolicyDataModel)
			return defaultPolicy, err
		}
	} else {
		_, err = impl.rbacPolicyDataRepository.UpdatePolicyDataForRoleWithTxn(rbacPolicyDataModel, tx)
		if err != nil {
			impl.logger.Errorw("error in updating default policy data entry", "err", err, "model", rbacPolicyDataModel)
			return defaultPolicy, err
		}
	}
	return defaultPolicy, nil
}

func (impl *RbacRoleServiceImpl) getDefaultRoleData(entityAccessType string, requestDto *bean.RbacRoleDto,
	roleResourceList []*repository.RbacRoleResourceDetail) (repository.PValDetailObj, string, error) {
	defaultRoleData := ""

	//creating default role
	defaultRole := &repository.RoleCacheDetailObj{}

	//we will update [entity, accessType, role, action] details separately
	//and remaining fields will be updated on the basis of db data

	//setting entity
	defaultRole.Entity = repository.PValDetailObj{
		Value: requestDto.Entity,
	}
	//setting accessType
	defaultRole.AccessType = repository.PValDetailObj{
		Value: requestDto.AccessType,
	}
	//setting action
	defaultRole.Action = repository.PValDetailObj{
		Value: requestDto.RoleName,
	}
	role := fmt.Sprintf("%s:%s", entityAccessType, requestDto.RoleName)
	roleIndexKeyMap := make(map[int]repository.PValUpdateKey)
	for _, roleResource := range roleResourceList {
		role += fmt.Sprint("_%")
		roleIndexKeyMap[len(role)-1] = roleResource.RoleResourceUpdateKey
		pValDetailObj := repository.PValDetailObj{
			Value:       bean.PValObjIndexReplacePlaceholder,
			IndexKeyMap: map[int]repository.PValUpdateKey{0: roleResource.RoleResourceUpdateKey},
		}
		//setting pValDetailObj
		reflect.ValueOf(defaultRole).Elem().FieldByName(roleResource.RoleResourceKey).Set(reflect.ValueOf(pValDetailObj))
	}
	rolePValObj := repository.PValDetailObj{
		Value:       role,
		IndexKeyMap: roleIndexKeyMap,
	}
	//setting role
	defaultRole.Role = rolePValObj
	//marshaling default role data
	marshalledDefaultRole, err := json.Marshal(defaultRole)
	if err != nil {
		impl.logger.Errorw("error in marshalling default role data", "err", err, "defaultRole", defaultRole)
		return rolePValObj, defaultRoleData, err
	}
	defaultRoleData = string(marshalledDefaultRole)
	return rolePValObj, defaultRoleData, nil
}

func (impl *RbacRoleServiceImpl) getDefaultPolicyDataObj(rolePValObj repository.PValDetailObj,
	resourceDetailList []*bean.RbacPolicyResource,
	policyResourceListMap map[string]*repository.RbacPolicyResourceDetail) (*repository.PolicyCacheDetailObj, error) {
	//creating default policy
	defaultPolicy := &repository.PolicyCacheDetailObj{
		Type: repository.PValDetailObj{
			Value: bean.PTypePolicy,
		},
		Sub: rolePValObj,
	}
	resActObjArr, err := impl.getResActObjArr(resourceDetailList, policyResourceListMap)
	if err != nil {
		impl.logger.Errorw("error in getting resActObjArr", "err", err)
		return defaultPolicy, err
	}
	defaultPolicy.ResActObjSet = resActObjArr
	return defaultPolicy, nil
}

func (impl *RbacRoleServiceImpl) getResActObjArr(resourceDetailList []*bean.RbacPolicyResource, policyResourceListMap map[string]*repository.RbacPolicyResourceDetail) ([]repository.ResActObj, error) {
	//making resource-action-object array with capacity = len of resource detail list because minimum one action per resource will be there
	resActObjArr := make([]repository.ResActObj, 0, len(resourceDetailList))
	for _, policyResourceDetail := range resourceDetailList {
		if policyResource, ok := policyResourceListMap[policyResourceDetail.Resource]; ok {
			var resource repository.PValDetailObj
			var resourceObj repository.PValDetailObj
			err := json.Unmarshal([]byte(policyResource.PolicyResourceValue), &resource)
			if err != nil {
				impl.logger.Errorw("error in unmarshalling policy resource value", "err", err)
				return nil, err
			}
			err = json.Unmarshal([]byte(policyResource.ResourceObject), &resourceObj)
			if err != nil {
				impl.logger.Errorw("error in unmarshalling policy resource value", "err", err)
				return nil, err
			}
			//map of action and resActObj
			resActObjArrForAPolicy := make([]repository.ResActObj, 0, len(policyResourceDetail.Actions))
			for _, action := range policyResourceDetail.Actions {
				resActObj := repository.ResActObj{
					Res: resource,
					Act: repository.PValDetailObj{
						Value: action,
					},
					Obj: resourceObj,
				}
				if action == bean.AllObjectAccessPlaceholder { // if action contains all then will ignore all other actions got in request for this resource
					resActObjArrForAPolicy = []repository.ResActObj{
						resActObj,
					}
					break
				} else {
					resActObjArrForAPolicy = append(resActObjArrForAPolicy, resActObj)
				}
			}
			resActObjArr = append(resActObjArr, resActObjArrForAPolicy...)
		}
	}
	return resActObjArr, nil
}

func getEntityAccessTypeString(entity, accessType string) string {
	entityAccessType := entity
	if len(accessType) > 0 {
		entityAccessType += fmt.Sprintf("/%s", accessType)
	}
	return entityAccessType
}

func (impl *RbacRoleServiceImpl) UpdateExistingUserOrGroupPolicies(oldPolicyDetailObj, newPolicyDetailObj *repository.PolicyCacheDetailObj,
	entity, accessType, role string) error {
	//getting difference between old and new policy set
	addedPolicies, deletedPolicies := getDiffBetweenPolicies(oldPolicyDetailObj.ResActObjSet, newPolicyDetailObj.ResActObjSet)
	//getting all roles for this entity, accessType and action(role)
	roles, err := impl.userAuthRepository.GetRolesByEntityAccessTypeAndAction(entity, accessType, role)
	if err != nil {
		impl.logger.Errorw("error in getting roles by entity, accessType and role", "err", err)
		return err
	}
	addedPoliciesObj := repository.PolicyCacheDetailObj{
		Type:         newPolicyDetailObj.Type,
		Sub:          newPolicyDetailObj.Sub,
		ResActObjSet: addedPolicies,
	}
	deletedPoliciesObj := repository.PolicyCacheDetailObj{
		Type:         newPolicyDetailObj.Type,
		Sub:          newPolicyDetailObj.Sub,
		ResActObjSet: deletedPolicies,
	}
	addedPoliciesLen := len(addedPolicies)
	deletedPoliciesLen := len(deletedPolicies)
	rolesLen := len(roles)
	addedPoliciesTotal := make([]casbin.Policy, 0, rolesLen*addedPoliciesLen)
	deletedPoliciesTotal := make([]casbin.Policy, 0, rolesLen*deletedPoliciesLen)
	for _, roleModel := range roles {
		pValUpdateMap := impl.userCommonService.GetPValUpdateMap(roleModel.Team, roleModel.EntityName,
			roleModel.Environment, roleModel.Entity, roleModel.Cluster, roleModel.Namespace, roleModel.Group, roleModel.Kind, roleModel.Resource, roleModel.Approver)
		if addedPoliciesLen > 0 {
			addedPoliciesObjCopy := addedPoliciesObj
			renderedAddedPolicies := impl.userCommonService.GetRenderedPolicy(addedPoliciesObjCopy, pValUpdateMap)
			addedPoliciesTotal = append(addedPoliciesTotal, renderedAddedPolicies...)
		}
		if deletedPoliciesLen > 0 {
			deletedPoliciesObjCopy := deletedPoliciesObj
			renderedDeletedPolicies := impl.userCommonService.GetRenderedPolicy(deletedPoliciesObjCopy, pValUpdateMap)
			deletedPoliciesTotal = append(deletedPoliciesTotal, renderedDeletedPolicies...)
		}
	}
	casbin.LoadPolicy()
	needToLoadPolicyAgain := false
	if len(addedPoliciesTotal) > 0 {
		err = casbin.AddPolicy(addedPoliciesTotal)
		if err != nil {
			impl.logger.Errorw("error in adding updated policies", "err", err)
			return err
		}
		needToLoadPolicyAgain = true
	}
	if len(deletedPoliciesTotal) > 0 {
		_ = casbin.RemovePolicy(deletedPoliciesTotal)
		needToLoadPolicyAgain = true
	}
	if needToLoadPolicyAgain {
		//loading policy for syncing orchestrator to casbin with  newly added or deleted policies
		casbin.LoadPolicy()
	}
	return nil
}

func getDiffBetweenPolicies(oldPolicies, newPolicies []repository.ResActObj) (addedPolicies, deletedPolicies []repository.ResActObj) {
	oldPolicyMap := make(map[string]bool)
	for _, oldPolicy := range oldPolicies {
		//converting all fields of data to a string
		data := fmt.Sprintf("res:%v,act:%v,obj:%v", oldPolicy.Res, oldPolicy.Act, oldPolicy.Obj)
		//creating entry for data, keeping false because if present in new policy
		//then will be set to true and will not be included in deletedPolicies
		oldPolicyMap[data] = false
	}
	for _, newPolicy := range newPolicies {
		//converting all fields of data to a string
		data := fmt.Sprintf("res:%v,act:%v,obj:%v", newPolicy.Res, newPolicy.Act, newPolicy.Obj)
		if _, ok := oldPolicyMap[data]; !ok {
			//data not present in old policy, to be included in addedPolicies
			addedPolicies = append(addedPolicies, newPolicy)
		} else {
			//data present in old policy; set old policy to true, so it does not get included in deletedPolicies
			oldPolicyMap[data] = true
		}
	}
	//checking oldPolicies for getting deletedPolicies
	for _, oldPolicy := range oldPolicies {
		data := fmt.Sprintf("res:%v,act:%v,obj:%v", oldPolicy.Res, oldPolicy.Act, oldPolicy.Obj)
		if presentInNew := oldPolicyMap[data]; !presentInNew {
			//data not present in old policy, to be included in addedPolicies
			deletedPolicies = append(deletedPolicies, oldPolicy)
		}
	}
	return addedPolicies, deletedPolicies
}
