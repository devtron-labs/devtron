package clusterTerminalAccess

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserTerminalAccessService interface {
	StartTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	UpdateTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	UpdateTerminalShellSession(request *models.UserTerminalShellSessionRequest) (*models.UserTerminalSessionResponse, error)
	FetchTerminalStatus(terminalAccessId int) (*models.UserTerminalSessionResponse, error)
	StopTerminalSession(userTerminalSessionId int) error
	DisconnectTerminalSession(userTerminalSessionId int) error
	DisconnectAllSessionsForUser(userId int32)
}

type UserTerminalAccessServiceImpl struct {
	TerminalAccessRepository     repository.TerminalAccessRepository
	Logger                       *zap.SugaredLogger
	Config                       *models.UserTerminalSessionConfig
	TerminalAccessSessionDataMap *map[int]*UserTerminalAccessSessionData
	TerminalAccessDataArrayMutex *sync.RWMutex
	PodStatusSyncCron            *cron.Cron
	k8sApplicationService        k8s.K8sApplicationService
	k8sClientService             application.K8sClientService
	terminalSessionHandler       terminal.TerminalSessionHandler
}

type UserTerminalAccessSessionData struct {
	sessionId                string
	latestActivityTime       time.Time
	terminalAccessDataEntity *models.UserTerminalAccessData
	terminateTriggered       bool
}

func GetTerminalAccessConfig() (*models.UserTerminalSessionConfig, error) {
	config := &models.UserTerminalSessionConfig{}
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func NewUserTerminalAccessServiceImpl(logger *zap.SugaredLogger, terminalAccessRepository repository.TerminalAccessRepository, config *models.UserTerminalSessionConfig,
	k8sApplicationService k8s.K8sApplicationService, k8sClientService application.K8sClientService, terminalSessionHandler terminal.TerminalSessionHandler) (*UserTerminalAccessServiceImpl, error) {
	//fetches all running and starting entities from db and start SyncStatus
	podStatusSyncCron := cron.New(cron.WithChain())
	terminalAccessDataArrayMutex := &sync.RWMutex{}
	map1 := make(map[int]*UserTerminalAccessSessionData)
	accessServiceImpl := &UserTerminalAccessServiceImpl{
		Logger:                       logger,
		TerminalAccessRepository:     terminalAccessRepository,
		Config:                       config,
		PodStatusSyncCron:            podStatusSyncCron,
		TerminalAccessDataArrayMutex: terminalAccessDataArrayMutex,
		k8sApplicationService:        k8sApplicationService,
		k8sClientService:             k8sClientService,
		TerminalAccessSessionDataMap: &map1,
		terminalSessionHandler:       terminalSessionHandler,
	}
	podStatusSyncCron.Start()
	_, err := podStatusSyncCron.AddFunc(fmt.Sprintf("@every %ds", config.TerminalPodStatusSyncTimeInSecs), accessServiceImpl.SyncPodStatus)
	if err != nil {
		logger.Errorw("error occurred while starting cron job", "time in secs", config.TerminalPodStatusSyncTimeInSecs)
		return nil, err
	}
	go accessServiceImpl.SyncRunningInstances()
	return accessServiceImpl, err
}

func (impl *UserTerminalAccessServiceImpl) StartTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal start request received for user", "request", request)
	userId := request.UserId
	// check for max session check
	err := impl.checkMaxSessionLimit(userId)
	if err != nil {
		return nil, err
	}
	maxIdForUser := impl.getMaxIdForUser(userId)
	podNameVar := impl.createPodName(request, maxIdForUser)
	terminalEntity, err := impl.createTerminalEntity(request, podNameVar)
	if err != nil {
		return nil, err
	}
	err = impl.startTerminalPod(podNameVar, request)
	return terminalEntity, err
}

