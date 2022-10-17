package clusterTerminalAccess

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
	"strings"
	"sync"
)

type UserTerminalAccessService interface {
	StartTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	UpdateTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	DisconnectTerminalSession(userTerminalSessionId int) error
}

type UserTerminalAccessServiceImpl struct {
	TerminalAccessRepository     repository.TerminalAccessRepository
	Logger                       *zap.SugaredLogger
	Config                       *models.UserTerminalSessionConfig
	TerminalAccessDataArray      []*models.UserTerminalAccessData
	TerminalAccessDataArrayMutex *sync.RWMutex
	PodStatusSyncCron            *cron.Cron
	k8sApplicationService        k8s.K8sApplicationService
	k8sClientService             application.K8sClientService
}

func NewUserTerminalAccessServiceImpl(logger *zap.SugaredLogger, terminalAccessRepository repository.TerminalAccessRepository, k8sApplicationService k8s.K8sApplicationService, k8sClientService application.K8sClientService) (*UserTerminalAccessServiceImpl, error) {
	config := &models.UserTerminalSessionConfig{}
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}
	//fetches all running and starting entities from db and start SyncStatus
	podStatusSyncCron := cron.New(
		cron.WithChain())
	terminalAccessDataArrayMutex := &sync.RWMutex{}
	accessServiceImpl := &UserTerminalAccessServiceImpl{
		Logger:                       logger,
		TerminalAccessRepository:     terminalAccessRepository,
		Config:                       config,
		PodStatusSyncCron:            podStatusSyncCron,
		TerminalAccessDataArrayMutex: terminalAccessDataArrayMutex,
		k8sApplicationService:        k8sApplicationService,
		k8sClientService:             k8sClientService,
	}
	podStatusSyncCron.Start()
	_, err = podStatusSyncCron.AddFunc(fmt.Sprintf("@every %ds", config.TerminalPodStatusSyncTimeInSecs), accessServiceImpl.SyncPodStatus)
	if err != nil {
		logger.Errorw("error occurred while starting cron job", "time in secs", config.TerminalPodStatusSyncTimeInSecs)
	}
	go accessServiceImpl.SyncRunningInstances()
	return accessServiceImpl, err
}

func (impl UserTerminalAccessServiceImpl) StartTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	userId := request.UserId
	terminalAccessDataList, err := impl.TerminalAccessRepository.GetUserTerminalAccessDataByUser(userId)
	if err != nil {
		impl.Logger.Errorw("error occurred while getting terminal access data for user id", "userId", userId, "err", err)
		return nil, err
	}
	// check for max session check
	maxSessionPerUser := impl.Config.MaxSessionPerUser
	userRunningSessionCount := len(terminalAccessDataList)
	if userRunningSessionCount >= maxSessionPerUser {
		errStr := fmt.Sprintf("cannot start new session more than configured %s", strconv.Itoa(maxSessionPerUser))
		impl.Logger.Errorw(errStr, "req", request)
		return nil, errors.New(errStr)
	}
	podName, err := impl.startTerminalPod(request, userRunningSessionCount)
	return impl.createTerminalEntity(request, podName, err)
}

func (impl UserTerminalAccessServiceImpl) createTerminalEntity(request *models.UserTerminalSessionRequest, podName string, err error) (*models.UserTerminalSessionResponse, error) {
	userAccessData := &models.UserTerminalAccessData{
		UserId:    request.UserId,
		ClusterId: request.ClusterId,
		NodeName:  request.NodeName,
		Status:    string(models.TerminalPodStarting),
		PodName:   podName,
		Metadata:  impl.extractMetadataString(request),
	}
	err = impl.TerminalAccessRepository.SaveUserTerminalAccessData(userAccessData)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving user terminal access data", "err", err)
		return nil, err
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	impl.TerminalAccessDataArray = append(impl.TerminalAccessDataArray, userAccessData)
	impl.TerminalAccessDataArrayMutex.Unlock()
	return &models.UserTerminalSessionResponse{
		UserTerminalSessionId: userAccessData.Id,
		UserId:                userAccessData.UserId,
		ShellName:             request.ShellName,
	}, nil
}

