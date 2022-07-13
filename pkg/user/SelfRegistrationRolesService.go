package user

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"go.uber.org/zap"
)

type SelfRegistrationRolesService interface {
	Check() (CheckResponse, error)
	SelfRegister(emailId string) (*bean.UserInfo, error)
	CheckAndCreateUserIfConfigured(emailId string) bool
}

type SelfRegistrationConfig struct {
	SelfRegistrationEnabled bool `env:"SELF_REGISTRATION_ENABLED" envDefault:"false"`
}

func GetSelfRegistrationConfig() (*SelfRegistrationConfig, error) {
	cfg := &SelfRegistrationConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type SelfRegistrationRolesServiceImpl struct {
	logger                          *zap.SugaredLogger
	selfRegistrationRolesRepository repository.SelfRegistrationRolesRepository
	userService                     UserService
	selfRegistrationConfig          *SelfRegistrationConfig
}

func NewSelfRegistrationRolesServiceImpl(logger *zap.SugaredLogger,
	selfRegistrationRolesRepository repository.SelfRegistrationRolesRepository, userService UserService, selfRegistrationConfig *SelfRegistrationConfig) *SelfRegistrationRolesServiceImpl {
	return &SelfRegistrationRolesServiceImpl{
		logger:                          logger,
		selfRegistrationRolesRepository: selfRegistrationRolesRepository,
		userService:                     userService,
		selfRegistrationConfig:          selfRegistrationConfig,
	}
}

func (impl *SelfRegistrationRolesServiceImpl) GetAll() ([]string, error) {
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

func (impl *SelfRegistrationRolesServiceImpl) Check() (CheckResponse, error) {
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
func (impl *SelfRegistrationRolesServiceImpl) SelfRegister(emailId string) (*bean.UserInfo, error) {

	roles, err := impl.Check()
	if err != nil || roles.Enabled == false {
		return nil, err
	}

	userInfo := &bean.UserInfo{
		EmailId:    emailId,
		Roles:      roles.Roles,
		SuperAdmin: false,
	}
	userInfos, err := impl.userService.SelfRegisterUserIfNotExists(userInfo)
	if err != nil {
		impl.logger.Errorw("error while register user", "error", err)
	}
	if len(userInfos) > 0 {
		return userInfos[0], nil
	} else {
		return nil, fmt.Errorf("user not created")
	}
}

func (impl *SelfRegistrationRolesServiceImpl) CheckAndCreateUserIfConfigured(emailId string) bool {
	exists := impl.userService.UserExists(emailId)
	if !exists {
		if impl.selfRegistrationConfig.SelfRegistrationEnabled {
			impl.logger.Infow("self registering user,  ", "email", emailId)
			user, err := impl.SelfRegister(emailId)
			if err != nil {
				impl.logger.Errorw("error while register user", "error", err)
			} else if user != nil && user.UserId > 0 {
				exists = true
			}
		}
	}
	impl.logger.Infow("user status", "email", emailId, "status", exists)
	return exists
}
