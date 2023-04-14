package mocks

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/stretchr/testify/assert"
	"testing"
)

var expectedMonitoringToolErr = util.ApiError{
	InternalMessage: "external-link-identifier-mapping failed to getting tools ",
	UserMessage:     "external-link-identifier-mapping failed to getting tools ",
}

func getExternalLinkService(t *testing.T) *externalLink.ExternalLinkServiceImpl {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	externalLinkRepositoryMocked := NewExternalLinkRepository(t)
	externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(t)
	externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(t)

	return externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
}

func TestExternalLinkServiceImpl_FetchAllActiveLinksByLinkIdentifier(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
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

	mockGlobLinks := make([]externalLink.ExternalLink, 0)
	mockGlobLinks = append(mockGlobLinks, externalLink.ExternalLink{
		Id:                           3,
		ExternalLinkMonitoringToolId: 1,
		Name:                         "name3",
		Url:                          "test-url3",
		IsEditable:                   true,
	})
	mockLinks = append(mockLinks, externalLink.ExternalLinkIdentifierMappingData{
		Id:                           4,
		ExternalLinkMonitoringToolId: 1,
		Name:                         "name4",
		Url:                          "test-url4",
		IsEditable:                   false,
		MappingId:                    0,
		Type:                         -1,
		ClusterId:                    0,
	})
	expectedResultLinks := make([]externalLink.ExternalLinkDto, 0)
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               mockLinks[0].Id,
		Name:             mockLinks[0].Name,
		Url:              mockLinks[0].Url,
		IsEditable:       mockLinks[0].IsEditable,
		MonitoringToolId: mockLinks[0].ExternalLinkMonitoringToolId,
		Identifiers: []externalLink.LinkIdentifier{
			{
				Type:      "cluster",
				ClusterId: mockLinks[0].ClusterId,
			},
			{
				Type:      "cluster",
				ClusterId: mockLinks[1].ClusterId,
			},
		},
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               mockLinks[2].Id,
		Name:             mockLinks[2].Name,
		Url:              mockLinks[2].Url,
		IsEditable:       mockLinks[2].IsEditable,
		MonitoringToolId: mockLinks[2].ExternalLinkMonitoringToolId,
		Identifiers: []externalLink.LinkIdentifier{
			{
				Type:       "external-helm-app",
				Identifier: mockLinks[2].Identifier,
			},
		},
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               mockLinks[3].Id,
		Name:             mockLinks[2].Name,
		Url:              mockLinks[2].Url,
		IsEditable:       mockLinks[2].IsEditable,
		MonitoringToolId: mockLinks[2].ExternalLinkMonitoringToolId,
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               mockGlobLinks[0].Id,
		Name:             mockGlobLinks[0].Name,
		Url:              mockGlobLinks[0].Url,
		IsEditable:       mockGlobLinks[0].IsEditable,
		MonitoringToolId: mockGlobLinks[0].ExternalLinkMonitoringToolId,
	})

	t.Run("Test GetLinks Global Links", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)

		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveLinkIdentifierData").Return(mockLinks, nil)
		externalLinkRepositoryMocked.On("FindAllClusterLinks").Return(mockGlobLinks, nil)

		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.Nil(tt, err)
		for i, testLink := range testResult {
			assert.Equal(tt, testLink.Id, expectedResultLinks[i].Id)
			assert.Equal(tt, testLink.MonitoringToolId, expectedResultLinks[i].MonitoringToolId)
			assert.Equal(tt, testLink.Name, expectedResultLinks[i].Name)
			assert.Equal(tt, testLink.Url, expectedResultLinks[i].Url)
			assert.Equal(tt, testLink.IsEditable, expectedResultLinks[i].IsEditable)
			assert.Equal(tt, testLink.Description, expectedResultLinks[i].Description)
			for j, identifier := range expectedResultLinks[i].Identifiers {
				assert.NotNil(tt, testLink.Identifiers[j])
				assert.Equal(tt, identifier.Type, testLink.Identifiers[j].Type)
				assert.Equal(tt, identifier.Identifier, testLink.Identifiers[j].Identifier)
				assert.Equal(tt, identifier.AppId, testLink.Identifiers[j].AppId)
				assert.Equal(tt, identifier.ClusterId, testLink.Identifiers[j].ClusterId)
				assert.Equal(tt, identifier.EnvId, testLink.Identifiers[j].EnvId)
			}
		}
	})

	t.Run("Test GetLinks Global Links Error check-1", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)
		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveLinkIdentifierData").Return(nil, fmt.Errorf("err"))
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.NotNil(tt, err)
		assert.Nil(tt, testResult)
	})

	t.Run("Test GetLinks Global Links Error check-2", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)
		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveLinkIdentifierData").Return(mockLinks, nil)
		externalLinkRepositoryMocked.On("FindAllClusterLinks").Return(nil, fmt.Errorf("err"))
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0)
		assert.NotNil(tt, err)
		assert.Nil(tt, testResult)
	})

	t.Run("Test GetLinks from App configuration", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)

		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByLinkIdentifier", linkIdentifierInput, 0).Return([]externalLink.ExternalLinkIdentifierMappingData{mockLinks[2]}, nil)
		externalLinkRepositoryMocked.On("FindAllClusterLinks").Return(mockGlobLinks, nil)
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0)
		assert.Nil(tt, err)
		assert.NotNil(tt, testResult)
		//assert.Equal(tt, 1, len(testResult))
		assert.Equal(tt, testResult[0].Id, expectedResultLinks[1].Id)
		assert.Equal(tt, testResult[0].MonitoringToolId, expectedResultLinks[1].MonitoringToolId)
		assert.Equal(tt, testResult[0].Name, expectedResultLinks[1].Name)
		assert.Equal(tt, testResult[0].Url, expectedResultLinks[1].Url)
		assert.Equal(tt, testResult[0].IsEditable, expectedResultLinks[1].IsEditable)
		assert.Equal(tt, testResult[0].Description, expectedResultLinks[1].Description)
		for j, identifier := range expectedResultLinks[1].Identifiers {
			assert.NotNil(tt, testResult[0].Identifiers[j])
			assert.Equal(tt, identifier.Type, testResult[0].Identifiers[j].Type)
			assert.Equal(tt, identifier.Identifier, testResult[0].Identifiers[j].Identifier)
			assert.Equal(tt, identifier.AppId, testResult[0].Identifiers[j].AppId)
			assert.Equal(tt, identifier.ClusterId, testResult[0].Identifiers[j].ClusterId)
			assert.Equal(tt, identifier.EnvId, testResult[0].Identifiers[j].EnvId)
		}
	})

	t.Run("Test GetLinks App Configuration Error check-1", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)
		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByLinkIdentifier", linkIdentifierInput, 0).Return(nil, fmt.Errorf("err"))
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0)
		assert.NotNil(tt, err)
		assert.Nil(tt, testResult)
	})

	t.Run("Test GetLinks App Configuration Error check-2", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)
		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveByLinkIdentifier", linkIdentifierInput, 0).Return([]externalLink.ExternalLinkIdentifierMappingData{mockLinks[2]}, nil)
		externalLinkRepositoryMocked.On("FindAllClusterLinks").Return(nil, fmt.Errorf("err"))
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0)
		assert.NotNil(tt, err)
		assert.Nil(tt, testResult)
	})

}

