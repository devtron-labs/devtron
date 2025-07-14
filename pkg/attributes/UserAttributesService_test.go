/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package attributes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMergeUserAttributesData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &UserAttributesServiceImpl{
		logger: logger.Sugar(),
	}

	t.Run("should merge cluster resources correctly", func(t *testing.T) {
		// Existing data with application resources
		existingData := map[string]interface{}{
			"computedAppTheme": "light",
			"resources": map[string]interface{}{
				"application/devtron-application": map[string]interface{}{
					"recently-visited": []interface{}{
						map[string]interface{}{"appId": "45", "appName": "pk-test-3"},
					},
				},
			},
			"viewPermittedEnvOnly": false,
		}

		// New data with cluster resources
		newData := map[string]interface{}{
			"resources": map[string]interface{}{
				"cluster": map[string]interface{}{
					"recently-visited": []interface{}{
						map[string]interface{}{"id": 1, "name": "default_cluster"},
						map[string]interface{}{"id": 40, "name": "ajay-test"},
					},
				},
			},
		}

		// Merge the data
		anyChanges := service.mergeUserAttributesData(existingData, newData)

		// Verify changes were detected
		assert.True(t, anyChanges)

		// Verify the structure
		resources := existingData["resources"].(map[string]interface{})

		// Check that cluster data was added
		clusterData := resources["cluster"].(map[string]interface{})
		clusterVisited := clusterData["recently-visited"].([]interface{})
		assert.Len(t, clusterVisited, 2)

		// Check that existing application data is preserved
		appData := resources["application/devtron-application"].(map[string]interface{})
		appVisited := appData["recently-visited"].([]interface{})
		assert.Len(t, appVisited, 1)

		// Check that other resource types are initialized
		assert.Contains(t, resources, "job")
		assert.Contains(t, resources, "app-group")

		// Check that non-resources fields are preserved
		assert.Equal(t, "light", existingData["computedAppTheme"])
		assert.Equal(t, false, existingData["viewPermittedEnvOnly"])
	})

	t.Run("should initialize resources structure when it doesn't exist", func(t *testing.T) {
		// Existing data without resources
		existingData := map[string]interface{}{
			"computedAppTheme": "dark",
		}

		// New data with job resources
		newData := map[string]interface{}{
			"resources": map[string]interface{}{
				"job": map[string]interface{}{
					"recently-visited": []interface{}{
						map[string]interface{}{"id": 1, "name": "job1"},
					},
				},
			},
		}

		// Merge the data
		anyChanges := service.mergeUserAttributesData(existingData, newData)

		// Verify changes were detected
		assert.True(t, anyChanges)

		// Verify the structure
		resources := existingData["resources"].(map[string]interface{})

		// Check that job data was added
		jobData := resources["job"].(map[string]interface{})
		jobVisited := jobData["recently-visited"].([]interface{})
		assert.Len(t, jobVisited, 1)

		// Check that all resource types are initialized
		assert.Contains(t, resources, "cluster")
		assert.Contains(t, resources, "app-group")
		assert.Contains(t, resources, "application/devtron-application")
	})

	t.Run("should handle non-resources fields normally", func(t *testing.T) {
		// Existing data
		existingData := map[string]interface{}{
			"computedAppTheme": "light",
		}

		// New data with theme change
		newData := map[string]interface{}{
			"computedAppTheme":     "dark",
			"viewPermittedEnvOnly": true,
		}

		// Merge the data
		anyChanges := service.mergeUserAttributesData(existingData, newData)

		// Verify changes were detected
		assert.True(t, anyChanges)

		// Verify the values
		assert.Equal(t, "dark", existingData["computedAppTheme"])
		assert.Equal(t, true, existingData["viewPermittedEnvOnly"])
	})

	t.Run("should detect no changes when data is identical", func(t *testing.T) {
		// Existing data
		existingData := map[string]interface{}{
			"computedAppTheme": "light",
			"resources": map[string]interface{}{
				"cluster": map[string]interface{}{
					"recently-visited": []interface{}{
						map[string]interface{}{"id": 1, "name": "default_cluster"},
					},
				},
			},
		}

		// Same data
		newData := map[string]interface{}{
			"resources": map[string]interface{}{
				"cluster": map[string]interface{}{
					"recently-visited": []interface{}{
						map[string]interface{}{"id": 1, "name": "default_cluster"},
					},
				},
			},
		}

		// Merge the data
		anyChanges := service.mergeUserAttributesData(existingData, newData)

		// Verify no changes were detected
		assert.False(t, anyChanges)
	})
}

func TestMergeResourcesData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &UserAttributesServiceImpl{
		logger: logger.Sugar(),
	}

	t.Run("should initialize all supported resource types", func(t *testing.T) {
		existingData := map[string]interface{}{}
		newResourcesValue := map[string]interface{}{
			"cluster": map[string]interface{}{
				"recently-visited": []interface{}{},
			},
		}

		anyChanges := service.mergeResourcesData(existingData, newResourcesValue)

		assert.True(t, anyChanges)
		resources := existingData["resources"].(map[string]interface{})

		// Check all supported types are initialized
		supportedTypes := []string{"cluster", "job", "app-group", "application/devtron-application"}
		for _, resourceType := range supportedTypes {
			assert.Contains(t, resources, resourceType)
			resourceData := resources[resourceType].(map[string]interface{})
			assert.Contains(t, resourceData, "recently-visited")
		}
	})
}
