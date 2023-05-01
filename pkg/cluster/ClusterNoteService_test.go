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
	Addr     string `env:"TEST_PG_ADDR" envDefault:"127.0.0.1"`
	Port     string `env:"TEST_PG_PORT" envDefault:"5432"`
	User     string `env:"TEST_PG_USER" envDefault:"postgres"`
	Password string `env:"TEST_PG_PASSWORD" envDefault:"postgrespw" secretData:"-"`
	Database string `env:"TEST_PG_DATABASE" envDefault:"orchestrator"`
	LogQuery bool   `env:"TEST_PG_LOG_QUERY" envDefault:"true"`
}

var clusterNoteService *ClusterNoteServiceImpl

func TestClusterNoteService_Save(t *testing.T) {
	if clusterNoteService == nil {
		InitClusterNoteService()
	}
	initialiseDb(t)
	testCases := []struct {
		name        string
		input       *ClusterNoteBean
		expectedErr bool
	}{
		{
			name: "TEST : successfully save the note",
			input: &ClusterNoteBean{
				Id:          0,
				ClusterId:   10000,
				Description: "Test Note",
				UpdatedBy:   1,
				UpdatedOn:   time.Now(),
			},
			expectedErr: false,
		},
		{
			name: "TEST : error while saving the existing note",
			input: &ClusterNoteBean{
				Id:          0,
				ClusterId:   10000,
				Description: "Test Note",
				UpdatedBy:   1,
				UpdatedOn:   time.Now(),
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
	//clean data in db
	cleanDb(t)
}

func TestClusterNoteServiceImpl_Update(t *testing.T) {
	if clusterNoteService == nil {
		InitClusterNoteService()
	}
	initialiseDb(t)
	// insert a cluster note in the database which will be updated later
	note := &ClusterNoteBean{
		ClusterId:   10001,
		Description: "test note",
		UpdatedBy:   1,
	}
	_, err := clusterNoteService.Save(note, 1)
	if err != nil {
		t.Fatalf("Error inserting record in database: %s", err.Error())
	}

	// define input for update function
	testCases := []struct {
		name        string
		input       *ClusterNoteBean
		expectedErr bool
	}{
		{
			name: "TEST : error while updating the existing note",
			input: &ClusterNoteBean{
				Id:        1,
				ClusterId: 100,
			},
			expectedErr: true,
		},
		{
			name: "TEST : successfully update the note",
			input: &ClusterNoteBean{
				Description: "Updated Text",
				ClusterId:   10001,
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			res, err := clusterNoteService.Update(tc.input, 1)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.NotEqual(tt, res.Id, 0)
			}
		})
	}

	//clean data in db
	cleanDb(t)
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
		Addr:     cfg.Addr + ":" + cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	}
	db = pg.Connect(&options)
	return db, nil
}

func cleanDb(tt *testing.T) {
	DB, _ := getDbConn()
	query := "DELETE FROM cluster_note_history WHERE note_id IN (SELECT id FROM cluster_note);\n" +
		"DELETE FROM cluster_note WHERE cluster_id IN (SELECT id FROM cluster);\n" +
		"DELETE FROM cluster WHERE id=10000 OR id=10001;\n"
	_, err := DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
}

func createClusterData(DB *pg.DB, bean *ClusterBean) error {
	model := &repository.Cluster{
		Id:          bean.Id,
		ClusterName: bean.ClusterName,
		ServerUrl:   bean.ServerUrl,
	}
	model.CreatedBy = 1
	model.UpdatedBy = 1
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	return DB.Insert(model)
}

func initialiseDb(tt *testing.T) {
	DB, _ := getDbConn()
	clusters := []ClusterBean{
		{
			Id:          10000,
			ClusterName: "test-cluster-1",
			ServerUrl:   "https://test1.cluster",
		},
		{
			Id:          10001,
			ClusterName: "test-cluster-2",
			ServerUrl:   "https://test2.cluster",
		},
	}
	for _, cluster := range clusters {
		err := createClusterData(DB, &cluster)
		assert.Nil(tt, err)
		if err != nil {
			return
		}
	}
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
	clusterNoteHistoryService := NewClusterNoteHistoryServiceImpl(clusterNoteHistoryRepository, logger)
	clusterNoteService = NewClusterNoteServiceImpl(clusterNoteRepository, clusterNoteHistoryService, logger)
}
