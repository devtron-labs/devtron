package user

import (
	"fmt"
	"testing"

	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/stretchr/testify/require"
)

func TestCheckRbacForARole(t *testing.T) {
	impl := UserCommonServiceImpl{}
	type authRequest struct {
		resource string
		token    string
		object   string
	}
	for _, tc := range []struct {
		name     string
		role     repository.RoleModel
		resource string
		object   string
	}{
		{name: "super admin", role: repository.RoleModel{Action: bean2.SUPER_ADMIN, Team: "team-1"}, resource: casbin.ResourceGlobal, object: "*"},
		{name: "helm", role: repository.RoleModel{AccessType: bean2.APP_ACCESS_TYPE_HELM, Team: "team-1"}, resource: casbin.ResourceGlobal, object: "*"},
		{name: "argo", role: repository.RoleModel{AccessType: bean2.APP_ACCESS_TYPE_ARGO, Team: "team-1"}, resource: casbin.ResourceGlobal, object: "*"},
		{name: "flux", role: repository.RoleModel{AccessType: bean2.APP_ACCESS_TYPE_FLUX, Team: "team-1"}, resource: casbin.ResourceGlobal, object: "*"},
		{name: "jobs", role: repository.RoleModel{Entity: bean2.EntityJobs, Team: "team-1"}, resource: casbin.ResourceGlobal, object: "*"},
		{name: "team", role: repository.RoleModel{Team: "team-1"}, resource: casbin.ResourceUser, object: "team-1"},
		{name: "team takes precedence over chart group", role: repository.RoleModel{Entity: bean2.CHART_GROUP_ENTITY, Team: "team-1"}, resource: casbin.ResourceUser, object: "team-1"},
		{
			name:     "cluster",
			role:     repository.RoleModel{Entity: bean2.CLUSTER_ENTITIY, Cluster: "CLUSTER-1", Namespace: "NS-1", Group: "apps", Kind: "deployments", Resource: "frontend"},
			resource: "cluster-1/ns-1/" + casbin.ResourceUser,
			object:   "apps/deployments/frontend",
		},
	} {
		for _, allowed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/allowed=%t", tc.name, allowed), func(t *testing.T) {
				var requests []authRequest
				managerAuth := func(resource, token, object string) bool {
					requests = append(requests, authRequest{resource, token, object})
					return allowed
				}

				result := impl.checkRbacForARole(&tc.role, "test-token", managerAuth)

				require.Equal(t, allowed, result)
				require.Equal(t, []authRequest{{tc.resource, "test-token", tc.object}}, requests)
			})
		}
	}
	for _, tc := range []struct {
		name string
		role repository.RoleModel
		want bool
	}{
		{name: "chart group", role: repository.RoleModel{Entity: bean2.CHART_GROUP_ENTITY}, want: true},
		{name: "unknown entity", role: repository.RoleModel{Entity: "unknown-entity"}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			result := impl.checkRbacForARole(&tc.role, "test-token", func(_, _, _ string) bool {
				calls++
				return !tc.want
			})

			require.Equal(t, tc.want, result)
			require.Zero(t, calls)
		})
	}
}

func TestMergeGroupsPreservesLiteralNames(t *testing.T) {
	impl := UserServiceImpl{}

	groups := impl.mergeGroups([]string{"team%v"}, []string{"team%0v", "team%v"})

	require.Equal(t, []string{"team%v", "team%0v"}, groups)
}
