package auth

import (
	"github.com/devtron-labs/authenticator/jwt"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig/bean"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig/repository"
	"github.com/go-pg/pg"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"strings"
	"time"
)

type GlobalAuthorisationConfigService interface {
	GetCacheDataForActiveConfigs() (map[string]bool, error)
	ReloadCache()
	CreateOrUpdateGlobalAuthConfig(globalAuthConfig bean.GlobalAuthorisationConfig, tx *pg.Tx) ([]*bean.GlobalAuthorisationConfigResponse, error)
	CreateOrUpdateGroupClaimsAuthConfig(tx *pg.Tx, authConfigType string, userId int32) ([]*bean.GlobalAuthorisationConfigResponse, error)
	GetAllActiveAuthorisationConfig() ([]*bean.GlobalAuthorisationConfigResponse, error)
	IsGroupClaimsConfigActive() bool
	IsDevtronSystemManagedConfigActive() bool
	GetEmailAndGroupsFromClaims(claims jwtv4.MapClaims) (string, []string)
}

type GlobalAuthorisationConfigServiceImpl struct {
	logger                              *zap.SugaredLogger
	globalAuthorisationConfigRepository repository.GlobalAuthorisationConfigRepository
	globalAuthActiveConfigCache         map[string]bool
}

func NewGlobalAuthorisationConfigServiceImpl(logger *zap.SugaredLogger,
	globalAuthorisationConfigRepository repository.GlobalAuthorisationConfigRepository) *GlobalAuthorisationConfigServiceImpl {
	globalAuthorisationConfigImpl := &GlobalAuthorisationConfigServiceImpl{
		logger:                              logger,
		globalAuthorisationConfigRepository: globalAuthorisationConfigRepository,
	}
	activeConfigCache, err := globalAuthorisationConfigImpl.GetCacheDataForActiveConfigs()
	if err != nil {
		globalAuthorisationConfigImpl.logger.Errorw("error in caching on start up", "err", err)
	}
	//Setting cache for active configs, will be set to empty map when error is caught
	globalAuthorisationConfigImpl.globalAuthActiveConfigCache = activeConfigCache

	return globalAuthorisationConfigImpl
}

// GetCacheDataForActiveConfigs Caches the active global authorisation config
func (impl *GlobalAuthorisationConfigServiceImpl) GetCacheDataForActiveConfigs() (map[string]bool, error) {
	activeConfigMap := make(map[string]bool)
	activeConfigs, err := impl.globalAuthorisationConfigRepository.GetAllActiveConfigs()
	if err != nil {
		impl.logger.Errorw("error in getting all active configs for cache", "err", "err")
		return activeConfigMap, err
	}
	for _, config := range activeConfigs {
		activeConfigMap[config.ConfigType] = config.Active
	}
	return activeConfigMap, err
}

// ReloadCache updates the cache and does not updates cache if any error is caught
func (impl *GlobalAuthorisationConfigServiceImpl) ReloadCache() {
	configCache, err := impl.GetCacheDataForActiveConfigs()
	if err != nil {
		impl.logger.Errorw("cache can not be reloaded, using previous cache", "err", err)
		return
	}
	//Setting new cache
	impl.globalAuthActiveConfigCache = configCache
}

func (impl *GlobalAuthorisationConfigServiceImpl) CreateOrUpdateGroupClaimsAuthConfig(tx *pg.Tx, authConfigType string, userId int32) ([]*bean.GlobalAuthorisationConfigResponse, error) {
	configType := []string{authConfigType}
	authConfig := bean.GlobalAuthorisationConfig{
		ConfigTypes: configType,
		UserId:      userId,
	}
	resp, err := impl.CreateOrUpdateGlobalAuthConfig(authConfig, tx)
	if err != nil {
		impl.logger.Errorw("error in CreateOrUpdateGroupClaimsAuthConfig", "err", err)
		return nil, err
	}
	return resp, nil

}