func (impl UserTerminalAccessServiceImpl) UpdateTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	userTerminalSessionId := request.Id
	terminalAccessData, err := impl.TerminalAccessRepository.GetUserTerminalAccessData(userTerminalSessionId)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching user terminal access data", "userTerminalSessionId", userTerminalSessionId, "err", err)
		return nil, err
	}
	podName := terminalAccessData.PodName
	err = impl.DeleteTerminalPod(terminalAccessData.ClusterId, podName)
	if err != nil {
		impl.Logger.Errorw("error occurred while deleting terminal pod", "userTerminalSessionId", userTerminalSessionId, "err", err)
		return nil, err
	}
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessPodTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching template", "template", models.TerminalAccessPodTemplateName, "err", err)
		return nil, err
	}
	err = impl.applyTemplateData(request, podName, terminalAccessPodTemplate)
	if err != nil {
		return nil, err
	}
	return impl.createTerminalEntity(request, podName, err)
}

func (impl UserTerminalAccessServiceImpl) checkTerminalExists(userTerminalSessionId int) (*models.UserTerminalAccessData, error) {
	terminalAccessData, err := impl.TerminalAccessRepository.GetUserTerminalAccessData(userTerminalSessionId)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching user terminal access data", "userTerminalSessionId", userTerminalSessionId, "err", err)
		return nil, err
	}
	terminalStatus := terminalAccessData.Status
	terminalPodStatus := models.TerminalPodStatus(terminalStatus)
	if terminalPodStatus == models.TerminalPodTerminated || terminalPodStatus == models.TerminalPodError {
		impl.Logger.Errorw("pod is already in terminated/error state", "userTerminalSessionId", userTerminalSessionId, "terminalPodStatus", terminalPodStatus)
		return nil, errors.New("pod already terminated")
	}
	return terminalAccessData, nil
}

func (impl UserTerminalAccessServiceImpl) DisconnectTerminalSession(userTerminalSessionId int) error {
	terminalAccessData, err := impl.checkTerminalExists(userTerminalSessionId)
	if err != nil {
		return err
	}
	// handle already terminated/not found cases
	err = impl.DeleteTerminalPod(terminalAccessData.ClusterId, terminalAccessData.PodName)
	if err != nil {
		impl.Logger.Errorw("error occurred while stopping terminal pod", "userTerminalSessionId", userTerminalSessionId, "err", err)
		return err
	}
	err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(userTerminalSessionId, string(models.TerminalPodTerminated))
	if err != nil {
		impl.Logger.Errorw("error occurred while updating terminal status in db", "userTerminalSessionId", userTerminalSessionId, "err", err)
	}
	return err
}

func (impl UserTerminalAccessServiceImpl) extractMetadataString(request *models.UserTerminalSessionRequest) string {
	metadata := make(map[string]string)
	metadata["BaseImage"] = request.BaseImage
	metadata["ShellName"] = request.ShellName
	metadataJsonBytes, err := json.Marshal(metadata)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting metadata to json", "request", request, "err", err)
		return "{}"
	}
	return string(metadataJsonBytes)
}

func (impl UserTerminalAccessServiceImpl) startTerminalPod(request *models.UserTerminalSessionRequest, runningCount int) (string, error) {
	podNameVar := impl.createPodName(request, runningCount)
	accessTemplates, err := impl.TerminalAccessRepository.FetchAllTemplates()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access templates", "err", err)
		return "", err
	}
	for _, accessTemplate := range accessTemplates {
		err = impl.applyTemplateData(request, podNameVar, accessTemplate)
		if err != nil {
			return "", err
		}
	}
	return podNameVar, err
}

