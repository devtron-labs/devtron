package adapter

import (
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"time"
)

func GetLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func GetCasbinGroupPolicy(emailId string, role string, expression string, expressionFormat string) casbin.Policy {
	return casbin.Policy{
		Type: "g",
		Sub:  casbin.Subject(emailId),
		Res:  casbin.Resource(expression),
		Act:  casbin.Action(expressionFormat),
		Obj:  casbin.Object(role),
	}
}
