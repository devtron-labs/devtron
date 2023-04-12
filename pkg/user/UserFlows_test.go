package user

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"log"
	"testing"
	"time"
)

const it = 100

func BenchmarkCreateDefaultPoliciesForAllTypesV2(b *testing.B) {
	config, err := sql.GetConfig()
	if err != nil {
		log.Fatal("error in sql config parsing")
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatal("error in getting logger")
	}
	dbConnection, err := sql.NewDbConnection(config, logger)
	userAuthRepository := repository2.NewUserAuthRepositoryImpl(dbConnection, logger, nil, nil)
	userRepo := repository2.NewUserRepositoryImpl(dbConnection, logger)
	defaultRbacPolicyRepo := repository2.NewRbacPolicyDataRepositoryImpl(logger, dbConnection)
	defaultRbacRoleRepo := repository2.NewRbacRoleDataRepositoryImpl(logger, dbConnection)
	defaultRbacCacheFactory := repository2.NewRbacDataCacheFactoryImpl(logger, defaultRbacPolicyRepo, defaultRbacRoleRepo)
	userCommonService := NewUserCommonServiceImpl(userAuthRepository, logger, userRepo, nil, nil, defaultRbacCacheFactory)
	teams := make(map[int]string, it)
	apps := make(map[int]string, it)
	envs := make(map[int]string, it)
	for i := 0; i < it; i++ {
		teams[i] = fmt.Sprintf("team-%d-%v", i, time.Now().Nanosecond())
		apps[i] = fmt.Sprintf("app-%d-%v", i, time.Now().Nanosecond())
		envs[i] = fmt.Sprintf("env-%d-%v", i, time.Now().Nanosecond())
	}
	entity := fmt.Sprintf("apps")
	accessType := fmt.Sprintf("devtron-app")
	action := fmt.Sprintf("manager")
	b.Run(fmt.Sprintf("BenchmarkCreateDefaultPoliciesForAllTypesV2"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userCommonService.CreateDefaultPoliciesForAllTypesV2(teams[i], apps[i], envs[i], entity, "", "", "", "", "", action, accessType)
		}
	})
}

func BenchmarkCreateDefaultPoliciesForAllTypes(b *testing.B) {
	config, err := sql.GetConfig()
	if err != nil {
		log.Fatal("error in sql config parsing")
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatal("error in getting logger")
	}
	dbConnection, err := sql.NewDbConnection(config, logger)
	defaultRoleRepo := repository2.NewDefaultAuthRoleRepositoryImpl(dbConnection, logger)
	defaultPolicyRepo := repository2.NewDefaultAuthPolicyRepositoryImpl(dbConnection, logger)
	userAuthRepository := repository2.NewUserAuthRepositoryImpl(dbConnection, logger, defaultPolicyRepo, defaultRoleRepo)
	teams := make(map[int]string, it)
	apps := make(map[int]string, it)
	envs := make(map[int]string, it)
	for i := 0; i < it; i++ {
		teams[i] = fmt.Sprintf("team-%d-%v", i, time.Now().Nanosecond())
		apps[i] = fmt.Sprintf("app-%d-%v", i, time.Now().Nanosecond())
		envs[i] = fmt.Sprintf("env-%d-%v", i, time.Now().Nanosecond())
	}
	entity := fmt.Sprintf("apps")
	accessType := fmt.Sprintf("devtron-app")
	action := fmt.Sprintf("manager")
	b.Run(fmt.Sprintf("BenchmarkCreateDefaultPoliciesForAllTypes"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userAuthRepository.CreateDefaultPoliciesForAllTypes(teams[i], apps[i], envs[i], entity, "", "", "", "", "", action, accessType, 1)
		}
	})

}
