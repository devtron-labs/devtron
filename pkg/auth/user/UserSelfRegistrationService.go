package user

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type UserSelfRegistrationService interface {
	CheckSelfRegistrationRoles() (CheckResponse, error)
	SelfRegister(emailId string, groups []string) (*bean.UserInfo, error)
	CheckAndCreateUserIfConfigured(claims jwt.MapClaims) bool
}

type UserSelfRegistrationServiceImpl struct {
	logger                           *zap.SugaredLogger
	selfRegistrationRolesRepository  repository.SelfRegistrationRolesRepository
	userService                      UserService
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService
}

func NewUserSelfRegistrationServiceImpl(logger *zap.SugaredLogger,
	selfRegistrationRolesRepository repository.SelfRegistrationRolesRepository, userService UserService,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService) *UserSelfRegistrationServiceImpl {
	return &UserSelfRegistrationServiceImpl{
		logger:                           logger,
		selfRegistrationRolesRepository:  selfRegistrationRolesRepository,
		userService:                      userService,
		globalAuthorisationConfigService: globalAuthorisationConfigService,
	}
}

func (impl *UserSelfRegistrationServiceImpl) GetAllSelfRegistrationRoles() ([]string, error) {
	roleEntries, err := impl.selfRegistrationRolesRepository.GetAll()
	if err != nil {
		impl.logger.Errorf("error fetching all role %+v", err)
		return nil, err
	}
	var roles []string
	for _, role := range roleEntries {
		if role.Role != "" {
			roles = append(roles, role.Role)
		}
	}
	return roles, nil
}

type CheckResponse struct {
	Enabled bool     `json:"enabled"`
	Roles   []string `json:"roles"`
}

func (impl *UserSelfRegistrationServiceImpl) CheckSelfRegistrationRoles() (CheckResponse, error) {
	roleEntries, err := impl.selfRegistrationRolesRepository.GetAll()
	var checkResponse CheckResponse
	if err != nil {
		impl.logger.Errorf("error fetching all role %+v", err)
		checkResponse.Enabled = false
		return checkResponse, err
	}
	//var roles []string
	if roleEntries != nil {
		for _, role := range roleEntries {
			if role.Role != "" {
				checkResponse.Roles = append(checkResponse.Roles, role.Role)
				checkResponse.Enabled = true
				//return checkResponse, err
			}
		}
		if checkResponse.Enabled == true {
			return checkResponse, err
		}
		checkResponse.Enabled = false
		return checkResponse, err
	}
	checkResponse.Enabled = false
	return checkResponse, nil
}

func (impl *UserSelfRegistrationServiceImpl) SelfRegister(emailId string, groups []string) (*bean.UserInfo, error) {
	toSelfRegisterUser := false
	var selfRegistrationRoles []string
	groupClaimsConfigActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	if groupClaimsConfigActive {
		toSelfRegisterUser = true
		selfRegistrationRoles = nil //just for easy readability
	} else {
		roles, err := impl.CheckSelfRegistrationRoles()
		if err != nil {
			return nil, err
		}
		if roles.Enabled {
			toSelfRegisterUser = true
			selfRegistrationRoles = roles.Roles
		}
	}
	if toSelfRegisterUser {
		impl.logger.Infow("self register start")
		userInfo := &bean.UserInfo{
			EmailId:    emailId,
			Roles:      selfRegistrationRoles,
			SuperAdmin: false,
		}
		userInfos, err := impl.userService.SelfRegisterUserIfNotExists(userInfo, groups, groupClaimsConfigActive)
		if err != nil {
			impl.logger.Errorw("error while register user", "error", err)
			return nil, err
		}
		impl.logger.Infow("self registered user", "user", userInfos)
		if len(userInfos) > 0 {
			return userInfos[0], nil
		} else {
			return nil, fmt.Errorf("user not created")
		}
	} else {
		return nil, nil
	}
}

func (impl *UserSelfRegistrationServiceImpl) CheckAndCreateUserIfConfigured(claims jwt.MapClaims) bool {
	emailId, groups := impl.globalAuthorisationConfigService.GetEmailAndGroupsFromClaims(claims)
	impl.logger.Info("check and create user if configured")
	exists := impl.userService.UserExists(emailId)
	isInactive, _, err := impl.userService.UserStatusCheckInDb(emailId)
	if err != nil {
		impl.logger.Errorw("skip this error and check for self registration", "error", err)
	}
	var id int32
	if !exists {
		impl.logger.Infow("self registering user,  ", "email", emailId)
		user, err := impl.SelfRegister(emailId, groups)
		if err != nil {
			impl.logger.Errorw("error while register user", "error", err)
		} else if user != nil && user.Id > 0 {
			id = user.Id
			exists = true
		}
	}
	if exists {
		groupClaimsConfigActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
		//user is active, need to update group claim data if needed
		if groupClaimsConfigActive {
			err := impl.userService.UpdateUserGroupMappingIfActiveUser(emailId, groups)
			if err != nil {
				impl.logger.Errorw("error in updating data for user group claims map", "err", err, "emailId", emailId)
				return exists
			}
		}
		impl.logger.Info("check and create user if configured - save audit", "isInactive", isInactive)
		if !isInactive {
			impl.userService.SaveLoginAudit(emailId, "localhost", id)
		}
	}
	impl.logger.Infow("user status", "email", emailId, "status", exists)
	return exists
}