func (impl *UserTerminalAccessServiceImpl) checkMaxSessionLimit(userId int32) error {
	maxSessionPerUser := impl.Config.MaxSessionPerUser
	activeSessionList := impl.getUserActiveSessionList(userId)
	userRunningSessionCount := len(activeSessionList)
	if userRunningSessionCount >= maxSessionPerUser {
		errStr := fmt.Sprintf("cannot start new session more than configured %s", strconv.Itoa(maxSessionPerUser))
		impl.Logger.Errorw(errStr, "userId", userId)
		return errors.New("session-limit-reached")
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) getMaxIdForUser(userId int32) int {
	accessSessionDataMap := impl.TerminalAccessSessionDataMap
	maxId := 0
	for _, userTerminalAccessSessionData := range *accessSessionDataMap {
		terminalAccessDataEntity := userTerminalAccessSessionData.terminalAccessDataEntity
		if terminalAccessDataEntity.UserId == userId {
			accessId := terminalAccessDataEntity.Id
			if accessId > maxId {
				maxId = accessId
			}
		}
	}
	return maxId
}

func (impl *UserTerminalAccessServiceImpl) getUserActiveSessionList(userId int32) []*UserTerminalAccessSessionData {
	var userTerminalAccessSessionDataArray []*UserTerminalAccessSessionData
	accessSessionDataMap := impl.TerminalAccessSessionDataMap
	for _, userTerminalAccessSessionData := range *accessSessionDataMap {
		terminalAccessDataEntity := userTerminalAccessSessionData.terminalAccessDataEntity
		if terminalAccessDataEntity.UserId == userId && userTerminalAccessSessionData.sessionId != "" {
			userTerminalAccessSessionDataArray = append(userTerminalAccessSessionDataArray, userTerminalAccessSessionData)
		}
	}
	return userTerminalAccessSessionDataArray
}

func (impl *UserTerminalAccessServiceImpl) createTerminalEntity(request *models.UserTerminalSessionRequest, podName string) (*models.UserTerminalSessionResponse, error) {
	userAccessData := &models.UserTerminalAccessData{
		UserId:    request.UserId,
		ClusterId: request.ClusterId,
		NodeName:  request.NodeName,
		Status:    string(models.TerminalPodStarting),
		PodName:   podName,
		Metadata:  impl.extractMetadataString(request),
	}
	err := impl.TerminalAccessRepository.SaveUserTerminalAccessData(userAccessData)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving user terminal access data", "err", err)
		return nil, err
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	terminalAccessDataArray := *impl.TerminalAccessSessionDataMap
	terminalAccessDataArray[userAccessData.Id] = &UserTerminalAccessSessionData{terminalAccessDataEntity: userAccessData, latestActivityTime: time.Now()}
	impl.TerminalAccessSessionDataMap = &terminalAccessDataArray
	return &models.UserTerminalSessionResponse{
		UserId:           userAccessData.UserId,
		PodName:          podName,
		TerminalAccessId: userAccessData.Id,
	}, nil
}

func (impl *UserTerminalAccessServiceImpl) UpdateTerminalShellSession(request *models.UserTerminalShellSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal update shell request received for user", "request", request)
	userTerminalAccessId := request.TerminalAccessId
	err := impl.StopTerminalSession(userTerminalAccessId)
	if err != nil {
		return nil, err
	}
	terminalAccessData, err := impl.TerminalAccessRepository.GetUserTerminalAccessData(userTerminalAccessId)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching user terminal access data", "userTerminalAccessId", userTerminalAccessId, "err", err)
		return nil, err
	}
	terminalAccessData.Metadata = impl.mergeToMetadataString(terminalAccessData.Metadata, request)
	err = impl.TerminalAccessRepository.UpdateUserTerminalAccessData(terminalAccessData)
	if err != nil {
		impl.Logger.Errorw("error occurred while updating terminal Access data ", "userTerminalAccessId", userTerminalAccessId, "err", err)
		return nil, err
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	terminalAccessSessionData := terminalAccessDataMap[terminalAccessData.Id]
	terminalAccessSessionData.terminalAccessDataEntity = terminalAccessData
	terminalAccessSessionData.latestActivityTime = time.Now()
	impl.TerminalAccessSessionDataMap = &terminalAccessDataMap

	return &models.UserTerminalSessionResponse{
		UserId:           terminalAccessData.UserId,
		PodName:          terminalAccessData.PodName,
		TerminalAccessId: terminalAccessData.Id,
	}, nil
}

func (impl *UserTerminalAccessServiceImpl) UpdateTerminalSession(request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal update request received for user", "request", request)
	userTerminalAccessId := request.Id
	err := impl.DisconnectTerminalSession(userTerminalAccessId)
	if err != nil {
		return nil, err
	}

	return impl.StartTerminalSession(request)
}

func (impl *UserTerminalAccessServiceImpl) DisconnectTerminalSession(userTerminalAccessId int) error {
	impl.Logger.Info("Disconnect terminal session request received", "userTerminalAccessId", userTerminalAccessId)
	err := impl.StopTerminalSession(userTerminalAccessId)
	if err != nil {
		return err
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	accessSessionDataMap := *impl.TerminalAccessSessionDataMap
	accessSessionData := accessSessionDataMap[userTerminalAccessId]
	terminalAccessData := accessSessionData.terminalAccessDataEntity
	err = impl.DeleteTerminalPod(terminalAccessData.ClusterId, terminalAccessData.PodName)
	if err != nil {
		if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == "NotFound" {
			accessSessionData.terminateTriggered = true
			err = nil
		}
	} else {
		accessSessionData.terminateTriggered = true
	}
	return err
}

func (impl *UserTerminalAccessServiceImpl) StopTerminalSession(userTerminalAccessId int) error {
	impl.Logger.Infow("terminal stop request received for user", "userTerminalAccessId", userTerminalAccessId)
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	accessSessionDataMap := *impl.TerminalAccessSessionDataMap
	accessSessionData, present := accessSessionDataMap[userTerminalAccessId]
	if present {
		impl.closeAndCleanTerminalSession(accessSessionData)
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) DisconnectAllSessionsForUser(userId int32) {
	impl.Logger.Infow("disconnecting all active session for user", "userId", userId)
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	activeSessionList := impl.getUserActiveSessionList(userId)
	for _, accessSessionData := range activeSessionList {
		impl.closeAndCleanTerminalSession(accessSessionData)
	}
}

func (impl *UserTerminalAccessServiceImpl) closeAndCleanTerminalSession(accessSessionData *UserTerminalAccessSessionData) {
	sessionId := accessSessionData.sessionId
	if sessionId != "" {
		userTerminalAccessId := accessSessionData.terminalAccessDataEntity.Id
		impl.Logger.Infow("closing socket connection", "userTerminalAccessId", userTerminalAccessId)
		impl.closeSession(sessionId)
		accessSessionData.sessionId = ""
		accessSessionData.latestActivityTime = time.Now()
	}
}

func (impl *UserTerminalAccessServiceImpl) closeSession(sessionId string) {
	impl.terminalSessionHandler.Close(sessionId, 1, "Process exited")
}

func (impl *UserTerminalAccessServiceImpl) extractMetadataString(request *models.UserTerminalSessionRequest) string {
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

func (impl *UserTerminalAccessServiceImpl) mergeToMetadataString(metadataJsonStr string, request *models.UserTerminalShellSessionRequest) string {
	metadataMap, err := impl.getMetadataMap(metadataJsonStr)
	if err != nil {
		metadataMap = make(map[string]string)
	}
	metadataMap["ShellName"] = request.ShellName
	metadataJsonBytes, err := json.Marshal(metadataMap)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting metadata to json", "request", request, "err", err)
		return "{}"
	}
	return string(metadataJsonBytes)
}

func (impl *UserTerminalAccessServiceImpl) getMetadataMap(metadata string) (map[string]string, error) {
	var metadataMap map[string]string
	err := json.Unmarshal([]byte(metadata), &metadataMap)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting metadata to map", "metadata", metadata, "err", err)
		return nil, err
	}
	return metadataMap, nil
}

func (impl *UserTerminalAccessServiceImpl) startTerminalPod(podNameVar string, request *models.UserTerminalSessionRequest) error {

	accessTemplates, err := impl.TerminalAccessRepository.FetchAllTemplates()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access templates", "err", err)
		return err
	}
	for _, accessTemplate := range accessTemplates {
		err = impl.applyTemplateData(request, podNameVar, accessTemplate, false)
		if err != nil {
			return err
		}
	}
	return err
}

func (impl *UserTerminalAccessServiceImpl) createPodName(request *models.UserTerminalSessionRequest, runningCount int) string {
	podNameVar := models.TerminalAccessPodNameTemplate
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessClusterIdTemplateVar, strconv.Itoa(request.ClusterId))
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	podNameVar = strings.ReplaceAll(podNameVar, models.TerminalAccessRandomIdVar, strconv.Itoa(runningCount+1))
	return podNameVar
}

