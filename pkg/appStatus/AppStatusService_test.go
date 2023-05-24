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

func TestUpdateStatusWithAppIdEnvId(t *testing.T) {
	t.SkipNow()
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("Test-1 error in getting app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testOutputContainer := appStatus.AppStatusContainer{
			AppId:  1,
			EnvId:  1,
			Status: "Healthy",
		}
		throwError := fmt.Errorf("get error")
		appStatusRepositoryMocked.On("Get", testOutputContainer.AppId, testOutputContainer.EnvId).Return(testOutputContainer, throwError)

		err = appStatusService.UpdateStatusWithAppIdEnvId(1, 1, "Progressing")
		assert.NotNil(tt, err)
		assert.Equal(tt, throwError.Error(), err.Error())
	})

	t.Run("Test-2 error in creating app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{}

		db, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", 1, 1).Return(testInputContainer, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		appStatusRepositoryMocked.On("Create", mock.AnythingOfTypeArgument("*pg.Tx"), appStatus.AppStatusContainer{AppId: 1, EnvId: 1, Status: "Progressing"}).Return(fmt.Errorf("create error"))

		err = appStatusService.UpdateStatusWithAppIdEnvId(1, 1, "Progressing")
		assert.NotNil(tt, err)
		assert.Equal(tt, "create error", err.Error())
	})

	t.Run("Test-3 success in creating app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{}
		testOutputContainerFromDb := appStatus.AppStatusContainer{
			AppId:  1,
			EnvId:  1,
			Status: "Progressing",
		}

		db, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", testOutputContainerFromDb.AppId, testOutputContainerFromDb.EnvId).Return(testInputContainer, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		appStatusRepositoryMocked.On("Create", mock.AnythingOfTypeArgument("*pg.Tx"), testOutputContainerFromDb).Return(nil)

		err = appStatusService.UpdateStatusWithAppIdEnvId(testOutputContainerFromDb.AppId, testOutputContainerFromDb.EnvId, testOutputContainerFromDb.Status)
		assert.Nil(tt, err)
	})

	t.Run("Test-4 No change in app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{
			AppId:  1,
			EnvId:  1,
			Status: "Progressing",
		}

		db, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", testInputContainer.AppId, testInputContainer.EnvId).Return(testInputContainer, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(db)

		err = appStatusService.UpdateStatusWithAppIdEnvId(testInputContainer.AppId, testInputContainer.EnvId, testInputContainer.Status)
		assert.Nil(tt, err)
	})

	t.Run("Test-5 error in updating app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testOutputContainerFromDb := appStatus.AppStatusContainer{
			AppId:  1,
			EnvId:  1,
			Status: "Healthy",
		}
		testInputContainer := appStatus.AppStatusContainer{
			AppId:  1,
			EnvId:  1,
			Status: "Progressing",
		}
		db, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", testOutputContainerFromDb.AppId, testOutputContainerFromDb.EnvId).Return(testOutputContainerFromDb, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		expectedTestError := fmt.Errorf("error in updating app-status")
		appStatusRepositoryMocked.On("Update", mock.AnythingOfTypeArgument("*pg.Tx"), testInputContainer).Return(expectedTestError)

		err = appStatusService.UpdateStatusWithAppIdEnvId(testInputContainer.AppId, testInputContainer.EnvId, testInputContainer.Status)
		assert.NotNil(tt, err)
		assert.Equal(tt, expectedTestError.Error(), err.Error())
	})

	t.Run("Test-6 success in updating app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testOutputContainerFromDb := appStatus.AppStatusContainer{
			AppId:  2,
			EnvId:  2,
			Status: "Healthy",
		}
		testInputContainer := appStatus.AppStatusContainer{
			AppId:  2,
			EnvId:  2,
			Status: "Progressing",
		}

		db, _ := getDbConn()
		appStatusRepositoryMocked.On("Get", testOutputContainerFromDb.AppId, testOutputContainerFromDb.EnvId).Return(testOutputContainerFromDb, nil)
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		appStatusRepositoryMocked.On("Update", mock.AnythingOfTypeArgument("*pg.Tx"), testInputContainer).Return(nil)

		err = appStatusService.UpdateStatusWithAppIdEnvId(testInputContainer.AppId, testInputContainer.EnvId, testInputContainer.Status)
		assert.Nil(tt, err)
	})
}

func TestDeleteWithAppIdEnvId(t *testing.T) {
	t.SkipNow()
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)

	t.Run("Test-1 error in deleting app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{
			AppId: 1,
			EnvId: 1,
		}
		db, _ := getDbConn()
		tx, _ := db.Begin()
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		expectedTestError := fmt.Errorf("error in deleting app-status")
		appStatusRepositoryMocked.On("Delete", mock.AnythingOfTypeArgument("*pg.Tx"), testInputContainer.AppId, testInputContainer.EnvId).Return(expectedTestError)

		err = appStatusService.DeleteWithAppIdEnvId(tx, testInputContainer.AppId, testInputContainer.EnvId)
		assert.NotNil(tt, err)
		assert.Equal(tt, expectedTestError.Error(), err.Error())
	})

	t.Run("Test-2 success in deleting app-status", func(tt *testing.T) {
		appStatusRepositoryMocked := mocks.NewAppStatusRepository(t)
		appStatusService := NewAppStatusServiceImpl(appStatusRepositoryMocked, logger, nil, nil)
		testInputContainer := appStatus.AppStatusContainer{
			AppId: 1,
			EnvId: 1,
		}

		db, _ := getDbConn()
		tx, _ := db.Begin()
		appStatusRepositoryMocked.On("GetConnection").Return(db)
		appStatusRepositoryMocked.On("Delete", mock.AnythingOfTypeArgument("*pg.Tx"), testInputContainer.AppId, testInputContainer.EnvId).Return(nil)

		err = appStatusService.DeleteWithAppIdEnvId(tx, testInputContainer.AppId, testInputContainer.EnvId)
		assert.Nil(tt, err)
	})
}

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
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	return dbConnection, nil
}
