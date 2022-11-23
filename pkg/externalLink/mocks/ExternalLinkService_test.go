package mocks

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/go-pg/pg"
	//"github.com/go-pg/pg/mocks"
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
	t.SkipNow()
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
	//tx := Tx{}
	//dbMocked := mocks.DB{}
	//dbMocked.On("Begin").Return(&tx, nil)
	//externalLinkRepositoryMocked.On("GetConnection").Return()
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
	t.SkipNow()
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
	//tx := Tx{}
	//dbMocked := mocks.DB{}
	//dbMocked.On("Begin").Return(&tx, nil)
	//externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
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
		Id:               2,
		Name:             "name2",
		Url:              "test-url2",
		IsEditable:       true,
		MonitoringToolId: 1,
		Identifiers: []externalLink.LinkIdentifier{
			{
				Type:       "external-helm-app",
				Identifier: "ext-helm-1",
			},
		},
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               4,
		Name:             "name4",
		Url:              "test-url4",
		IsEditable:       false,
		MonitoringToolId: 1,
	})
	expectedResultLinks = append(expectedResultLinks, externalLink.ExternalLinkDto{
		Id:               3,
		Name:             "name3",
		Url:              "test-url3",
		IsEditable:       true,
		MonitoringToolId: 1,
	})

	t.Run("Test GetLinks Global Links", func(tt *testing.T) {
		externalLinkRepositoryMocked := NewExternalLinkRepository(tt)
		externalLinkIdentifierMappingRepositoryMocked := NewExternalLinkIdentifierMappingRepository(tt)
		externalLinkMonitoringToolRepository := NewExternalLinkMonitoringToolRepository(tt)

		externalLinkService := externalLink.NewExternalLinkServiceImpl(logger, externalLinkMonitoringToolRepository, externalLinkIdentifierMappingRepositoryMocked, externalLinkRepositoryMocked)
		externalLinkIdentifierMappingRepositoryMocked.On("FindAllActiveLinkIdentifierData").Return(mockLinks, nil)
		externalLinkRepositoryMocked.On("FindAllClusterLinks").Return(mockGlobLinks, nil)

		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 2)
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
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 2)
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
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, 0, externalLink.SUPER_ADMIN_ROLE, 2)
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
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0, externalLink.ADMIN_ROLE, 2)
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
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0, externalLink.ADMIN_ROLE, 2)
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
		testResult, err := externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifierInput, 0, externalLink.ADMIN_ROLE, 2)
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
		expectedErr := &util.ApiError{
			InternalMessage: "external-link-identifier-mapping failed to getting tools ",
			UserMessage:     "external-link-identifier-mapping failed to getting tools ",
		}
		externalLinkMonitoringToolRepository.On("FindAllActive").Return(nil, fmt.Errorf("err"))
		mockToolDtosResponse, err := externalLinkService.GetAllActiveTools()
		assert.NotNil(tt, err)
		assert.Nil(tt, mockToolDtosResponse)
		assert.Equal(tt, expectedErr, err)
	})
}

func TestExternalLinkServiceImpl_Update(t *testing.T) {
	t.SkipNow()
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
	//tx := Tx{}
	//dbMocked := mocks.DB{}
	//dbMocked.On("Begin").Return(&tx, nil)
	//externalLinkRepositoryMocked.On("GetConnection").Return(&dbMocked)
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
