/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package environment

import (
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGetEnvironmentListForAutocomplete(t *testing.T) {

	t.Run("FilteredAppDeploymentConfigReturned", func(t *testing.T) {
		environmentRepoMock := mocks.NewEnvironmentRepository(t)
		attributesRepoMock := mocks2.NewAttributesRepository(t)
		impl := EnvironmentServiceImpl{
			environmentRepository: environmentRepoMock,
			attributesRepository:  attributesRepoMock,
		}

		mockError := error(nil)
		mockModels := []*repository3.Environment{}
		mockModel := repository3.Environment{
			Id: 1,
			Cluster: &repository.Cluster{
				ClusterName: "demo",
				CdArgoSetup: true,
			},
		}
		mockModels = append(mockModels, &mockModel)
		environmentRepoMock.On("FindAllActive").Return(mockModels, mockError)

		mockDeploymentConfigConfig := &repository2.Attributes{
			Id:       1,
			Key:      "1",
			Value:    "{\"argo_cd\": false, \"helm\": true}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError = error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)
		environmentList, _ := impl.GetEnvironmentListForAutocomplete(true)
		assert.Equal(t, []string{"helm"}, environmentList[0].AllowedDeploymentTypes)
	})

	t.Run("InvalidDeploymentConfigPresentInAttributesTable", func(t *testing.T) {
		environmentRepoMock := mocks.NewEnvironmentRepository(t)
		attributesRepoMock := mocks2.NewAttributesRepository(t)
		impl := EnvironmentServiceImpl{
			environmentRepository: environmentRepoMock,
			attributesRepository:  attributesRepoMock,
		}

		mockError := error(nil)
		mockModels := []*repository3.Environment{}
		mockModel := repository3.Environment{
			Id: 1,
			Cluster: &repository.Cluster{
				ClusterName: "demo",
				CdArgoSetup: true,
			},
		}
		mockModels = append(mockModels, &mockModel)
		environmentRepoMock.On("FindAllActive").Return(mockModels, mockError)

		mockDeploymentConfigConfig := &repository2.Attributes{
			Id:       1,
			Key:      "1",
			Value:    "{\"bdsjd\": false, \"dnjsds\": true}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError = error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)
		environmentList, _ := impl.GetEnvironmentListForAutocomplete(true)
		assert.Equal(t, []string{"helm", "argo_cd"}, environmentList[0].AllowedDeploymentTypes)
	})

	t.Run("BothValidDeploymentConfigIsFalse", func(t *testing.T) {
		environmentRepoMock := mocks.NewEnvironmentRepository(t)
		attributesRepoMock := mocks2.NewAttributesRepository(t)
		impl := EnvironmentServiceImpl{
			environmentRepository: environmentRepoMock,
			attributesRepository:  attributesRepoMock,
		}

		mockError := error(nil)
		mockModels := []*repository3.Environment{}
		mockModel := repository3.Environment{
			Id: 1,
			Cluster: &repository.Cluster{
				ClusterName: "demo",
				CdArgoSetup: true,
			},
		}
		mockModels = append(mockModels, &mockModel)
		environmentRepoMock.On("FindAllActive").Return(mockModels, mockError)

		mockDeploymentConfigConfig := &repository2.Attributes{
			Id:       1,
			Key:      "1",
			Value:    "{\"helm\": false, \"argo_cd\": false}",
			Active:   false,
			AuditLog: sql.AuditLog{},
		}
		mockError = error(nil)
		attributesRepoMock.On("FindByKey", mock.Anything).Return(mockDeploymentConfigConfig, mockError)
		environmentList, _ := impl.GetEnvironmentListForAutocomplete(true)
		assert.Equal(t, []string{"helm", "argo_cd"}, environmentList[0].AllowedDeploymentTypes)
	})

	t.Run("IsDeploymentParamSetToFalse", func(t *testing.T) {
		environmentRepoMock := mocks.NewEnvironmentRepository(t)
		impl := EnvironmentServiceImpl{
			environmentRepository: environmentRepoMock,
		}

		mockError := error(nil)
		mockModels := []*repository3.Environment{}
		mockModel := repository3.Environment{
			Id:   1,
			Name: "demo-devtron",
			Cluster: &repository.Cluster{
				ClusterName: "demo",
				CdArgoSetup: true,
			},
		}
		mockModels = append(mockModels, &mockModel)
		environmentRepoMock.On("FindAllActive").Return(mockModels, mockError)

		environmentList, _ := impl.GetEnvironmentListForAutocomplete(false)
		assert.Equal(t, "demo-devtron", environmentList[0].Environment)
	})
}