func (impl *UserTerminalAccessServiceImpl) applyTemplateData(request *models.UserTerminalSessionRequest, podNameVar string,
	terminalTemplate *models.TerminalAccessTemplates, isUpdate bool) error {
	templateName := terminalTemplate.TemplateName
	templateData := terminalTemplate.TemplateData
	clusterId := request.ClusterId
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessClusterIdTemplateVar, strconv.Itoa(clusterId))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessNodeNameVar, request.NodeName)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessBaseImageVar, request.BaseImage)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessNamespaceVar, impl.Config.TerminalPodDefaultNamespace)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessPodNameVar, podNameVar)
	err := impl.applyTemplate(clusterId, terminalTemplate.TemplateData, templateData, isUpdate)
	if err != nil {
		impl.Logger.Errorw("error occurred while applying template ", "name", templateName, "err", err)
		return err
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) SyncPodStatus() {
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	for _, terminalAccessSessionData := range terminalAccessDataMap {
		sessionId := terminalAccessSessionData.sessionId
		if sessionId != "" {
			validSession := impl.terminalSessionHandler.ValidateSession(sessionId)
			if validSession {
				continue
			} else {
				impl.closeAndCleanTerminalSession(terminalAccessSessionData)
			}
		}
		//check remaining running which are active from last x minutes
		timeGapInMinutes := time.Since(terminalAccessSessionData.latestActivityTime).Minutes()
		if impl.Config.TerminalPodInActiveDurationInMins < int(timeGapInMinutes) {
			terminalAccessData := terminalAccessSessionData.terminalAccessDataEntity
			existingStatus := terminalAccessData.Status
			terminalPodStatusString := existingStatus
			impl.deleteClusterTerminalTemplates(terminalAccessData.ClusterId, terminalAccessData.PodName)
			err := impl.DeleteTerminalPod(terminalAccessData.ClusterId, terminalAccessData.PodName)
			if err != nil {
				if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == "NotFound" {
					terminalPodStatusString = string(models.TerminalPodTerminated)
				} else {
					continue
				}
			}
			terminalAccessSessionData.terminateTriggered = true
			if existingStatus != terminalPodStatusString {
				terminalAccessId := terminalAccessData.Id
				err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(terminalAccessId, terminalPodStatusString)
				if err != nil {
					impl.Logger.Errorw("error occurred while updating terminal status", "terminalAccessId", terminalAccessId, "err", err)
					continue
				}
				terminalAccessData.Status = terminalPodStatusString
			}
		}
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	for _, terminalAccessSessionData := range terminalAccessDataMap {
		terminalAccessData := terminalAccessSessionData.terminalAccessDataEntity
		if terminalAccessData.Status != string(models.TerminalPodStarting) && terminalAccessData.Status != string(models.TerminalPodRunning) {
			// check if this is the last data for this cluster and user then delete terminal resource
			delete(terminalAccessDataMap, terminalAccessData.Id)
		}
	}
	impl.TerminalAccessSessionDataMap = &terminalAccessDataMap
}

func (impl *UserTerminalAccessServiceImpl) checkAndStartSession(terminalAccessData *models.UserTerminalAccessData) (string, error) {
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessPodTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching template", "template", models.TerminalAccessPodTemplateName, "err", err)
		return "", err
	}
	clusterId := terminalAccessData.ClusterId
	terminalAccessPodName := terminalAccessData.PodName
	terminalPodStatusString, err := impl.getPodStatus(clusterId, terminalAccessPodName, terminalAccessPodTemplate.TemplateData)
	if err != nil {
		return "", err
	}
	sessionID := ""
	terminalAccessId := terminalAccessData.Id
	if terminalPodStatusString == string(models.TerminalPodRunning) {
		err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(terminalAccessId, terminalPodStatusString)
		if err != nil {
			impl.Logger.Errorw("error occurred while updating terminal status", "terminalAccessId", terminalAccessId, "err", err)
			return "", err
		}
		terminalAccessData.Status = terminalPodStatusString
		//create terminal session if status is Running and store sessionId
		metadata := terminalAccessData.Metadata
		metadataMap, err := impl.getMetadataMap(metadata)
		if err != nil {
			impl.Logger.Errorw("error occurred while converting metadata json to map", "metadata", metadata, "err", err)
			return "", err
		}
		request := &terminal.TerminalSessionRequest{
			Shell:     metadataMap["ShellName"],
			Namespace: impl.Config.TerminalPodDefaultNamespace,
			PodName:   terminalAccessPodName,
			ClusterId: clusterId,
		}
		_, terminalMessage, err := impl.terminalSessionHandler.GetTerminalSession(request)
		if err != nil {
			impl.Logger.Errorw("error occurred while creating terminal session", "terminalAccessId", terminalAccessId, "err", err)
			return "", err
		}
		sessionID = terminalMessage.SessionID
	}
	return sessionID, err
}

