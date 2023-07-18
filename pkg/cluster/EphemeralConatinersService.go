package cluster

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	"time"
)

type EphemeralContainerRequest struct {
	BasicData    *EphemeralContainerBasicData    `json:"basicData"`
	AdvancedData *EphemeralContainerAdvancedData `json:"advancedData"`
	Namespace    string                          `json:"namespace"`
	ClusterId    int                             `json:"clusterId"`
	PodName      string                          `json:"podName"`
	UserId       int32                           `json:"-"`
}

type EphemeralContainerAdvancedData struct {
	Manifest string `json:"manifest"`
}

type EphemeralContainerBasicData struct {
	ContainerName       string `json:"containerName"`
	TargetContainerName string `json:"targetContainerName"`
	Image               string `json:"image"`
}

func (request EphemeralContainerRequest) getContainerBean() repository.EphemeralContainerBean {
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

type EphemeralContainerService interface {
	AuditEphemeralContainerAction(model EphemeralContainerRequest, actionType repository.ContainerAction) error
}

type EphemeralContainerServiceImpl struct {
	repository repository.EphemeralContainersRepository
	logger     *zap.SugaredLogger
}

func NewEphemeralContainerServiceImpl(repository repository.EphemeralContainersRepository, logger *zap.SugaredLogger) *EphemeralContainerServiceImpl {
	return &EphemeralContainerServiceImpl{
		repository: repository,
		logger:     logger,
	}
}

func (impl *EphemeralContainerServiceImpl) AuditEphemeralContainerAction(model EphemeralContainerRequest, actionType repository.ContainerAction) error {

	container, err := impl.repository.FindContainerByName(model.ClusterId, model.Namespace, model.PodName, model.BasicData.ContainerName)
	if err != nil {
		impl.logger.Errorw("error in finding ephemeral container in the database", "err", err, "ClusterId", model.ClusterId, "Namespace", model.Namespace, "PodName", model.PodName, "ContainerName", model.BasicData.ContainerName)
		return err
	}

	if container != nil && actionType == repository.ActionCreate {
		impl.logger.Errorw("Container already present in the provided pod", "ClusterId", model.ClusterId, "Namespace", model.Namespace, "PodName", model.PodName, "ContainerName", model.BasicData.ContainerName)
		return errors.New("container already present in the provided pod")
	}

	tx, err := impl.repository.StartTx()
	defer func() {
		err = impl.repository.RollbackTx(tx)
		if err != nil {
			impl.logger.Infow("error in rolling back transaction", "err", err, "ClusterId", model.ClusterId, "Namespace", model.Namespace, "PodName", model.PodName, "ContainerName", model.BasicData.ContainerName)
		}
	}()

	if err != nil {
		impl.logger.Errorw("error in creating transaction", "err", err)
		return err
	}

	var auditLogBean repository.EphemeralContainerAction
	if container == nil {
		bean := model.getContainerBean()
		if actionType != repository.ActionCreate {
			// if a container is not present in database and the user is trying to access/terminate it means it is externally created
			bean.IsExternallyCreated = true
		}
		err = impl.repository.SaveData(tx, &bean)
		if err != nil {
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
		impl.logger.Errorw("Failed to save ephemeral container", "error", err)
		return err
	}

	err = impl.repository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err, "req", model)
		return err
	}
	impl.logger.Errorw("transaction committed successfully")
	return nil
}
