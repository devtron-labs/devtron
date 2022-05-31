package user

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"net/http"
)

type SelfRegistrationRolesHandler interface {
	SelfRegister(w http.ResponseWriter, r *http.Request)
	SelfRegisterCheck(w http.ResponseWriter, r *http.Request)
}

type SelfRegistrationRolesHandlerImpl struct {
	logger                       *zap.SugaredLogger
	selfRegistrationRolesService user.SelfRegistrationRolesService
	userService                  user.UserService
}

func NewSelfRegistrationRolesHandlerImpl(logger *zap.SugaredLogger,
	selfRegistrationRolesService user.SelfRegistrationRolesService,
	userService user.UserService) *SelfRegistrationRolesHandlerImpl {
	return &SelfRegistrationRolesHandlerImpl{
		logger:                       logger,
		selfRegistrationRolesService: selfRegistrationRolesService,
		userService:                  userService,
	}
}

func (impl *SelfRegistrationRolesHandlerImpl) SelfRegister(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	email, err := impl.userService.GetEmailFromToken(token)

	if err != nil {
		impl.logger.Errorw("request err, selfRegister", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.selfRegistrationRolesService.SelfRegister(email)
	common.WriteJsonResp(w, nil, map[string]string{"status": "ok"}, http.StatusOK)
}

func (impl *SelfRegistrationRolesHandlerImpl) SelfRegisterCheck(w http.ResponseWriter, r *http.Request) {
	res, err := impl.selfRegistrationRolesService.Check()
	if err != nil {
		impl.logger.Errorw("service err, Check", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, map[string]bool{"enabled": res.Enabled}, http.StatusOK)
}
