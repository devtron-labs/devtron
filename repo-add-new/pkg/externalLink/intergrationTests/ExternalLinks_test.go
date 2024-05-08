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
	Port            string `env:"TEST_PG_PORT" envDefault:"55000"`
	User            string `env:"TEST_PG_USER" envDefault:"postgres"`
	Password        string `env:"TEST_PG_PASSWORD" envDefault:"postgrespw" secretData:"-"`
	Database        string `env:"TEST_PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"TEST_APP" envDefault:"orchestrator"`
	LogQuery        bool   `env:"TEST_PG_LOG_QUERY" envDefault:"true"`
}

var externalLinkService *externalLink.ExternalLinkServiceImpl

func TestExternalLinkServiceImpl_Create(t *testing.T) {
	t.SkipNow()
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
	t.Run("Test Create API With Valid Payload", func(tt *testing.T) {
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//fetch data via GET API

		outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.Equal(tt, 1, len(outputData))
		assert.Equal(tt, inputData[0].Name, outputData[0].Name)
		assert.Equal(tt, inputData[0].Description, outputData[0].Description)
		assert.Equal(tt, inputData[0].Type, outputData[0].Type)
		assert.Equal(tt, inputData[0].Url, outputData[0].Url)
		assert.Equal(tt, inputData[0].IsEditable, outputData[0].IsEditable)
		assert.NotNil(tt, outputData[0].Identifiers)
		assert.Equal(tt, 2, len(outputData[0].Identifiers))
		for i, idf := range inputData[0].Identifiers {
			assert.Equal(tt, idf.Type, outputData[0].Identifiers[i].Type)
			assert.Equal(tt, idf.Identifier, outputData[0].Identifiers[i].Identifier)
			assert.Equal(tt, idf.ClusterId, outputData[0].Identifiers[i].ClusterId)
		}
		//clean created data
		cleanDb(tt)
	})

	t.Run("Test Create API With InValid Payload - 1", func(tt *testing.T) {
		inputData[0].Identifiers[0].Identifier = "1a"
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.NotNil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, false, res.Success)

		//clean created data
		cleanDb(tt)
	})

	t.Run("Test Create API With InValid Payload - 2", func(tt *testing.T) {
		inputData[0].Type = "invalidType"
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.NotNil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, false, res.Success)

		//clean created data
		cleanDb(tt)
	})
}

