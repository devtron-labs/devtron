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

package trigger

import (
	"testing"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/stretchr/testify/assert"
)

func TestHandlerServiceImpl_isWorkflowInCriticalPhase(t *testing.T) {
	// Create a handler service instance for testing
	handlerService := &HandlerServiceImpl{}

	t.Run("Starting workflow should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        1,
			Status:    "Starting",
			StartedOn: time.Now(),
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Starting workflow should not be in critical phase")
	})

	t.Run("Pending workflow should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        2,
			Status:    "Pending",
			StartedOn: time.Now(),
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Pending workflow should not be in critical phase")
	})

	t.Run("Recently started Running workflow should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        3,
			Status:    "Running",
			StartedOn: time.Now().Add(-1 * time.Minute), // Started 1 minute ago
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Recently started Running workflow should not be in critical phase")
	})

	t.Run("Long running workflow should be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        4,
			Status:    "Running",
			StartedOn: time.Now().Add(-5 * time.Minute), // Started 5 minutes ago
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.True(t, result, "Long running workflow should be in critical phase")
	})

	t.Run("Running workflow with zero StartedOn should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        5,
			Status:    "Running",
			StartedOn: time.Time{}, // Zero time
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Running workflow with zero StartedOn should not be in critical phase")
	})
}

func TestCiTriggerRequest_HasPipelineId(t *testing.T) {
	t.Run("CiTriggerRequest should contain PipelineId", func(t *testing.T) {
		triggerRequest := &types.CiTriggerRequest{
			PipelineId:  123,
			TriggeredBy: 1,
		}
		assert.Equal(t, 123, triggerRequest.PipelineId, "CiTriggerRequest should have PipelineId field")
		assert.Equal(t, int32(1), triggerRequest.TriggeredBy, "CiTriggerRequest should have TriggeredBy field")
	})
}

func TestCiPipeline_HasAutoAbortField(t *testing.T) {
	t.Run("CiPipeline should have AutoAbortPreviousBuilds field", func(t *testing.T) {
		pipeline := &pipelineConfig.CiPipeline{
			Id:                      1,
			AutoAbortPreviousBuilds: true,
		}
		assert.True(t, pipeline.AutoAbortPreviousBuilds, "CiPipeline should have AutoAbortPreviousBuilds field")
	})
}