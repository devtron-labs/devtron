package clusterTerminalAccess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	terminal2 "github.com/devtron-labs/devtron/internal/sql/repository/terminal"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserTerminalAccessService interface {
	StartTerminalSession(ctx context.Context, request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	UpdateTerminalSession(ctx context.Context, request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error)
	UpdateTerminalShellSession(ctx context.Context, request *models.UserTerminalShellSessionRequest) (*models.UserTerminalSessionResponse, error)
	FetchTerminalStatus(ctx context.Context, terminalAccessId int) (*models.UserTerminalSessionResponse, error)
	StopTerminalSession(ctx context.Context, userTerminalAccessId int)
	DisconnectTerminalSession(ctx context.Context, userTerminalAccessId int) error
	DisconnectAllSessionsForUser(ctx context.Context, userId int32)
	FetchPodManifest(ctx context.Context, userTerminalAccessId int) (resp *application.ManifestResponse, err error)
	FetchPodEvents(ctx context.Context, userTerminalAccessId int) (*application.EventsResponse, error)
}

type UserTerminalAccessServiceImpl struct {
	TerminalAccessRepository     terminal2.TerminalAccessRepository
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

func NewUserTerminalAccessServiceImpl(logger *zap.SugaredLogger, terminalAccessRepository terminal2.TerminalAccessRepository, config *models.UserTerminalSessionConfig,
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

func (impl *UserTerminalAccessServiceImpl) StartTerminalSession(ctx context.Context, request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
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
	err = impl.startTerminalPod(ctx, podNameVar, request)
	return terminalEntity, err
}

func (impl *UserTerminalAccessServiceImpl) checkMaxSessionLimit(userId int32) error {
	maxSessionPerUser := impl.Config.MaxSessionPerUser
	activeSessionList := impl.getUserActiveSessionList(userId)
	userRunningSessionCount := len(activeSessionList)
	if userRunningSessionCount >= maxSessionPerUser {
		errStr := fmt.Sprintf("cannot start new session more than configured %s", strconv.Itoa(maxSessionPerUser))
		impl.Logger.Errorw(errStr, "userId", userId)
		return errors.New(models.MaxSessionLimitReachedMsg)
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

func (impl *UserTerminalAccessServiceImpl) UpdateTerminalShellSession(ctx context.Context, request *models.UserTerminalShellSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal update shell request received for user", "request", request)
	userTerminalAccessId := request.TerminalAccessId
	impl.StopTerminalSession(ctx, userTerminalAccessId)
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

func (impl *UserTerminalAccessServiceImpl) UpdateTerminalSession(ctx context.Context, request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal update request received for user", "request", request)
	userTerminalAccessId := request.Id
	err := impl.DisconnectTerminalSession(ctx, userTerminalAccessId)
	if err != nil {
		return nil, err
	}

	return impl.StartTerminalSession(ctx, request)
}

func (impl *UserTerminalAccessServiceImpl) DisconnectTerminalSession(ctx context.Context, userTerminalAccessId int) error {
	impl.Logger.Info("Disconnect terminal session request received", "userTerminalAccessId", userTerminalAccessId)
	impl.StopTerminalSession(ctx, userTerminalAccessId)
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	accessSessionDataMap := *impl.TerminalAccessSessionDataMap
	accessSessionData := accessSessionDataMap[userTerminalAccessId]
	terminalAccessData := accessSessionData.terminalAccessDataEntity
	metadata := terminalAccessData.Metadata
	metadataMap, err := impl.getMetadataMap(metadata)
	if err != nil {
		return err
	}
	namespace := metadataMap["Namespace"]
	err = impl.DeleteTerminalPod(ctx, terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
	if err != nil {
		if isResourceNotFoundErr(err) {
			accessSessionData.terminateTriggered = true
			err = nil
		}
	} else {
		accessSessionData.terminateTriggered = true
	}
	return err
}

func isResourceNotFoundErr(err error) bool {
	if errStatus, ok := err.(*k8sErrors.StatusError); ok && errStatus.Status().Reason == metav1.StatusReasonNotFound {
		return true
	}
	return false
}

func (impl *UserTerminalAccessServiceImpl) StopTerminalSession(ctx context.Context, userTerminalAccessId int) {
	impl.Logger.Infow("terminal stop request received for user", "userTerminalAccessId", userTerminalAccessId)
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
	accessSessionDataMap := *impl.TerminalAccessSessionDataMap
	accessSessionData, present := accessSessionDataMap[userTerminalAccessId]
	if present {
		impl.closeAndCleanTerminalSession(accessSessionData)
	}
}

func (impl *UserTerminalAccessServiceImpl) DisconnectAllSessionsForUser(ctx context.Context, userId int32) {
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
	metadata["Namespace"] = request.Namespace
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

func (impl *UserTerminalAccessServiceImpl) startTerminalPod(ctx context.Context, podNameVar string, request *models.UserTerminalSessionRequest) error {

	accessTemplates, err := impl.TerminalAccessRepository.FetchAllTemplates()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access templates", "err", err)
		return err
	}
	for _, accessTemplate := range accessTemplates {
		err = impl.applyTemplateData(ctx, request, podNameVar, accessTemplate, false)
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

func (impl *UserTerminalAccessServiceImpl) applyTemplateData(ctx context.Context, request *models.UserTerminalSessionRequest, podNameVar string,
	terminalTemplate *models.TerminalAccessTemplates, isUpdate bool) error {
	templateName := terminalTemplate.TemplateName
	templateData := terminalTemplate.TemplateData
	clusterId := request.ClusterId
	namespace := request.Namespace
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessClusterIdTemplateVar, strconv.Itoa(clusterId))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessNodeNameVar, request.NodeName)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessBaseImageVar, request.BaseImage)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessNamespaceVar, namespace)
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessPodNameVar, podNameVar)
	err := impl.applyTemplate(ctx, clusterId, terminalTemplate.TemplateData, templateData, isUpdate, namespace)
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
			metadata := terminalAccessData.Metadata
			metadataMap, err := impl.getMetadataMap(metadata)
			if err != nil {
				continue
			}
			namespace := metadataMap["Namespace"]
			impl.deleteClusterTerminalTemplates(context.Background(), terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
			err = impl.DeleteTerminalPod(context.Background(), terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
			if err != nil {
				if isResourceNotFoundErr(err) {
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

func (impl *UserTerminalAccessServiceImpl) checkAndStartSession(ctx context.Context, terminalAccessData *models.UserTerminalAccessData) (string, error) {
	clusterId := terminalAccessData.ClusterId
	terminalAccessPodName := terminalAccessData.PodName
	metadata := terminalAccessData.Metadata
	metadataMap, err := impl.getMetadataMap(metadata)
	if err != nil {
		return "", err
	}
	namespace := metadataMap["Namespace"]
	terminalPodStatusString, err := impl.getPodStatus(ctx, clusterId, terminalAccessPodName, namespace)
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
		request := &terminal.TerminalSessionRequest{
			Shell:     metadataMap["ShellName"],
			Namespace: namespace,
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

func (impl *UserTerminalAccessServiceImpl) FetchTerminalStatus(ctx context.Context, terminalAccessId int) (*models.UserTerminalSessionResponse, error) {
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
	terminalAccessData, err := impl.validateTerminalAccessFromDb(ctx, terminalAccessId, terminalAccessData, terminalSessionId, terminalAccessSessionData, terminalAccessDataMap)
	if err != nil {
		return nil, err
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

func (impl *UserTerminalAccessServiceImpl) validateTerminalAccessFromDb(ctx context.Context, terminalAccessId int, terminalAccessData *models.UserTerminalAccessData, terminalSessionId string, terminalAccessSessionData *UserTerminalAccessSessionData, terminalAccessDataMap map[int]*UserTerminalAccessSessionData) (*models.UserTerminalAccessData, error) {
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
		terminalSessionId, err = impl.checkAndStartSession(ctx, existingTerminalAccessData)
		if err != nil {
			return nil, err
		}
		if terminalAccessSessionData == nil {
			terminalAccessSessionData = &UserTerminalAccessSessionData{}
		}
		impl.TerminalAccessDataArrayMutex.Lock()
		terminalAccessSessionData.sessionId = terminalSessionId
		terminalAccessSessionData.terminalAccessDataEntity = existingTerminalAccessData
		terminalAccessDataMap[terminalAccessId] = terminalAccessSessionData
		impl.TerminalAccessDataArrayMutex.Unlock()

	}
	return terminalAccessData, nil
}

func (impl *UserTerminalAccessServiceImpl) DeleteTerminalPod(ctx context.Context, clusterId int, terminalPodName string, namespace string) error {
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessPodTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching template", "template", models.TerminalAccessPodTemplateName, "err", err)
		return err
	}
	gvkDataString := terminalAccessPodTemplate.TemplateData
	err = impl.DeleteTerminalResource(ctx, clusterId, terminalPodName, gvkDataString, namespace)
	return err
}

func (impl *UserTerminalAccessServiceImpl) DeleteTerminalResource(ctx context.Context, clusterId int, terminalResourceName string, resourceTemplateString string, namespace string) error {
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(resourceTemplateString), nil, nil)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "resourceTemplateString", resourceTemplateString, "err", err)
		return err
	}

	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		return err
	}

	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name:      terminalResourceName,
			Namespace: namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   groupVersionKind.Group,
				Version: groupVersionKind.Version,
				Kind:    groupVersionKind.Kind,
			},
		},
	}
	_, err = impl.k8sClientService.DeleteResource(ctx, restConfig, k8sRequest)
	if err != nil {
		impl.Logger.Errorw("error occurred while deleting resource for pod", "podName", terminalResourceName, "err", err)
	}
	return err
}

func (impl *UserTerminalAccessServiceImpl) applyTemplate(ctx context.Context, clusterId int, gvkDataString string, templateData string, isUpdate bool, namespace string) error {
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(ctx, clusterId)
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
			Namespace: namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   groupVersionKind.Group,
				Version: groupVersionKind.Version,
				Kind:    groupVersionKind.Kind,
			},
		},
	}

	if isUpdate {
		k8sRequest.Patch = templateData
		_, err = impl.k8sClientService.UpdateResource(ctx, restConfig, k8sRequest)
	} else {
		_, err = impl.k8sClientService.CreateResource(ctx, restConfig, k8sRequest, templateData)
	}
	if err != nil {
		if errStatus, ok := err.(*k8sErrors.StatusError); !(ok && errStatus.Status().Reason == metav1.StatusReasonAlreadyExists) {
			impl.Logger.Errorw("error in creating resource", "err", err, "request", k8sRequest)
			return err
		}
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) getPodStatus(ctx context.Context, clusterId int, podName string, namespace string) (string, error) {
	response, err := impl.getPodManifest(ctx, clusterId, podName, namespace)
	if err != nil {
		if err.Error() == string(models.TerminalPodTerminated) {
			return string(models.TerminalPodTerminated), nil
		} else {
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

func (impl *UserTerminalAccessServiceImpl) getPodManifest(ctx context.Context, clusterId int, podName string, namespace string) (*application.ManifestResponse, error) {
	request, err := impl.getPodRequestBean(clusterId, podName, namespace)
	if err != nil {
		return nil, err
	}
	response, err := impl.k8sApplicationService.GetResource(ctx, request)
	if err != nil {
		if isResourceNotFoundErr(err) {
			return nil, errors.New(string(models.TerminalPodTerminated))
		} else {
			impl.Logger.Errorw("error occurred while fetching resource info for pod", "podName", podName)
			return nil, err
		}
	}
	return response, nil
}

func (impl *UserTerminalAccessServiceImpl) getPodRequestBean(clusterId int, podName string, namespace string) (*k8s.ResourceRequestBean, error) {
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessPodTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching template", "template", models.TerminalAccessPodTemplateName, "err", err)
		return nil, err
	}
	gvkDataString := terminalAccessPodTemplate.TemplateData
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(gvkDataString), nil, nil)
	if err != nil {
		impl.Logger.Errorw("error occurred while extracting data for gvk", "gvkDataString", gvkDataString, "err", err)
		return nil, err
	}
	request := &k8s.ResourceRequestBean{
		ClusterId: clusterId,
		AppIdentifier: &client.AppIdentifier{
			ClusterId: clusterId,
		},
		K8sRequest: &application.K8sRequestBean{
			ResourceIdentifier: application.ResourceIdentifier{
				Name:      podName,
				Namespace: namespace,
				GroupVersionKind: schema.GroupVersionKind{
					Group:   groupVersionKind.Group,
					Version: groupVersionKind.Version,
					Kind:    groupVersionKind.Kind,
				},
			},
		},
	}
	return request, nil
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

func (impl *UserTerminalAccessServiceImpl) deleteClusterTerminalTemplates(ctx context.Context, clusterId int, podName string, namespace string) {
	templateData, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessClusterRoleBindingTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access template", "err", err)
		return
	}
	templateName := strings.ReplaceAll(models.TerminalAccessClusterRoleBindingTemplate, models.TerminalAccessPodNameTemplate, podName)
	impl.DeleteTerminalResource(ctx, clusterId, templateName, templateData.TemplateData, namespace)

	templateData, err = impl.TerminalAccessRepository.FetchTerminalAccessTemplate(models.TerminalAccessServiceAccountTemplateName)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access template", "err", err)
		return
	}
	templateName = strings.ReplaceAll(models.TerminalAccessServiceAccountTemplate, models.TerminalAccessPodNameTemplate, podName)
	impl.DeleteTerminalResource(ctx, clusterId, templateName, templateData.TemplateData, namespace)
}

func (impl *UserTerminalAccessServiceImpl) FetchPodManifest(ctx context.Context, userTerminalAccessId int) (resp *application.ManifestResponse, err error) {
	terminalAccessData, err := impl.getTerminalAccessDataForId(userTerminalAccessId)
	if err != nil {
		return nil, errors.New("unable to fetch manifest")
	}
	if terminalAccessData.Status == string(models.TerminalPodTerminated) {
		return nil, errors.New("pod-terminated")
	}
	metadataMap, err := impl.getMetadataMap(terminalAccessData.Metadata)
	if err != nil {
		return nil, err
	}
	namespace := metadataMap["Namespace"]
	manifest, err := impl.getPodManifest(ctx, terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
	if err == errors.New(string(models.TerminalPodTerminated)) {
		return nil, errors.New("pod-terminated")
	}
	return manifest, err
}

func (impl *UserTerminalAccessServiceImpl) FetchPodEvents(ctx context.Context, userTerminalAccessId int) (*application.EventsResponse, error) {
	terminalAccessData, err := impl.getTerminalAccessDataForId(userTerminalAccessId)
	if err != nil {
		return nil, errors.New("unable to fetch pod event")
	}
	if terminalAccessData.Status == string(models.TerminalPodTerminated) {
		return nil, errors.New("pod-terminated")
	}
	metadataMap, err := impl.getMetadataMap(terminalAccessData.Metadata)
	if err != nil {
		return nil, err
	}
	namespace := metadataMap["Namespace"]
	podRequestBean, err := impl.getPodRequestBean(terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
	return impl.k8sApplicationService.ListEvents(ctx, podRequestBean)
}

func (impl *UserTerminalAccessServiceImpl) getTerminalAccessDataForId(userTerminalAccessId int) (*models.UserTerminalAccessData, error) {
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	terminalAccessSessionData, present := terminalAccessDataMap[userTerminalAccessId]
	var terminalAccessData *models.UserTerminalAccessData
	var err error
	if present {
		terminalAccessData = terminalAccessSessionData.terminalAccessDataEntity
	} else {
		terminalAccessData, err = impl.TerminalAccessRepository.GetUserTerminalAccessData(userTerminalAccessId)
		if err != nil {
			impl.Logger.Errorw("error occurred while fetching terminal access data ", "userTerminalAccessId", userTerminalAccessId, "err", err)
			return nil, err
		}
	}
	return terminalAccessData, err
}
