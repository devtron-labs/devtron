package noop

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean"
	"go.uber.org/zap"
	"net/http"
)

type NoopUserService struct {
	logger *zap.SugaredLogger
}

func NewNoopUserService(logger *zap.SugaredLogger) *NoopUserService {
	return &NoopUserService{
		logger: logger,
	}
}

func (impl NoopUserService) CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource string, token string, object string) bool) ([]*bean.UserInfo, []bean.RestrictedGroup, error) {
	return nil, nil, nil
}

func (impl NoopUserService) SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource string, token string, object string) bool) (*bean.UserInfo, bool, bool, []bean.RestrictedGroup, error) {
	return nil, false, false, nil, nil
}

func (impl NoopUserService) GetById(id int32) (*bean.UserInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (impl NoopUserService) GetAll() ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) GetAllWithFilters(request *bean.ListingRequest) (*bean.UserListingResponse, error) {
	return nil, nil
}

func (impl NoopUserService) GetAllDetailedUsers() ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) GetEmailFromToken(token string) (string, error) {
	return "", nil
}

func (impl NoopUserService) GetEmailAndVersionFromToken(token string) (string, string, error) {
	return "", "", nil
}

func (impl NoopUserService) GetEmailById(userId int32) (string, error) {
	return "", nil
}

func (impl NoopUserService) GetLoggedInUser(r *http.Request) (int32, error) {
	return 0, nil
}

func (impl NoopUserService) GetByIds(ids []int32) ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) DeleteUser(userInfo *bean.UserInfo) (bool, error) {
	return false, nil
}

func (impl NoopUserService) BulkDeleteUsers(request *bean.BulkDeleteRequest) (bool, error) {
	return false, nil
}

func (impl NoopUserService) CheckUserRoles(id int32) ([]string, error) {
	return nil, nil
}

func (impl NoopUserService) SyncOrchestratorToCasbin() (bool, error) {
	return false, nil
}

func (impl NoopUserService) GetUserByToken(context context.Context, token string) (int32, string, error) {
	return 0, "", nil
}

func (impl NoopUserService) GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) UserExists(emailId string) bool {
	return true
}

func (impl NoopUserService) UpdateTriggerPolicyForTerminalAccess() (err error) {
	return err
}

func (impl NoopUserService) GetRoleFiltersByUserRoleGroups(userRoleGroups []bean.UserRoleGroup) ([]bean.RoleFilter, error) {
	return nil, nil
}

func (impl NoopUserService) SaveLoginAudit(emailId, clientIp string, id int32) {

}

func (impl NoopUserService) CheckIfTokenIsValid(email string, version string) error {
	return nil
}
