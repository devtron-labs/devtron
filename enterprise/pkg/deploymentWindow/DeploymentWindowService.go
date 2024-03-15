package deploymentWindow

import (
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"go.uber.org/zap"
	"time"
	_ "time/tzdata"
)

type DeploymentWindowService interface {

	//CRUD
	CreateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error)
	UpdateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error)
	GetDeploymentWindowProfileForId(profileId int) (*DeploymentWindowProfile, error)
	DeleteDeploymentWindowProfileForId(profileId int, userId int32) error
	ListDeploymentWindowProfiles() ([]*DeploymentWindowProfileMetadata, error)

	//Overview
	GetDeploymentWindowProfileOverview(appId int, envIds []int) (*DeploymentWindowResponse, error)

	//State
	GetStateForAppEnv(targetTime time.Time, appId int, envId int, userId int32) (UserActionState, *EnvironmentState, error)
	GetDeploymentWindowProfileState(targetTime time.Time, appId int, envIds []int, filterForDays int, userId int32) (*DeploymentWindowResponse, error)
	GetDeploymentWindowProfileStateAppGroup(targetTime time.Time, selectors []AppEnvSelector, filterForDays int, userId int32) (*DeploymentWindowAppGroupResponse, error)
}

type DeploymentWindowServiceImpl struct {
	logger                 *zap.SugaredLogger
	cfg                    *DeploymentWindowConfig
	userService            user.UserService
	resourceMappingService resourceQualifiers.QualifierMappingService
	timeWindowService      timeoutWindow.TimeoutWindowService
	globalPolicyManager    globalPolicy.GlobalPolicyDataManager
}

func NewDeploymentWindowServiceImpl(
	logger *zap.SugaredLogger,
	resourceMappingService resourceQualifiers.QualifierMappingService,
	timeWindowService timeoutWindow.TimeoutWindowService,
	globalPolicyManager globalPolicy.GlobalPolicyDataManager,
	userService user.UserService,
) (*DeploymentWindowServiceImpl, error) {
	cfg, err := GetDeploymentWindowConfig()
	if err != nil {
		return nil, err
	}
	return &DeploymentWindowServiceImpl{
		cfg:                    cfg,
		logger:                 logger,
		resourceMappingService: resourceMappingService,
		timeWindowService:      timeWindowService,
		globalPolicyManager:    globalPolicyManager,
		userService:            userService,
	}, nil
}

type DeploymentWindowConfig struct {
	DeploymentWindowFetchDays int `env:"DEPLOYMENT_WINDOW_FETCH_DAYS" envDefault:"90"`
}

func GetDeploymentWindowConfig() (*DeploymentWindowConfig, error) {
	cfg := &DeploymentWindowConfig{}
	err := env.Parse(cfg)
	return cfg, err
}
