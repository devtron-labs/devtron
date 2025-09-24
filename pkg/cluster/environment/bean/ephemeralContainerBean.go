package bean

import (
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
)

type EphemeralContainerRequest struct {
	BasicData                        *EphemeralContainerBasicData    `json:"basicData" validate:"required"`
	AdvancedData                     *EphemeralContainerAdvancedData `json:"advancedData"`
	Namespace                        string                          `json:"namespace" validate:"required"`
	ClusterId                        int                             `json:"clusterId" validate:"gt=0"`
	PodName                          string                          `json:"podName"   validate:"required"`
	ExternalArgoApplicationName      string                          `json:"externalArgoApplicationName,omitempty"`
	ExternalArgoApplicationNamespace string                          `json:"externalArgoApplicationNamespace,omitempty"`
	ExternalArgoAppIdentifier        *bean.ArgoAppIdentifier         `json:"externalArgoAppIdentifier"`
	UserId                           int32                           `json:"-"`
}

type EphemeralContainerAdvancedData struct {
	Manifest string `json:"manifest" validate:"required"`
}

type EphemeralContainerBasicData struct {
	ContainerName       string `json:"containerName" validate:"required"`
	TargetContainerName string `json:"targetContainerName" validate:"required"`
	Image               string `json:"image" validate:"required"`
}

func (request EphemeralContainerRequest) GetContainerBean() repository.EphemeralContainerBean {
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

const EXTERNAL_EPHIMERAL_CONTAINER_ERR string = "externally created ephemeral containers cannot be removed\""
