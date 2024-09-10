package noop

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
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

func (impl NoopUserService) CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource string, token string, object string) bool) ([]*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource string, token string, object string) bool) (*bean.UserInfo, bool, bool, []string, error) {
	return nil, false, false, nil, nil
}

func (impl NoopUserService) GetById(id int32) (*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) GetAll() ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) GetAllDetailedUsers() ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) GetEmailFromToken(token string) (string, error) {
	return "", nil
}

func (impl NoopUserService) GetLoggedInUser(r *http.Request) (int32, error) {
	return 1, nil //TODO return admin userId
}

func (impl NoopUserService) GetByIds(ids []int32) ([]bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) DeleteUser(userInfo *bean.UserInfo) (bool, error) {
	return false, nil
}

func (impl NoopUserService) CheckUserRoles(id int32) ([]string, error) {
	return nil, nil
}

func (impl NoopUserService) SyncOrchestratorToCasbin() (bool, error) {
	return false, nil
}

func (impl NoopUserService) GetUserByToken(token string) (int32, string, error) {
	return 1, "", nil
}

func (impl NoopUserService) IsSuperAdmin(userId int) (bool, error) {
	return true, nil
}

func (impl NoopUserService) GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error) {
	return nil, nil
}

func (impl NoopUserService) UserExists(emailId string) bool {
	return true
}

func (impl NoopUserService) UpdateTriggerPolicyForTerminalAccess() (err error) {
	return nil
}

func (impl NoopUserService) GetRoleFiltersByGroupNames(groupNames []string) ([]bean.RoleFilter, error) {
	return make([]bean.RoleFilter, 0), nil
}

func (impl NoopUserService) SaveLoginAudit(emailId, clientIp string, id int32) {

}

func (impl NoopUserService) FetchRolesFromGroup(userId int32) ([]*repository.RoleModel, error) {
	impl.logger.Warnw("method not impl for FetchRolesFromGroup")
	return make([]*repository.RoleModel, 0), nil
}
