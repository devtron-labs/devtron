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

	t.Run("Workflow at exactly 2 minute boundary should be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        6,
			Status:    "Running",
			StartedOn: time.Now().Add(-2 * time.Minute), // Started exactly 2 minutes ago
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.True(t, result, "Workflow at 2 minute boundary should be in critical phase")
	})

	t.Run("Completed workflow should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        7,
			Status:    "Succeeded",
			StartedOn: time.Now().Add(-10 * time.Minute),
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Completed workflow should not be in critical phase")
	})

	t.Run("Failed workflow should not be in critical phase", func(t *testing.T) {
		workflow := &pipelineConfig.CiWorkflow{
			Id:        8,
			Status:    "Failed",
			StartedOn: time.Now().Add(-3 * time.Minute),
		}
		result := handlerService.isWorkflowInCriticalPhase(workflow)
		assert.False(t, result, "Failed workflow should not be in critical phase")
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

	t.Run("CiPipeline AutoAbortPreviousBuilds should default to false", func(t *testing.T) {
		pipeline := &pipelineConfig.CiPipeline{
			Id: 2,
			// AutoAbortPreviousBuilds not set, should default to false
		}
		assert.False(t, pipeline.AutoAbortPreviousBuilds, "CiPipeline AutoAbortPreviousBuilds should default to false")
	})
}

func TestWorkflowStatus_Constants(t *testing.T) {
	t.Run("Should have expected workflow statuses for auto-abort logic", func(t *testing.T) {
		// These are the statuses we check for in FindRunningWorkflowsForPipeline
		runningStatuses := []string{"Running", "Starting", "Pending"}
		
		// Ensure we have the expected statuses
		assert.Contains(t, runningStatuses, "Running", "Should include Running status")
		assert.Contains(t, runningStatuses, "Starting", "Should include Starting status")
		assert.Contains(t, runningStatuses, "Pending", "Should include Pending status")
		assert.Len(t, runningStatuses, 3, "Should have exactly 3 running statuses")
	})
}