package appStatus

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type Config struct {
	Addr            string `env:"TEST_PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"TEST_PG_PORT" envDefault:"55000"`
	User            string `env:"TEST_PG_USER" envDefault:"postgres"`
	Password        string `env:"TEST_PG_PASSWORD" envDefault:"postgrespw" secretData:"-"`
	Database        string `env:"TEST_PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"TEST_APP" envDefault:"orchestrator"`
	LogQuery        bool   `env:"TEST_PG_LOG_QUERY" envDefault:"true"`
}

func getDbConn() (*pg.DB, error) {
	//if db != nil {
	//	return db, nil
	//}
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		//Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	return dbConnection, nil
}

func TestUpdateStatusWithAppIdEnvId(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("Test-1 error in getting app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testOutputContainer := appStatus.AppStatusContainer{
			AppId:          0,
			EnvId:          1,
			InstalledAppId: 1,
			Status:         "Healthy",
		}
		appStatusRepositoryMocked.On("Get", 1, 1).Return(testOutputContainer, fmt.Errorf("get error"))

		err = appStatusService.UpdateStatusWithAppIdEnvId(1, 1, "Progressing")
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), "get error")
	})

	t.Run("Test-2 error in creating app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{}

		tx, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", 1, 1).Return(testInputContainer, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(tx)
		appStatusRepositoryMocked.On("Create", mock.AnythingOfType("struct"), mock.AnythingOfType("struct")).Return(fmt.Errorf("create error"))

		err = appStatusService.UpdateStatusWithAppIdEnvId(1, 1, "Progressing")
		assert.NotNil(tt, err)
		assert.Equal(tt, "create error", err.Error())
	})

}

func TestDeleteWithAppIdEnvId(t *testing.T) {

}
