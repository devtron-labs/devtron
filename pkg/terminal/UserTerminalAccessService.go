package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type UserTerminalAccessService interface {
	StartTerminalSession(request *bean.UserTerminalSessionRequest) (*bean.UserTerminalSessionResponse, error)
	UpdateTerminalSession(request *bean.UserTerminalSessionRequest) (*bean.UserTerminalSessionResponse, error)
	DisconnectTerminalSession(userTerminalSessionId int) error
}

type UserTerminalAccessServiceImpl struct {
	TerminalAccessRepository repository.TerminalAccessRepository
	Logger                   *zap.SugaredLogger
	Config                   *bean.UserTerminalSessionConfig
	terminalAccessDataArray  []*models.UserTerminalAccessData
}

func NewUserTerminalAccessServiceImpl(logger *zap.SugaredLogger, terminalAccessRepository repository.TerminalAccessRepository) (*UserTerminalAccessServiceImpl, error) {
	config := &bean.UserTerminalSessionConfig{}
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}
	accessServiceImpl := &UserTerminalAccessServiceImpl{
		Logger:                   logger,
		TerminalAccessRepository: terminalAccessRepository,
		Config:                   config,
	}
	//fetches all running and starting entities from db
	return accessServiceImpl, nil
}

func (impl UserTerminalAccessServiceImpl) StartTerminalSession(request *bean.UserTerminalSessionRequest) (*bean.UserTerminalSessionResponse, error) {
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
		errStr := fmt.Sprintf("cannot create more session than configured %s", strconv.Itoa(maxSessionPerUser))
		impl.Logger.Errorw(errStr, "req", request)
		return nil, errors.New(errStr)
	}
	podName, err := impl.startTerminalPod(request, userRunningSessionCount)
	return impl.createTerminalEntity(request, podName, err)
}

func (impl UserTerminalAccessServiceImpl) createTerminalEntity(request *bean.UserTerminalSessionRequest, podName string, err error) (*bean.UserTerminalSessionResponse, error) {
	userAccessData := &models.UserTerminalAccessData{
		UserId:    request.UserId,
		ClusterId: request.ClusterId,
		NodeName:  request.NodeName,
		Status:    string(bean.Starting),
		PodName:   podName,
		Metadata:  impl.extractMetadataString(request),
	}
	err = impl.TerminalAccessRepository.SaveUserTerminalAccessData(userAccessData)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving user terminal access data", "err", err)
		return nil, err
	}
	return &bean.UserTerminalSessionResponse{
		UserTerminalSessionId: userAccessData.Id,
		UserId:                userAccessData.UserId,
		ShellName:             request.ShellName,
	}, nil
}

func (impl UserTerminalAccessServiceImpl) UpdateTerminalSession(request *bean.UserTerminalSessionRequest) (*bean.UserTerminalSessionResponse, error) {
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
	terminalAccessPodTemplate, err := impl.TerminalAccessRepository.FetchTerminalAccessTemplate(bean.TerminalAccessPodTemplateName)
	err = impl.applyTemplateData(request, terminalAccessPodTemplate.TemplateData, podName, terminalAccessPodTemplate.TemplateName)
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
	terminalPodStatus := bean.TerminalPodStatus(terminalStatus)
	if terminalPodStatus == bean.Terminated || terminalPodStatus == bean.Error {
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
	err = impl.TerminalAccessRepository.UpdateUserTerminalStatus(userTerminalSessionId, string(bean.Terminated))
	if err != nil {
		impl.Logger.Errorw("error occurred while updating terminal status in db", "userTerminalSessionId", userTerminalSessionId, "err", err)
	}
	return err
}

func (impl UserTerminalAccessServiceImpl) extractMetadataString(request *bean.UserTerminalSessionRequest) string {
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

func (impl UserTerminalAccessServiceImpl) startTerminalPod(request *bean.UserTerminalSessionRequest, runningCount int) (string, error) {
	podNameVar := impl.createPodName(request, runningCount)
	accessTemplates, err := impl.TerminalAccessRepository.FetchAllTemplates()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching terminal access templates", "err", err)
		return "", err
	}
	for _, accessTemplate := range accessTemplates {
		templateData := accessTemplate.TemplateData
		err = impl.applyTemplateData(request, templateData, podNameVar, accessTemplate.TemplateName)
		if err != nil {
			return "", err
		}
	}
	return podNameVar, err
}

func (impl UserTerminalAccessServiceImpl) createPodName(request *bean.UserTerminalSessionRequest, runningCount int) string {
	podNameVar := bean.TerminalAccessPodNameTemplate
	podNameVar = strings.ReplaceAll(podNameVar, bean.TerminalAccessClusterIdTemplateVar, strconv.Itoa(request.ClusterId))
	podNameVar = strings.ReplaceAll(podNameVar, bean.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	podNameVar = strings.ReplaceAll(podNameVar, bean.TerminalAccessRandomIdVar, strconv.Itoa(runningCount+1))
	return podNameVar
}

func (impl UserTerminalAccessServiceImpl) applyTemplateData(request *bean.UserTerminalSessionRequest, templateData string, podNameVar string, templateName string) error {
	clusterId := request.ClusterId
	templateData = strings.ReplaceAll(templateData, bean.TerminalAccessClusterIdTemplateVar, strconv.Itoa(clusterId))
	templateData = strings.ReplaceAll(templateData, bean.TerminalAccessUserIdTemplateVar, strconv.FormatInt(int64(request.UserId), 10))
	templateData = strings.ReplaceAll(templateData, bean.TerminalAccessPodNameVar, podNameVar)
	err := impl.applyTemplate(clusterId, templateData)
	if err != nil {
		impl.Logger.Errorw("error occurred while applying template ", "name", templateName, "err", err)
		return err
	}
	return nil
}

func (impl UserTerminalAccessServiceImpl) syncPodStatus() {
	// set starting/running pods in memory and fetch status of those pods and update their status in Db
	for _, terminalAccessData := range impl.terminalAccessDataArray {
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
		}
	}
}

func (impl UserTerminalAccessServiceImpl) DeleteTerminalPod(clusterId int, terminalPodName string) error {
	//make pod delete request, handle errors if pod does  not exists
	return nil
}

func (impl UserTerminalAccessServiceImpl) applyTemplate(clusterId int, templateData string) error {
	return nil
}

func (impl UserTerminalAccessServiceImpl) getPodStatus(clusterId int, podName string) (bean.TerminalPodStatus, error) {
	// return terminated if pod does not exist
	return bean.Running, nil
}