func (impl *UserTerminalAccessServiceImpl) FetchTerminalStatus(terminalAccessId int) (*models.UserTerminalSessionResponse, error) {
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	terminalAccessSessionData, present := terminalAccessDataMap[terminalAccessId]
	var terminalSessionId = ""
	var terminalAccessData *models.UserTerminalAccessData
	if present {
		if terminalAccessSessionData.terminateTriggered {
			accessDataEntity := terminalAccessSessionData.terminalAccessDataEntity
			return &models.UserTerminalSessionResponse{
				TerminalAccessId:      terminalAccessId,
				UserId:                accessDataEntity.UserId,
				Status:                models.TerminalPodStatus(accessDataEntity.Status),
				PodName:               accessDataEntity.PodName,
				UserTerminalSessionId: terminalSessionId,
			}, nil
		} else {
			terminalSessionId = terminalAccessSessionData.sessionId
			validSession := impl.terminalSessionHandler.ValidateSession(terminalSessionId)
			if validSession {
				terminalAccessData = terminalAccessSessionData.terminalAccessDataEntity
			}
		}
	}
	if terminalAccessData == nil {
		existingTerminalAccessData, err := impl.TerminalAccessRepository.GetUserTerminalAccessData(terminalAccessId)
		if err != nil {
			impl.Logger.Errorw("error occurred while fetching terminal status", "terminalAccessId", terminalAccessId, "err", err)
			return nil, err
		}
		terminalAccessData = existingTerminalAccessData
		if existingTerminalAccessData.Status == string(models.TerminalPodTerminated) {
			return nil, errors.New("pod-terminated")
		}
		err = impl.checkMaxSessionLimit(existingTerminalAccessData.UserId)
		if err != nil {
			return nil, err
		}
		terminalSessionId, err = impl.checkAndStartSession(existingTerminalAccessData)
		if err != nil {
			return nil, err
		}
		impl.TerminalAccessDataArrayMutex.Lock()
		terminalAccessSessionData.sessionId = terminalSessionId
		terminalAccessSessionData.terminalAccessDataEntity = existingTerminalAccessData
		terminalAccessDataMap[terminalAccessId] = terminalAccessSessionData
		impl.TerminalAccessDataArrayMutex.Unlock()
	}
	terminalAccessDataId := terminalAccessData.Id

	terminalAccessResponse := &models.UserTerminalSessionResponse{
		TerminalAccessId:      terminalAccessDataId,
		UserId:                terminalAccessData.UserId,
		Status:                models.TerminalPodStatus(terminalAccessData.Status),
		PodName:               terminalAccessData.PodName,
		UserTerminalSessionId: terminalSessionId,
	}
	return terminalAccessResponse, nil
}

