package pipeline

import (
	repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"log"
	"testing"

	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/util"
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

var (
	dockerRegistryConfig *DockerRegistryConfigImpl
	validInputForSaving  = &DockerArtifactStoreBean{
		Id:                     "integration-test-store-1",
		PluginId:               "cd.go.artifact.docker.registry",
		RegistryType:           "docker-hub",
		IsDefault:              true,
		RegistryURL:            "docker.io",
		Username:               "test-user",
		Password:               "test-password",
		IsOCICompliantRegistry: true,
		OCIRegistryConfig: map[string]string{
			"CHART":     "PULL/PUSH",
			"CONTAINER": "PULL/PUSH",
		},
		DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
			Id:                   0,
			CredentialType:       "SAME_AS_REGISTRY",
			AppliedClusterIdsCsv: "",
			IgnoredClusterIdsCsv: "-1",
		},
	}
)

func TestRegistryConfigService_Save(t *testing.T) {
	if dockerRegistryConfig == nil {
		InitDockerRegistryConfig()
	}
	testCases := []struct {
		name        string
		input       *DockerArtifactStoreBean
		expectedErr bool
	}{
		{
			name:        "TEST : successfully save the registry",
			input:       validInputForSaving,
			expectedErr: false,
		},
		{
			name:        "TEST : error while saving the registry, record already exists",
			input:       validInputForSaving,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			tc.input.User = 1
			res, err := dockerRegistryConfig.Create(tc.input)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, res.Id, "integration-test-store-1")
				assert.True(tt, res.IsOCICompliantRegistry)
				assert.Equal(tt, res.OCIRegistryConfig["CONTAINER"], "PULL/PUSH")
				assert.Equal(tt, res.OCIRegistryConfig["CHART"], "PULL/PUSH")
				assert.Equal(tt, res.DockerRegistryIpsConfig.CredentialType, repository.DockerRegistryIpsCredentialType("SAME_AS_REGISTRY"))
			}
		})
	}
	//clean data in db
	cleanDb(t)
}

func TestRegistryConfigService_Update(t *testing.T) {
	if dockerRegistryConfig == nil {
		InitDockerRegistryConfig()
	}
	// insert a cluster note in the database which will be updated later
	validInputForSaving.User = 1
	savedRegisrty, err := dockerRegistryConfig.Create(validInputForSaving)
	if err != nil {
		t.Fatalf("Error inserting record in database: %s", err.Error())
	}
	delete(savedRegisrty.OCIRegistryConfig, "CHART")
	// define input for update function
	testCases := []struct {
		name        string
		input       *DockerArtifactStoreBean
		expectedErr bool
	}{
		{
			name: "TEST : error while updating a non-existing registry",
			input: &DockerArtifactStoreBean{
				Id:           "non-existing-registry",
				PluginId:     "cd.go.artifact.docker.registry",
				RegistryType: "docker-hub",
				IsDefault:    true,
				RegistryURL:  "docker.io",
			},
			expectedErr: true,
		},
		{
			name:        "TEST : successfully update the note",
			input:       savedRegisrty,
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			tc.input.User = 1
			res, err := dockerRegistryConfig.Update(tc.input)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				assert.Nil(tt, err)
				assert.Equal(tt, res.Id, "integration-test-store-1")
				assert.Equal(tt, res.OCIRegistryConfig["CONTAINER"], "PULL/PUSH")
				_, containerStorageActionExists := res.OCIRegistryConfig["CHART"]
				assert.False(tt, containerStorageActionExists)
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

	query := "DELETE FROM oci_registry_config WHERE docker_artifact_store_id = ?;\n" +
		"DELETE FROM docker_registry_ips_config WHERE docker_artifact_store_id = ?;\n" +
		"DELETE FROM docker_artifact_store WHERE id = ?;\n"
	_, err := DB.Exec(query, "integration-test-store-1", "integration-test-store-1", "integration-test-store-1")
	assert.Nil(tt, err)
	if err != nil {
		return
	}
}

func InitDockerRegistryConfig() {
	if dockerRegistryConfig != nil {
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

	dockerArtifactStoreRepository := repository.NewDockerArtifactStoreRepositoryImpl(conn)
	dockerRegistryIpsConfigRepository := repository.NewDockerRegistryIpsConfigRepositoryImpl(conn)
	ociRegistryConfigRepository := repository.NewOCIRegistryConfigRepositoryImpl(conn)
	dockerRegistryConfig = NewDockerRegistryConfigImpl(logger, dockerArtifactStoreRepository, dockerRegistryIpsConfigRepository, ociRegistryConfigRepository)
}
