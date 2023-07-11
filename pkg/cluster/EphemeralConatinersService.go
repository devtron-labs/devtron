package cluster

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

const (
	ActionCreate   = 0
	ActionAccessed = 1
	ActionDelete   = 2
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
	SaveEphemeralContainer(tx *pg.Tx, model EphemeralContainerRequest) error
	UpdateDeleteEphemeralContainer(tx *pg.Tx, model EphemeralContainerRequest, actionType int) error
	// send action type 1 in case of used and 2 in case of terminated
}

type EphemeralContainerServiceImpl struct {
	repository repository.EphemeralContainersRepository
	logger     *zap.SugaredLogger
}

func (impl *EphemeralContainerServiceImpl) SaveEphemeralContainer(tx *pg.Tx, model EphemeralContainerRequest) error {

	container, err := impl.repository.FindContainerByName(model.ClusterId, model.Namespace, model.PodName, model.BasicData.ContainerName)
	if err != nil {
		_ = tx.Rollback() // Rollback the transaction if an error occurs during FindContainerByName
		return err
	}
	if container != nil {
		_ = tx.Rollback() // Rollback the transaction if the container already exists in the provided pod
		impl.logger.Errorw("Container already present in the provided pod")
		return errors.New("container already present in the provided pod")
	}

	bean := ConvertToEphemeralContainerBean(model)
	err = impl.repository.SaveData(tx, &bean)
	if err != nil {
		_ = tx.Rollback() // Rollback the transaction if an error occurs during SaveData
		impl.logger.Errorw("Failed to save ephemeral container", "error", err)
		return err
	}

	var auditLogBean repository.EphemeralContainerAction
	auditLogBean.EphemeralContainerId = bean.Id
	auditLogBean.ActionType = ActionCreate
	auditLogBean.PerformedAt = time.Now()
	auditLogBean.PerformedBy = model.UserId
	err = impl.repository.SaveAction(tx, &auditLogBean)
	if err != nil {
		_ = tx.Rollback() // Rollback the transaction if an error occurs during SaveAction
		impl.logger.Errorw("Failed to save ephemeral container", "error", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (impl *EphemeralContainerServiceImpl) UpdateDeleteEphemeralContainer(tx *pg.Tx, model EphemeralContainerRequest, actionType int) error {

	container, err := impl.repository.FindContainerByName(model.ClusterId, model.Namespace, model.PodName, model.BasicData.ContainerName)
	if err != nil {
		_ = tx.Rollback() // Rollback the transaction if an error occurs during FindContainerByName
		return err
	}

	var auditLogBean repository.EphemeralContainerAction
	if container == nil {
		bean := ConvertToEphemeralContainerBean(model)
		bean.IsExternallyCreated = true
		err = impl.repository.SaveData(tx, &bean)
		if err != nil {
			_ = tx.Rollback() // Rollback the transaction if an error occurs during SaveData
			impl.logger.Errorw("Failed to save ephemeral container", "error", err)
			return err
		}
		auditLogBean.EphemeralContainerId = bean.Id
	} else {
		auditLogBean.EphemeralContainerId = container.Id
	}

	auditLogBean.ActionType = actionType
	auditLogBean.PerformedAt = time.Now()
	auditLogBean.PerformedBy = model.UserId

	err = impl.repository.SaveAction(tx, &auditLogBean)
	if err != nil {
		_ = tx.Rollback() // Rollback the transaction if an error occurs during SaveAction
		impl.logger.Errorw("Failed to save ephemeral container", "error", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func ConvertToEphemeralContainerBean(request EphemeralContainerRequest) repository.EphemeralContainerBean {
	return repository.EphemeralContainerBean{
		Name:                request.BasicData.ContainerName,
		ClusterId:           request.ClusterId,
		Namespace:           request.Namespace,
		PodName:             request.PodName,
		TargetContainer:     request.BasicData.TargetContainerName,
		Config:              request.AdvancedData.Manifest,
		IsExternallyCreated: false,
	}
}
