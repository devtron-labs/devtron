package repository

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type RbacDataCacheFactory interface {
	GetDefaultRoleDataAndPolicyByEntityAccessTypeAndRoleType(entity, accessType, roleType string) (RoleCacheDetailObj, PolicyCacheDetailObj, error)
	SyncPolicyCache()
	SyncRoleDataCache()
}

type RbacDataCacheFactoryImpl struct {
	logger                          *zap.SugaredLogger
	policyCache                     map[string]PolicyCacheDetailObj
	roleCache                       map[string]RoleCacheDetailObj
	mutex                           sync.Mutex
	defaultRbacPolicyDataRepository RbacPolicyDataRepository
	defaultRbacRoleDataRepository   RbacRoleDataRepository
}

type PolicyCacheDetailObj struct {
	Type         PValDetailObj `json:"type"`
	Sub          PValDetailObj `json:"sub"`
	ResActObjSet []ResActObj   `json:"resActObjSet"`
}

type RoleCacheDetailObj struct {
	Role        PValDetailObj `json:"role"`
	Entity      PValDetailObj `json:"entity"`
	Team        PValDetailObj `json:"team"`
	EntityName  PValDetailObj `json:"entityName"`
	Environment PValDetailObj `json:"environment"`
	Action      PValDetailObj `json:"action"`
	AccessType  PValDetailObj `json:"accessType"`
	Cluster     PValDetailObj `json:"cluster"`
	Namespace   PValDetailObj `json:"namespace"`
	Group       PValDetailObj `json:"group"`
	Kind        PValDetailObj `json:"kind"`
	Resource    PValDetailObj `json:"resource"`
	Approver    PValDetailObj `json:"approver"`
}

type ResActObj struct {
	Res PValDetailObj `json:"res"`
	Act PValDetailObj `json:"act"`
	Obj PValDetailObj `json:"obj"`
}

type PValDetailObj struct {
	Value       string                `json:"value"`
	IndexKeyMap map[int]PValUpdateKey `json:"indexKeyMap"` //map of index at which replacement is to be done and name of key that is to for updating value
}

type PValResolvedValue struct {
	object string
}

func NewPValResolvedValue(obj string) PValResolvedValue {
	return PValResolvedValue{object: obj}
}

func (value PValResolvedValue) String() string {
	return value.object
}

func (value PValResolvedValue) Boolean() bool {
	parsedVal, err := strconv.ParseBool(value.object)
	if err != nil {
		return false
	}
	return parsedVal
}

func NewRbacDataCacheFactoryImpl(logger *zap.SugaredLogger,
	defaultRbacPolicyDataRepository RbacPolicyDataRepository,
	defaultRbacRoleDataRepository RbacRoleDataRepository) *RbacDataCacheFactoryImpl {
	policyCache := initialisePolicyDataCache()
	roleCache := initialiseRoleDataCache()
	return &RbacDataCacheFactoryImpl{
		logger:                          logger,
		policyCache:                     policyCache,
		roleCache:                       roleCache,
		defaultRbacPolicyDataRepository: defaultRbacPolicyDataRepository,
		defaultRbacRoleDataRepository:   defaultRbacRoleDataRepository,
	}
}

func (impl *RbacDataCacheFactoryImpl) GetDefaultRoleDataAndPolicyByEntityAccessTypeAndRoleType(entity, accessType, roleType string) (RoleCacheDetailObj, PolicyCacheDetailObj, error) {
	defaultPolicyData := PolicyCacheDetailObj{}
	defaultRoleData := RoleCacheDetailObj{}

	//getting key for cache map
	keyForMap := getCacheMapKey(entity, accessType, roleType)

	//checking and getting default policy data from cache
	if val, ok := impl.policyCache[keyForMap]; ok {
		defaultPolicyData = val
	} else {
		impl.logger.Errorw("default role not found", "entity", entity, "accessType", accessType, "roleType", roleType)
		return defaultRoleData, defaultPolicyData, fmt.Errorf("default policy not found")
	}

	//checking and getting default role data from cache
	if val, ok := impl.roleCache[keyForMap]; ok {
		defaultRoleData = val
	} else {
		impl.logger.Errorw("default policy not found", "entity", entity, "accessType", accessType, "roleType", roleType)
		return defaultRoleData, defaultPolicyData, fmt.Errorf("default role not found")
	}

	return defaultRoleData, defaultPolicyData, nil
}

func (impl *RbacDataCacheFactoryImpl) SyncPolicyCache() {
	//getting all default policies
	defaultRbacPolicies, err := impl.defaultRbacPolicyDataRepository.GetPolicyDataForAllRoles()
	if err != nil {
		return
	}
	for _, defaultRbacPolicy := range defaultRbacPolicies {
		policyData := defaultRbacPolicy.PolicyData
		var policyDataObj PolicyCacheDetailObj
		err = json.Unmarshal([]byte(policyData), &policyDataObj)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling policy data", "err", err)
			continue
		}
		keyForMap := getCacheMapKey(defaultRbacPolicy.Entity, defaultRbacPolicy.AccessType, defaultRbacPolicy.Role)
		impl.mutex.Lock()
		impl.policyCache[keyForMap] = policyDataObj
		impl.mutex.Unlock()
	}
}

func (impl *RbacDataCacheFactoryImpl) SyncRoleDataCache() {
	//getting all default policies
	defaultRbacRoles, err := impl.defaultRbacRoleDataRepository.GetRoleDataForAllRoles()
	if err != nil {
		return
	}
	for _, defaultRbacRole := range defaultRbacRoles {
		roleData := defaultRbacRole.RoleData
		var roleDataObj RoleCacheDetailObj
		err = json.Unmarshal([]byte(roleData), &roleDataObj)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling role data", "err", err)
			continue
		}
		keyForMap := getCacheMapKey(defaultRbacRole.Entity, defaultRbacRole.AccessType, defaultRbacRole.Role)
		impl.mutex.Lock()
		impl.roleCache[keyForMap] = roleDataObj
		impl.mutex.Unlock()
	}
}

func initialisePolicyDataCache() map[string]PolicyCacheDetailObj {
	c := make(map[string]PolicyCacheDetailObj)
	return c
}

func initialiseRoleDataCache() map[string]RoleCacheDetailObj {
	c := make(map[string]RoleCacheDetailObj)
	return c
}

func getCacheMapKey(entity, accessType, roleType string) string {
	return fmt.Sprintf("%s_%s_%s", entity, accessType, roleType)
}
