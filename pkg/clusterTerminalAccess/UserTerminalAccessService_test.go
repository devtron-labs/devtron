package clusterTerminalAccess

import (
	"errors"
	"github.com/devtron-labs/devtron/client/k8s/application"
	mocks4 "github.com/devtron-labs/devtron/client/k8s/application/mocks"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/terminal"
	mocks2 "github.com/devtron-labs/devtron/pkg/terminal/mocks"
	mocks3 "github.com/devtron-labs/devtron/util/k8s/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestNewUserTerminalAccessService(t *testing.T) {
	//t.SkipNow()
	podJson := "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"${pod_name}\"},\"spec\":{\"serviceAccountName\":\"${pod_name}-sa\",\"nodeSelector\":{\"kubernetes.io/hostname\":\"${node_name}\"},\"containers\":[{\"name\":\"internal-kubectl\",\"image\":\"${base_image}\",\"command\":[\"/bin/bash\",\"-c\",\"--\"],\"args\":[\"while true; do sleep 30; done;\"]}]}}"
	t.Run("CheckMaxSessionLimit", func(t *testing.T) {
		logger, err := util.InitLogger()
		assert.Nil(t, err)
		userTerminalSessionConfig, err := GetTerminalAccessConfig()
		assert.Nil(t, err)
		userTerminalSessionConfig.MaxSessionPerUser = 1
		terminalAccessRepository := mocks.NewTerminalAccessRepository(t)
		terminalSessionHandler := mocks2.NewTerminalSessionHandler(t)
		k8sApplicationService := mocks3.NewK8sApplicationService(t)
		k8sClientService := mocks4.NewK8sClientService(t)
		terminalAccessRepository.On("GetAllRunningUserTerminalData").Return(nil, nil)
		terminalAccessServiceImpl, err := NewUserTerminalAccessServiceImpl(logger, terminalAccessRepository, userTerminalSessionConfig, k8sApplicationService, k8sClientService, terminalSessionHandler)
		assert.Nil(t, err)

		terminalAccessDataId := 1
		var savedTerminalAccessData *models.UserTerminalAccessData
		terminalAccessRepository.On("SaveUserTerminalAccessData", mock.AnythingOfType("*models.UserTerminalAccessData")).
			Return(func(data *models.UserTerminalAccessData) error {
				data.Id = terminalAccessDataId
				savedTerminalAccessData = data
				return nil
			})
		terminalAccessRepository.On("FetchAllTemplates").Return(nil, nil)
		mockedClusterId := 1
		mockedShellName := "bash"
		mockedUserId := int32(1)
		mockedNodeName := "random1"
		request := &models.UserTerminalSessionRequest{UserId: mockedUserId, ClusterId: mockedClusterId, NodeName: mockedNodeName, BaseImage: "random2", ShellName: mockedShellName}
		terminalSessionResponse1, err := terminalAccessServiceImpl.StartTerminalSession(request)
		assert.Nil(t, err)
		terminalAccessId1 := terminalSessionResponse1.TerminalAccessId
		assert.NotZero(t, terminalAccessId1)
		assert.Equal(t, terminalSessionResponse1.UserId, request.UserId)
		podTemplate := &models.TerminalAccessTemplates{TemplateData: podJson}
		podStatus := "Running"
		k8sApplicationService.On("GetResource", mock.AnythingOfType("*k8s.ResourceRequestBean")).Return(&application.ManifestResponse{Manifest: unstructured.Unstructured{Object: map[string]interface{}{"status": map[string]interface{}{"phase": podStatus}}}}, nil)
		terminalAccessRepository.On("FetchTerminalAccessTemplate", models.TerminalAccessPodTemplateName).Return(podTemplate, nil)
		terminalAccessRepository.On("GetUserTerminalAccessData", terminalAccessId1).Return(savedTerminalAccessData, nil)
		terminalAccessRepository.On("UpdateUserTerminalStatus", mock.AnythingOfType("int"), mock.AnythingOfType("string")).
			Return(func(id int, status string) error {
				assert.Equal(t, terminalAccessDataId, id)
				assert.Equal(t, podStatus, status)
				return nil
			})
		terminalSessionHandler.On("ValidateSession", "").Return(false)
		randomSessionId := "randomSessionId"
		terminalSessionHandler.On("GetTerminalSession", mock.Anything).
			Return(200, func(req *terminal.TerminalSessionRequest) *terminal.TerminalMessage {
				assert.Equal(t, mockedClusterId, req.ClusterId)
				assert.Equal(t, mockedShellName, req.Shell)
				assert.Equal(t, terminalSessionResponse1.PodName, req.PodName)
				terminalMsg := &terminal.TerminalMessage{SessionID: randomSessionId}
				return terminalMsg
			}, nil)
		terminalSessionStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(terminalAccessId1)
		assert.Nil(t, err)
		assert.Equal(t, podStatus, string(terminalSessionStatus.Status))
		assert.Equal(t, randomSessionId, terminalSessionStatus.UserTerminalSessionId)
		terminalSessionResponse2, err := terminalAccessServiceImpl.StartTerminalSession(request)
		assert.Equal(t, errors.New(models.MaxSessionLimitReachedMsg), err)
		assert.Nil(t, terminalSessionResponse2)
	})
}
