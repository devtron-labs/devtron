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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	utils1 "github.com/devtron-labs/devtron/pkg/clusterTerminalAccess/clusterTerminalUtils"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"github.com/yannh/kubeconform/pkg/resource"
	"github.com/yannh/kubeconform/pkg/validator"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	FetchTerminalStatus(ctx context.Context, terminalAccessId int, namespace string, containerName string, shellName string) (*models.UserTerminalSessionResponse, error)
	StopTerminalSession(ctx context.Context, userTerminalAccessId int)
	DisconnectTerminalSession(ctx context.Context, userTerminalAccessId int) error
	DisconnectAllSessionsForUser(ctx context.Context, userId int32)
	FetchPodManifest(ctx context.Context, userTerminalAccessId int) (resp *application.ManifestResponse, err error)
	FetchPodEvents(ctx context.Context, userTerminalAccessId int) (*models.UserTerminalPodEvents, error)
	ValidateShell(podName, namespace, shellName, containerName string, clusterId int) (bool, string, error)
	EditTerminalPodManifest(ctx context.Context, request *models.UserTerminalSessionRequest, override bool) (ManifestEditResponse, error)
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
	K8sCapacityService           k8s.K8sCapacityService
}

type UserTerminalAccessSessionData struct {
	sessionId                string
	latestActivityTime       time.Time
	terminalAccessDataEntity *models.UserTerminalAccessData
	terminateTriggered       bool
}
type ManifestEditResponse struct {
	ErrorComments    string                        `json:"errors,omitempty"`
	ManifestResponse *application.ManifestResponse `json:"manifestResponse"`
	models.UserTerminalSessionResponse
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
	k8sApplicationService k8s.K8sApplicationService, k8sClientService application.K8sClientService, terminalSessionHandler terminal.TerminalSessionHandler,
	K8sCapacityService k8s.K8sCapacityService) (*UserTerminalAccessServiceImpl, error) {
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
		K8sCapacityService:           K8sCapacityService,
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
func (impl *UserTerminalAccessServiceImpl) ValidateShell(podName, namespace, shellName, containerName string, clusterId int) (bool, string, error) {
	impl.Logger.Infow("Inside validateShell method", "UserTerminalAccessServiceImpl")
	if containerName == "" {
		containerName = "devtron-debug-terminal"
	}
	req := &terminal.TerminalSessionRequest{
		PodName:       podName,
		Namespace:     namespace,
		Shell:         shellName,
		ClusterId:     clusterId,
		ContainerName: containerName,
	}
	if shellName == models.AutoSelectShell {
		shell, err := impl.terminalSessionHandler.AutoSelectShell(req)
		if err != nil {
			return false, shell, err
		}
		return true, shell, err
	}
	res, err := impl.terminalSessionHandler.ValidateShell(req)
	if err != nil && err.Error() == terminal.CommandExecutionFailed {
		return res, shellName, errors.New(fmt.Sprintf(models.ShellNotSupported, shellName))
	}
	return res, shellName, err
}
func (impl *UserTerminalAccessServiceImpl) StartTerminalSession(ctx context.Context, request *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {
	impl.Logger.Infow("terminal start request received for user", "request", request)
	//if request.Manifest not empty, requested from edit-manifest page to start terminal session with edited manifest.
	if request.Manifest != "" && !request.DebugNode {
		res, err := impl.EditTerminalPodManifest(ctx, request, true)
		return &models.UserTerminalSessionResponse{
			TerminalAccessId:      res.TerminalAccessId,
			UserTerminalSessionId: res.UserTerminalSessionId,
			Status:                res.Status,
			PodName:               res.PodName,
			NodeName:              res.NodeName,
			ShellName:             res.ShellName,
			Containers:            res.Containers,
			PodExists:             res.PodExists,
		}, err
	}
	//if DebugNode is true requested to start terminal session for node-debug pod
	if request.DebugNode {
		return impl.StartNodeDebug(request)
	}

	//start terminal session
	podNameVar, err := impl.getPodNameVar(request)
	if err != nil {
		return nil, err
	}
	terminalEntity, err := impl.createTerminalEntity(request, *podNameVar)
	if err != nil {
		return nil, err
	}
	isAutoSelect := false
	if request.NodeName == models.AUTO_SELECT_NODE {
		isAutoSelect = true
	}
	err = impl.startTerminalPod(ctx, terminalEntity.PodName, request, isAutoSelect)
	terminalEntity.DebugNode = false
	return terminalEntity, err
}

func (impl *UserTerminalAccessServiceImpl) getPodNameVar(request *models.UserTerminalSessionRequest) (*string, error) {
	userId := request.UserId
	// check for max session check
	err := impl.checkMaxSessionLimit(userId)
	if err != nil {
		return nil, err
	}
	maxIdForUser := impl.getMaxIdForUser(userId)
	podNameVar := impl.createPodName(request, maxIdForUser)
	return &podNameVar, err
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
	updateTerminalShellResponse := &models.UserTerminalSessionResponse{
		UserId:           terminalAccessData.UserId,
		PodName:          terminalAccessData.PodName,
		TerminalAccessId: terminalAccessData.Id,
		ShellName:        request.ShellName,
	}
	statusAndReason := strings.Split(terminalAccessData.Status, "/")
	if statusAndReason[0] == string(models.TerminalPodTerminated) {
		updateTerminalShellResponse.Status = models.TerminalPodTerminated
		updateTerminalShellResponse.ErrorReason = statusAndReason[1]
		return updateTerminalShellResponse, nil
	}

	if models.TerminalPodStatus(terminalAccessData.Status) == models.TerminalPodRunning {
		isValidShell, shellName, err := impl.ValidateShell(terminalAccessData.PodName, request.NameSpace, request.ShellName, request.ContainerName, terminalAccessData.ClusterId)
		podStatus := models.TerminalPodStatus(terminalAccessData.Status)
		if err != nil && err.Error() == terminal.PodNotFound {
			podStatus = models.TerminalPodTerminated
		}
		if !isValidShell {
			impl.Logger.Infow("shell is not supported", "podName", terminalAccessData.PodName, "namespace", request.NameSpace, "shell", request.ShellName, "reason", err)
			updateTerminalShellResponse.Status = podStatus
			updateTerminalShellResponse.ErrorReason = err.Error()
			updateTerminalShellResponse.IsValidShell = isValidShell
			//have to get shellName from validate shell , because we can auto-select the shell
			updateTerminalShellResponse.ShellName = shellName
			return updateTerminalShellResponse, nil
		}
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

	updateTerminalShellResponse.IsValidShell = true
	updateTerminalShellResponse.Status = models.TerminalPodStatus(statusAndReason[0])
	return updateTerminalShellResponse, nil
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
	if accessSessionDataMap == nil {
		return nil
	}
	accessSessionData := accessSessionDataMap[userTerminalAccessId]
	if accessSessionData == nil {
		return nil
	}
	terminalAccessData := accessSessionData.terminalAccessDataEntity
	if terminalAccessData == nil {
		return nil
	}
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

func getErrorDetailedMessage(err error) string {
	if errStatus, ok := err.(*k8sErrors.StatusError); ok {
		return errStatus.Status().Message
	}
	return ""
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
	if accessSessionDataMap == nil {
		return
	}
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

func (impl *UserTerminalAccessServiceImpl) startTerminalPod(ctx context.Context, podNameVar string, request *models.UserTerminalSessionRequest, isAutoSelect bool) error {

	accessTemplates, err := impl.TerminalAccessRepository.FetchAllTemplates()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access templates", "err", err)
		return err
	}
	for _, accessTemplate := range accessTemplates {
		//do not apply the node debug pod template
		if accessTemplate.TemplateName == utils1.TerminalNodeDebugPodName {
			continue
		}
		err = impl.applyTemplateData(ctx, request, podNameVar, accessTemplate, false, isAutoSelect)
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

//template data use kubernetes object
func (impl *UserTerminalAccessServiceImpl) applyTemplateData(ctx context.Context, request *models.UserTerminalSessionRequest, podNameVar string,
	terminalTemplate *models.TerminalAccessTemplates, isUpdate bool, isAutoSelect bool) error {
	templateName := terminalTemplate.TemplateName
	templateData := terminalTemplate.TemplateData
	clusterId := request.ClusterId
	namespace := request.Namespace
	templateData = strings.ReplaceAll(templateData, models.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	templateData, err := utils1.ReplaceTemplateData(templateData, podNameVar, namespace, request.NodeName, request.BaseImage, isAutoSelect, request.NodeTaints)
	if err != nil {
		impl.Logger.Errorw("error occurred while updating template data", "name", templateName, "err", err)
		return err
	}

	err = impl.applyTemplate(ctx, clusterId, terminalTemplate.TemplateData, templateData, isUpdate, namespace)
	if err != nil {
		impl.Logger.Errorw("error occurred while applying template ", "name", templateName, "err", err)
		return err
	}
	return nil
}

func (impl *UserTerminalAccessServiceImpl) SyncPodStatus() {
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	impl.TerminalAccessDataArrayMutex.Lock()
	defer impl.TerminalAccessDataArrayMutex.Unlock()
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
					errorDetailedMessage := getErrorDetailedMessage(err)
					terminalPodStatusString = fmt.Sprintf("%s/%s", string(models.TerminalPodTerminated), errorDetailedMessage)
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

	for _, terminalAccessSessionData := range terminalAccessDataMap {
		terminalAccessData := terminalAccessSessionData.terminalAccessDataEntity
		if terminalAccessData.Status != string(models.TerminalPodStarting) && terminalAccessData.Status != string(models.TerminalPodRunning) {
			// check if this is the last data for this cluster and user then delete terminal resource
			delete(terminalAccessDataMap, terminalAccessData.Id)
		}
	}
	impl.TerminalAccessSessionDataMap = &terminalAccessDataMap
}

func (impl *UserTerminalAccessServiceImpl) checkAndStartSession(ctx context.Context, terminalAccessData *models.UserTerminalAccessData) (string, string, error) {
	clusterId := terminalAccessData.ClusterId
	terminalAccessPodName := terminalAccessData.PodName
	metadata := terminalAccessData.Metadata
	metadataMap, err := impl.getMetadataMap(metadata)
	if err != nil {
		return "", "", err
	}
	namespace := metadataMap["Namespace"]
	terminalPodStatusString, nodeName, err := impl.getPodStatus(ctx, clusterId, terminalAccessPodName, namespace)
	if err != nil {
		return "", "", err
	}
	sessionID := ""
	terminalAccessId := terminalAccessData.Id
	if terminalPodStatusString == string(models.TerminalPodRunning) {
		err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(terminalAccessId, terminalPodStatusString)
		if err != nil {
			impl.Logger.Errorw("error occurred while updating terminal status", "terminalAccessId", terminalAccessId, "err", err)
			return "", "", err
		}
		impl.TerminalAccessDataArrayMutex.Lock()
		terminalAccessData.Status = terminalPodStatusString
		impl.TerminalAccessDataArrayMutex.Unlock()
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
			return "", "", err
		}
		sessionID = terminalMessage.SessionID
	}
	return sessionID, nodeName, err
}

func (impl *UserTerminalAccessServiceImpl) FetchTerminalStatus(ctx context.Context, terminalAccessId int, namespace string, containerName string, shellName string) (*models.UserTerminalSessionResponse, error) {
	terminalAccessDataMap := *impl.TerminalAccessSessionDataMap
	terminalAccessSessionData, present := terminalAccessDataMap[terminalAccessId]
	var terminalSessionId = ""
	var terminalAccessData *models.UserTerminalAccessData
	if present {
		if terminalAccessSessionData.terminateTriggered {
			accessDataEntity := terminalAccessSessionData.terminalAccessDataEntity
			response := &models.UserTerminalSessionResponse{
				TerminalAccessId:      terminalAccessId,
				UserId:                accessDataEntity.UserId,
				Status:                models.TerminalPodStatus(accessDataEntity.Status),
				PodName:               accessDataEntity.PodName,
				UserTerminalSessionId: terminalSessionId,
				ShellName:             shellName,
			}
			if models.TerminalPodStatus(accessDataEntity.Status) == models.TerminalPodRunning {
				isValid, _, err := impl.ValidateShell(accessDataEntity.PodName, namespace, shellName, containerName, accessDataEntity.ClusterId)
				response.IsValidShell = isValid
				if err != nil {
					if err.Error() == terminal.PodNotFound {
						response.Status = models.TerminalPodTerminated
						impl.TerminalAccessDataArrayMutex.Lock()
						terminalAccessSessionData.terminalAccessDataEntity.Status = fmt.Sprintf("%s/%s", models.TerminalPodTerminated, terminal.PodNotFound)
						impl.TerminalAccessDataArrayMutex.Unlock()
					}
					response.ErrorReason = err.Error()
				}
			}
			return response, nil
		} else {
			terminalSessionId = terminalAccessSessionData.sessionId
			validSession := impl.terminalSessionHandler.ValidateSession(terminalSessionId)
			if validSession {
				impl.TerminalAccessDataArrayMutex.Lock()
				terminalAccessData = terminalAccessSessionData.terminalAccessDataEntity
				impl.TerminalAccessDataArrayMutex.Unlock()
			}
		}
	}
	terminalAccessData, err := impl.validateTerminalAccessFromDb(ctx, terminalAccessId, terminalAccessData, terminalSessionId, terminalAccessSessionData, terminalAccessDataMap)
	if err != nil {
		if strings.Contains(err.Error(), "pod-terminated") {
			return &models.UserTerminalSessionResponse{
				TerminalAccessId: terminalAccessId,
				Status:           models.TerminalPodTerminated,
				ErrorReason:      err.Error(),
			}, nil
		} else {
			return nil, err
		}
	}
	terminalAccessDataId := terminalAccessData.Id
	terminalAccessResponse := &models.UserTerminalSessionResponse{
		TerminalAccessId:      terminalAccessDataId,
		UserId:                terminalAccessData.UserId,
		Status:                models.TerminalPodStatus(terminalAccessData.Status),
		PodName:               terminalAccessData.PodName,
		UserTerminalSessionId: terminalSessionId,
		NodeName:              terminalAccessData.NodeName,
		ShellName:             shellName,
	}
	if models.TerminalPodStatus(terminalAccessData.Status) == models.TerminalPodRunning {
		isValid, _, err := impl.ValidateShell(terminalAccessData.PodName, namespace, shellName, containerName, terminalAccessData.ClusterId)
		terminalAccessResponse.IsValidShell = isValid
		if err != nil {
			if err.Error() == terminal.PodNotFound {
				terminalAccessResponse.Status = models.TerminalPodTerminated
				impl.TerminalAccessDataArrayMutex.Lock()
				terminalAccessSessionData.terminalAccessDataEntity.Status = fmt.Sprintf("%s/%s", models.TerminalPodTerminated, terminal.PodNotFound)
				impl.TerminalAccessDataArrayMutex.Unlock()
			}
			terminalAccessResponse.ErrorReason = err.Error()
		}
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
		statusAndReason := strings.Split(existingTerminalAccessData.Status, "/")
		if statusAndReason[0] == string(models.TerminalPodTerminated) {
			impl.TerminalAccessDataArrayMutex.Lock()
			if terminalAccessSessionData != nil && terminalAccessSessionData.terminalAccessDataEntity != nil {
				terminalAccessSessionData.terminalAccessDataEntity.Status = string(models.TerminalPodTerminated)
			}
			impl.TerminalAccessDataArrayMutex.Unlock()
			return nil, errors.New(fmt.Sprintf("pod-terminated(%s)", statusAndReason[1]))
		}
		err = impl.checkMaxSessionLimit(existingTerminalAccessData.UserId)
		if err != nil {
			return nil, err
		}
		var nodeName = terminalAccessData.NodeName
		terminalSessionId, nodeName, err = impl.checkAndStartSession(ctx, existingTerminalAccessData)
		if err != nil {
			return nil, err
		}
		if nodeName != "" {
			terminalAccessData.NodeName = nodeName
			existingTerminalAccessData.NodeName = nodeName
		}
		impl.TerminalAccessDataArrayMutex.Lock()
		if terminalAccessSessionData == nil {
			terminalAccessSessionData = &UserTerminalAccessSessionData{}
		}
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

func (impl *UserTerminalAccessServiceImpl) getPodStatus(ctx context.Context, clusterId int, podName string, namespace string) (string, string, error) {
	response, err := impl.getPodManifest(ctx, clusterId, podName, namespace)
	if err != nil {
		statusReason := strings.Split(err.Error(), "/")
		if statusReason[0] == string(models.TerminalPodTerminated) {
			return err.Error(), "", nil
		} else {
			return "", "", err
		}
	}
	status := ""
	nodeName := ""
	if response != nil {
		manifest := response.Manifest
		for key, value := range manifest.Object {
			if key == "status" {
				statusData := value.(map[string]interface{})
				status = statusData["phase"].(string)
			}
			if key == "spec" {
				specData := value.(map[string]interface{})
				if _, ok := specData["nodeName"]; ok {
					nodeName = specData["nodeName"].(string)
				}
			}
		}
	}
	impl.Logger.Debug("pod status", "podName", podName, "status", status)
	return status, nodeName, nil
}

func (impl *UserTerminalAccessServiceImpl) getPodManifest(ctx context.Context, clusterId int, podName string, namespace string) (*application.ManifestResponse, error) {
	request, err := impl.getPodRequestBean(clusterId, podName, namespace)
	if err != nil {
		return nil, err
	}
	response, err := impl.k8sApplicationService.GetResource(ctx, request)
	if err != nil {
		if isResourceNotFoundErr(err) {
			errorDetailedMessage := getErrorDetailedMessage(err)
			terminalPodStatusString := fmt.Sprintf("%s/%s", string(models.TerminalPodTerminated), errorDetailedMessage)
			return nil, errors.New(terminalPodStatusString)
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
	statusReason := strings.Split(terminalAccessData.Status, "/")
	if statusReason[0] == string(models.TerminalPodTerminated) {
		return nil, errors.New(fmt.Sprintf("pod-terminated(%s)", statusReason[1]))
	}
	metadataMap, err := impl.getMetadataMap(terminalAccessData.Metadata)
	if err != nil {
		return nil, err
	}
	namespace := metadataMap["Namespace"]
	manifest, err := impl.getPodManifest(ctx, terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
	if err != nil {
		statusReason = strings.Split(err.Error(), "/")
		if statusReason[0] == string(models.TerminalPodTerminated) {
			return nil, errors.New(fmt.Sprintf("pod-terminated(%s)", statusReason[1]))
		}
	}

	return manifest, err
}

func (impl *UserTerminalAccessServiceImpl) FetchPodEvents(ctx context.Context, userTerminalAccessId int) (*models.UserTerminalPodEvents, error) {
	terminalAccessData, err := impl.getTerminalAccessDataForId(userTerminalAccessId)
	if err != nil {
		return nil, errors.New("unable to fetch pod event")
	}

	metadataMap, err := impl.getMetadataMap(terminalAccessData.Metadata)
	if err != nil {
		return nil, err
	}
	namespace := metadataMap["Namespace"]
	podRequestBean, err := impl.getPodRequestBean(terminalAccessData.ClusterId, terminalAccessData.PodName, namespace)
	podEvents, err := impl.k8sApplicationService.ListEvents(ctx, podRequestBean)
	status := string(terminalAccessData.Status)
	statusReason := strings.Split(terminalAccessData.Status, "/")
	errorReason := ""
	if statusReason[0] == string(models.TerminalPodTerminated) {
		status = string(models.TerminalPodTerminated)
		errorReason = fmt.Sprintf("pod-terminated(%s)", statusReason[1])
	}
	return &models.UserTerminalPodEvents{
		EventsResponse: interface{}(podEvents),
		ErrorReason:    errorReason,
		Status:         status,
	}, err
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

func (impl *UserTerminalAccessServiceImpl) EditTerminalPodManifest(ctx context.Context, editManifestRequest *models.UserTerminalSessionRequest, override bool) (ManifestEditResponse, error) {

	manifestRequest := editManifestRequest.Manifest
	userTerminalAccessId := editManifestRequest.Id
	impl.Logger.Infow("Reached EditPodManifest method", "userTerminalAccessId", userTerminalAccessId, "manifest", manifestRequest)

	result := ManifestEditResponse{}

	manifestResponse := &application.ManifestResponse{}
	manifestMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(manifestRequest), &manifestMap)
	if err != nil {
		impl.Logger.Errorw("error in unmarshalling manifest request", "err", err)
		return result, err
	}

	manifestResponse.Manifest.SetUnstructuredContent(manifestMap)
	result.ManifestResponse = manifestResponse

	//return if not a pod yaml
	if manifestMap != nil {
		if manifestMap["kind"] != "Pod" {
			err := errors.New("manifest should be of kind \"Pod\"")
			impl.Logger.Errorw("given manifest is not a pod manifest", "manifest", manifestMap, "err", err)
			return result, err
		}
	}

	//construct validator
	v, err := validator.New(nil, validator.Opts{Strict: true})
	if err != nil {
		impl.Logger.Errorw("failed initializing validator", "err", err)
		return result, err
	}
	//construct validate request
	YamlResource := resource.Resource{
		Bytes: []byte(manifestRequest),
	}
	//validate the yaml
	validatorResponse := v.ValidateResource(YamlResource)
	if validatorResponse.Err != nil {
		result.ManifestResponse = manifestResponse
		result.ErrorComments = validatorResponse.Err.Error()
		return result, nil
	}
	// valid pod yaml found
	impl.Logger.Infow("pod manifest yaml validated using \"kubeconform\" validator", "podManifest", manifestMap)

	//convert manifestMap to v1.Pod object
	podObject := v1.Pod{}
	err = runtime.DefaultUnstructuredConverter.
		FromUnstructured(manifestResponse.Manifest.Object, &podObject)
	if err != nil {
		impl.Logger.Errorw("error in converting manifest request to k8s Pod object", "userTerminalAccessId", userTerminalAccessId, "err", err, "manifest", manifestRequest)
		return result, err
	}

	if override {
		//override pod variables with requested variables
		podObject.Namespace = editManifestRequest.Namespace
		if editManifestRequest.NodeName != models.AUTO_SELECT_NODE {
			podObject.Spec.NodeName = editManifestRequest.NodeName
			for _, taint := range editManifestRequest.NodeTaints {
				podObject.Spec.Tolerations = append(podObject.Spec.Tolerations, v1.Toleration{
					Effect:   v1.TaintEffect(taint.Effect),
					Key:      taint.Key,
					Operator: v1.TolerationOpExists,
				})
			}
		}
		//set base image to the requested container
		for i, container := range podObject.Spec.Containers {
			if container.Name == editManifestRequest.ContainerName {
				podObject.Spec.Containers[i].Image = editManifestRequest.BaseImage
			}
		}

	}

	if podObject.Spec.NodeName != "" {
		_, err := impl.K8sCapacityService.GetNode(context.Background(), editManifestRequest.ClusterId, podObject.Spec.NodeName)
		if err != nil {
			impl.Logger.Errorw("failed to get node details for requested node", "err", err, "userId", editManifestRequest.UserId, "nodeName", podObject.Spec.NodeName, "clusterId", editManifestRequest.ClusterId)
			return result, err
		}
	}
	if len(podObject.Namespace) == 0 {
		podObject.Namespace = utils1.DefaultNamespace
	}
	terminalAccessData, err := impl.getTerminalAccessDataForId(userTerminalAccessId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while fetching user terminal access data", "userTerminalAccessId", userTerminalAccessId, "err", err)
		return result, err
	}
	if terminalAccessData == nil || podObject.Name != terminalAccessData.PodName {
		if !editManifestRequest.ForceDelete && impl.checkOtherPodExists(ctx, podObject.Name, podObject.Namespace, editManifestRequest.ClusterId) {
			result.PodExists = true
			result.PodName = podObject.Name
			result.NameSpace = podObject.Namespace
			//log the pod exists info
			//return, to warn user that pod with this name already exists in the given namespace
			return result, nil
		}
	}

	//delete if (user chooses to force delete) or (old pod and new pod have same name)
	//if reached this point, force delete the existing pod and create new
	impl.forceDeletePod(ctx, podObject.Name, podObject.Namespace, editManifestRequest.ClusterId, editManifestRequest.UserId)

	//determine request to  createTerminalEntity method
	editManifestRequest.NodeName = podObject.Spec.NodeName
	editManifestRequest.Namespace = podObject.Namespace
	terminalStartResponse, err := impl.createTerminalEntity(editManifestRequest, podObject.Name)
	if err != nil {
		impl.Logger.Errorw("failed to create terminal entity", "userTerminalAccessId", userTerminalAccessId, "err", err)
		return result, err
	}
	//delete resourceVersion before applying the pod yaml got from user, we are deleting this field because we never update the pod,we always create a new pod.
	podObject.ResourceVersion = ""
	//create podTemplate from PodObject
	podTemplateBytes, err := json.Marshal(&podObject)
	podTemplate := string(podTemplateBytes)
	//start new session with provided Pod manifest
	err = impl.applyTemplate(ctx, editManifestRequest.ClusterId, podTemplate, podTemplate, false, editManifestRequest.Namespace)
	if err != nil {
		impl.Logger.Errorw("failed to start terminal session", "userTerminalAccessId", userTerminalAccessId, "err", err)
		//send podObject data
		return result, err
	}
	result.PodExists = false
	result.DebugNode = utils1.IsNodeDebugPod(&podObject)
	var containers []models.Container
	for _, con := range podObject.Spec.Containers {
		containers = append(containers, models.Container{
			ContainerName: con.Name,
			Image:         con.Image,
		})
	}
	result.TerminalAccessId = terminalStartResponse.TerminalAccessId
	result.Containers = containers
	result.PodName = terminalStartResponse.PodName
	result.NodeName = terminalStartResponse.NodeName
	result.ShellName = editManifestRequest.ShellName
	result.Status = ""
	result.NodeName = podObject.Spec.NodeName
	result.NameSpace = podObject.Namespace
	return result, nil
}

func (impl *UserTerminalAccessServiceImpl) checkOtherPodExists(ctx context.Context, podName, namespace string, clusterId int) bool {
	podRequestBean, _ := impl.getPodRequestBean(clusterId, podName, namespace)
	res, _ := impl.k8sApplicationService.GetResource(ctx, podRequestBean)
	if res != nil {
		return true
	}
	return false
}

func (impl *UserTerminalAccessServiceImpl) forceDeletePod(ctx context.Context, podName, namespace string, clusterId int, userId int32) bool {
	//add grace period 0 to force delete
	podRequestBean, err := impl.getPodRequestBean(clusterId, podName, namespace)
	if err != nil {
		impl.Logger.Errorw("error occurred in getting the pod request bean", "podName", podName, "nameSpace", namespace, "clusterId", clusterId)
		return false
	}
	podRequestBean.K8sRequest.ForceDelete = true
	_, err = impl.k8sApplicationService.DeleteResource(ctx, podRequestBean, userId)
	if err != nil && !isResourceNotFoundErr(err) {
		return false
	}
	return true
}

func (impl *UserTerminalAccessServiceImpl) StartNodeDebug(userTerminalRequest *models.UserTerminalSessionRequest) (*models.UserTerminalSessionResponse, error) {

	if userTerminalRequest.NodeName == models.AUTO_SELECT_NODE {
		return nil, errors.New("node-name is not valid, node-name : " + userTerminalRequest.NodeName)
	}
	if userTerminalRequest.NodeName == "" || userTerminalRequest.ShellName == "" {
		return nil, errors.New("node-name or shell cannot be empty, node-name : " + userTerminalRequest.NodeName + ", shell : " + userTerminalRequest.ShellName)
	}
	nodeInfo, err := impl.K8sCapacityService.GetNode(context.Background(), userTerminalRequest.ClusterId, userTerminalRequest.NodeName)
	if err != nil {
		impl.Logger.Errorw("failed to get node details for requested node", "err", err, "userId", userTerminalRequest.UserId, "nodeName", userTerminalRequest.NodeName, "clusterId", userTerminalRequest.ClusterId)
		return nil, err
	}
	taints := make([]models.NodeTaints, 0)
	for _, taint := range nodeInfo.Spec.Taints {
		taints = append(taints, models.NodeTaints{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: string(taint.Effect),
		})
	}
	userTerminalRequest.NodeTaints = taints
	podObject, err := impl.GenerateNodeDebugPod(userTerminalRequest)
	if err != nil {
		impl.Logger.Errorw("failed to create node-debug pod", "err", err, "userId", userTerminalRequest.UserId, "userTerminalRequest", userTerminalRequest)
		return nil, err
	}
	result, err := impl.createTerminalEntity(userTerminalRequest, podObject.Name)
	if err != nil {
		impl.Logger.Errorw("failed to create terminal entity", "err", err, "userId", userTerminalRequest.UserId, "userTerminalRequest", userTerminalRequest)
		return nil, err
	}
	result.PodExists = false
	result.DebugNode = true
	var containers []models.Container
	for _, con := range podObject.Spec.Containers {
		containers = append(containers, models.Container{
			ContainerName: con.Name,
			Image:         con.Image,
		})
	}
	result.ShellName = userTerminalRequest.ShellName
	result.Containers = containers
	result.Status = ""
	result.NodeName = podObject.Spec.NodeName
	return result, nil
}

func (impl *UserTerminalAccessServiceImpl) GenerateNodeDebugPod(o *models.UserTerminalSessionRequest) (*v1.Pod, error) {
	nodeName := o.NodeName
	pn := fmt.Sprintf("node-debugger-%s-%s", nodeName, util.Generate(5))

	impl.Logger.Infow("Creating node debugging pod ", "podName", pn, "nodeName", nodeName)
	debugNodePodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(utils1.TerminalNodeDebugPodName)
	if err != nil {
		impl.Logger.Errorw("error in fetching debugNodePodTemplate by name from terminal_access_templates table ", "template_name", utils1.TerminalNodeDebugPodName, "err", err)
		return nil, err
	}

	serviceAccountTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(utils1.TerminalAccessServiceAccount)
	if err != nil {
		impl.Logger.Errorw("error in fetching debugNodePodTemplate by name from terminal_access_templates table ", "template_name", utils1.TerminalNodeDebugPodName, "err", err)
		return nil, err
	}

	clusterRoleBindingTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(utils1.TerminalAccessRoleBinding)
	if err != nil {
		impl.Logger.Errorw("error in fetching debugNodePodTemplate by name from terminal_access_templates table ", "template_name", utils1.TerminalNodeDebugPodName, "err", err)
		return nil, err
	}

	debugPod := &v1.Pod{}

	podTemplate, err := utils1.ReplaceTemplateData(debugNodePodTemplate.TemplateData, pn, o.Namespace, nodeName, o.BaseImage, false, o.NodeTaints)
	if err != nil {
		impl.Logger.Errorw("error in converting pod object into pod yaml", "err", err, "pod", debugPod)
		return debugPod, err
	}
	err = json.Unmarshal([]byte(podTemplate), &debugPod)
	if err != nil {
		impl.Logger.Errorw("error occurred while unmarshaling template data into coreV1 Pod Object", "template_name", utils1.TerminalNodeDebugPodName, "err", err)
		return nil, errors.New("internal server error occurred while creating node debug pod")
	}
	SATemplate, err := utils1.ReplaceTemplateData(serviceAccountTemplate.TemplateData, pn, o.Namespace, "", "", false, nil)
	if err != nil {
		impl.Logger.Errorw("error in converting pod object into pod yaml", "err", err, "pod", debugPod)
		return debugPod, err
	}
	err = impl.applyTemplate(context.Background(), o.ClusterId, SATemplate, SATemplate, false, o.Namespace)
	if err != nil {
		return debugPod, err
	}
	RoleBindingTemplate, err := utils1.ReplaceTemplateData(clusterRoleBindingTemplate.TemplateData, pn, o.Namespace, "", "", false, nil)
	if err != nil {
		impl.Logger.Errorw("error in converting pod object into pod yaml", "err", err, "pod", debugPod)
		return debugPod, err
	}
	err = impl.applyTemplate(context.Background(), o.ClusterId, RoleBindingTemplate, RoleBindingTemplate, false, o.Namespace)
	if err != nil {
		return debugPod, err
	}

	err = impl.applyTemplate(context.Background(), o.ClusterId, podTemplate, podTemplate, false, o.Namespace)
	return debugPod, err
}
