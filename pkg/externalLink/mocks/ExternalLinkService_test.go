package mocks

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Tx struct {
}

func (tx *Tx) Commit() error {
	return nil
}
func (tx *Tx) Rollback() error {
	return nil
}
func getExternalLinkService(t *testing.T) *externalLink.ExternalLinkServiceImpl {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	return externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
}

func TestExternalLinkServiceImpl_Create(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
	inputRequests := make([]*externalLink.ExternalLinkDto, 0)
	inputRequests = append(inputRequests, &externalLink.ExternalLinkDto{
		Name:        "test1",
		Url:         "https://www.google.com",
		Type:        "clusterLevel",
		Identifiers: nil,
	})
	inputRequests = append(inputRequests, &externalLink.ExternalLinkDto{
		Name:        "test2",
		Url:         "https://www.abc.com",
		Type:        "appLevel",
		Identifiers: nil,
	})
	outputResponse1 := &externalLink.ExternalLinkApiResponse{
		Success: true,
	}
	tx := Tx{}
	dbMocked := mocks.DB{}
	dbMocked.On("Begin").Return(&tx, nil)
	externalLinkRepositoryMocked.On("GetConnection").Return()
	//test1
	externalLinkRepositoryMocked.On("Save", nil).Return(nil)
	externalLinkIdentifierMappingRepositoryMocked.On("Save", nil, &pg.Tx{}).Return("error")
	testResult, err := externalLinkService.Create(inputRequests, 1, "admin")
	assert.Nil(t, err)
	assert.NotNil(t, testResult)
	assert.Equal(t, testResult.Success, outputResponse1.Success)
	inputRequests = append(inputRequests, &externalLink.ExternalLinkDto{
		Name:        "test2",
		Url:         "https://www.abc.com",
		Type:        "appLevel",
		Identifiers: []externalLink.LinkIdentifier{},
	})
	//test2
	externalLinkRepositoryMocked.On("Save", inputRequests[1], nil).Return("error")
	externalLinkIdentifierMappingRepositoryMocked.On("Save", nil).Return(nil)
	testResult, err = externalLinkService.Create(inputRequests, 1, "admin")
	outputResponse2 := &externalLink.ExternalLinkApiResponse{
		Success: false,
	}
	assert.NotNil(t, err)
	assert.Equal(t, "error", err)
	assert.NotNil(t, testResult)
	assert.Equal(t, outputResponse2.Success, testResult.Success)

	inputRequests[1].Identifiers = append(inputRequests[1].Identifiers, externalLink.LinkIdentifier{
		Type:       "devtron-app",
		Identifier: "abc",
	})
	//test3
	externalLinkRepositoryMocked.On("Save", inputRequests[1], nil).Return(nil)
	externalLinkIdentifierMappingRepositoryMocked.On("Save", nil).Return(nil)
	testResult, err = externalLinkService.Create(inputRequests, 1, "admin")
	outputResponse3 := &externalLink.ExternalLinkApiResponse{
		Success: false,
	}
	assert.NotNil(t, outputResponse3)
	assert.NotNil(t, err)
	assert.Equal(t, testResult.Success, outputResponse3.Success)
	inputRequests[1].Identifiers = append(inputRequests[0].Identifiers, externalLink.LinkIdentifier{
		Type:       "devtron-app",
		Identifier: "1",
	})
	//test4
	externalLinkRepositoryMocked.On("Save", inputRequests[1], nil).Return(nil)
	externalLinkIdentifierMappingRepositoryMocked.On("Save", nil).Return(nil)
	testResult, err = externalLinkService.Create(inputRequests, 1, "admin")
	outputResponse4 := &externalLink.ExternalLinkApiResponse{
		Success: false,
	}
	assert.NotNil(t, outputResponse4)
	assert.NotNil(t, err)
	assert.Equal(t, testResult.Success, outputResponse4.Success)

}

func TestExternalLinkServiceImpl_DeleteLink(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
	mockLink := externalLink.ExternalLink{
		Id:         1,
		IsEditable: false,
	}
	tx := Tx{}
	dbMocked := mocks.DB{}
	dbMocked.On("Begin").Return(&tx, nil)
	externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
	mockExternalLinkMappings := make([]externalLink.ExternalLinkIdentifierMapping, 0)
	mockExternalLinkMappings = append(mockExternalLinkMappings, externalLink.ExternalLinkIdentifierMapping{})
	externalLinkRepositoryMocked.On("FindOne", 1).Return(mockLink)
	externalLinkRepositoryMocked.On("Update", nil, nil).Return(nil)
	externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByExternalLinkId", 1).Return(mockExternalLinkMappings)
	externalLinkIdentifierMappingRepositoryMocked.On("Update").Return(nil)
	res, err := externalLinkService.DeleteLink(1, 2, "admin")
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("user not allowed to perform update or delete"))
	assert.NotNil(t, res)
	assert.Equal(t, res.Success, false)
	res, err = externalLinkService.DeleteLink(1, 2, "superAdmin")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.Success, true)
}

