package adapter

import (
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
)

func GetCasbinGroupPolicy(emailId string, role string, twcDto *bean2.TimeoutWindowConfigDto) bean.Policy {
	return bean.Policy{
		Type: "g",
		Sub:  bean.Subject(emailId),
		Obj:  bean.Object(role),
	}
}
