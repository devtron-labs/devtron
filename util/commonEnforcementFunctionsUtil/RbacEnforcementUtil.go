package commonEnforcementFunctionsUtil

import (
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"strings"
)

type CommonEnforcementUtil interface {
	CheckAuthorizationForGlobalEnvironment(token string, object string) bool
	CheckAuthorizationByEmailInBatchForGlobalEnvironment(token string, object []string) map[string]bool
	CheckAuthorisationForEnvAutocomplete(token string, environments []bean2.EnvironmentBean) []bean2.EnvironmentBean
	CheckAuthorisationOnApp(token string, projectWiseApps []*app.TeamAppBean) []*app.TeamAppBean
	CheckRbacForMangerAndAboveAccess(token string, userId int32) (bool, error)
}
type CommonEnforcementUtilImpl struct {
	enforcer          casbin.Enforcer
	enforcerUtil      rbac.EnforcerUtil
	logger            *zap.SugaredLogger
	userService       user.UserService
	userCommonService user.UserCommonService
}

func NewCommonEnforcementUtilImpl(enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	logger *zap.SugaredLogger,
	userService user.UserService,
	userCommonService user.UserCommonService) *CommonEnforcementUtilImpl {
	return &CommonEnforcementUtilImpl{
		enforcer:          enforcer,
		enforcerUtil:      enforcerUtil,
		logger:            logger,
		userService:       userService,
		userCommonService: userCommonService,
	}
}

func (impl *CommonEnforcementUtilImpl) CheckAuthorizationForGlobalEnvironment(token string, object string) bool {
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, object); !ok {
		return false
	}
	return true
}

func (impl *CommonEnforcementUtilImpl) CheckAuthorizationByEmailInBatchForGlobalEnvironment(token string, object []string) map[string]bool {
	var objectResult map[string]bool
	if len(object) > 0 {
		objectResult = impl.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, object)
	}
	return objectResult
}

func (impl *CommonEnforcementUtilImpl) CheckAuthorisationForEnvAutocomplete(token string, environments []bean2.EnvironmentBean) []bean2.EnvironmentBean {
	grantedEnvironment := make([]bean2.EnvironmentBean, 0, len(environments))
	// RBAC enforcer applying
	var envIdentifierList []string
	for _, item := range environments {
		envIdentifierList = append(envIdentifierList, strings.ToLower(item.EnvironmentIdentifier))
	}

	result := impl.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
	for _, item := range environments {

		var hasAccess bool
		EnvironmentIdentifier := item.ClusterName + "__" + item.Namespace
		if item.EnvironmentIdentifier != EnvironmentIdentifier {
			// fix for futuristic case
			hasAccess = result[strings.ToLower(EnvironmentIdentifier)] || result[strings.ToLower(item.EnvironmentIdentifier)]
		} else {
			hasAccess = result[strings.ToLower(item.EnvironmentIdentifier)]
		}
		if hasAccess {
			grantedEnvironment = append(grantedEnvironment, item)
		}
	}
	return grantedEnvironment
}
func (impl *CommonEnforcementUtilImpl) CheckAuthorisationOnApp(token string, projectWiseApps []*app.TeamAppBean) []*app.TeamAppBean {
	isActionUserSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	for _, project := range projectWiseApps {
		var accessedApps []*app.AppBean
		for _, app := range project.AppList {
			if isActionUserSuperAdmin {
				accessedApps = append(accessedApps, app)
				continue
			}
			object := impl.enforcerUtil.GetAppRBACNameByAppAndProjectName(project.ProjectName, app.Name)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); ok {
				accessedApps = append(accessedApps, app)
			}
		}
		if len(accessedApps) == 0 {
			accessedApps = make([]*app.AppBean, 0)
		}
		project.AppList = accessedApps
	}
	return projectWiseApps
}
