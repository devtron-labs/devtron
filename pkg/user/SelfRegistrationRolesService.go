package user

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"go.uber.org/zap"
)

type SelfRegistrationRolesService interface {
	Check() (CheckResponse, error)
	SelfRegister(emailId string)
}

type SelfRegistrationRolesServiceImpl struct {
	logger                          *zap.SugaredLogger
	selfRegistrationRolesRepository repository.SelfRegistrationRolesRepository
	userService                     UserService
}

func NewSelfRegistrationRolesServiceImpl(logger *zap.SugaredLogger,
	selfRegistrationRolesRepository repository.SelfRegistrationRolesRepository, userService UserService) *SelfRegistrationRolesServiceImpl {
	return &SelfRegistrationRolesServiceImpl{
		logger:                          logger,
		selfRegistrationRolesRepository: selfRegistrationRolesRepository,
		userService:                     userService,
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
func (impl *SelfRegistrationRolesServiceImpl) SelfRegister(emailId string) {

	roles, err := impl.Check()
	if err != nil || roles.Enabled == false {
		return
	}

	userInfo := &bean.UserInfo{
		EmailId:    emailId,
		Roles:      roles.Roles,
		SuperAdmin: false,
	}
	_, err = impl.userService.SelfRegisterUserIfNotExists(userInfo)
	if err != nil {
		impl.logger.Errorw("error while register user", "error", err)
	}
}
