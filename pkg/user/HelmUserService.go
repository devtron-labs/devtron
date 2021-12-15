package user

import (
	"fmt"
	"github.com/devtron-labs/authenticator/jwt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/user/dto"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type HelmUserService interface {
	GetLoggedInUser(r *http.Request) (int32, error)
	CreateUser(userInfo *bean2.UserInfo) ([]*bean2.UserInfo, error)
	UpdateUser(userInfo *bean2.UserInfo) (*bean2.UserInfo, error)
	GetById(id int32) (*bean2.UserInfo, error)
	GetAll() ([]bean2.UserInfo, error)
}

type HelmUserServiceImpl struct {
	logger              *zap.SugaredLogger
	userRepository      HelmUserRepository
	userRoleRepository  HelmUserRoleRepository
	roleGroupRepository HelmRoleGroupRepository
	sessionManager      *middleware.SessionManager
}

func NewHelmUserServiceImpl(logger *zap.SugaredLogger,
	userRepository HelmUserRepository, userRoleRepository HelmUserRoleRepository,
	sessionManager *middleware.SessionManager) *HelmUserServiceImpl {
	serviceImpl := &HelmUserServiceImpl{
		logger:             logger,
		userRepository:     userRepository,
		userRoleRepository: userRoleRepository,
		sessionManager:     sessionManager,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}
func (impl HelmUserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
	token := r.Header.Get("token")
	return impl.getUserByToken(token)
}
func (impl HelmUserServiceImpl) CreateUser(userInfo *bean2.UserInfo) ([]*bean2.UserInfo, error) {
	var pass []string
	var userResponse []*bean2.UserInfo
	emailIds := strings.Split(userInfo.EmailId, ",")
	for _, emailId := range emailIds {
		dbUser, err := impl.userRepository.FetchActiveOrDeletedUserByEmail(emailId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}

		//if found, update it with new roles
		if dbUser != nil && dbUser.Id > 0 {
			userInfo, err = impl.updateUserIfExists(userInfo, dbUser, emailId)
			if err != nil {
				impl.logger.Errorw("error while create user if exists in db", "error", err)
				return nil, err
			}
		}
		// if not found, create new user
		if err == pg.ErrNoRows {
			userInfo, err = impl.createUserIfNotExists(userInfo, emailId)
			if err != nil {
				impl.logger.Errorw("error while create user if not exists in db", "error", err)
				return nil, err
			}
		}
		pass = append(pass, emailId)
		userInfo.EmailId = emailId
		userResponse = append(userResponse, &bean2.UserInfo{Id: userInfo.Id, EmailId: emailId, Groups: userInfo.Groups, RoleFilters: userInfo.RoleFilters, SuperAdmin: userInfo.SuperAdmin})
	}

	return userResponse, nil
}

func (impl HelmUserServiceImpl) UpdateUser(userInfo *bean2.UserInfo) (*bean2.UserInfo, error) {
	model := &HelmUserModel{}
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userInfo.UserId
	model.Active = true
	model, err := impl.userRepository.UpdateUser(model, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	return nil, nil
}

func (impl HelmUserServiceImpl) createUserIfNotExists(userInfo *bean2.UserInfo, emailId string) (*bean2.UserInfo, error) {
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

		isSuperAdmin := false
		roles, err := impl.checkUserRoles(userInfo.UserId)
		if err != nil {
			return nil, err
		}
		for _, item := range roles {
			if item == bean.SUPERADMIN {
				isSuperAdmin = true
			}
		}
		if isSuperAdmin == false {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			return nil, err
		}

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

func (impl HelmUserServiceImpl) validateUserRequest(userInfo *bean2.UserInfo) (bool, error) {
	if len(userInfo.RoleFilters) == 1 &&
		userInfo.RoleFilters[0].Team == "" && userInfo.RoleFilters[0].Environment == "" && userInfo.RoleFilters[0].Action == "" {
		//skip
	} else {
		invalid := false
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.Team) > 0 && len(roleFilter.Action) > 0 {
				//
			} else if len(roleFilter.Entity) > 0 {
				//
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

func (impl HelmUserServiceImpl) updateUserIfExists(userInfo *bean2.UserInfo, dbUser *HelmUserModel, emailId string) (*bean2.UserInfo, error) {
	updateUserInfo, err := impl.GetById(dbUser.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if dbUser.Active == false {
		updateUserInfo = &bean2.UserInfo{Id: dbUser.Id}
	}
	updateUserInfo.RoleFilters = impl.mergeRoleFilter(updateUserInfo.RoleFilters, userInfo.RoleFilters)
	updateUserInfo.Groups = impl.mergeGroups(updateUserInfo.Groups, userInfo.Groups)
	updateUserInfo.UserId = userInfo.UserId
	updateUserInfo.EmailId = emailId // override case sensitivity
	updateUserInfo, err = impl.UpdateUser(updateUserInfo)
	if err != nil {
		impl.logger.Errorw("error while update user", "error", err)
		return nil, err
	}
	return userInfo, nil
}

func (impl HelmUserServiceImpl) checkUserRoles(id int32) ([]string, error) {
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

func (impl HelmUserServiceImpl) mergeRoleFilter(oldR []bean2.RoleFilter, newR []bean2.RoleFilter) []bean2.RoleFilter {
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

func (impl HelmUserServiceImpl) mergeGroups(oldGroups []string, newGroups []string) []string {
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

func (impl HelmUserServiceImpl) getUserByToken(token string) (int32, error) {
	if token == "" {
		impl.logger.Infow("no token provided", "token", token)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return http.StatusUnauthorized, err
	}

	//claims, err := impl.sessionManager.VerifyToken(token)
	claims, err := impl.sessionManager.VerifyToken(token)

	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "failed to verify token",
			UserMessage:     fmt.Sprintf("token verification failed while getting logged in user: %s", token),
		}
		return http.StatusUnauthorized, err
	}
	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		impl.logger.Errorw("failed to MapClaims", "error", err)
		return http.StatusUnauthorized, err
	}

	email := jwt.GetField(mapClaims, "email")
	sub := jwt.GetField(mapClaims, "sub")

	if email == "" && sub == "admin" {
		email = sub
	}

	userInfo, err := impl.getUserByEmail(email)
	if err != nil {
		impl.logger.Errorw("unable to fetch user from db", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNotFoundForToken,
			InternalMessage: "user not found for token",
			UserMessage:     fmt.Sprintf("no user found against provided token: %s", token),
		}
		return http.StatusUnauthorized, err
	}
	return userInfo.Id, nil
}

func (impl HelmUserServiceImpl) getUserByEmail(emailId string) (*bean2.UserInfo, error) {
	model, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roles, err := impl.userRoleRepository.GetRolesByUserId(model.Id)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	var roleFilters []bean2.RoleFilter
	for _, role := range roles {
		roleFilters = append(roleFilters, bean2.RoleFilter{
			Entity:      role.Entity,
			Team:        role.Team,
			Environment: role.Environment,
			EntityName:  role.EntityName,
			Action:      role.Action,
		})
	}

	response := &bean2.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		AccessToken: model.AccessToken,
		RoleFilters: roleFilters,
	}

	return response, nil
}
func (impl HelmUserServiceImpl) GetById(id int32) (*bean2.UserInfo, error) {
	model, err := impl.userRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roles, err := impl.userRoleRepository.GetRolesByUserId(model.Id)
	if err != nil {
		impl.logger.Debugw("No Roles Found for user", "id", model.Id)
	}
	isSuperAdmin := false
	var roleFilters []bean2.RoleFilter
	roleFilterMap := make(map[string]*bean2.RoleFilter)
	for _, role := range roles {
		key := ""
		if len(role.Team) > 0 {
			key = fmt.Sprintf("%s_%s", role.Team, role.Action)
		} else if len(role.Entity) > 0 {
			key = fmt.Sprintf("%s_%s", role.Entity, role.Action)
		}
		if _, ok := roleFilterMap[key]; ok {
			envArr := strings.Split(roleFilterMap[key].Environment, ",")
			if containsArr(envArr, AllEnvironment) {
				roleFilterMap[key].Environment = AllEnvironment
			} else if !containsArr(envArr, role.Environment) {
				roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, role.Environment)
			}
			entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
			if !containsArr(entityArr, role.EntityName) {
				roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, role.EntityName)
			}
		} else {
			roleFilterMap[key] = &bean2.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
			}

		}
		if role.Role == bean.SUPERADMIN {
			isSuperAdmin = true
		}
	}
	for _, v := range roleFilterMap {
		if v.Action == "super-admin" {
			continue
		}
		roleFilters = append(roleFilters, *v)
	}

	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}

	var filterGroups []string
	for _, item := range groups {
		if strings.Contains(item, "group:") {
			filterGroups = append(filterGroups, item)
		}
	}

	filterGroupsModels, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(filterGroups)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	filterGroups = nil
	for _, item := range filterGroupsModels {
		filterGroups = append(filterGroups, item.Name)
	}
	if len(filterGroups) == 0 {
		filterGroups = make([]string, 0)
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean2.RoleFilter, 0)
	}
	response := &bean2.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		RoleFilters: roleFilters,
		Groups:      filterGroups,
		SuperAdmin:  isSuperAdmin,
	}

	return response, nil
}

func (impl HelmUserServiceImpl) GetAll() ([]bean2.UserInfo, error) {
	model, err := impl.userRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var response []bean2.UserInfo
	for _, m := range model {
		response = append(response, bean2.UserInfo{
			Id:          m.Id,
			EmailId:     m.EmailId,
			RoleFilters: make([]bean2.RoleFilter, 0),
			Groups:      make([]string, 0),
		})
	}
	if len(response) == 0 {
		response = make([]bean2.UserInfo, 0)
	}
	return response, nil
}
