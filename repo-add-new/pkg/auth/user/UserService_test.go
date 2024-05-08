package user

import (
	"testing"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	repomock2 "github.com/devtron-labs/devtron/pkg/auth/user/repository/RepositoryMocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
)

func TestUserUpdateService(t *testing.T) {

	t.Run("UpdateApiCase1", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		userRepositoryMocked := repomock2.NewUserRepository(t)
		userAuthRepositoryMocked := repomock2.NewUserAuthRepository(t)
		roleGroupRepositoryMocked := repomock2.NewRoleGroupRepository(t)

		//enforcer := casbinmock.NewEnforcer(t)
		//
		//userRestHandler := user2.NewUserRestHandlerImpl(
		//	nil,
		//	nil,
		//	nil,
		//	enforcer,
		//	nil)

		groups := []string{"livspace"}

		roleFilters := []bean.RoleFilter{}

		roleFilters = append(roleFilters, bean.RoleFilter{
			Entity:      "",
			Team:        "devtron-demo",
			EntityName:  "ajayclone,ajayclone2",
			Environment: "default_cluster__bulk,devtron-demo,default_cluster__test1,default_cluster__test2,demo1__demo1-env,default_cluster__5",
			Action:      "admin",
			AccessType:  "",
		})

		roleFilters = append(roleFilters, bean.RoleFilter{
			Entity:      "chart-group",
			Team:        "",
			EntityName:  "",
			Environment: "",
			Action:      "view",
			AccessType:  "",
		})

		userInfo := bean.UserInfo{
			Id:           24,
			EmailId:      "pawan@devtron.ai",
			Roles:        nil,
			AccessToken:  "",
			UserType:     "",
			LastUsedAt:   time.Now(),
			LastUsedByIp: "",
			Exist:        false,
			UserId:       18,
			RoleFilters:  roleFilters,
			Status:       "",
			Groups:       groups,
			SuperAdmin:   false,
		}

		model := repository2.UserModel{
			Id:          24,
			EmailId:     "test12@devtron.ai",
			AccessToken: "",
			Active:      true,
			UserType:    "",
		}

		userRoleModels := []*repository2.UserRoleModel{}

		userRoleModelOne := repository2.UserRoleModel{TableName: struct{}{},
			Id:     738,
			UserId: 24,
			RoleId: 1372,
			User: repository2.UserModel{
				TableName:   struct{}{},
				Id:          0,
				EmailId:     "",
				AccessToken: "",
				Active:      false,
				UserType:    "",
				AuditLog: sql.AuditLog{
					CreatedOn: time.Time{},
					CreatedBy: 0,
					UpdatedOn: time.Time{},
					UpdatedBy: 0,
				},
			},
			AuditLog: sql.AuditLog{
				CreatedOn: time.Time{},
				CreatedBy: 0,
				UpdatedOn: time.Time{},
				UpdatedBy: 0,
			},
		}

		userRoleModelTwo := repository2.UserRoleModel{TableName: struct{}{},
			Id:     739,
			UserId: 24,
			RoleId: 1052,
			User: repository2.UserModel{
				TableName:   struct{}{},
				Id:          0,
				EmailId:     "",
				AccessToken: "",
				Active:      false,
				UserType:    "",
				AuditLog: sql.AuditLog{
					CreatedOn: time.Time{},
					CreatedBy: 0,
					UpdatedOn: time.Time{},
					UpdatedBy: 0,
				},
			},
			AuditLog: sql.AuditLog{
				CreatedOn: time.Time{},
				CreatedBy: 0,
				UpdatedOn: time.Time{},
				UpdatedBy: 0,
			},
		}

		userRoleModels = append(userRoleModels, &userRoleModelOne)
		userRoleModels = append(userRoleModels, &userRoleModelTwo)

		roleModelOne := repository2.RoleModel{
			TableName:   struct{}{},
			Id:          1372,
			Role:        "role:admin_devtron-demo_default_cluster__bulk_ajayclone",
			Entity:      "",
			Team:        "devtron-demo",
			EntityName:  "ajayclone",
			Environment: "default_cluster__bulk",
			Action:      "admin",
			AccessType:  "",
			AuditLog:    sql.AuditLog{},
		}

		roleModelTwo := repository2.RoleModel{
			TableName:   struct{}{},
			Id:          1052,
			Role:        "role:admin_devtron-demo_default_cluster__bulk_ajayclone2",
			Entity:      "",
			Team:        "devtron-demo",
			EntityName:  "ajayclone2",
			Environment: "default_cluster__bulk",
			Action:      "admin",
			AccessType:  "",
			AuditLog:    sql.AuditLog{},
		}

		userGroup := repository2.RoleGroup{
			TableName:   struct{}{},
			Id:          1,
			Name:        "test",
			CasbinName:  "group:test",
			Description: "sample description 2",
			Active:      true,
			AuditLog:    sql.AuditLog{},
		}

		userRepositoryMocked.On("GetByIdIncludeDeleted", 24).Return(model)
		userAuthRepositoryMocked.On("GetUserRoleMappingByUserId", 24).Return(userRoleModels)
		userAuthRepositoryMocked.On("GetRoleByFilter", "", "devtron-demo", "ajayclone", "default_cluster__bulk", "admin", "").Return(roleModelOne)
		userAuthRepositoryMocked.On("GetRoleByFilter", "", "devtron-demo", "ajayclone2", "default_cluster__bulk", "admin", "").Return(roleModelTwo)
		roleGroupRepositoryMocked.On("GetRoleGroupByName", "test").Return(userGroup)

		dbConnection := userRepositoryMocked.GetConnection()
		tx, err := dbConnection.Begin()

		userRepositoryMocked.On("UpdateUser", model, tx).Return(model, nil)

		userServiceImpl := NewUserServiceImpl(userAuthRepositoryMocked,
			sugaredLogger,
			userRepositoryMocked,
			roleGroupRepositoryMocked,
			nil,
			nil,
			nil)

		token := ""
		_, isRolesChanged, isGroupsModified, restrictedGroups, _ := userServiceImpl.UpdateUser(&userInfo, token, nil)

		assert.Equal(t, isRolesChanged, false)
		assert.Equal(t, isGroupsModified, false)
		assert.Equal(t, restrictedGroups, 2)

	})

}