func TestExternalLinkServiceImpl_GetAllActiveTools(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)

	mockTools := make([]externalLink.ExternalLinkMonitoringTool, 0)
	mockToolsExpectedResult := make([]externalLink.ExternalLinkMonitoringToolDto, 0)
	t.Run("TestCase logic check : get monitoring tools", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)

		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
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

		externalLinkMonitoringToolRepository.On("FindAllActive").Return(mockTools, nil)
		mockToolDtosResponse, err := externalLinkService.GetAllActiveTools()
		assert.Nil(tt, err)
		for i, mockToolDto := range mockToolDtosResponse {
			assert.NotNil(tt, mockToolDto)
			assert.Equal(tt, mockToolDto.Id, mockToolsExpectedResult[i].Id)
			assert.Equal(tt, mockToolDto.Name, mockToolsExpectedResult[i].Name)
			assert.Equal(tt, mockToolDto.Icon, mockToolsExpectedResult[i].Icon)
			assert.Equal(tt, mockToolDto.Category, mockToolsExpectedResult[i].Category)
		}
	})
	t.Run("TestCase Error check : get monitoring tools", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)

		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)

		externalLinkMonitoringToolRepository.On("FindAllActive").Return(nil, fmt.Errorf("err"))

		mockToolDtosResponse, err := externalLinkService.GetAllActiveTools()
		assert.NotNil(tt, err)
		assert.Nil(tt, mockToolDtosResponse)
		assert.Equal(tt, &expectedMonitoringToolErr, err)
	})
}
