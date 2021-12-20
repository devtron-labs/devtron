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
	roleGroupRepository HelmUserRoleGroupRepository
	sessionManager      *middleware.SessionManager
	helmUserUtil        HelmUserUtil
}

func NewHelmUserServiceImpl(logger *zap.SugaredLogger,
	userRepository HelmUserRepository, userRoleRepository HelmUserRoleRepository,
	sessionManager *middleware.SessionManager, helmUserUtil HelmUserUtil) *HelmUserServiceImpl {
	serviceImpl := &HelmUserServiceImpl{
		logger:             logger,
		userRepository:     userRepository,
		userRoleRepository: userRoleRepository,
		sessionManager:     sessionManager,
		helmUserUtil:       helmUserUtil,
	}
	return serviceImpl
}
func (impl HelmUserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
	token := r.Header.Get("token")
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
	model, err := impl.userRepository.FetchActiveUserByEmail(email)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNotFoundForToken,
			InternalMessage: "user not found for token",
			UserMessage:     fmt.Sprintf("no user found against provided token: %s", token),
		}
		return http.StatusUnauthorized, err
	}
	return model.Id, nil
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
			userInfo, err = impl.helmUserUtil.CreateUserIfNotExists(userInfo, emailId)
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

func (impl HelmUserServiceImpl) updateUserIfExists(userInfo *bean2.UserInfo, dbUser *HelmUserModel, emailId string) (*bean2.UserInfo, error) {
	updateUserInfo, err := impl.GetById(dbUser.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if dbUser.Active == false {
		updateUserInfo = &bean2.UserInfo{Id: dbUser.Id}
	}
	updateUserInfo.RoleFilters = impl.helmUserUtil.MergeRoleFilter(updateUserInfo.RoleFilters, userInfo.RoleFilters)
	updateUserInfo.Groups = impl.helmUserUtil.MergeGroups(updateUserInfo.Groups, userInfo.Groups)
	updateUserInfo.UserId = userInfo.UserId
	updateUserInfo.EmailId = emailId // override case sensitivity
	updateUserInfo, err = impl.UpdateUser(updateUserInfo)
	if err != nil {
		impl.logger.Errorw("error while update user", "error", err)
		return nil, err
	}
	return userInfo, nil
}