func TestExternalLinkServiceImpl_Update(t *testing.T) {
	t.SkipNow()
	if externalLinkService == nil {
		InitExternalLinkService()
	}

	//update apps to apps
	t.Run("TEST : update link from apps to apps", func(tt *testing.T) {
		outputData := CreateAndGetAppLevelExternalLink(tt)

		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.IsEditable = false

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update it via update API
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//clean data in db
		cleanDb(tt)
	})

	//update app to all apps

	t.Run("TEST : update link from app to all apps", func(tt *testing.T) {

		outputData := CreateAndGetAppLevelExternalLink(tt)
		//change update fields
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)
		createdLink.IsEditable = false

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update it via update API
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//test with admin user
		res, err = externalLinkService.Update(createdLink, externalLink.ADMIN_ROLE)
		assert.NotNil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, false, res.Success)
		//clean data in db
		cleanDb(tt)
	})

	//update 1app to 1cluster
	t.Run("TEST : update link from app to cluster", func(tt *testing.T) {
		outputData := CreateAndGetAppLevelExternalLink(tt)

		//change fields to update
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "clusterLevel"
		createdLink.Identifiers = []externalLink.LinkIdentifier{{
			Type:       "cluster",
			Identifier: "1",
			ClusterId:  1,
		}}

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, 1, len(outputDataAfterUpdate[0].Identifiers))
		assert.Equal(tt, expectedResultLink.Identifiers[0].Type, outputDataAfterUpdate[0].Identifiers[0].Type)
		assert.Equal(tt, expectedResultLink.Identifiers[0].ClusterId, outputDataAfterUpdate[0].Identifiers[0].ClusterId)

		//clean data in db
		cleanDb(tt)
	})

	//update 1app to all cluster
	t.Run("TEST : update link from app to all cluster", func(tt *testing.T) {
		outputData := CreateAndGetAppLevelExternalLink(tt)

		//change fields to update
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "clusterLevel"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//clean data in db
		cleanDb(tt)
	})

	//update 1cluster to 1 app
	t.Run("TEST : update link from cluster to app", func(tt *testing.T) {
		outputData := CreateAndGetClusterLevelExternalLink(tt)

		//change fields to be updated
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "appLevel"
		createdLink.Identifiers = []externalLink.LinkIdentifier{{
			Type:       "devtron-app",
			Identifier: "1",
		}}

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, 1, len(outputDataAfterUpdate[0].Identifiers))
		assert.Equal(tt, expectedResultLink.Identifiers[0].Type, outputDataAfterUpdate[0].Identifiers[0].Type)
		assert.Equal(tt, len(expectedResultLink.Identifiers[0].Identifier), len(outputDataAfterUpdate[0].Identifiers[0].Identifier))

		//clean data in db
		cleanDb(tt)
	})

	//update 1cluster to all cluster
	t.Run("TEST : update link from cluster to all clusters", func(tt *testing.T) {
		outputData := CreateAndGetClusterLevelExternalLink(tt)

		//change fields to be updated
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "clusterLevel"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))
		cleanDb(tt)
	})

	//update 1cluster to all apps
	t.Run("TEST : update link from cluster to all apps", func(tt *testing.T) {
		outputData := CreateAndGetClusterLevelExternalLink(tt)

		//change fields to be updated
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "appLevel"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err := externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//clean data in db
		cleanDb(tt)
	})

	//all apps to all cluster
	t.Run("TEST : update link from all app to all clusters", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "appLevel",
			Identifiers:      make([]externalLink.LinkIdentifier, 0),
			Url:              "http://integration-test.com",
			IsEditable:       true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)

		//get created data
		outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.Equal(tt, 1, len(outputData))

		//change fields to be updated
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "clusterLevel"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err = externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//clean data in db
		cleanDb(tt)
	})

	//all cluster to all apps
	t.Run("TEST : update link from all clusters to all apps", func(tt *testing.T) {
		inputData := make([]*externalLink.ExternalLinkDto, 0)
		inp1 := externalLink.ExternalLinkDto{
			MonitoringToolId: 4,
			Name:             "IntegrationTest-1",
			Description:      "integration test link description",
			Type:             "clusterLevel",
			Identifiers:      make([]externalLink.LinkIdentifier, 0),
			Url:              "http://integration-test.com",
			IsEditable:       true,
		}
		inputData = append(inputData, &inp1)
		res, err := externalLinkService.Create(inputData, 1, externalLink.ADMIN_ROLE)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, true, res.Success)

		//get created data
		outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.Equal(tt, 1, len(outputData))

		//change fields to be updated
		createdLink := outputData[0]
		createdLink.Name = "IntegrationTest-1-update"
		createdLink.Type = "appLevel"
		createdLink.Identifiers = make([]externalLink.LinkIdentifier, 0)

		var expectedResultLink externalLink.ExternalLinkDto
		Copy(&expectedResultLink, createdLink)

		//update the link
		res, err = externalLinkService.Update(createdLink, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//test if it's updated properly
		outputDataAfterUpdate, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, outputDataAfterUpdate)
		assert.Equal(tt, 1, len(outputDataAfterUpdate))
		assert.Equal(tt, expectedResultLink.Id, outputDataAfterUpdate[0].Id)
		assert.Equal(tt, expectedResultLink.Name, outputDataAfterUpdate[0].Name)
		assert.Equal(tt, expectedResultLink.Type, outputDataAfterUpdate[0].Type)
		assert.Equal(tt, expectedResultLink.Description, outputDataAfterUpdate[0].Description)
		assert.Equal(tt, expectedResultLink.Url, outputDataAfterUpdate[0].Url)
		assert.Equal(tt, expectedResultLink.MonitoringToolId, outputDataAfterUpdate[0].MonitoringToolId)
		assert.Equal(tt, len(expectedResultLink.Identifiers), len(outputDataAfterUpdate[0].Identifiers))

		//clean data in db
		cleanDb(tt)
	})
}

