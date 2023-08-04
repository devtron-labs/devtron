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
	storeIds             = []string{"integration-test-store-1", "integration-test-store-2", "integration-test-store-3"}
	validInput1          = &DockerArtifactStoreBean{
		Id:                     storeIds[0],
		PluginId:               "cd.go.artifact.docker.registry",
		RegistryType:           "docker-hub",
		IsDefault:              true,
		RegistryURL:            "docker.io",
		Username:               "test-user",
		Password:               "test-password",
		IsOCICompliantRegistry: false,
		DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
			Id:                   0,
			CredentialType:       "SAME_AS_REGISTRY",
			AppliedClusterIdsCsv: "",
			IgnoredClusterIdsCsv: "-1",
		},
	}
	validInput2 = &DockerArtifactStoreBean{
		Id:                     storeIds[1],
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
		IsPublic:       false,
		RepositoryList: []string{"username/test", "username/chart"},
		DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
			Id:                   0,
			CredentialType:       "SAME_AS_REGISTRY",
			AppliedClusterIdsCsv: "",
			IgnoredClusterIdsCsv: "-1",
		},
	}
	validInput3 = &DockerArtifactStoreBean{
		Id:                     storeIds[2],
		PluginId:               "cd.go.artifact.docker.registry",
		RegistryType:           "docker-hub",
		RegistryURL:            "docker.io",
		IsOCICompliantRegistry: true,
		OCIRegistryConfig: map[string]string{
			"CHART": "PULL",
		},
		IsPublic: true,
	}
)

func TestRegistryConfigService_Save(t *testing.T) {
	t.SkipNow()
	if dockerRegistryConfig == nil {
		InitDockerRegistryConfig()
	}
	testCases := []struct {
		name        string
		input       *DockerArtifactStoreBean
		expectedErr bool
	}{
		{
			name:        "TEST1 : successfully save the registry",
			input:       validInput1,
			expectedErr: false,
		}, {
			name:        "TEST2 : successfully save the registry",
			input:       validInput2,
			expectedErr: false,
		}, {
			name:        "TEST3 : successfully save the registry",
			input:       validInput3,
			expectedErr: false,
		}, {
			name:        "TEST4 : error while saving the registry, record already exists",
			input:       validInput1,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		//clean data in db
		t.Cleanup(cleanDb)
		t.Run(tc.name, func(tt *testing.T) {
			tc.input.User = 1
			res, err := dockerRegistryConfig.Create(tc.input)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				store, err := dockerRegistryConfig.FetchOneDockerAccount(tc.input.Id)
				if err != nil {
					t.Fatalf("Error inserting record in database: %s", err.Error())
				}
				assert.Nil(tt, err)
				assert.Equal(tt, res.Id, tc.input.Id)
				assert.Equal(tt, store.Id, tc.input.Id)
				assert.Equal(tt, res.IsPublic, tc.input.IsPublic)
				assert.Equal(tt, store.IsPublic, tc.input.IsPublic)
				assert.True(tt, res.IsOCICompliantRegistry)
				assert.True(tt, store.IsOCICompliantRegistry)
				if tc.input.OCIRegistryConfig != nil {
					if _, inputStorageActionExists := res.OCIRegistryConfig["CONTAINER"]; !inputStorageActionExists {
						_, containerStorageActionExists := res.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
						_, containerStorageActionExists = store.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
					} else {
						assert.Equal(tt, res.OCIRegistryConfig["CONTAINER"], tc.input.OCIRegistryConfig["CONTAINER"])
						assert.Equal(tt, store.OCIRegistryConfig["CONTAINER"], tc.input.OCIRegistryConfig["CONTAINER"])
					}

					if _, inputStorageActionExists := res.OCIRegistryConfig["CHART"]; !inputStorageActionExists {
						_, chartStorageActionExists := res.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
						_, chartStorageActionExists = store.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
					} else {
						assert.Equal(tt, res.OCIRegistryConfig["CHART"], tc.input.OCIRegistryConfig["CHART"])
						assert.Equal(tt, store.OCIRegistryConfig["CHART"], tc.input.OCIRegistryConfig["CHART"])
					}
				}
				if tc.input.DockerRegistryIpsConfig != nil {
					assert.Equal(tt, res.DockerRegistryIpsConfig.CredentialType, tc.input.DockerRegistryIpsConfig.CredentialValue)
					assert.Equal(tt, store.DockerRegistryIpsConfig.CredentialType, tc.input.DockerRegistryIpsConfig.CredentialValue)
				}
			}
		})
	}
}

