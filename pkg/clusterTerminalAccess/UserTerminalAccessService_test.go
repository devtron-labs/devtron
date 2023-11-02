package clusterTerminalAccess

import (
	"context"
	"errors"
	util2 "github.com/devtron-labs/common-lib-private/utils/k8s"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	mocks3 "github.com/devtron-labs/devtron/pkg/k8s/application/mocks"
	"github.com/devtron-labs/devtron/pkg/terminal"
	mocks2 "github.com/devtron-labs/devtron/pkg/terminal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestNewUserTerminalAccessService(t *testing.T) {
	//t.SkipNow()
	podJson := "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"${pod_name}\"},\"spec\":{\"serviceAccountName\":\"${pod_name}-sa\",\"nodeSelector\":{\"kubernetes.io/hostname\":\"${node_name}\"},\"containers\":[{\"name\":\"internal-kubectl\",\"image\":\"${base_image}\",\"command\":[\"/bin/bash\",\"-c\",\"--\"],\"args\":[\"while true; do sleep 30; done;\"]}]}}"
	t.Run("CheckMaxSessionLimit", func(tt *testing.T) {
		terminalAccessRepository, terminalSessionHandler, k8sApplicationService, terminalAccessServiceImpl := loadUserTerminalAccessService(tt)
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
		terminalSessionResponse1, err := terminalAccessServiceImpl.StartTerminalSession(context.Background(), request)
		assert.Nil(tt, err)
		terminalAccessId1 := terminalSessionResponse1.TerminalAccessId
		assert.NotZero(tt, terminalAccessId1)
		assert.Equal(tt, terminalSessionResponse1.UserId, request.UserId)
		podTemplate := &models.TerminalAccessTemplates{TemplateData: podJson}
		podStatus := "Running"
		k8sApplicationService.On("GetResource", mock.AnythingOfType("*k8s.ResourceRequestBean")).Return(&util2.ManifestResponse{Manifest: unstructured.Unstructured{Object: map[string]interface{}{"status": map[string]interface{}{"phase": podStatus}}}}, nil)
		terminalAccessRepository.On("FetchTerminalAccessTemplate", models.TerminalAccessPodTemplateName).Return(podTemplate, nil)
		terminalAccessRepository.On("GetUserTerminalAccessData", terminalAccessId1).Return(savedTerminalAccessData, nil)
		terminalAccessRepository.On("UpdateUserTerminalStatus", mock.AnythingOfType("int"), mock.AnythingOfType("string")).
			Return(func(id int, status string) error {
				assert.Equal(tt, terminalAccessDataId, id)
				assert.Equal(tt, podStatus, status)
				return nil
			})
		terminalSessionHandler.On("ValidateSession", "").Return(false)
		randomSessionId := "randomSessionId"
		terminalSessionHandler.On("GetTerminalSession", mock.AnythingOfType("*terminal.TerminalSessionRequest")).
			Return(200, func(req *terminal.TerminalSessionRequest) *terminal.TerminalMessage {
				assert.Equal(tt, mockedClusterId, req.ClusterId)
				assert.Equal(tt, mockedShellName, req.Shell)
				assert.Equal(tt, terminalSessionResponse1.PodName, req.PodName)
				terminalMsg := &terminal.TerminalMessage{SessionID: randomSessionId}
				return terminalMsg
			}, nil)
		terminalSessionStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(context.Background(), terminalAccessId1, "default", "", "sh")
		assert.Nil(tt, err)
		assert.Equal(tt, podStatus, string(terminalSessionStatus.Status))
		assert.Equal(tt, randomSessionId, terminalSessionStatus.UserTerminalSessionId)
		terminalSessionResponse2, err := terminalAccessServiceImpl.StartTerminalSession(context.Background(), request)
		assert.Equal(tt, errors.New(models.MaxSessionLimitReachedMsg), err)
		assert.Nil(tt, terminalSessionResponse2)
	})

	t.Run("K8sResourceErrorCase", func(tt *testing.T) {
		terminalAccessRepository, _, k8sApplicationService, terminalAccessServiceImpl := loadUserTerminalAccessService(tt)
		terminalAccessId := 1
		randomUserId := int32(2)
		randomClusterId := 3
		randomPodName := "randomName"
		terminalAccessData := &models.UserTerminalAccessData{
			Status:    string(models.TerminalPodRunning),
			UserId:    randomUserId,
			ClusterId: randomClusterId,
			PodName:   randomPodName,
		}
		terminalAccessRepository.On("GetUserTerminalAccessData", terminalAccessId).Return(terminalAccessData, nil)
		podTemplate := &models.TerminalAccessTemplates{TemplateData: podJson}
		terminalAccessRepository.On("FetchTerminalAccessTemplate", models.TerminalAccessPodTemplateName).Return(podTemplate, nil)
		failedMsg := &k8sErrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonForbidden}}
		k8sApplicationService.On("GetResource", mock.AnythingOfType("*k8s.ResourceRequestBean")).Return(nil, failedMsg)
		terminalSessionStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(context.Background(), terminalAccessId, "default", "", "sh")
		assert.Nil(tt, terminalSessionStatus)
		assert.NotNil(tt, err)
		assert.Equal(tt, failedMsg, err)
	})

	t.Run("DbSaveOperationFailed", func(tt *testing.T) {
		terminalAccessRepository, _, _, terminalAccessServiceImpl := loadUserTerminalAccessService(tt)
		mockedClusterId := 1
		mockedShellName := "bash"
		mockedUserId := int32(1)
		mockedNodeName := "random1"
		queryExecutionErr := errors.New("query execution failed")
		terminalAccessRepository.On("SaveUserTerminalAccessData", mock.AnythingOfType("*models.UserTerminalAccessData")).
			Return(func(data *models.UserTerminalAccessData) error {
				assert.Equal(tt, mockedClusterId, data.ClusterId)
				assert.Equal(tt, mockedUserId, data.UserId)
				assert.Equal(tt, mockedNodeName, data.NodeName)
				assert.NotEmpty(tt, data.Metadata)
				return queryExecutionErr
			})

		request := &models.UserTerminalSessionRequest{UserId: mockedUserId, ClusterId: mockedClusterId, NodeName: mockedNodeName, BaseImage: "random2", ShellName: mockedShellName}
		terminalSessionResponse, err := terminalAccessServiceImpl.StartTerminalSession(context.Background(), request)
		assert.Nil(tt, terminalSessionResponse)
		assert.Equal(tt, queryExecutionErr, err)
	})

	t.Run("WrongPodTemplate", func(tt *testing.T) {
		terminalAccessRepository, _, _, terminalAccessServiceImpl := loadUserTerminalAccessService(tt)
		terminalAccessId := 1
		randomUserId := int32(2)
		randomClusterId := 3
		randomPodName := "randomName"
		terminalAccessData := &models.UserTerminalAccessData{
			Status:    string(models.TerminalPodRunning),
			UserId:    randomUserId,
			ClusterId: randomClusterId,
			PodName:   randomPodName,
		}
		terminalAccessRepository.On("GetUserTerminalAccessData", terminalAccessId).Return(terminalAccessData, nil)
		podTemplate := &models.TerminalAccessTemplates{TemplateData: "wrong-pod-json"}
		terminalAccessRepository.On("FetchTerminalAccessTemplate", models.TerminalAccessPodTemplateName).Return(podTemplate, nil)
		terminalSessionStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(context.Background(), terminalAccessId, "default", "", "sh")
		assert.Nil(tt, terminalSessionStatus)
		assert.NotNil(tt, err)
	})

	t.Run("Pod Manifest : invalid manifest structure Test", func(tt *testing.T) {
		_, _, _, terminalAccessServiceImpl := loadUserTerminalAccessService(tt)
		editedManifest := "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":1},:{\"serviceAccountName\":\"hello\"}}"
		request := &models.UserTerminalSessionRequest{
			UserId:    int32(2),
			ClusterId: 1,
			BaseImage: "ubuntu",
			ShellName: "sh",
			NodeName:  "demo-new",
			Manifest:  editedManifest,
		}
		res, err := terminalAccessServiceImpl.EditTerminalPodManifest(context.Background(), request, false)
		assert.NotNil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, len(res.ErrorComments), 0, res.ErrorComments)

		editedManifest = "{\"apiVersion\":\"v1\",\"kind\":\"Random\",\"metadata\":{\"name\":1},\"spec\":{\"serviceAccountName\":\"hello\"}}"
		request.Manifest = editedManifest
		res, err = terminalAccessServiceImpl.EditTerminalPodManifest(context.Background(), request, false)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "manifest should be of kind \"Pod\"")
		assert.NotNil(t, res)
		assert.Equal(t, len(res.ErrorComments), 0, res.ErrorComments)
	})
}

func loadUserTerminalAccessService(t *testing.T) (*mocks.TerminalAccessRepository, *mocks2.TerminalSessionHandler, *mocks3.K8sApplicationService, *UserTerminalAccessServiceImpl) {
	logger, err := util.InitLogger()
	assert.Nil(t, err)
	userTerminalSessionConfig, err := GetTerminalAccessConfig()
	assert.Nil(t, err)
	userTerminalSessionConfig.MaxSessionPerUser = 1
	terminalAccessRepository := mocks.NewTerminalAccessRepository(t)
	terminalSessionHandler := mocks2.NewTerminalSessionHandler(t)
	k8sApplicationService := mocks3.NewK8sApplicationService(t)
	terminalAccessRepository.On("GetAllRunningUserTerminalData").Return(nil, nil)
	terminalAccessServiceImpl, err := NewUserTerminalAccessServiceImpl(logger, terminalAccessRepository, userTerminalSessionConfig, nil, terminalSessionHandler, nil, nil)
	assert.Nil(t, err)
	return terminalAccessRepository, terminalSessionHandler, k8sApplicationService, terminalAccessServiceImpl
}
