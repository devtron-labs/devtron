package user

import "github.com/devtron-labs/devtron/api/bean"

func HidePermissions(roleGroup *bean.RoleGroup) {
	// setting empty role filters to hide permissions
	roleGroup.RoleFilters = make([]bean.RoleFilter, 0)
}
