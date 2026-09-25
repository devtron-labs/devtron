package user

import (
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"testing"
)

func TestCheckRbacForARole(t *testing.T) {
	impl := UserCommonServiceImpl{}

	t.Run("super admin role authorized", func(t *testing.T) {
		role := &repository.RoleModel{
			Action: bean2.SUPER_ADMIN,
		}
		managerAuth := func(resource, token, object string) bool {
			return true
		}
		authorized := impl.checkRbacForARole(role, "test-token", managerAuth)
		if !authorized {
			t.Errorf("expected role to be authorized, got false")
		}
	})

	t.Run("chart group entity authorized unconditionally", func(t *testing.T) {
		role := &repository.RoleModel{
			Entity: bean2.CHART_GROUP_ENTITY,
		}
		managerAuth := func(resource, token, object string) bool {
			return false
		}
		authorized := impl.checkRbacForARole(role, "test-token", managerAuth)
		if !authorized {
			t.Errorf("expected chart group entity to be authorized, got false")
		}
	})

	t.Run("team role authorization check", func(t *testing.T) {
		role := &repository.RoleModel{
			Team: "devtron-team",
		}
		managerAuth := func(resource, token, object string) bool {
			return object == "devtron-team"
		}
		authorized := impl.checkRbacForARole(role, "test-token", managerAuth)
		if !authorized {
			t.Errorf("expected team role to be authorized for matching team")
		}
	})

	t.Run("default entity unauthorized", func(t *testing.T) {
		role := &repository.RoleModel{
			Entity: "unknown-entity",
		}
		managerAuth := func(resource, token, object string) bool {
			return true
		}
		authorized := impl.checkRbacForARole(role, "test-token", managerAuth)
		if authorized {
			t.Errorf("expected unknown entity to be unauthorized, got true")
		}
	})
}
