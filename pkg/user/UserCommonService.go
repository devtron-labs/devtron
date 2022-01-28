package user

import (
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"strings"
)

type UserCommonService interface {
	RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo, existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel, tx *pg.Tx) ([]casbin2.Policy, error)
	RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx) ([]casbin2.Policy, error)
}

type UserCommonServiceImpl struct {
	userAuthRepository  repository2.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository2.UserRepository
	roleGroupRepository repository2.RoleGroupRepository
	sessionManager2     *middleware.SessionManager
}

func NewUserCommonServiceImpl(userAuthRepository repository2.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository2.UserRepository,
	userGroupRepository repository2.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager) *UserCommonServiceImpl {
	serviceImpl := &UserCommonServiceImpl{
		userAuthRepository:  userAuthRepository,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: userGroupRepository,
		sessionManager2:     sessionManager2,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo, existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel, tx *pg.Tx) ([]casbin2.Policy, error) {
	var eliminatedPolicies []casbin2.Policy
	// DELETE Removed Items
	for _, roleFilter := range userInfo.RoleFilters {
		if roleFilter.EntityName == "" {
			roleFilter.EntityName = "NONE"
		}
		if roleFilter.Environment == "" {
			roleFilter.Environment = "NONE"
		}
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		for _, environment := range environments {
			for _, entityName := range entityNames {
				if entityName == "NONE" {
					entityName = ""
				}
				if environment == "NONE" {
					environment = ""
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
				if err != nil {
					impl.logger.Errorw("Error in fetching roles by filter", "user", userInfo)
					return nil, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
					userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
					continue
				}
				if _, ok := existingRoleIds[roleModel.Id]; ok {
					delete(eliminatedRoleIds, roleModel.Id)
				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request

	for _, userRoleModel := range eliminatedRoleIds {
		_, err := impl.userAuthRepository.DeleteUserRoleMapping(userRoleModel, tx)
		if err != nil {
			impl.logger.Errorw("Error in delete user role mapping", "user", userInfo)
			return nil, err
		}
		role, err := impl.userAuthRepository.GetRoleById(userRoleModel.RoleId)
		if err != nil {
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(role.Role)})
	}
	// DELETE ENDS
	return eliminatedPolicies, nil
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx) ([]casbin2.Policy, error) {
	// Filter out removed items in current request
	//var policies []casbin2.Policy
	for _, roleFilter := range request.RoleFilters {
		if roleFilter.EntityName == "" {
			roleFilter.EntityName = "NONE"
		}
		if roleFilter.Environment == "" {
			roleFilter.Environment = "NONE"
		}
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		for _, environment := range environments {
			for _, entityName := range entityNames {
				if entityName == "NONE" {
					entityName = ""
				}
				if environment == "NONE" {
					environment = ""
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
				if err != nil {
					impl.logger.Errorw("Error in fetching roles by filter", "user", request)
					return nil, err
				}
				if roleModel.Id == 0 {
					impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
					request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
					continue
				}
				//roleModel := roleModels[0]
				if _, ok := existingRoles[roleModel.Id]; ok {
					delete(eliminatedRoles, roleModel.Id)
				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	var eliminatedPolicies []casbin2.Policy
	for _, model := range eliminatedRoles {
		_, err := impl.roleGroupRepository.DeleteRoleGroupRoleMapping(model, tx)
		if err != nil {
			return nil, err
		}
		role, err := impl.userAuthRepository.GetRoleById(model.RoleId)
		if err != nil {
			return nil, err
		}
		policyGroup, err := impl.roleGroupRepository.GetRoleGroupById(model.RoleGroupId)
		if err != nil {
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(policyGroup.CasbinName), Obj: casbin2.Object(role.Role)})
	}
	return eliminatedPolicies, nil
}

func containsArr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
