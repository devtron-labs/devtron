package user

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
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
}

func NewSelfRegistrationRolesHandlerImpl(logger *zap.SugaredLogger,
	selfRegistrationRolesService user.SelfRegistrationRolesService) *SelfRegistrationRolesHandlerImpl {
	return &SelfRegistrationRolesHandlerImpl{
		logger:                       logger,
		selfRegistrationRolesService: selfRegistrationRolesService,
	}
}

func (impl *SelfRegistrationRolesHandlerImpl) SelfRegister(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var userInfo bean.UserInfo
	err := decoder.Decode(&userInfo)
	if err != nil {
		impl.logger.Errorw("request err, selfRegister", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.selfRegistrationRolesService.SelfRegister(userInfo.EmailId)
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