func (impl *UserTerminalAccessServiceImpl) DeleteTerminalPod(clusterId int, terminalPodName string) error {
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessPodTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching template", "template", models.TerminalAccessPodTemplateName, "err", err)
		return err
	}
	gvkDataString := terminalAccessPodTemplate.TemplateData
	err = impl.DeleteTerminalResource(clusterId, terminalPodName, gvkDataString)
	return err
}

func (impl *UserTerminalAccessServiceImpl) DeleteTerminalResource(clusterId int, terminalResourceName string, resourceTemplateString string) error {
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(resourceTemplateString), nil, nil)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "resourceTemplateString", resourceTemplateString, "err", err)
		return err
	}

	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(clusterId)
	if err != nil {
		return err
	}

	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name:      terminalResourceName,
			Namespace: impl.Config.TerminalPodDefaultNamespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   groupVersionKind.Group,
				Version: groupVersionKind.Version,
				Kind:    groupVersionKind.Kind,
			},
		},
	}
	_, err = impl.k8sClientService.DeleteResource(restConfig, k8sRequest)
	if err != nil {
		impl.Logger.Errorw("error occurred while deleting resource for pod", "podName", terminalResourceName, "err", err)
	}
	return err
}

func (impl *UserTerminalAccessServiceImpl) applyTemplate(clusterId int, gvkDataString string, templateData string, isUpdate bool) error {
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(clusterId)
	if err != nil {
		return err
	}

	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(gvkDataString), nil, nil)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "gvkDataString", gvkDataString, "err", err)
		return err
	}

	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Namespace: impl.Config.TerminalPodDefaultNamespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   groupVersionKind.Group,
				Version: groupVersionKind.Version,
				Kind:    groupVersionKind.Kind,
			},
		},
	}

	if isUpdate {
		k8sRequest.Patch = templateData
		_, err = impl.k8sClientService.UpdateResource(restConfig, k8sRequest)
	} else {
		_, err = impl.k8sClientService.CreateResource(restConfig, k8sRequest, templateData)
	}
	if err != nil {
		if errStatus, ok := err.(*k8sErrors.StatusError); !(ok && errStatus.Status().Reason == "AlreadyExists") {
			impl.Logger.Errorw("error in creating resource", "err", err, "request", k8sRequest)
			return err
		}
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) getPodStatus(clusterId int, podName string, gvkDataString string) (string, error) {
	// return terminated if pod does not exist
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(gvkDataString), nil, nil)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "gvkDataString", gvkDataString, "err", err)
		return "", err
	}
	request := &k8s.ResourceRequestBean{
		AppIdentifier: &client.AppIdentifier{
			ClusterId: clusterId,
		},
		K8sRequest: &application.K8sRequestBean{
			ResourceIdentifier: application.ResourceIdentifier{
				Name:      podName,
				Namespace: impl.Config.TerminalPodDefaultNamespace,
				GroupVersionKind: schema.GroupVersionKind{
					Group:   groupVersionKind.Group,
					Version: groupVersionKind.Version,
					Kind:    groupVersionKind.Kind,
				},
			},
		},
	}
	response, err := impl.k8sApplicationService.GetResource(request)
	if err != nil {
		if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == "NotFound" {
			return string(models.TerminalPodTerminated), nil
		} else {
			impl.Logger.Errorw("error occurred while fetching resource info for pod", "podName", podName)
			return "", err
		}
	}
	status := ""
	if response != nil {
		manifest := response.Manifest
		for key, value := range manifest.Object {
			if key == "status" {
				statusData := value.(map[string]interface{})
				status = statusData["phase"].(string)
			}
		}
	}
	impl.Logger.Debug("pod status", "podName", podName, "status", status)
	return status, nil
}

