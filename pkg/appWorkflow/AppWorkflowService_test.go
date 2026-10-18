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

package appWorkflow

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	"go.uber.org/zap"
)

// Only the two methods FindAllAppWorkflowMapping calls are implemented; the
// embedded interfaces satisfy the rest and panic if anything else is reached.
type stubAppWorkflowRepository struct {
	appWorkflow.AppWorkflowRepository
	mappings []*appWorkflow.AppWorkflowMapping
}

func (s stubAppWorkflowRepository) FindByWorkflowIds(_ []int) ([]*appWorkflow.AppWorkflowMapping, error) {
	return s.mappings, nil
}

type stubPipelineRepository struct {
	pipelineConfig.PipelineRepository
	deleteRequested map[int]bool
}

func (s stubPipelineRepository) FilterDeploymentDeleteRequestedPipelineIds(_ []int) (map[int]bool, error) {
	return s.deleteRequested, nil
}

func TestFindAllAppWorkflowMappingDeleteRequest(t *testing.T) {
	const workflowId = 1

	mappings := []*appWorkflow.AppWorkflowMapping{
		{Id: 1, ComponentId: 100, AppWorkflowId: workflowId, Type: bean4.CI_PIPELINE_TYPE},
		{Id: 2, ComponentId: 200, AppWorkflowId: workflowId, Type: bean4.CD_PIPELINE_TYPE, ParentId: 100, ParentType: bean4.CI_PIPELINE_TYPE},
		{Id: 3, ComponentId: 201, AppWorkflowId: workflowId, Type: bean4.CD_PIPELINE_TYPE, ParentId: 200, ParentType: bean4.CD_PIPELINE_TYPE},
	}

	tests := []struct {
		name            string
		deleteRequested map[int]bool
		want            map[int]bool // component id -> expected DeploymentAppDeleteRequest
	}{
		{
			name:            "no pipeline has a delete request",
			deleteRequested: map[int]bool{},
			want:            map[int]bool{100: false, 200: false, 201: false},
		},
		{
			name:            "one cd pipeline has a delete request",
			deleteRequested: map[int]bool{200: true},
			want:            map[int]bool{100: false, 200: true, 201: false},
		},
		{
			name:            "both cd pipelines have a delete request",
			deleteRequested: map[int]bool{200: true, 201: true},
			want:            map[int]bool{100: false, 200: true, 201: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := AppWorkflowServiceImpl{
				Logger:                zap.NewNop().Sugar(),
				appWorkflowRepository: stubAppWorkflowRepository{mappings: mappings},
				pipelineRepository:    stubPipelineRepository{deleteRequested: tt.deleteRequested},
			}

			result, err := impl.FindAllAppWorkflowMapping([]int{workflowId})
			if err != nil {
				t.Fatalf("FindAllAppWorkflowMapping returned error: %v", err)
			}

			got := map[int]bool{}
			for _, dto := range result[workflowId] {
				got[dto.ComponentId] = dto.DeploymentAppDeleteRequest
			}
			for componentId, want := range tt.want {
				if got[componentId] != want {
					t.Errorf("component %d: DeploymentAppDeleteRequest = %v, want %v",
						componentId, got[componentId], want)
				}
			}

			// IsLast is set on the local before the append in the same loop, so
			// it arriving correctly shows the DTOs really were populated.
			for _, dto := range result[workflowId] {
				if dto.ComponentId == 201 && !dto.IsLast {
					t.Errorf("component 201: IsLast = false, want true")
				}
			}
		})
	}
}