func TestRegistryConfigService_Update(t *testing.T) {
	t.SkipNow()
	if dockerRegistryConfig == nil {
		InitDockerRegistryConfig()
	}
	//clean data in db
	t.Cleanup(cleanDb)

	// insert a cluster note in the database which will be updated later
	validInput2.User = 1
	savedRegisrty, err := dockerRegistryConfig.Create(validInput2)
	if err != nil {
		t.Fatalf("Error inserting record in database: %s", err.Error())
	}

	assert.Equal(t, savedRegisrty.OCIRegistryConfig["CONTAINER"], validInput2.OCIRegistryConfig["CONTAINER"])
	assert.Equal(t, savedRegisrty.OCIRegistryConfig["CHART"], validInput2.OCIRegistryConfig["CHART"])

	delete(savedRegisrty.OCIRegistryConfig, "CHART")
	// define input for update function
	testCases := []struct {
		name        string
		input       *DockerArtifactStoreBean
		expectedErr bool
	}{
		{
			name:        "TEST : error while updating a non-existing registry",
			input:       validInput1,
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
				store, err := dockerRegistryConfig.FetchOneDockerAccount(validInput2.Id)
				if err != nil {
					t.Fatalf("Error inserting record in database: %s", err.Error())
				}
				assert.Nil(tt, err)
				assert.Equal(tt, res.Id, tc.input.Id)
				assert.Equal(tt, store.Id, tc.input.Id)
				assert.Equal(tt, res.IsPublic, tc.input.IsPublic)
				assert.Equal(tt, store.IsPublic, tc.input.IsPublic)
				if tc.input.OCIRegistryConfig != nil {
					if _, inputStorageActionExists := res.OCIRegistryConfig["CONTAINER"]; !inputStorageActionExists {
						_, containerStorageActionExists := res.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
						_, containerStorageActionExists = store.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
					} else {
						assert.Equal(tt, res.OCIRegistryConfig["CONTAINER"], tc.input.OCIRegistryConfig["CONTAINER"])
						assert.Equal(tt, store.OCIRegistryConfig["CONTAINER"], tc.input.OCIRegistryConfig["CONTAINER"])
					}

					if _, inputStorageActionExists := res.OCIRegistryConfig["CHART"]; !inputStorageActionExists {
						_, chartStorageActionExists := res.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
						_, chartStorageActionExists = store.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
					} else {
						assert.Equal(tt, res.OCIRegistryConfig["CHART"], tc.input.OCIRegistryConfig["CHART"])
						assert.Equal(tt, store.OCIRegistryConfig["CHART"], tc.input.OCIRegistryConfig["CHART"])
					}
				}
			}
		})
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
		Addr:     cfg.Addr + ":" + cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	}
	db = pg.Connect(&options)
	return db, nil
}

func cleanDb() {
	log.Println("Cleaning Up...")
	DB, _ := getDbConn()
	query := "DELETE FROM oci_registry_config WHERE docker_artifact_store_id IN (?) ;\n" +
		"DELETE FROM docker_registry_ips_config WHERE docker_artifact_store_id IN (?) ;\n" +
		"DELETE FROM docker_artifact_store WHERE id IN (?) ;\n"
	_, err := DB.Query(nil, query, pg.In(storeIds), pg.In(storeIds), pg.In(storeIds))
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
	dockerRegistryConfig = NewDockerRegistryConfigImpl(logger, nil, dockerArtifactStoreRepository, dockerRegistryIpsConfigRepository, ociRegistryConfigRepository)
}
