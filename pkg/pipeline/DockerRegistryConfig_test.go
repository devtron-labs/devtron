package pipeline

import (
	"context"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/helm-app/mocks"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
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
	dockerRegistryConfig          *DockerRegistryConfigImpl
	dockerArtifactStoreRepository *repository.DockerArtifactStoreRepositoryImpl
	helmAppServiceMock            *mocks.HelmAppService
	storeIds                      = []string{"integration-test-store-1", "integration-test-store-2", "integration-test-store-3"}
	validInput1                   = DockerArtifactStoreBean{
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
			CredentialType:       dockerRegistry.IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY,
			AppliedClusterIdsCsv: "",
			IgnoredClusterIdsCsv: "-1",
			Active:               true,
		},
	}
	validInput2 = DockerArtifactStoreBean{
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
			CredentialType:       dockerRegistry.IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY,
			AppliedClusterIdsCsv: "",
			IgnoredClusterIdsCsv: "-1",
			Active:               true,
		},
	}
	validInput3 = DockerArtifactStoreBean{
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
		InitDockerRegistryConfig(t)
	}
	testCases := []struct {
		name        string
		input       DockerArtifactStoreBean
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
	//clean data in db
	t.Cleanup(cleanDb)
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			tc.input.User = 1
			res, err := dockerRegistryConfig.Create(&tc.input)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				store, err := dockerRegistryConfig.FetchOneDockerAccount(tc.input.Id)
				if err != nil {
					t.Fatalf("Error inserting record in database: %s", err.Error())
				}
				assert.Nil(tt, err)
				assert.Equal(tt, tc.input.Id, res.Id)
				assert.Equal(tt, tc.input.Id, store.Id)
				assert.Equal(tt, tc.input.IsPublic, res.IsPublic)
				assert.Equal(tt, tc.input.IsPublic, store.IsPublic)
				assert.Equal(tt, tc.input.IsOCICompliantRegistry, res.IsOCICompliantRegistry)
				assert.Equal(tt, tc.input.IsOCICompliantRegistry, store.IsOCICompliantRegistry)
				if tc.input.OCIRegistryConfig != nil {
					if _, inputStorageActionExists := res.OCIRegistryConfig["CONTAINER"]; !inputStorageActionExists {
						_, containerStorageActionExists := res.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
						_, containerStorageActionExists = store.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
					} else {
						assert.Equal(tt, tc.input.OCIRegistryConfig["CONTAINER"], res.OCIRegistryConfig["CONTAINER"])
						assert.Equal(tt, tc.input.OCIRegistryConfig["CONTAINER"], store.OCIRegistryConfig["CONTAINER"])
					}

					if _, inputStorageActionExists := res.OCIRegistryConfig["CHART"]; !inputStorageActionExists {
						_, chartStorageActionExists := res.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
						_, chartStorageActionExists = store.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
					} else {
						assert.Equal(tt, tc.input.OCIRegistryConfig["CHART"], res.OCIRegistryConfig["CHART"])
						assert.Equal(tt, tc.input.OCIRegistryConfig["CHART"], store.OCIRegistryConfig["CHART"])
					}
				}
				if tc.input.DockerRegistryIpsConfig != nil {
					assert.NotZero(tt, res.DockerRegistryIpsConfig.Id)
					assert.NotZero(tt, store.DockerRegistryIpsConfig.Id)
					assert.Equal(tt, res.DockerRegistryIpsConfig.Id, store.DockerRegistryIpsConfig.Id)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialType, res.DockerRegistryIpsConfig.CredentialType)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialType, store.DockerRegistryIpsConfig.CredentialType)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialValue, res.DockerRegistryIpsConfig.CredentialValue)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialValue, store.DockerRegistryIpsConfig.CredentialValue)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.AppliedClusterIdsCsv, res.DockerRegistryIpsConfig.AppliedClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.AppliedClusterIdsCsv, store.DockerRegistryIpsConfig.AppliedClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.IgnoredClusterIdsCsv, res.DockerRegistryIpsConfig.IgnoredClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.IgnoredClusterIdsCsv, store.DockerRegistryIpsConfig.IgnoredClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.Active, res.DockerRegistryIpsConfig.Active)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.Active, store.DockerRegistryIpsConfig.Active)
				}
			}
		})
	}
	t.Run(fmt.Sprintf("TEST%v : successfully fetch all chart providers", len(testCases)+1), func(t *testing.T) {
		list := make([]string, 0)
		for _, testcase := range testCases {
			if !testcase.expectedErr {
				chartConfig, chartConfigExists := testcase.input.OCIRegistryConfig["CHART"]
				if testcase.input.IsOCICompliantRegistry && chartConfigExists && chartConfig == "PUSH" {
					continue
				}
				list = append(list, testcase.input.Id)
			}
		}
		providerList := make([]string, 0)
		chartProviders, err := dockerArtifactStoreRepository.FindAllChartProviders()
		for _, provider := range chartProviders {
			providerList = append(providerList, provider.Id)
		}
		assert.Nil(t, err)
		for _, registry := range list {
			assert.Contains(t, providerList, registry)
		}
	})
}

