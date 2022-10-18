package user

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	repomock "github.com/devtron-labs/devtron/pkg/user/repository/RepositoryMocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUserUpdateService(t *testing.T) {

	t.Run("UpdateApiCase1", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		userRepositoryMocked := repomock.NewUserRepository(t)
		userAuthRepositoryMocked := repomock.NewUserAuthRepository(t)
		roleGroupRepositoryMocked := repomock.NewRoleGroupRepository(t)

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

		model := repository.UserModel{
			Id:          24,
			EmailId:     "test12@devtron.ai",
			AccessToken: "",
			Active:      true,
			UserType:    "",
		}

		userRoleModels := []*repository.UserRoleModel{}

		userRoleModelOne := repository.UserRoleModel{TableName: struct{}{},
			Id:     738,
			UserId: 24,
			RoleId: 1372,
			User: repository.UserModel{
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

		userRoleModelTwo := repository.UserRoleModel{TableName: struct{}{},
			Id:     739,
			UserId: 24,
			RoleId: 1052,
			User: repository.UserModel{
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

		roleModelOne := repository.RoleModel{
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

		roleModelTwo := repository.RoleModel{
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

		userGroup := repository.RoleGroup{
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

		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZhNmE3Y2Y4ZTkzNDE1Njk3N2U4MmVlYzQ4M2I2OTM3NjRmMTU4NjcifQ.eyJpc3MiOiJodHRwczovL2RlbW8uZGV2dHJvbi5pbmZvOjMyNDQzL29yY2hlc3RyYXRvci9hcGkvZGV4Iiwic3ViIjoiQ2hVeE1ESTFOVGd3TURjNU5qa3pNelk1T0Rrek5USVNCbWR2YjJkc1pRIiwiYXVkIjoiYXJnby1jZCIsImV4cCI6MTY2NjEyMDMzNCwiaWF0IjoxNjY2MDMzOTM0LCJhdF9oYXNoIjoiUGtNdDJmYzlGQ1VXU2NRNEFuY0Y3USIsImNfaGFzaCI6ImN3WVh2bnhWbnVmdkFtWjdWZXV1elEiLCJlbWFpbCI6ImF5dXNoQGRldnRyb24uYWkiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkF5dXNoIE1haGVzaHdhcmkifQ.vqUl-bJJc8Eg5fiIo-dOTIQfS-VZTHvVXSNFA24XJT1ITqgXYNC7RuQT-IF-ZR2c5R_2mmiQ7gjZsAe9aztHUD0pkeHNThoVU-BFiqp5Wv1IkQpoRNzFen-qlKJccTKogOXFV22whvdi0n-EHzr7d0km8IuOihaYP00W6xG1wEk1WpDFMqusKXj3iQCzA_yV6Y6V-GgkngpKtkSeLtTUzLtXkyMhA3VOY4HMGaLDZ2kI29alN8j7TrGukhBSfKRxYz-WOk746tW2e3NLeT0nR9AHHtthr6fwPpmMsF4uYZzp385cjprsk0jiugMBFJsGljIJvrDwqQx3hYX_wFNmFw"
		_, isRolesChanged, isGroupsModified, restrictedGroups, _ := userServiceImpl.UpdateUser(&userInfo, token, nil)

		assert.Equal(t, isRolesChanged, false)
		assert.Equal(t, isGroupsModified, false)
		assert.Equal(t, restrictedGroups, 2)

	})

}
