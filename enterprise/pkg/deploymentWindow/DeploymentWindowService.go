package deploymentWindow

import (
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
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
	CheckTriggerAllowedState(targetTime time.Time, appId int, envId int, userId int32) (bool, error)
	GetDeploymentWindowProfileState(targetTime time.Time, appId int, envIds []int, userId int32) (*DeploymentWindowResponse, error)
	GetDeploymentWindowProfileStateAppGroup(targetTime time.Time, selectors []AppEnvSelector, userId int32) (*DeploymentWindowAppGroupResponse, error)
}

type DeploymentWindowServiceImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	cfg          *DeploymentWindowConfig
	timeZoneMap  map[string]*time.Location
	userService  user.UserService

	resourceMappingService      resourceQualifiers.QualifierMappingService
	timeoutWindowMappingService timeoutWindow.TimeoutWindowResourceMappingService
	globalPolicyManager         globalPolicy.GlobalPolicyDataManager
}

func NewDeploymentWindowServiceImpl(
	logger *zap.SugaredLogger,
	timeZoneMap map[string]*time.Location,
	resourceMappingService resourceQualifiers.QualifierMappingService,
	timeoutWindowMappingService timeoutWindow.TimeoutWindowResourceMappingService,
	globalPolicyManager globalPolicy.GlobalPolicyDataManager,
	dbConnection *pg.DB,
	userService user.UserService,
) (*DeploymentWindowServiceImpl, error) {
	cfg, err := GetDeploymentWindowConfig()
	if err != nil {
		return nil, err
	}
	return &DeploymentWindowServiceImpl{
		cfg:                         cfg,
		logger:                      logger,
		timeZoneMap:                 timeZoneMap,
		resourceMappingService:      resourceMappingService,
		timeoutWindowMappingService: timeoutWindowMappingService,
		globalPolicyManager:         globalPolicyManager,
		dbConnection:                dbConnection,
		userService:                 userService,
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

func (impl DeploymentWindowServiceImpl) StartATransaction() (*pg.Tx, error) {
	tx, err := impl.dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning a transaction", "err", err)
		return nil, err
	}
	return tx, nil
}

func (impl DeploymentWindowServiceImpl) CommitATransaction(tx *pg.Tx) error {
	err := tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commiting a transaction", "err", err)
		return err
	}
	return nil
}

func (impl DeploymentWindowServiceImpl) getTimeZoneData(timeZone string) (*time.Location, error) {
	var location *time.Location
	var err error
	if data, ok := impl.timeZoneMap[timeZone]; ok && data != nil {
		return data, nil
	} else {
		location, err = time.LoadLocation(timeZone)
		if err != nil {
			return nil, err
		}
		impl.timeZoneMap[timeZone] = location
	}
	return location, nil
}
