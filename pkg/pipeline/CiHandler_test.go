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