func TestExternalLinkServiceImpl_Delete(t *testing.T) {
	t.SkipNow()
	if externalLinkService == nil {
		InitExternalLinkService()
	}

	t.Run("Test To Delete app level links", func(tt *testing.T) {
		outputData := CreateAndGetAppLevelExternalLink(tt)
		//delete the created link
		res, err := externalLinkService.DeleteLink(outputData[0].Id, 1, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//get links and check we get 0 links
		res1, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.Equal(tt, 0, len(res1))

		//clean created data
		cleanDb(tt)
	})

	t.Run("Test To Delete cluster level links", func(tt *testing.T) {
		outputData := CreateAndGetClusterLevelExternalLink(tt)

		//delete the created link
		res, err := externalLinkService.DeleteLink(outputData[0].Id, 1, externalLink.SUPER_ADMIN_ROLE)
		assert.Nil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, true, res.Success)

		//get links and check we get 0 links
		res1, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		assert.Equal(tt, 0, len(res1))

		//clean created data
		cleanDb(tt)

		//try to delete non-editable link with admin
		outputData = CreateAndGetClusterLevelExternalLink(tt)
		res, err = externalLinkService.DeleteLink(outputData[0].Id, 1, externalLink.ADMIN_ROLE)
		assert.NotNil(tt, err)
		assert.NotNil(tt, res)
		assert.Equal(tt, false, res.Success)

		//clean created data
		cleanDb(tt)
	})

}

func TestExternalLinkServiceImpl_FetchAllActiveLinksByLinkIdentifier(t *testing.T) {
	t.SkipNow()
	if externalLinkService == nil {
		InitExternalLinkService()
	}
	expectedLinksForDevtronAppMap, expectedLinksForDevtronInstalledAppMap, expectedLinksForExternalHelmAppMap := CreateTestDataInDbToTestFetchAllActiveLinksByLinkIdentifier()

	//test the links of devtron-app with expected devtron-app links
	t.Run("Test Links for devtron-app", func(tt *testing.T) {
		linkIdentifier := externalLink.LinkIdentifier{
			Type:       "devtron-app",
			Identifier: "1",
		}
		resultDevtronAppLinks, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(&linkIdentifier, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, resultDevtronAppLinks)

		resultDevtronAppLinksMap := map[string]externalLink.ExternalLinkDto{}
		for _, resultLink := range resultDevtronAppLinks {
			resultDevtronAppLinksMap[resultLink.Name] = *resultLink
		}

		//test the result links with expected links
		for key, expectedValue := range expectedLinksForDevtronAppMap {
			val, ok := resultDevtronAppLinksMap[key]
			assert.Equal(tt, true, ok)
			assert.NotNil(tt, val)
			assert.Equal(tt, expectedValue.MonitoringToolId, val.MonitoringToolId)
			assert.Equal(tt, expectedValue.Name, val.Name)
			assert.Equal(tt, expectedValue.Description, val.Description)
			assert.Equal(tt, expectedValue.Url, val.Url)
			assert.Equal(tt, len(expectedValue.Identifiers), len(val.Identifiers))
			for i, ExpIdf := range expectedValue.Identifiers {
				assert.Equal(tt, ExpIdf.Type, val.Identifiers[i].Type)
			}
		}
	})

	//test the links of devtron-installed-app with expected devtron-installed-app links
	t.Run("Test Links for devtron-installed-app", func(tt *testing.T) {
		linkIdentifier := externalLink.LinkIdentifier{
			Type:       "devtron-installed-app",
			Identifier: "12",
		}
		resultDevtronInstalledAppLinks, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(&linkIdentifier, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, resultDevtronInstalledAppLinks)

		resultDevtronInstalledAppLinksMap := map[string]externalLink.ExternalLinkDto{}
		for _, resultLink := range resultDevtronInstalledAppLinks {
			resultDevtronInstalledAppLinksMap[resultLink.Name] = *resultLink
		}

		//test the result links with expected links
		for key, expectedValue := range expectedLinksForDevtronInstalledAppMap {
			val, ok := resultDevtronInstalledAppLinksMap[key]
			assert.Equal(tt, true, ok)
			assert.NotNil(tt, val)
			assert.Equal(tt, expectedValue.MonitoringToolId, val.MonitoringToolId)
			assert.Equal(tt, expectedValue.Name, val.Name)
			assert.Equal(tt, expectedValue.Description, val.Description)
			assert.Equal(tt, expectedValue.Url, val.Url)
			assert.Equal(tt, len(expectedValue.Identifiers), len(val.Identifiers))
			for i, ExpIdf := range expectedValue.Identifiers {
				assert.Equal(tt, ExpIdf.Type, val.Identifiers[i].Type)
			}
		}
	})

	//test the links of external-helm-app with expected external-helm-app links
	t.Run("Test Links for external-helm-app", func(tt *testing.T) {
		linkIdentifier := externalLink.LinkIdentifier{
			Type:       "external-helm-app",
			Identifier: "helm-app-test",
		}
		resultExternalHelmAppLinks, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(&linkIdentifier, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, resultExternalHelmAppLinks)

		resultExternalHelmAppLinksMap := map[string]externalLink.ExternalLinkDto{}
		for _, resultLink := range resultExternalHelmAppLinks {
			resultExternalHelmAppLinksMap[resultLink.Name] = *resultLink
		}

		//test the result links with expected links
		for key, expectedValue := range expectedLinksForExternalHelmAppMap {
			val, ok := resultExternalHelmAppLinksMap[key]
			assert.Equal(tt, true, ok)
			assert.NotNil(tt, val)
			assert.Equal(tt, expectedValue.MonitoringToolId, val.MonitoringToolId)
			assert.Equal(tt, expectedValue.Name, val.Name)
			assert.Equal(tt, expectedValue.Description, val.Description)
			assert.Equal(tt, expectedValue.Url, val.Url)
			assert.Equal(tt, len(expectedValue.Identifiers), len(val.Identifiers))
			for i, ExpIdf := range expectedValue.Identifiers {
				assert.Equal(tt, ExpIdf.Type, val.Identifiers[i].Type)
			}
		}
	})
	cleanDb(t)
}