func TestExternalLinkServiceImpl_FetchAllActiveLinksByLinkIdentifier(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
	linkIdentifierInput := &externalLink.LinkIdentifier{
		Type:       "external-helm-app",
		Identifier: "ext-helm-1",
	}

	mockLinks := make([]externalLink.ExternalLinkIdentifierMappingData, 0)
	mockLinks = append(mockLinks, externalLink.ExternalLinkIdentifierMappingData{
		Id:                           1,
		ExternalLinkMonitoringToolId: 1,
		Name:                         "name1",
		Url:                          "test-url1",
		IsEditable:                   true,
		MappingId:                    1,
		Type:                         0,
		ClusterId:                    1,
	})
	mockLinks = append(mockLinks, externalLink.ExternalLinkIdentifierMappingData{
		Id:                           1,
		ExternalLinkMonitoringToolId: 1,
		Name:                         "name1",
		Url:                          "test-url1",
		IsEditable:                   true,
		MappingId:                    2,
		Type:                         0,
		ClusterId:                    4,
	})
	mockLinks = append(mockLinks, externalLink.ExternalLinkIdentifierMappingData{
		Id:                           2,
		ExternalLinkMonitoringToolId: 1,
		Name:                         "name2",
		Url:                          "test-url2",
		IsEditable:                   true,
		MappingId:                    3,
		Type:                         3,
		Identifier:                   "ext-helm-1",
	})
	tx := Tx{}
	dbMocked := mocks.DB{}
	dbMocked.On("Begin").Return(&tx, nil)
	externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
	externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveLinkIdentifierData").Return(mockLinks)
	expectedResultLinks := make([]externalLink.ExternalLinkDto, 0)
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               1,
		Name:             "name1",
		Url:              "test-url1",
		IsEditable:       true,
		MonitoringToolId: 1,
		Identifiers: []externalLink.LinkIdentifier{
			{
				Type:      "cluster",
				ClusterId: 1,
			},
			{
				Type:      "cluster",
				ClusterId: 4,
			},
		},
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:         2,
		Name:       "name2",
		Url:        "test-url2",
		IsEditable: true,
		Identifiers: []externalLink.LinkIdentifier{
			{
				Type:       "external-helm-app",
				Identifier: "ext-helm-1",
			},
		},
	})

	testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 2)
	assert.Nil(t, err)
	for i, testLink := range testResult {
		assert.Equal(t, testLink.Id, expectedResultLinks[i].Id)
		assert.Equal(t, testLink.MonitoringToolId, expectedResultLinks[i].MonitoringToolId)
		assert.Equal(t, testLink.Name, expectedResultLinks[i].Name)
		assert.Equal(t, testLink.Url, expectedResultLinks[i].Url)
		assert.Equal(t, testLink.IsEditable, expectedResultLinks[i].IsEditable)
		assert.Equal(t, testLink.Description, expectedResultLinks[i].Description)
		for j, identifier := range expectedResultLinks[i].Identifiers {
			assert.NotNil(t, testLink.Identifiers[j])
			assert.Equal(t, identifier.Type, testLink.Identifiers[j].Type)
			assert.Equal(t, identifier.Identifier, testLink.Identifiers[j].Identifier)
			assert.Equal(t, identifier.AppId, testLink.Identifiers[j].AppId)
			assert.Equal(t, identifier.ClusterId, testLink.Identifiers[j].ClusterId)
			assert.Equal(t, identifier.EnvId, testLink.Identifiers[j].EnvId)
		}
	}

	externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByLinkIdentifier").Return([]externalLink.ExternalLinkIdentifierMappingData{mockLinks[2]})
	testResult, err = externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.ADMIN_ROLE, 2)
	assert.Nil(t, testResult)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("user role is not super_admin"))

	testResult, err = externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0, externalLink.ADMIN_ROLE, 2)
	assert.Nil(t, err)
	assert.NotNil(t, testResult)
	assert.Equal(t, 1, len(testResult))
	assert.Equal(t, testResult[0].Id, expectedResultLinks[1].Id)
	assert.Equal(t, testResult[0].MonitoringToolId, expectedResultLinks[1].MonitoringToolId)
	assert.Equal(t, testResult[0].Name, expectedResultLinks[1].Name)
	assert.Equal(t, testResult[0].Url, expectedResultLinks[1].Url)
	assert.Equal(t, testResult[0].IsEditable, expectedResultLinks[1].IsEditable)
	assert.Equal(t, testResult[0].Description, expectedResultLinks[1].Description)
	for j, identifier := range expectedResultLinks[1].Identifiers {
		assert.NotNil(t, testResult[0].Identifiers[j])
		assert.Equal(t, identifier.Type, testResult[0].Identifiers[j].Type)
		assert.Equal(t, identifier.Identifier, testResult[0].Identifiers[j].Identifier)
		assert.Equal(t, identifier.AppId, testResult[0].Identifiers[j].AppId)
		assert.Equal(t, identifier.ClusterId, testResult[0].Identifiers[j].ClusterId)
		assert.Equal(t, identifier.EnvId, testResult[0].Identifiers[j].EnvId)
	}

}

