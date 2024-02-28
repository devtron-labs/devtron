package globalPolicy

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type Config struct {
	Addr     string `env:"TEST_PG_ADDR" envDefault:"127.0.0.1"`
	Port     string `env:"TEST_PG_PORT" envDefault:"5432"`
	User     string `env:"TEST_PG_USER" envDefault:"postgres"`
	Password string `env:"TEST_PG_PASSWORD" envDefault:"postgrespw" secretData:"-"`
	Database string `env:"TEST_PG_DATABASE" envDefault:"orchestrator"`
	LogQuery bool   `env:"TEST_PG_LOG_QUERY" envDefault:"true"`
}

func setupTestDatabase() (*pg.DB, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	}
	db := pg.Connect(&options)
	return db, nil
}

func TestCreatePolicyIntegration(t *testing.T) {

	db, err := setupTestDatabase()
	defer db.Close()
	if err != nil {

	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}

	globalPolicyRepo := repository.NewGlobalPolicyRepositoryImpl(logger, db)
	globalPolicySearchableFieldRepo := repository.NewGlobalPolicySearchableFieldRepositoryImpl(logger, db)

	globalPolicyManager := NewGlobalPolicyDataManagerImpl(logger, globalPolicyRepo, globalPolicySearchableFieldRepo)

	testPolicy := &bean.GlobalPolicyDataModel{
		GlobalPolicyBaseModel: bean.GlobalPolicyBaseModel{
			Name:          "Random Name",
			Enabled:       true,
			PolicyOf:      bean.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW,
			PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
			JsonData:      "{\n  \"deploymentWindowProfile\": {\n    \"Id\": 0,\n    \"Name\": \"string\",\n    \"Description\": \"string\",\n    \"DisplayMessage\": \"string\",\n    \"UserOverrideList\": [\n      0\n    ],\n    \"UsersOverride\": true,\n    \"SuperAdminOverride\": true,\n    \"Type\": \"Blackout\",\n    \"DeploymentWindowList\": [\n      {\n        \"Id\": 0,\n        \"TimeFrom\": \"2024-02-26T11:37:25.200Z\",\n        \"HourMinuteFrom\": \"string\",\n        \"HourMinuteTo\": \"string\",\n        \"DayFrom\": 0,\n        \"DayTo\": 0,\n        \"TimeTo\": \"2024-02-26T11:37:25.200Z\",\n        \"WeekdayFrom\": \"Sunday\"\n      }\n    ],\n    \"TimeZone\": \"string\",\n    \"Enabled\": true\n  }\n}",
			Active:        true,
			UserId:        1,
		},
		SearchableFields: []bean.SearchableField{

			{
				FieldName: "Name", FieldValue: "Random Name", FieldType: bean.StringType,
			},
		},
	}

	createdPolicy, err := globalPolicyManager.CreatePolicy(testPolicy, nil)

	assert.NoError(t, err, "Unexpected error creating policy")
	assert.NotNil(t, createdPolicy, "Expected created policy to be not nil")

}