func TestExternalLinkMonitoringToolRepository_FindAllActive(t *testing.T) {
	t.SkipNow()
	//Get Test data
	query := "select * from external_link_monitoring_tool where active = true"
	var expectedTools []externalLink.ExternalLinkMonitoringTool
	DB, err := getDbConn()
	if err != nil {
		log.Fatalf("error in getting db connection, err : %s", err)
	}
	_, err = DB.Query(&expectedTools, query)
	if err != nil {
		log.Fatalf("error in getting monitoring tools from data base, err : %s", err)
	}
	expectedToolsMap := map[int]externalLink.ExternalLinkMonitoringTool{}
	for _, tool := range expectedTools {
		expectedToolsMap[tool.Id] = tool
	}

	//Get Data from test service
	repoToBeTested := externalLink.NewExternalLinkMonitoringToolRepositoryImpl(DB)
	returnedTools, err := repoToBeTested.FindAllActive()
	assert.Nil(t, err)
	assert.NotNil(t, returnedTools)
	assert.Equal(t, len(expectedTools), len(returnedTools))

	returnedToolsMap := map[int]externalLink.ExternalLinkMonitoringTool{}
	for _, tool := range returnedTools {
		returnedToolsMap[tool.Id] = tool
	}

	//test the returned-links with expected-links
	for key, expectedVal := range expectedToolsMap {
		val, ok := returnedToolsMap[key]
		assert.Equal(t, true, ok)
		assert.Equal(t, expectedVal, val)
	}
}