func (impl UserTerminalAccessServiceImpl) createPodName(request *models.UserTerminalSessionRequest, runningCount int) string {
	podNameVar := models.TerminalAccessPodNameTemplate
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessClusterIdTemplateVar, strconv.Itoa(request.ClusterId))
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessRandomIdVar, strconv.Itoa(runningCount+1))
	return podNameVar
}

func (impl UserTerminalAccessServiceImpl) applyTemplateData(request *models.UserTerminalSessionRequest, podNameVar string, terminalTemplate *models.TerminalAccessTemplates) error {

	templateData := terminalTemplate.TemplateData
	clusterId := request.ClusterId
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessClusterIdTemplateVar, strconv.Itoa(clusterId))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessPodNameVar, podNameVar)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessBaseImageVar, request.BaseImage)
	err := impl.applyTemplate(clusterId, terminalTemplate.TemplateKindData, templateData)
	if err != nil {
		impl.Logger.Errorw("error occurred while applying template ", "name", terminalTemplate.TemplateName, "err", err)
		return err
	}
	return nil
}

func (impl UserTerminalAccessServiceImpl) SyncPodStatus() {
	// set starting/running pods in memory and fetch status of those pods and update their status in Db
	for _, terminalAccessData := range impl.TerminalAccessDataArray {
		clusterId := terminalAccessData.ClusterId
		terminalAccessPodName := terminalAccessData.PodName
		terminalPodStatus, err := impl.getPodStatus(clusterId, terminalAccessPodName)
		if err != nil {
			continue
		}
		terminalPodStatusString := string(terminalPodStatus)
		if terminalAccessData.Status != terminalPodStatusString {
			terminalAccessId := terminalAccessData.Id
			err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(terminalAccessId, terminalPodStatusString)
			if err != nil {
				impl.Logger.Errorw("error occurred while updating terminal status", "terminalAccessId", terminalAccessId, "err", err)
				continue
			}
			terminalAccessData.Status = terminalPodStatusString
		}
	}
	var newArray []*models.UserTerminalAccessData
	for _, terminalAccessData := range impl.TerminalAccessDataArray {
		if terminalAccessData.Status != string(models.TerminalPodTerminated) && terminalAccessData.Status != string(models.TerminalPodError) {
			newArray = append(newArray, terminalAccessData)
		}
	}
	impl.TerminalAccessDataArray = newArray
}

func (impl UserTerminalAccessServiceImpl) DeleteTerminalPod(clusterId int, terminalPodName string) error {
	//make pod delete request, handle errors if pod does  not exists
	return nil
}

func (impl UserTerminalAccessServiceImpl) applyTemplate(clusterId int, gvkDataString string, templateData string) error {
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(clusterId)
	if err != nil {
		return err
	}

	var gvkData map[string]string
	err = json.Unmarshal([]byte(gvkDataString), &gvkData)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "gvkDataString", gvkDataString, "err", err)
		return err
	}

	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			GroupVersionKind: schema.GroupVersionKind{
				Group:   gvkData["group"],
				Version: gvkData["version"],
				Kind:    gvkData["kind"],
			},
		},
	}

	_, err = impl.k8sClientService.CreateResource(restConfig, k8sRequest, templateData)
	if err != nil && err.(*k8sErrors.StatusError).Status().Reason != "AlreadyExists" {
		impl.Logger.Errorw("error in creating resource", "err", err, "request", k8sRequest)
		return err
	}
	return nil
}

func (impl UserTerminalAccessServiceImpl) getPodStatus(clusterId int, podName string) (models.TerminalPodStatus, error) {
	// return terminated if pod does not exist
	return models.TerminalPodRunning, nil
}

func (impl UserTerminalAccessServiceImpl) SyncRunningInstances() {
	terminalAccessData, err := impl.TerminalAccessRepository.GetAllUserTerminalAccessData()
	if err != nil {
		impl.Logger.Fatalw("error occurred while fetching all running/starting data", "err", err)
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	impl.TerminalAccessDataArray = terminalAccessData
	impl.TerminalAccessDataArrayMutex.Unlock()
}
