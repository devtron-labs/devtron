package rbac

import (
	"github.com/argoproj/argo-cd/util/session"
	"github.com/casbin/casbin"
	jsonadapter "github.com/casbin/json-adapter"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestEnforcerImpl_enforceByEmail(t *testing.T) {
	type fields struct {
		Enforcer       *casbin.Enforcer
		SessionManager *session.SessionManager
		logger         *zap.SugaredLogger
	}
	type args struct {
		vals []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "basic test",
			fields: fields{
				Enforcer: newEnforcer(getMangerPolicies()),
				logger:   &zap.SugaredLogger{},
			},
			args: args{vals: toInterface([]string{"abc@abc.com", ResourceUser, ActionCreate, "*"})},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EnforcerImpl{
				Enforcer:       tt.fields.Enforcer,
				SessionManager: tt.fields.SessionManager,
				logger:         tt.fields.logger,
			}
			if got := e.enforceByEmail(tt.fields.Enforcer, tt.args.vals...); got != tt.want {
				t.Errorf("enforceByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newEnforcer(policies [][]string) *casbin.Enforcer {
	var b []byte
	a := jsonadapter.NewAdapter(&b)
	enforcer := casbin.NewEnforcer("../../auth_model.conf", a, true)
	addPolicies(enforcer, policies)
	return enforcer
}

func addPolicies(enforcer *casbin.Enforcer, policies [][]string) {
	enforcer.LoadPolicy()
	for _, policy := range policies {
		if strings.EqualFold(policy[0], "p") {
			enforcer.AddPolicy(policy[1], policy[2], policy[3], policy[4], "allow")
		} else if strings.EqualFold(policy[0], "g") {
			enforcer.AddGroupingPolicy(policy[1], policy[2])
		}
	}
	enforcer.SavePolicy()
	enforcer.LoadPolicy()
}

func getMangerPolicies() [][]string {
	return [][]string{
		[]string{"p", "role:manager_dev_devtron-demo_", "applications", "*", "dev/*"},
		[]string{"p", "role:manager_dev_devtron-demo_", "environment", "*", "devtron-demo/*"},
		[]string{"p", "role:manager_dev_devtron-demo_", "team", "*", "dev"},
		[]string{"p", "role:manager_dev_devtron-demo_", "user", "*", "dev"},
		[]string{"p", "role:manager_dev_devtron-demo_", "notification", "*", "dev"},
		[]string{"p", "role:manager_dev_devtron-demo_", "global-environment", "*", "devtron-demo"},
		[]string{"g", "abc@abc.com", "role:manager_dev_devtron-demo_"},
	}
}

func toInterface(input []string) []interface{} {
	out := make([]interface{}, len(input))
	for index, in := range input {
		out[index] = in
	}
	return out
}