func (impl *GlobalAuthorisationConfigServiceImpl) CreateOrUpdateGlobalAuthConfig(globalAuthConfig bean.GlobalAuthorisationConfig, tx *pg.Tx) ([]*bean.GlobalAuthorisationConfigResponse, error) {
	configs, err := impl.globalAuthorisationConfigRepository.GetByConfigTypes(globalAuthConfig.ConfigTypes)
	if err != nil {
		impl.logger.Errorw("error in checking global auth config exist by config type", "err", err, "configTypes", globalAuthConfig.ConfigTypes)
		return nil, err
	}
	existingConfigMap := make(map[string]*repository.GlobalAuthorisationConfig)
	for _, config := range configs {
		if len(config.ConfigType) > 0 && config.Id > 0 {
			existingConfigMap[config.ConfigType] = config
		}
	}
	createConfigModels := make([]*repository.GlobalAuthorisationConfig, 0, len(globalAuthConfig.ConfigTypes))
	updateConfigModels := make([]*repository.GlobalAuthorisationConfig, 0, len(globalAuthConfig.ConfigTypes))
	for _, cfg := range globalAuthConfig.ConfigTypes {
		if cfgModel, ok := existingConfigMap[cfg]; !ok {
			// Create config model for bulk operations
			configModel := impl.createConfigModel(globalAuthConfig.UserId, cfg)
			createConfigModels = append(createConfigModels, configModel)
		} else {
			// update to active model for bulk operations
			updateConfigModels = append(updateConfigModels, impl.updateConfigModel(cfgModel, globalAuthConfig.UserId))
		}
	}
	//Checking if transaction exists, if not make a new transaction
	isNewTransaction := false
	if tx == nil {
		tx, err = impl.globalAuthorisationConfigRepository.StartATransaction()
		if err != nil {
			impl.logger.Errorw("error in starting a transaction", "err", err)
			return nil, err
		}
		isNewTransaction = true
		// Rollback tx on error.
		defer tx.Rollback()
	}
	if len(createConfigModels) > 0 {
		err = impl.globalAuthorisationConfigRepository.CreateConfig(tx, createConfigModels)
		if err != nil {
			impl.logger.Errorw("error in creating configs in bulk", "err", err, "createConfigModels", createConfigModels)
			return nil, err
		}
	}
	if len(updateConfigModels) > 0 {
		err = impl.globalAuthorisationConfigRepository.UpdateConfig(tx, updateConfigModels)
		if err != nil {
			impl.logger.Errorw("error in updating configs in bulk", "err", err, "updateConfigModels", updateConfigModels)
			return nil, err
		}
	}

	// Updating all configs to inactive except those got in request(transactional)
	err = impl.globalAuthorisationConfigRepository.SetConfigsToInactiveExceptGivenConfigs(tx, globalAuthConfig.ConfigTypes)
	if err != nil {
		impl.logger.Errorw("error in SetConfigsToInactiveExceptGivenConfigs", "err", err, "configTypes", globalAuthConfig.ConfigTypes)
		return nil, err
	}

	if isNewTransaction {
		err = impl.globalAuthorisationConfigRepository.CommitATransaction(tx)
		if err != nil {
			impl.logger.Errorw("error in committing a transaction", "err", err)
			return nil, err
		}
	}
	allConfigModel := append(createConfigModels, updateConfigModels...)
	globalConfigResponse := make([]*bean.GlobalAuthorisationConfigResponse, 0, len(allConfigModel))
	for _, authConfig := range allConfigModel {
		config := &bean.GlobalAuthorisationConfigResponse{
			Id:         authConfig.Id,
			ConfigType: authConfig.ConfigType,
			Active:     authConfig.Active,
		}
		globalConfigResponse = append(globalConfigResponse, config)
	}

	// Reloading Local Cache on Save/Update
	impl.ReloadCache()
	return globalConfigResponse, err
}

func (impl *GlobalAuthorisationConfigServiceImpl) createConfigModel(userId int32, configType string) *repository.GlobalAuthorisationConfig {
	configModel := &repository.GlobalAuthorisationConfig{}
	configModel.ConfigType = configType
	configModel.Active = true
	configModel.CreatedOn = time.Now()
	configModel.CreatedBy = userId
	configModel.UpdatedOn = time.Now()
	configModel.UpdatedBy = userId

	return configModel
}

func (impl *GlobalAuthorisationConfigServiceImpl) updateConfigModel(configModel *repository.GlobalAuthorisationConfig, userId int32) *repository.GlobalAuthorisationConfig {
	configModel.Active = true
	configModel.UpdatedOn = time.Now()
	configModel.UpdatedBy = userId
	return configModel
}

func (impl *GlobalAuthorisationConfigServiceImpl) GetAllActiveAuthorisationConfig() ([]*bean.GlobalAuthorisationConfigResponse, error) {
	configs, err := impl.globalAuthorisationConfigRepository.GetAllActiveConfigs()
	if err != nil {
		impl.logger.Errorw("error in getting authorisation config by config type ", "err", err)
		return nil, err
	}
	var authConfigsResponse []*bean.GlobalAuthorisationConfigResponse
	for _, config := range configs {
		authConfig := &bean.GlobalAuthorisationConfigResponse{
			Id:         config.Id,
			ConfigType: config.ConfigType,
			Active:     config.Active,
		}
		authConfigsResponse = append(authConfigsResponse, authConfig)
	}

	return authConfigsResponse, nil
}

func (impl *GlobalAuthorisationConfigServiceImpl) GetEmailAndGroupsFromClaims(claims jwtv4.MapClaims) (string, []string) {
	email := jwt.GetField(claims, "email")
	sub := jwt.GetField(claims, "sub")
	if email == "" && (sub == "admin" || sub == "admin:login") {
		email = "admin"
	}
	groups := make([]string, 0)
	groupClaimsIf := jwt.GetFieldInterface(claims, "groups")
	if groupClaimsArr, ok := groupClaimsIf.([]interface{}); ok {
		for _, groupClaim := range groupClaimsArr {
			if group, ok := groupClaim.(string); ok {
				groups = append(groups, group)
			}
		}
	} else if groupClaimStr, ok := groupClaimsIf.(string); ok {
		splitGroups := strings.Split(groupClaimStr, ",")
		groups = splitGroups
	}
	return email, groups
}

func (impl *GlobalAuthorisationConfigServiceImpl) IsGroupClaimsConfigActive() bool {
	if _, ok := impl.globalAuthActiveConfigCache[string(bean.GroupClaims)]; ok {
		return true
	}
	return false
}

func (impl *GlobalAuthorisationConfigServiceImpl) IsDevtronSystemManagedConfigActive() bool {
	if _, ok := impl.globalAuthActiveConfigCache[string(bean.DevtronSystemManaged)]; ok {
		return true
	}
	return false
}
