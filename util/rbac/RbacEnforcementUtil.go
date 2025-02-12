package rbac

import "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"

type RbacEnforcementUtil interface {
	CheckAuthorizationForGlobalEnvironment(token string, object string) bool
	CheckAuthorizationByEmailInBatchForGlobalEnvironment(token string, object []string) map[string]bool
}
type RbacEnforcementUtilImpl struct {
	enforcer casbin.Enforcer
}

func NewRbacEnforcementUtilImpl(enforcer casbin.Enforcer) *RbacEnforcementUtilImpl {
	return &RbacEnforcementUtilImpl{
		enforcer: enforcer,
	}
}

func (impl *RbacEnforcementUtilImpl) CheckAuthorizationForGlobalEnvironment(token string, object string) bool {
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, object); !ok {
		return false
	}
	return true
}

func (impl *RbacEnforcementUtilImpl) CheckAuthorizationByEmailInBatchForGlobalEnvironment(token string, object []string) map[string]bool {
	var objectResult map[string]bool
	if len(object) > 0 {
		objectResult = impl.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, object)
	}
	return objectResult
}
