package scanningResultsParser

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/scanningResultsParser"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ScanningResultRestHandler interface {
	ScanResults(w http.ResponseWriter, r *http.Request)
}

type ScanningResultRestHandlerImpl struct {
	logger       *zap.SugaredLogger
	userService  user.UserService
	scanService  scanningResultsParser.Service
	enforcer     casbin.Enforcer
	enforcerUtil rbac.EnforcerUtil
	validator    *validator.Validate
}

func NewScanningResultRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	scanService scanningResultsParser.Service,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
) *ScanningResultRestHandlerImpl {
	return &ScanningResultRestHandlerImpl{
		logger:       logger,
		userService:  userService,
		scanService:  scanService,
		enforcer:     enforcer,
		enforcerUtil: enforcerUtil,
		validator:    validator,
	}
}
func (impl ScanningResultRestHandlerImpl) ScanResults(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	var ciWorkflowId, appId, envId, installedAppId int
	if appIdStr := v.Get("appId"); appIdStr != "" {
		appId, err = strconv.Atoi(appIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	if installedAppIdStr := v.Get("installedAppId"); installedAppIdStr != "" {
		installedAppId, err = strconv.Atoi(installedAppIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	if envIdStr := v.Get("envId"); envIdStr != "" {
		envId, err = strconv.Atoi(envIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	if ciWorkflowIdStr := v.Get("ciWorkflowId"); ciWorkflowIdStr != "" {
		ciWorkflowId, err = strconv.Atoi(ciWorkflowIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	//if ciWorkflowId == 0 && (envId == 0 || appId == 0) {
	//	common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	//	return
	//}
	// RBAC
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = impl.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC
	resp, err := impl.scanService.GetScanResults(appId, envId, ciWorkflowId, installedAppId)
	if err != nil {
		impl.logger.Errorw("service err, scan results", "err", err, "appId %d envId %d", appId, envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)

		return
	}

	//sanitize resp
	resp = impl.sanitizeResponse(resp)

	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}

func (impl ScanningResultRestHandlerImpl) sanitizeResponse(resp scanningResultsParser.Response) scanningResultsParser.Response {
	if resp.CodeScan.License != nil && len(resp.CodeScan.License.Licenses) == 0 {
		resp.CodeScan.License = nil
	}
	if resp.CodeScan.Vulnerability != nil && len(resp.CodeScan.Vulnerability.Vulnerabilities) == 0 {
		resp.CodeScan.Vulnerability = nil
	}

	if resp.CodeScan.ExposedSecrets != nil && len(resp.CodeScan.ExposedSecrets.ExposedSecrets) == 0 {
		resp.CodeScan.ExposedSecrets = nil
	}

	if resp.CodeScan.MisConfigurations != nil && len(resp.CodeScan.MisConfigurations.MisConfigurations) == 0 {
		resp.CodeScan.MisConfigurations = nil
	}

	if resp.ImageScan.License != nil && len(resp.ImageScan.License.List) == 0 {
		resp.ImageScan.License = nil
	}
	if resp.ImageScan.Vulnerability != nil && len(resp.ImageScan.Vulnerability.List) == 0 {
		resp.ImageScan.Vulnerability = nil
	}
	return resp
}
