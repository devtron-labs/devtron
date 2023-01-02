package tests

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestAppStatusRepositoryImpl_Get(t *testing.T) {
	repo := initAppStatusRepo()
	testData := getTestdata()[0]

	//insert dummy data
	insertTestData(testData)

	//get the data with the API
	dataFromAPI, err := repo.Get(testData.AppId, testData.EnvId)
	assert.Nil(t, err)
	assert.NotNil(t, dataFromAPI)
	assert.Equal(t, testData.AppId, dataFromAPI.AppId)
	assert.Equal(t, testData.EnvId, dataFromAPI.EnvId)
	assert.Equal(t, testData.Status, dataFromAPI.Status)

	//delete the test data
	deleteTestdata()
}

func TestAppStatusRepositoryImpl_Create(t *testing.T) {
	repo := initAppStatusRepo()
	testData := getTestdata()
	//create dummy data in the db
	tx, _ := db.Begin()
	for _, data := range testData {
		err := repo.Create(tx, data)
		assert.NotNil(t, err)
	}
	err := tx.Commit()
	if err != nil {
		log.Fatal("error in committing data in db", "err", err)
	}
	//verify the inserted data
	for _, expectedData := range testData {
		dataFromDb, err := repo.Get(expectedData.AppId, expectedData.EnvId)
		assert.Nil(t, err)
		assert.NotNil(t, dataFromDb)
		assert.Equal(t, expectedData.AppId, dataFromDb.AppId)
		assert.Equal(t, expectedData.EnvId, dataFromDb.EnvId)
		assert.Equal(t, expectedData.Status, dataFromDb.Status)
	}
	//delete the data from db
	deleteTestdata()
}
func TestAppStatusRepositoryImpl_Update(t *testing.T) {
	repo := initAppStatusRepo()

}
func TestAppStatusRepositoryImpl_Delete(t *testing.T) {
	repo := initAppStatusRepo()

}

func TestAppStatusRepositoryImpl_DeleteWithAppId(t *testing.T) {
	repo := initAppStatusRepo()

}
func TestAppStatusRepositoryImpl_DeleteWithEnvId(t *testing.T) {
	repo := initAppStatusRepo()

}

func getTestdata() []appStatus.AppStatusContainer {
	testDataArray := make([]appStatus.AppStatusContainer, 0)
	testDataArray = append(testDataArray, appStatus.AppStatusContainer{
		AppId:  1,
		EnvId:  1,
		Status: "Healthy",
	})

	testDataArray = append(testDataArray, appStatus.AppStatusContainer{
		AppId:  1,
		EnvId:  2,
		Status: "Progressing",
	})

	testDataArray = append(testDataArray, appStatus.AppStatusContainer{
		AppId:  2,
		EnvId:  1,
		Status: "Degraded",
	})

	testDataArray = append(testDataArray, appStatus.AppStatusContainer{
		AppId:  2,
		EnvId:  2,
		Status: "Missing",
	})

	return testDataArray
}

//utilities
func insertTestData(testData appStatus.AppStatusContainer) {
	model := appStatus.AppStatusDto{}
	query := "insert into" +
		" app_status(app_id,env_id,status,updated_on)" +
		" values(?,?,?,now());"
	_, err := db.Query(model, query, testData.AppId, testData.EnvId, testData.Status)
	if err != nil {
		log.Fatal("error in inserting data in db", "err", err)
	}

}

func deleteTestdata() {
	model := appStatus.AppStatusDto{}
	query := "DELETE " +
		"FROM app_status;"
	_, err := db.Query(&model, query)
	if err != nil {
		log.Fatal("error in deleting data from db", "err", err)
	}

	return
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

var db *pg.DB

func getDbConn() (*pg.DB, error) {
	if db != nil {
		return db, nil
	}
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	return dbConnection, nil
}

var appStatusRepository *appStatus.AppStatusRepositoryImpl

func initAppStatusRepo() *appStatus.AppStatusRepositoryImpl {
	if appStatusRepository != nil {
		return appStatusRepository
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	conn, err := getDbConn()
	if err != nil {
		log.Fatalf("error in db connection initialization %s, %s", "err", err)
	}
	err = createAppStatusTable()
	if err != nil {
		log.Fatalf("error in creating app_status table %s, %s", "err", err)
	}
	appStatusRepository = appStatus.NewAppStatusRepositoryImpl(conn, logger)
	return appStatusRepository
}

func createAppStatusTable() error {
	cmd := "CREATE TABLE IF NOT EXISTS public.app_status" +
		"(\"app_id\" integer,\"env_id\" integer," +
		"\"status\" varchar(255)," +
		"\"updated_on\" timestamp with time zone NOT NULL," +
		"PRIMARY KEY (\"app_id\",\"env_id\"),"

	_, err := db.Exec(cmd)
	return err
}
