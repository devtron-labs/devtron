package user

import (
	"fmt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/user/dto"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type HelmUserUtil interface {
	CreateUserIfNotExists(userInfo *bean2.UserInfo, emailId string) (*bean2.UserInfo, error)
	MergeRoleFilter(oldR []bean2.RoleFilter, newR []bean2.RoleFilter) []bean2.RoleFilter
	MergeGroups(oldGroups []string, newGroups []string) []string
}

type HelmUserUtilImpl struct {
	logger              *zap.SugaredLogger
	userRepository      HelmUserRepository
	userRoleRepository  HelmUserRoleRepository
	roleGroupRepository HelmUserRoleGroupRepository
	sessionManager      *middleware.SessionManager
}

func NewHelmUserUtilImpl(logger *zap.SugaredLogger,
	userRepository HelmUserRepository, userRoleRepository HelmUserRoleRepository,
	sessionManager *middleware.SessionManager) *HelmUserUtilImpl {
	serviceImpl := &HelmUserUtilImpl{
		logger:             logger,
		userRepository:     userRepository,
		userRoleRepository: userRoleRepository,
		sessionManager:     sessionManager,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl HelmUserUtilImpl) CreateUserIfNotExists(userInfo *bean2.UserInfo, emailId string) (*bean2.UserInfo, error) {
	// if not found, create new user
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	_, err = impl.validateUserRequest(userInfo)
	if err != nil {
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
		return nil, err
	}

	//create new user in our db on d basis of info got from google api or hex. assign a basic role
	model := &HelmUserModel{
		EmailId:     emailId,
		AccessToken: userInfo.AccessToken,
	}
	model.Active = true
	model.CreatedBy = userInfo.UserId
	model.UpdatedBy = userInfo.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model, err = impl.userRepository.CreateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new user", "error", err)
		err = &util.ApiError{
			Code:            constants.UserCreateDBFailed,
			InternalMessage: "failed to create new user in db",
			UserMessage:     fmt.Sprintf("requested by %d", userInfo.UserId),
		}
		return nil, err
	}
	userInfo.Id = model.Id

	//Starts Role and Mapping
	var policies []casbin2.Policy
	if userInfo.SuperAdmin == false {
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
					roleModel, err := impl.userRoleRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						//userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + roleFilter.Environment + "," + roleFilter.Application + "," + roleFilter.Action

						//TODO - create roles from here
						if len(roleFilter.Team) > 0 {
							flag, err := impl.userRoleRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userRoleRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not found for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 {
							flag, err := impl.userRoleRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userRoleRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not found for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else {
							continue
						}
					}
					//roleModel := roleModels[0]
					if roleModel.Id > 0 {
						userRoleModel := &HelmUserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
						userRoleModel, err = impl.userRoleRepository.CreateUserRoleMapping(userRoleModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}

		// START GROUP POLICY
		for _, item := range userInfo.Groups {
			userGroup, err := impl.roleGroupRepository.GetRoleGroupByName(item)
			if err != nil {
				return nil, err
			}
			//object := "group:" + strings.ReplaceAll(item, " ", "_")
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(emailId), Obj: casbin2.Object(userGroup.CasbinName)})
		}
		// END GROUP POLICY
	} else if userInfo.SuperAdmin == true {

		isSuperAdmin := impl.IsSuperAddmin(userInfo.UserId)
		if isSuperAdmin == false {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			return nil, err
		}
		// create super admin policy if found missing in casbin db.
		flag, err := impl.userRoleRepository.CreateUpdateDefaultPoliciesForSuperAdmin(tx)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userRoleRepository.GetRoleByFilter("", "", "", "", "super-admin")
		if err != nil {
			impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &HelmUserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
			userRoleModel, err = impl.userRoleRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, err
			}
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
		}

	}
	if len(policies) > 0 {
		pRes := casbin2.AddPolicy(policies)
		println(pRes)
	}
	//Ends

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (impl HelmUserUtilImpl) validateUserRequest(userInfo *bean2.UserInfo) (bool, error) {
	if len(userInfo.RoleFilters) == 1 &&
		userInfo.RoleFilters[0].Team == "" && userInfo.RoleFilters[0].Environment == "" && userInfo.RoleFilters[0].Action == "" {
		//skip
	} else {
		invalid := false
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.Team) > 0 && len(roleFilter.Action) > 0 {
				// request has dawf or hawf filters
			} else if len(roleFilter.Entity) > 0 {
				// request has chart group role
			} else if len(roleFilter.Cluster) > 0 {
				// request has devtops cluster role
			} else {
				invalid = true
			}
		}
		if invalid {
			err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
			return false, err
		}
	}
	return true, nil
}

func (impl HelmUserUtilImpl) checkUserRoles(id int32) ([]string, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "id", model.Id)
		return nil, err
	}

	return groups, nil
}

func (impl HelmUserUtilImpl) IsSuperAddmin(id int32) bool {
	isSuperAdmin := false
	roles, err := impl.checkUserRoles(id)
	if err != nil {
		return false
	}
	for _, item := range roles {
		if item == bean.SUPERADMIN {
			isSuperAdmin = true
		}
	}
	return isSuperAdmin
}

func (impl HelmUserUtilImpl) MergeRoleFilter(oldR []bean2.RoleFilter, newR []bean2.RoleFilter) []bean2.RoleFilter {
	var roleFilters []bean2.RoleFilter
	keysMap := make(map[string]bool)
	for _, role := range oldR {
		roleFilters = append(roleFilters, bean2.RoleFilter{
			Entity:      role.Entity,
			Team:        role.Team,
			Environment: role.Environment,
			EntityName:  role.EntityName,
			Action:      role.Action,
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action)
		if _, ok := keysMap[key]; !ok {
			roleFilters = append(roleFilters, bean2.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
			})
		}
	}
	return roleFilters
}

func (impl HelmUserUtilImpl) MergeGroups(oldGroups []string, newGroups []string) []string {
	var groups []string
	keysMap := make(map[string]bool)
	for _, group := range oldGroups {
		groups = append(groups, group)
		key := fmt.Sprintf(group)
		keysMap[key] = true
	}
	for _, group := range newGroups {
		key := fmt.Sprintf(group)
		if _, ok := keysMap[key]; !ok {
			groups = append(groups, group)
		}
	}
	return groups
}
