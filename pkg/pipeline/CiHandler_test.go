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

package pipeline

import (
	"testing"
)

func TestCiHanlder(t *testing.T) {

	//TODO - fix it
	/*
		t.Run("UpdateCiWorkflowStatusFailure", func(t *testing.T) {
			sugaredLogger, _ := util.NewSugardLogger()
			//assert.True(t, err == nil, err)
			ciWorkflowRepositoryMocked := mocks2.NewCiWorkflowRepository(t)
			var ciWorkflows []*pipelineConfig.CiWorkflow
			dbEntity := &pipelineConfig.CiWorkflow{
				Id:           1,
				Name:         "test-wf-1",
				Status:       Running,
				PodStatus:    Running,
				StartedOn:    time.Now(),
				CiPipelineId: 0,
				PodName:      "test-pod-1",
			}
			ciWorkflows = append(ciWorkflows, dbEntity)
			ciWorkflowRepositoryMocked.On("FindByStatusesIn", Running).Return(ciWorkflows, nil)
			ciHanlder := NewCiHandlerImpl(sugaredLogger, nil, nil, nil, ciWorkflowRepositoryMocked, nil, nil,
				nil, nil, nil, nil, nil, nil, nil, nil)
			_ = ciHanlder.UpdateCiWorkflowStatusFailure(15)
			//assert.Nil(t, err)
		})*/
}
