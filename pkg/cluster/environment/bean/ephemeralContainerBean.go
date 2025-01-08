package bean

import "github.com/devtron-labs/devtron/pkg/cluster/repository"

type EphemeralContainerRequest struct {
	BasicData                   *EphemeralContainerBasicData    `json:"basicData"`
	AdvancedData                *EphemeralContainerAdvancedData `json:"advancedData"`
	Namespace                   string                          `json:"namespace" validate:"required"`
	ClusterId                   int                             `json:"clusterId" validate:"gt=0"`
	PodName                     string                          `json:"podName"   validate:"required"`
	ExternalArgoApplicationName string                          `json:"externalArgoApplicationName,omitempty"`
	UserId                      int32                           `json:"-"`
}

type EphemeralContainerAdvancedData struct {
	Manifest string `json:"manifest"`
}

type EphemeralContainerBasicData struct {
	ContainerName       string `json:"containerName"`
	TargetContainerName string `json:"targetContainerName"`
	Image               string `json:"image"`
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