func (impl *UserTerminalAccessServiceImpl) SyncRunningInstances() {
	terminalAccessData, err := impl.TerminalAccessRepository.GetAllRunningUserTerminalData()
	if err != nil {
		impl.Logger.Fatalw("error occurred while fetching all running/starting data", "err", err)
	}
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	for _, accessData := range terminalAccessData {
		terminalAccessDataMap[accessData.Id] = &UserTerminalAccessSessionData{
			terminalAccessDataEntity: accessData,
			latestActivityTime:       time.Now(),
		}
	}
	impl.TerminalAccessSessionDataMap = &terminalAccessDataMap
	impl.Logger.Infow("all running/starting terminal pod loaded", "size", len(terminalAccessDataMap))
}

func (impl *UserTerminalAccessServiceImpl) deleteClusterTerminalTemplates(clusterId int, podName string) {
	templateData, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessClusterRoleBindingTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access template", "err", err)
		return
	}
	templateName := strings.ReplaceAll(models.TerminalAccessClusterRoleBindingTemplate, models.TerminalAccessPodNameTemplate, podName)
	impl.DeleteTerminalResource(clusterId, templateName, templateData.TemplateData)

	templateData, err = impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessServiceAccountTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access template", "err", err)
		return
	}
	templateName = strings.ReplaceAll(models.TerminalAccessServiceAccountTemplate, models.TerminalAccessPodNameTemplate, podName)
	impl.DeleteTerminalResource(clusterId, templateName, templateData.TemplateData)
}