func TestRegistryConfigService_Update(t *testing.T) {
	t.SkipNow()
	if dockerRegistryConfig == nil {
		InitDockerRegistryConfig(t)
	}
	//clean data in db
	t.Cleanup(cleanDb)

	// insert a cluster note in the database which will be updated later
	updateRgistry1 := validInput2
	updateRgistry1.User = 1

	updateRgistry3 := validInput2
	updateRgistry3.User = 1

	updateRgistry5 := validInput2
	updateRgistry5.User = 1

	savedRegisrty, err := dockerRegistryConfig.Create(&updateRgistry1)
	if err != nil {
		t.Fatalf("Error inserting record in database: %s", err.Error())
	}
	assert.Equal(t, updateRgistry1.OCIRegistryConfig["CONTAINER"], savedRegisrty.OCIRegistryConfig["CONTAINER"])
	assert.Equal(t, updateRgistry1.OCIRegistryConfig["CHART"], savedRegisrty.OCIRegistryConfig["CHART"])

	updateRgistry2 := *savedRegisrty
	updateRgistry2.User = 1
	updateRgistry2.OCIRegistryConfig = map[string]string{
		"CONTAINER": "PULL/PUSH",
	}
	updateRgistry1.RepositoryList = make([]string, 0)

	updateRgistry4 := *savedRegisrty
	updateRgistry4.User = 1
	updateRgistry4.OCIRegistryConfig = map[string]string{
		"CHART": "PULL/PUSH",
	}
	updateRgistry4.DockerRegistryIpsConfig = nil

	// define input for update function
	testCases := []struct {
		name        string
		input       DockerArtifactStoreBean
		expectedErr bool
	}{
		{
			name:        "TEST1 : error while updating a non-existing registry",
			input:       validInput1,
			expectedErr: true,
		}, {
			name:        "TEST2 : successfully update the note",
			input:       updateRgistry2,
			expectedErr: false,
		}, {
			name:        "TEST3 : successfully update the note",
			input:       updateRgistry3,
			expectedErr: false,
		}, {
			name:        "TEST4 : successfully update the note",
			input:       updateRgistry4,
			expectedErr: false,
		}, {
			name:        "TEST5 : successfully update the note",
			input:       updateRgistry5,
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			tc.input.User = 1
			// Set expectations for the mock method call
			helmAppServiceMock.On("ValidateOCIRegistry", context.Background(), &client.RegistryCredential{
				RegistryUrl:  tc.input.RegistryURL,
				Username:     tc.input.Username,
				Password:     tc.input.Password,
				AwsRegion:    tc.input.AWSRegion,
				AccessKey:    tc.input.AWSAccessKeyId,
				SecretKey:    tc.input.AWSSecretAccessKey,
				RegistryType: string(tc.input.RegistryType),
				IsPublic:     tc.input.IsPublic,
			}).Return(true)
			res, err := dockerRegistryConfig.Update(&tc.input)
			if tc.expectedErr {
				assert.NotNil(tt, err)
			} else {
				store, err := dockerRegistryConfig.FetchOneDockerAccount(validInput2.Id)
				if err != nil {
					t.Fatalf("Error fetching record from database: %s", err.Error())
				}
				assert.Nil(tt, err)
				assert.Equal(tt, res.Id, tc.input.Id)
				assert.Equal(tt, store.Id, tc.input.Id)
				assert.Equal(tt, res.IsPublic, tc.input.IsPublic)
				assert.Equal(tt, store.IsPublic, tc.input.IsPublic)
				assert.Equal(tt, tc.input.IsOCICompliantRegistry, res.IsOCICompliantRegistry)
				assert.Equal(tt, tc.input.IsOCICompliantRegistry, store.IsOCICompliantRegistry)
				if tc.input.OCIRegistryConfig != nil {
					if _, inputStorageActionExists := res.OCIRegistryConfig["CONTAINER"]; !inputStorageActionExists {
						_, containerStorageActionExists := res.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
						_, containerStorageActionExists = store.OCIRegistryConfig["CONTAINER"]
						assert.False(tt, containerStorageActionExists)
					} else {
						assert.Equal(tt, tc.input.OCIRegistryConfig["CONTAINER"], res.OCIRegistryConfig["CONTAINER"])
						assert.Equal(tt, tc.input.OCIRegistryConfig["CONTAINER"], store.OCIRegistryConfig["CONTAINER"])
					}

					if _, inputStorageActionExists := res.OCIRegistryConfig["CHART"]; !inputStorageActionExists {
						_, chartStorageActionExists := res.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
						_, chartStorageActionExists = store.OCIRegistryConfig["CHART"]
						assert.False(tt, chartStorageActionExists)
					} else {
						assert.Equal(tt, tc.input.OCIRegistryConfig["CHART"], res.OCIRegistryConfig["CHART"])
						assert.Equal(tt, tc.input.OCIRegistryConfig["CHART"], store.OCIRegistryConfig["CHART"])
					}
				} else {
					assert.Nil(tt, res.OCIRegistryConfig)
					assert.Nil(tt, store.OCIRegistryConfig)
				}
				if tc.input.DockerRegistryIpsConfig != nil {
					assert.NotZero(tt, res.DockerRegistryIpsConfig.Id)
					assert.NotZero(tt, store.DockerRegistryIpsConfig.Id)
					assert.Equal(tt, res.DockerRegistryIpsConfig.Id, store.DockerRegistryIpsConfig.Id)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialType, res.DockerRegistryIpsConfig.CredentialType)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialType, store.DockerRegistryIpsConfig.CredentialType)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialValue, res.DockerRegistryIpsConfig.CredentialValue)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.CredentialValue, store.DockerRegistryIpsConfig.CredentialValue)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.AppliedClusterIdsCsv, res.DockerRegistryIpsConfig.AppliedClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.AppliedClusterIdsCsv, store.DockerRegistryIpsConfig.AppliedClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.IgnoredClusterIdsCsv, res.DockerRegistryIpsConfig.IgnoredClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.IgnoredClusterIdsCsv, store.DockerRegistryIpsConfig.IgnoredClusterIdsCsv)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.Active, res.DockerRegistryIpsConfig.Active)
					assert.Equal(tt, tc.input.DockerRegistryIpsConfig.Active, store.DockerRegistryIpsConfig.Active)
				} else {
					assert.Nil(tt, res.DockerRegistryIpsConfig)
					assert.Nil(tt, store.DockerRegistryIpsConfig)
				}
			}
		})
	}
}

var db *pg.DB

func captureQuery(event *pg.QueryProcessedEvent) {
	fmt.Println(event.FormattedQuery())
}

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

	// Add the query event listener to capture SQL queries.
	db.OnQueryProcessed(captureQuery)

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

func InitDockerRegistryConfig(t *testing.T) {
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

	dockerArtifactStoreRepository = repository.NewDockerArtifactStoreRepositoryImpl(conn)
	dockerRegistryIpsConfigRepository := repository.NewDockerRegistryIpsConfigRepositoryImpl(conn)
	ociRegistryConfigRepository := repository.NewOCIRegistryConfigRepositoryImpl(conn)
	//Mock helm service
	helmAppServiceMock = mocks.NewHelmAppService(t)
	dockerRegistryConfig = NewDockerRegistryConfigImpl(logger, helmAppServiceMock, dockerArtifactStoreRepository, dockerRegistryIpsConfigRepository, ociRegistryConfigRepository)
}