func CreateTestDataInDbToTestFetchAllActiveLinksByLinkIdentifier() (a, b, c map[string]externalLink.ExternalLinkDto) {
	linksForDevtronAppMap := map[string]externalLink.ExternalLinkDto{}
	linksForDevtronInstalledAppMap := map[string]externalLink.ExternalLinkDto{}
	linksForExternalHelmAppMap := map[string]externalLink.ExternalLinkDto{}
	inputData := make([]*externalLink.ExternalLinkDto, 0)
	itr := 0
	//link for only one cluster
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "cluster-link-1",
		Description:      "integration test link description",
		Type:             "clusterLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "cluster", Identifier: "1", ClusterId: 1},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	})
	itr++
	//link for another cluster
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "cluster-link-2",
		Description:      "integration test link description",
		Type:             "clusterLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "cluster", Identifier: "2", ClusterId: 2},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	})
	itr++

	//create all-cluster link
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "all-cluster-link",
		Description:      "integration test link description",
		Type:             "clusterLevel",
		Url:              "http://integration-test.com",
		IsEditable:       true,
	})
	linksForDevtronAppMap[inputData[itr].Name] = *inputData[itr]
	linksForDevtronInstalledAppMap[inputData[itr].Name] = *inputData[itr]
	linksForExternalHelmAppMap[inputData[itr].Name] = *inputData[itr]
	itr++

	//create all-app link
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "all-app-link",
		Description:      "integration test link description",
		Type:             "appLevel",
		Url:              "http://integration-test.com",
		IsEditable:       true,
	})
	linksForDevtronAppMap[inputData[itr].Name] = *inputData[itr]
	linksForDevtronInstalledAppMap[inputData[itr].Name] = *inputData[itr]
	linksForExternalHelmAppMap[inputData[itr].Name] = *inputData[itr]
	itr++

	//create for devtron-app
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "devtron-app-link",
		Description:      "integration test link description",
		Type:             "appLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "devtron-app", Identifier: "1", ClusterId: 0},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	})
	linksForDevtronAppMap[inputData[itr].Name] = *inputData[itr]
	itr++

	//create for devtron-installed-app
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "devtron-installed-app-link",
		Description:      "integration test link description",
		Type:             "appLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "devtron-installed-app", Identifier: "12", ClusterId: 0},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	})
	linksForDevtronInstalledAppMap[inputData[itr].Name] = *inputData[itr]
	itr++

	//create for external-helm-app
	inputData = append(inputData, &externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "external-helm-app-link",
		Description:      "integration test link description",
		Type:             "appLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "external-helm-app", Identifier: "helm-app-test", ClusterId: 0},
		},
		Url:        "http://integration-test.com",
		IsEditable: true,
	})
	linksForExternalHelmAppMap[inputData[itr].Name] = *inputData[itr]

	//create the above links in db using create service
	_, err := externalLinkService.Create(inputData, 1, externalLink.SUPER_ADMIN_ROLE)
	if err != nil {
		log.Fatalf("error occured while creating external links payload, err : %s", err)
	}
	return linksForDevtronAppMap, linksForDevtronInstalledAppMap, linksForExternalHelmAppMap
}

func CreateAndGetClusterLevelExternalLink(tt *testing.T) []*externalLink.ExternalLinkDto {
	inputData := make([]*externalLink.ExternalLinkDto, 0)
	inp1 := externalLink.ExternalLinkDto{
		MonitoringToolId: 4,
		Name:             "IntegrationTest-1",
		Description:      "integration test link description",
		Type:             "clusterLevel",
		Identifiers: []externalLink.LinkIdentifier{
			{Type: "cluster", Identifier: "1", ClusterId: 1},
			{Type: "cluster", Identifier: "2", ClusterId: 2},
		},
		Url:        "http://integration-test.com",
		IsEditable: false,
	}
	inputData = append(inputData, &inp1)

	res, err := externalLinkService.Create(inputData, 1, externalLink.SUPER_ADMIN_ROLE)
	assert.Nil(tt, err)
	assert.NotNil(tt, res)
	assert.Equal(tt, true, res.Success)

	outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
	assert.Nil(tt, err)
	assert.Equal(tt, 1, len(outputData))
	return outputData
}

func CreateAndGetAppLevelExternalLink(tt *testing.T) []*externalLink.ExternalLinkDto {

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

	res, err := externalLinkService.Create(inputData, 1, externalLink.SUPER_ADMIN_ROLE)
	assert.Nil(tt, err)
	assert.NotNil(tt, res)
	assert.Equal(tt, true, res.Success)

	outputData, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
	assert.Nil(tt, err)
	assert.Equal(tt, 1, len(outputData))
	return outputData
}

func Copy(to *externalLink.ExternalLinkDto, from *externalLink.ExternalLinkDto) {
	to.Type = from.Type
	to.Id = from.Id
	to.MonitoringToolId = from.MonitoringToolId
	to.Name = from.Name
	to.Url = from.Url
	to.IsEditable = from.IsEditable
	to.Description = from.Description
	to.Identifiers = from.Identifiers
}
func cleanDb(tt *testing.T) {
	DB, _ := getDbConn()
	query := "truncate external_link;"
	_, err := DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
	}
	query = "truncate external_link_identifier_mapping;"
	_, err = DB.Exec(query)
	assert.Nil(tt, err)
	if err != nil {
		return
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
