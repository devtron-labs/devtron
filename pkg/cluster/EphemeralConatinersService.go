package cluster

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"time"
)

type EphemeralContainerRequest struct {
	BasicData    *EphemeralContainerBasicData    `json:"basicData"`
	AdvancedData *EphemeralContainerAdvancedData `json:"advancedData"`
	Namespace    string                          `json:"namespace"`
	ClusterId    int                             `json:"clusterId"`
	PodName      string                          `json:"podName"`
	UserId       int                             `json:"-"`
}

type EphemeralContainerAdvancedData struct {
	Manifest string `json:"manifest"`
}

type EphemeralContainerBasicData struct {
	ContainerName       string `json:"containerName"`
	TargetContainerName string `json:"targetContainerName"`
	Image               string `json:"image"`
}

type EphemeralContainerService interface {
	SaveEphemeralContainer(model EphemeralContainerRequest) error
	UpdateDeleteEphemeralContainer(model EphemeralContainerRequest, actionType int) error
	// send action type 1 in case of used and 2 in case of terminated
}

type EphemeralContainerServiceImpl struct {
	repository repository.EphemeralContainersRepository
}

func (impl *EphemeralContainerServiceImpl) SaveEphemeralContainer(model EphemeralContainerRequest) error {
	err := impl.repository.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = impl.repository.Rollback()
		} else {
			err = impl.repository.Commit()
		}
	}()

	container, err := impl.repository.FindContainerByName(model.ClusterId, model.Namespace, model.PodName, model.BasicData.ContainerName)
	if err != nil {
		return err
	}
	if container != nil {
		return errors.New("container already present in the provided pod")
	}
	bean := ConvertToEphemeralContainerBean(model)
	err = impl.repository.SaveData(&bean)
	if err != nil {
		return err
	}
	var auditLogBean repository.EphemeralContainerAction
	auditLogBean.EphemeralContainerID = bean.Id
	auditLogBean.ActionType = 0
	auditLogBean.PerformedAt = time.Now()
	auditLogBean.PerformedBy = model.UserId
	err = impl.repository.SaveAction(&auditLogBean)
	if err != nil {
		return err
	}
	return nil
}

func (impl *EphemeralContainerServiceImpl) UpdateDeleteEphemeralContainer(model EphemeralContainerRequest, actionType int) error {
	err := impl.repository.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = impl.repository.Rollback()
		} else {
			err = impl.repository.Commit()
		}
	}()

	container, err := impl.repository.FindContainerByName(model.ClusterId, model.Namespace, model.PodName, model.BasicData.ContainerName)
	if err != nil {
		return err
	}

	var auditLogBean repository.EphemeralContainerAction
	if container == nil {
		bean := ConvertToEphemeralContainerBean(model)
		bean.IsExternallyCreated = true
		err = impl.repository.SaveData(&bean)
		if err != nil {
			return err
		}
		auditLogBean.EphemeralContainerID = bean.Id
	} else {
		auditLogBean.EphemeralContainerID = container.Id
	}

	auditLogBean.ActionType = actionType
	auditLogBean.PerformedAt = time.Now()
	auditLogBean.PerformedBy = model.UserId

	err = impl.repository.SaveAction(&auditLogBean)
	if err != nil {
		return err
	}

	return nil
}

func ConvertToEphemeralContainerBean(request EphemeralContainerRequest) repository.EphemeralContainerBean {
	return repository.EphemeralContainerBean{
		Name:                request.BasicData.ContainerName,
		ClusterID:           request.ClusterId,
		Namespace:           request.Namespace,
		PodName:             request.PodName,
		TargetContainer:     request.BasicData.TargetContainerName,
		Config:              request.AdvancedData.Manifest,
		IsExternallyCreated: false,
	}
}
