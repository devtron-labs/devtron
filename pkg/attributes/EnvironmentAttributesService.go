package attributes

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
)

type EnvironmentDigestEnforcementConfig map[string]map[string]bool

type ClusterDetail struct {
	ClusterName  string `json:"clusterName"`
	Environments []int  `json:"environments"`
}

type ClusterDetailRequest struct {
	ClusterDetails []ClusterDetail `json:"clusterDetails"`
}

type EnvironmentAttributesDto struct {
	EmailId string `json:"emailId"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	UserId  int32  `json:"-"`
}

type EnvironmentAttributesService interface {
	AddPullImageUsingDigestConfig(request *EnvironmentAttributesDto) (*EnvironmentAttributesDto, error)
}

type EnvironmentAttributesServiceImpl struct {
	logger               *zap.SugaredLogger
	attributesRepository repository.UserAttributesRepository
}

func NewEnvironmentAttributesServiceImpl(logger *zap.SugaredLogger,
	attributesRepository repository.UserAttributesRepository) *UserAttributesServiceImpl {
	serviceImpl := &UserAttributesServiceImpl{
		logger:               logger,
		attributesRepository: attributesRepository,
	}
	return serviceImpl
}

func (impl EnvironmentAttributesServiceImpl) AddPullImageUsingDigestConfig(request *EnvironmentAttributesDto) (*EnvironmentAttributesDto, error) {

}
