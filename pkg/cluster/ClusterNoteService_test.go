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
	Database string `env:"TEST_PG_DATABASE" envDefault:"postgres"`
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
				ClusterId:   1,
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
				ClusterId:   1,
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

func TestClusterNoteServiceImpl_Update_InvalidFields(t *testing.T) {
	if clusterNoteService == nil {
		InitClusterNoteService()
	}
	initialiseDb(t)
	// insert a cluster note in the database which will be updated later
	note := &ClusterNoteBean{
		ClusterId:   100,
		Description: "test note",
		UpdatedBy:   1,
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
	dbConnection := pg.Connect(&options)
	return dbConnection, nil
}

func cleanDb(tt *testing.T) {
	DB, _ := getDbConn()
	query := "truncate cluster cascade;"
	_, err := DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
}

func initialiseDb(tt *testing.T) {
	DB, _ := getDbConn()
	query := "CREATE TABLE IF NOT EXISTS public.cluster (\n    id integer NOT NULL PRIMARY KEY,\n    cluster_name character varying(250) NOT NULL,\n    active boolean NOT NULL,\n    created_on timestamp with time zone NOT NULL,\n    created_by integer NOT NULL,\n    updated_on timestamp with time zone NOT NULL,\n    updated_by integer NOT NULL,\n    server_url character varying(250),\n    config json,\n    prometheus_endpoint character varying(250),\n    cd_argo_setup boolean DEFAULT false,\n    p_username character varying(250),\n    p_password character varying(250),\n    p_tls_client_cert text,\n    p_tls_client_key text\n);\n\n\nALTER TABLE public.cluster OWNER TO postgres;"
	_, err := DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
	query = "-- Sequence and defined type for cluster_note\nCREATE SEQUENCE IF NOT EXISTS id_seq_cluster_note;\n\n-- Sequence and defined type for cluster_note_history\nCREATE SEQUENCE IF NOT EXISTS id_seq_cluster_note_history;\n\n-- cluster_note Table Definition\nCREATE TABLE IF NOT EXISTS public.cluster_note\n(\n    \"id\"                        int4         NOT NULL DEFAULT nextval('id_seq_cluster_note'::regclass),\n    \"cluster_id\"                int4         NOT NULL,\n    \"description\"               TEXT         NOT NULL,\n    \"created_on\"                timestamptz  NOT NULL,\n    \"created_by\"                int4         NOT NULL,\n    \"updated_on\"                timestamptz,\n    \"updated_by\"                int4,\n    PRIMARY KEY (\"id\")\n);\n\n-- add foreign key\nALTER TABLE \"public\".\"cluster_note\"\n    ADD FOREIGN KEY (\"cluster_id\") REFERENCES \"public\".\"cluster\" (\"id\");\n\n-- cluster_note_history Table Definition\nCREATE TABLE IF NOT EXISTS public.cluster_note_history\n(\n    \"id\"                        int4         NOT NULL DEFAULT nextval('id_seq_cluster_note_history'::regclass),\n    \"note_id\"                   int4         NOT NULL,\n    \"description\"               TEXT         NOT NULL,\n    \"created_on\"                timestamptz  NOT NULL,\n    \"created_by\"                int4         NOT NULL,\n    \"updated_on\"                timestamptz,\n    \"updated_by\"                int4,\n    PRIMARY KEY (\"id\")\n);\n\n-- add foreign key\nALTER TABLE \"public\".\"cluster_note_history\"\n    ADD FOREIGN KEY (\"note_id\") REFERENCES \"public\".\"cluster_note\" (\"id\");"
	_, err = DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
	query = "INSERT INTO public.cluster (\nid, cluster_name, active, created_on, created_by, updated_on, updated_by) VALUES (\n'1'::integer, 'test'::character varying, true::boolean, '2009-10-31T01:48:52Z '::timestamp with time zone, '1'::integer, '2009-10-31T01:48:52Z '::timestamp with time zone, '1'::integer)\n returning id;"
	_, err = DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
	query = "INSERT INTO public.cluster (\nid, cluster_name, active, created_on, created_by, updated_on, updated_by) VALUES (\n'100'::integer, 'test-cluster'::character varying, true::boolean, '2009-10-31T01:48:52Z '::timestamp with time zone, '1'::integer, '2009-10-31T01:48:52Z '::timestamp with time zone, '1'::integer)\n returning id;"
	_, err = DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
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