func TestExternalLinkServiceImpl_GetAllActiveTools(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
	mockTools := make([]externalLink.ExternalLinkMonitoringTool, 0)
	mockToolsExpectedResult := make([]externalLink.ExternalLinkMonitoringToolDto, 0)
	mockToolsExpectedResult = append(mockToolsExpectedResult, externalLink.ExternalLinkMonitoringToolDto{
		Id:       1,
		Name:     "grafana",
		Icon:     "icon1",
		Category: 1,
	})
	mockTools = append(mockTools, externalLink.ExternalLinkMonitoringTool{
		Id:       1,
		Name:     "grafana",
		Icon:     "icon1",
		Category: 1,
		Active:   true,
	})
	mockToolsExpectedResult = append(mockToolsExpectedResult, externalLink.ExternalLinkMonitoringToolDto{
		Id:       2,
		Name:     "kibana",
		Icon:     "icon2",
		Category: 2,
	})
	mockTools = append(mockTools, externalLink.ExternalLinkMonitoringTool{
		Id:       2,
		Name:     "kibana",
		Icon:     "icon2",
		Category: 2,
		Active:   true,
	})
	tx := Tx{}
	dbMocked := mocks.DB{}
	dbMocked.On("Begin").Return(&tx, nil)
	externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
	externalLinkMonitoringToolRepository.On("FindAllActive").Return(mockTools)
	mockToolDtosResponse, err := externalLinkService.GetAllActiveTools()
	assert.Nil(t, err)
	for i, mockToolDto := range mockToolDtosResponse {
		assert.NotNil(t, mockToolDto)
		assert.Equal(t, mockToolDto.Id, mockToolsExpectedResult[i].Id)
		assert.Equal(t, mockToolDto.Name, mockToolsExpectedResult[i].Name)
		assert.Equal(t, mockToolDto.Icon, mockToolsExpectedResult[i].Icon)
		assert.Equal(t, mockToolDto.Category, mockToolsExpectedResult[i].Category)
	}

}

func TestExternalLinkServiceImpl_Update(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	activeMappings := []externalLink.LinkIdentifier{
		{
			Type:      "cluster",
			ClusterId: 1,
		},
		{
			Type:      "cluster",
			ClusterId: 4,
		},
	}
	externalLinkDtoInput := externalLink.ExternalLinkDto{
		Id:               1,
		Name:             "name2",
		Url:              "test-url1",
		IsEditable:       true,
		MonitoringToolId: 1,
		Identifiers:      activeMappings,
	}

	externalLinkOutput := externalLink.ExternalLink{
		Id:                           1,
		Name:                         "name1",
		Url:                          "test-url1",
		IsEditable:                   true,
		ExternalLinkMonitoringToolId: 1,
	}
	externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
	tx := Tx{}
	dbMocked := mocks.DB{}
	dbMocked.On("Begin").Return(&tx, nil)
	externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
	externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByExternalLinkId").Return(activeMappings)
	externalLinkIdentifierMappingRepositoryMocked.On("Update").Return(nil)
	externalLinkIdentifierMappingRepositoryMocked.On("Save").Return(nil)
	externalLinkRepositoryMocked.On("FindOne", 1).Return(externalLinkOutput)
	externalLinkRepositoryMocked.On("Update").Return(nil)

	res, err := externalLinkService.Update(&externalLinkDtoInput, externalLink.SUPER_ADMIN_ROLE)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.Success, true)

}
