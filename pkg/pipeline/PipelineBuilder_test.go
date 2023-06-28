package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestPipelineBuilderImpl_validateDeploymentAppType(t *testing.T) {

	t.Run("DeploymentConfigDoesNotExist", func(t *testing.T) {
		attributesRepoMock := mocks.NewAttributesRepository(t)

		impl := PipelineBuilderImpl{
			attributesRepository: attributesRepoMock, // Provide a mock implementation of attributesRepository
		}
		pipeline := &bean.CDPipelineConfigObject{
			EnvironmentId:     123,
			DeploymentAppType: "SomeAppType",
		}

		mockDeploymentConfigConfig := &repository.Attributes{
			Id:       1,
			Key:      "2",
			Value:    "fsfdfz",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError := error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)

		err := impl.validateDeploymentAppType(pipeline)
		fmt.Print(err)
		apiErr, _ := err.(*util.ApiError)
		assert.Equal(t, http.StatusInternalServerError, apiErr.HttpStatusCode)
	})

	t.Run("AllDeploymentConfigTrue", func(t *testing.T) {
		attributesRepoMock := mocks.NewAttributesRepository(t)

		impl := PipelineBuilderImpl{
			attributesRepository: attributesRepoMock, // Provide a mock implementation of attributesRepository
		}
		pipeline := &bean.CDPipelineConfigObject{
			EnvironmentId:     123,
			DeploymentAppType: "SomeAppType",
		}

		mockDeploymentConfig := &repository.Attributes{
			Id:       1,
			Key:      "123",
			Value:    "{\"argo_cd\": true, \"helm\": true}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError := error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfig, mockError)

		err := impl.validateDeploymentAppType(pipeline)

		assert.Nil(t, err)
	})

	t.Run("ValidDeploymentConfigReceived", func(t *testing.T) {
		attributesRepoMock := mocks.NewAttributesRepository(t)

		impl := PipelineBuilderImpl{
			attributesRepository: attributesRepoMock, // Provide a mock implementation of attributesRepository
		}
		pipeline := &bean.CDPipelineConfigObject{
			EnvironmentId:     123,
			DeploymentAppType: "helm",
		}

		mockDeploymentConfigConfig := &repository.Attributes{
			Id:       1,
			Key:      "123",
			Value:    "{\"argo_cd\": false, \"helm\": true}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError := error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)

		err := impl.validateDeploymentAppType(pipeline)

		assert.Nil(t, err)
	})

	t.Run("InvalidDeploymentConfigReceived", func(t *testing.T) {
		attributesRepoMock := mocks.NewAttributesRepository(t)

		impl := PipelineBuilderImpl{
			attributesRepository: attributesRepoMock, // Provide a mock implementation of attributesRepository
		}
		pipeline := &bean.CDPipelineConfigObject{
			EnvironmentId:     123,
			DeploymentAppType: "SomeOtherAppType",
		}

		mockDeploymentConfigConfig := &repository.Attributes{
			Id:       1,
			Key:      "123",
			Value:    "{\"argo_cd\": false, \"helm\": true}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}

		mockError := error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)

		err := impl.validateDeploymentAppType(pipeline)
		apiErr, _ := err.(*util.ApiError)
		assert.Equal(t, http.StatusBadRequest, apiErr.HttpStatusCode)
	})
}
