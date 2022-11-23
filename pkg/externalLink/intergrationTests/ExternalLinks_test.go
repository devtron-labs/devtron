package intergrationTests

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type Config struct {
	Addr            string `env:"TEST_PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"TEST_PG_PORT" envDefault:"8085"`
	User            string `env:"TEST_PG_USER" envDefault:""`
	Password        string `env:"TEST_PG_PASSWORD" envDefault:"" secretData:"-"`
	Database        string `env:"TEST_PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"TEST_APP" envDefault:"orchestrator"`
	LogQuery        bool   `env:"TEST_PG_LOG_QUERY" envDefault:"true"`
}

var externalLinkService *externalLink.ExternalLinkServiceImpl

func TestExternalLinkServiceImpl_Create(t *testing.T) {
	if externalLinkService == nil {
		InitExternalLinkService()
	}
	inputData := make([]*externalLink.ExternalLinkDto, 0)
	inp1 := externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "IntegrationTest-1",
		Description:      "integration test link description",
		Type:             "appLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "devtron-app", Identifier: "1", ClusterId: 0},
			{Type: "devtron-app", Identifier: "103", ClusterId: 0},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	}
	//data created via CREATE API
	inputData = append(inputData, &inp1)
	res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, true, res.Success)

	//fetch data via GET API

	outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(outputData))
	assert.Equal(t, inputData[0].Name, outputData[0].Name)
	assert.Equal(t, inputData[0].Description, outputData[0].Description)
	assert.Equal(t, inputData[0].Type, outputData[0].Type)
	assert.Equal(t, inputData[0].Url, outputData[0].Url)
	assert.Equal(t, inputData[0].IsEditable, outputData[0].IsEditable)
	assert.NotNil(t, outputData[0].Identifiers)
	assert.Equal(t, 2, outputData[0].Identifiers)
	for i, idf := range inputData[0].Identifiers {
		assert.Equal(t, idf.Type, outputData[0].Identifiers[i].Type)
		assert.Equal(t, idf.Identifier, outputData[0].Identifiers[i].Identifier)
		assert.Equal(t, idf.ClusterId, outputData[0].Identifiers[i].ClusterId)
	}
	//clean created data
	cleanDb()
}

func TestExternalLinkServiceImpl_Update(t *testing.T) {
	//update app to all apps

	t.Run("TEST : update link from app to all apps", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)

		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 1)
		assert.Nil(tt, err)
		assert.Equal(tt, 1, len(outputData))

		//change update fields
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update it via update API
		res, err = externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 1)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, expectedResultLink.Identifiers, outputDataAfterUpdate[0].Identifiers)

		//clean data in db
		cleanDb()
	})

	//update 1app to 1cluster
	t.Run("TEST : update link from app to cluster", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)

		//get created link
		outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 1)
		assert.Nil(tt, err)
		assert.Equal(tt, 1, len(outputData))

		//run tests
		cleanDb()
	})
	//update 1app to all cluster
	t.Run("TEST : update link from app to all cluster", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
	//update 1cluster to 1 app
	t.Run("TEST : update link from cluster to app", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
	//update 1cluster to all cluster
	t.Run("TEST : update link from app to all apps", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
	//update 1cluster to all apps
	t.Run("TEST : update link from cluster to all apps", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
	//all apps to all cluster
	t.Run("TEST : update link from all app to all clusters", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
	//all cluster to all apps
	t.Run("TEST : update link from all app to all clusters", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers: []externalLink.LinkIdentifier{
				{Type: "devtron-app", Identifier: "1", ClusterId: 0},
				{Type: "devtron-app", Identifier: "103", ClusterId: 0},
			},
			Url:        "http://integration-test.com",
			IsEditable: true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)
		//run tests
		cleanDb()
	})
}

func Copy(to *externalLink.ExternalLinkDto, from *externalLink.ExternalLinkDto) {
	to.Type = from.Type
	to.Id = from.Id
	to.MonitoringToolId = from.MonitoringToolId
	to.Name = from.Name
	to.Url = from.Url
	to.IsEditable = from.IsEditable
	to.Description = from.Description
}
func cleanDb() {
	var inf interface{}
	tx, _ := db.Begin()
	defer tx.Rollback()
	query := "DELETE FROM external_link WHERE id IS NOT NULL;"
	_, err := tx.Query(inf, query)
	if err != nil {
		return
	}
	query = "DELETE FROM external_link_identifier_mapping WHERE id IS NOT NULL;"
	_, err = tx.Query(inf, query)
	if err != nil {
		return
	}
	tx.Commit()
}

var db *pg.DB

func getDbConn() (*pg.DB, error) {
	if db != nil {
		return db, nil
	}
	cfg := Config{}
	err := env.Parse(cfg)
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

func InitExternalLinkService() {
	if externalLinkService != nil {
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

	externalLinkMonitoringToolRepository := externalLink.NewExternalLinkMonitoringToolRepositoryImpl(conn)
	externalLinkRepository := externalLink.NewExternalLinkRepositoryImpl(conn)
	externalLinkidentifierMappingRepository := externalLink.NewExternalLinkIdentifierMappingRepositoryImpl(conn)

	externalLinkService = externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkidentifierMappingRepository, externalLinkRepository)
}
