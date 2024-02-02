package bean

import util5 "github.com/devtron-labs/common-lib/utils/k8s"

type PodRotateRequest struct {
	AppId               int                        `json:"appId" validate:"required"`
	EnvironmentId       int                        `json:"environmentId" validate:"required"`
	ResourceIdentifiers []util5.ResourceIdentifier `json:"resources" validate:"required"`
	UserId              int32                      `json:"-"`
}
