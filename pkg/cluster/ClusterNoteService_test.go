package cluster

import (
	"log"
	"testing"
	"time"

	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
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

var clusterNoteService *ClusterNoteServiceImpl

func TestClusterNoteService_Save(t *testing.T) {
	t.SkipNow()
	if clusterNoteService == nil {
		InitClusterNoteService()
	}

	testCases := []struct {
		name        string
		input       *ClusterNoteBean
		expectedErr bool
	}{
		{
			name: "TEST : successfully save the note",
			input: &ClusterNoteBean{
				Id:          0,
				ClusterId:   1,
				Description: "Test Note",
				CreatedBy:   1,
				CreatedOn:   time.Now(),
			},
			expectedErr: false,
		},
		{
			name: "TEST : error while saving the existing note",
			input: &ClusterNoteBean{
				Id:          0,
				ClusterId:   1,
				Description: "Test Note",
				CreatedBy:   1,
				CreatedOn:   time.Now(),
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			res, err := clusterNoteService.Save(tc.input, 1)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.NotEqual(tt, res.Id, 0)
			}
		})
	}
}

func TestClusterNoteServiceImpl_Update_InvalidFields(t *testing.T) {
	t.SkipNow()
	if clusterNoteService == nil {
		InitClusterNoteService()
	}

	// insert a cluster note in the database which will be updated later
	note := &ClusterNoteBean{
		ClusterId:   100,
		Description: "test note",
		CreatedBy:   1,
	}
	_, err := clusterNoteService.Save(note, 1)
	if err != nil {
		t.Fatalf("Error inserting record in database: %s", err.Error())
	}

	// define input for update function
	input := &ClusterNoteBean{
		Id:        1,
		ClusterId: 100,
	}

	// try updating the record with invalid fields and check if it returns error
	_, err = clusterNoteService.Update(input, 1)
	if err == nil {
		t.Fatal("Expected an error on updating record with invalid fields, but got nil")
	}
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

func InitClusterNoteService() {
	if clusterNoteService != nil {
		return
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	conn, err := getDbConn()
	if err != nil {
		log.Fatalf("error in db connection initialization %s, %s", "err", err)
	}

	clusterNoteHistoryRepository := repository.NewClusterNoteHistoryRepositoryImpl(conn, logger)
	clusterNoteRepository := repository.NewClusterNoteRepositoryImpl(conn, logger)

	clusterNoteService = NewClusterNoteServiceImpl(clusterNoteRepository, clusterNoteHistoryRepository, logger)
}
