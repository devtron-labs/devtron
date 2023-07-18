package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const ephemeralContainerNotFoundError = "ephemeral container not found container"

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
	repository     repository.EphemeralContainersRepository
	clusterService ClusterService
	K8sUtil        *util.K8sUtil
	logger         *zap.SugaredLogger
}

func NewEphemeralContainerServiceImpl(repository repository.EphemeralContainersRepository, logger *zap.SugaredLogger, clusterService ClusterService, K8sUtil *util.K8sUtil) *EphemeralContainerServiceImpl {
	return &EphemeralContainerServiceImpl{
		repository:     repository,
		clusterService: clusterService,
		K8sUtil:        K8sUtil,
		logger:         logger,
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
		if actionType != repository.ActionCreate {
			// ActionCreate is happening through devtron, the model will contain all the required fields, set by the fn caller
			// for other actions,if we dodn't find the container in db, we should get the ephemeralContainer data from pod manifest
			err = impl.getEphemeralContainerDataFromManifest(&model)
			if err != nil {
				if (actionType == repository.ActionAccessed) && strings.Contains(err.Error(), ephemeralContainerNotFoundError) {
					impl.logger.Errorw("skipping auditing as terminal access requested for non ephemeral container", "error", err)
					return nil
				}
				return err
			}
		}

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
	impl.logger.Infow("transaction committed successfully")
	return nil
}

func (impl *EphemeralContainerServiceImpl) getEphemeralContainerDataFromManifest(req *EphemeralContainerRequest) error {
	clusterBean, err := impl.clusterService.FindById(req.ClusterId)
	if err != nil {
		impl.logger.Errorw("error occurred in finding clusterBean by Id", "clusterId", req.ClusterId, "err", err)
		return err
	}

	clusterConfig := clusterBean.GetClusterConfig()
	v1Client, err := impl.K8sUtil.GetClient(&clusterConfig)
	if err != nil {
		//not logging clusterConfig as it contains sensitive data
		impl.logger.Errorw("error occurred in getting v1Client with cluster config", "err", err, "clusterId", req.ClusterId)
		return err
	}
	pod, err := impl.K8sUtil.GetPodByName(req.Namespace, req.PodName, v1Client)
	if err != nil {
		impl.logger.Errorw("error in getting pod", "clusterId", req.ClusterId, "namespace", req.Namespace, "podName", req.PodName, "err", err)
		return err
	}
	var ephemeralContainer *corev1.EphemeralContainer
	for _, ec := range pod.Spec.EphemeralContainers {
		if ec.Name == req.BasicData.ContainerName {
			ephemeralContainer = &ec
			break
		}
	}
	if ephemeralContainer == nil {
		impl.logger.Errorw("terminal session requested for non ephemeral container,so not auditing the terminal access", "clusterId", req.ClusterId, "namespace", req.Namespace, "podName", req.PodName)
		return errors.New(fmt.Sprintf("%s: %s , pod: %s", ephemeralContainerNotFoundError, req.BasicData.ContainerName, req.PodName))
	}
	ephemeralContainerJson, err := json.Marshal(ephemeralContainer)
	if err != nil {
		impl.logger.Errorw("error occurred while marshaling ephemeralContainer object", "err", err, "ephemeralContainer", ephemeralContainer)
		return err
	}
	req.BasicData.TargetContainerName = ephemeralContainer.TargetContainerName
	req.BasicData.Image = ephemeralContainer.Image
	req.AdvancedData = &EphemeralContainerAdvancedData{
		Manifest: string(ephemeralContainerJson),
	}
	return nil
}
