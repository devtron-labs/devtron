/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"fmt"
	globalUtil "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/config"
	infraErrors "github.com/devtron-labs/devtron/pkg/infraConfig/errors"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type InfraConfigService interface {
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*v1.ProfileBeanDto, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	// This also takes care of the platform update and delete, the default platform cannot be deleted.
	UpdateProfile(userId int32, profileName string, profileBean *v1.ProfileBeanDto) error
	// Deprecated: UpdateProfileV0 is deprecated in favor of UpdateProfile
	UpdateProfileV0(userId int32, profileName string, profileToUpdate *v1.ProfileBeanDto) error
	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error)
	// GetConfigurationsByScopeAndTargetPlatforms fetches the infra configurations for the given scope and targetPlatforms.
	GetConfigurationsByScopeAndTargetPlatforms(scope resourceQualifiers.Scope, targetPlatformsList []string) (map[string]*v1.InfraConfig, error)
	HandleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error
	InfraConfigServiceEnt
}

type InfraConfigServiceImpl struct {
	logger            *zap.SugaredLogger
	infraProfileRepo  repository.InfraConfigRepository
	infraConfig       *v1.InfraConfig
	infraConfigClient config.InfraConfigClient

	appService                     app.AppService
	dtResourceSearchableKeyService read.DevtronResourceSearchableKeyService
	qualifierMappingService        resourceQualifiers.QualifierMappingService
	ciConfig                       *types.CiConfig
	attributesService              attributes.AttributesService
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo repository.InfraConfigRepository,
	appService app.AppService,
	dtResourceSearchableKeyService read.DevtronResourceSearchableKeyService,
	qualifierMappingService resourceQualifiers.QualifierMappingService,
	attributesService attributes.AttributesService,
	infraConfigClient config.InfraConfigClient,
	variables *util2.EnvironmentVariables) (*InfraConfigServiceImpl, error) {
	envConfig, err := types.GetCiConfig()
	if err != nil {
		return nil, fmt.Errorf("error retrieving CiConfig: %v", err)
	}
	infraConfiguration, err := getDefaultInfraConfigFromEnv(envConfig)
	if err != nil {
		return nil, fmt.Errorf("error retrieving default infra config: %v", err)
	}
	infraProfileService := &InfraConfigServiceImpl{
		logger:                         logger,
		infraProfileRepo:               infraProfileRepo,
		infraConfig:                    infraConfiguration,
		appService:                     appService,
		dtResourceSearchableKeyService: dtResourceSearchableKeyService,
		qualifierMappingService:        qualifierMappingService,
		attributesService:              attributesService,
		infraConfigClient:              infraConfigClient,
		ciConfig:                       envConfig,
	}
	if !variables.InternalEnvVariables.IsDevelopmentEnv() {
		err = infraProfileService.loadDefaultProfile()
		if err != nil {
			return nil, fmt.Errorf("error loading default profile: %v", err)
		}
	}
	return infraProfileService, err
}

func (impl *InfraConfigServiceImpl) GetProfileByName(name string) (*v1.ProfileBeanDto, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(name)
	if err != nil && !globalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching profile", "profileName", name, "error", err)
		return nil, err
	} else if globalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("profile does not exist", "profileName", name, "error", err)
		return nil, globalUtil.NewApiError(http.StatusNotFound, infraErrors.ProfileDoNotExists, infraErrors.ProfileDoNotExists)
	}
	profileBean := adapter.ConvertToProfileBean(infraProfile)
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileIds(sliceUtil.GetSliceOf(infraProfile.Id))
	if err != nil {
		impl.logger.Errorw("error in fetching configurations using profileId", "profileId", infraProfile.Id, "error", err)
		return nil, err
	}
	platformsList, err := impl.infraProfileRepo.GetPlatformListByProfileId(infraProfile.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching platforms using profileId", "profileId", infraProfile.Id)
		return nil, err
	}
	configurationBeans := make(map[string][]*v1.ConfigurationBean)
	if infraConfigurations != nil {
		configurationBeans, err = impl.infraConfigClient.GetConfigurationBeansForProfile(infraConfigurations, profileBean.GetName())
		if err != nil {
			impl.logger.Errorw("error in converting infraConfigurations into platformMap", "profileName", profileBean.GetName(), "error", err)
			return nil, err
		}
	}

	for _, platform := range platformsList {
		if _, exists := configurationBeans[platform]; !exists {
			configurationBeans[platform] = []*v1.ConfigurationBean{}
		}
	}

	profileBean.Configurations = configurationBeans
	return profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileToUpdate *v1.ProfileBeanDto) error {
	if !util.IsValidProfileNameRequested(profileName, profileToUpdate.GetName()) {
		impl.logger.Errorw("error in validating profile name change", "profileName", profileName, "profileToUpdate", profileToUpdate)
		return globalUtil.NewApiError(http.StatusBadRequest, infraErrors.InvalidProfileNameChangeRequested, infraErrors.InvalidProfileNameChangeRequested)
	}
	err := impl.validateUpdateRequest(profileToUpdate, profileName)
	if err != nil {
		impl.logger.Errorw("error in validating payload", "profileName", profileName, "error", err)
		return globalUtil.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
	}
	existingPlatforms, err := impl.infraProfileRepo.GetPlatformsByProfileName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching list of the platforms for the profile", "profileName", profileName, "error", err)
		return err
	}
	updatableProfilePlatforms, creatableProfilePlatforms, err := impl.getCreatableAndUpdatableProfilePlatforms(userId, profileToUpdate, profileName, existingPlatforms)
	if err != nil {
		impl.logger.Errorw("Error in getCreatableAndUpdatableProfilePlatforms", "profile", profileToUpdate, "err", err)
		return err
	}
	profileFromDb, err := impl.infraProfileRepo.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching profile", "profileName", profileName, "error", err)
		return err
	}
	profileToUpdate.Id = profileFromDb.Id
	infraProfileEntity := adapter.ConvertToInfraProfileEntity(profileToUpdate)
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true
	sanitizedUpdatableInfraConfigurations, sanitizedCreatableInfraConfigurations, err := impl.sanitizeAndGetUpdatableAndCreatableConfigurationEntities(userId, profileName, profileToUpdate, existingPlatforms)
	if err != nil {
		// if sanity failed throw the error
		impl.logger.Errorw("Error in sanitizeAndGetUpdatableAndCreatableConfigurationEntities", "profile", profileToUpdate, "err", err)
		return err
	}
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileBean", profileToUpdate, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfileEntity.UpdateAuditLog(userId)
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfileEntity)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "infraProfile", infraProfileEntity, "error", err)
		return err
	}
	// Update existing platform mappings in the database
	if len(updatableProfilePlatforms) > 0 {
		err = impl.infraProfileRepo.UpdatePlatformProfileMapping(tx, updatableProfilePlatforms)
		if err != nil {
			impl.logger.Errorw("Error updating profile platform mappings", "profile", profileToUpdate, "err", err)
			return err
		}
	}
	// Create new platform mappings in the database
	if len(creatableProfilePlatforms) > 0 {
		err = impl.infraProfileRepo.CreatePlatformProfileMapping(tx, creatableProfilePlatforms)
		if err != nil {
			impl.logger.Errorw("Error creating profile platform mappings", "profile", profileToUpdate, "err", err)
			return err
		}
	}

	sanitizedUpdatableInfraConfigurations = adapter.UpdatePlatformMappingInConfigEntities(sanitizedUpdatableInfraConfigurations, updatableProfilePlatforms)
	if len(sanitizedUpdatableInfraConfigurations) > 0 {
		err = impl.infraProfileRepo.UpdateConfigurations(tx, sanitizedUpdatableInfraConfigurations)
		if err != nil {
			impl.logger.Errorw("error in creating configurations", "updatableInfraConfigurations", sanitizedUpdatableInfraConfigurations, "error", err)
			return err
		}
		err = impl.infraConfigClient.HandlePostUpdateOperations(tx, sanitizedUpdatableInfraConfigurations)
		if err != nil {
			impl.logger.Errorw("error in handling post update operations", "updatableInfraConfigurations", sanitizedUpdatableInfraConfigurations, "error", err)
			return err
		}
	}

	sanitizedCreatableInfraConfigurations = adapter.UpdatePlatformMappingInConfigEntities(sanitizedCreatableInfraConfigurations, creatableProfilePlatforms)
	if len(sanitizedCreatableInfraConfigurations) > 0 {
		err = impl.infraProfileRepo.CreateConfigurations(tx, sanitizedCreatableInfraConfigurations)
		if err != nil {
			impl.logger.Errorw("error in creating configurations", "creatableInfraConfigurations", sanitizedCreatableInfraConfigurations, "error", err)
			return err
		}
		err = impl.infraConfigClient.HandlePostCreateOperations(tx, sanitizedCreatableInfraConfigurations)
		if err != nil {
			impl.logger.Errorw("error in handling post create operations", "creatableInfraConfigurations", sanitizedCreatableInfraConfigurations, "error", err)
			return err
		}
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) UpdateProfileV0(userId int32, profileName string, profileToUpdate *v1.ProfileBeanDto) error {
	if profileName == v1.DEFAULT_PROFILE_NAME {
		profileName = v1.GLOBAL_PROFILE_NAME
	}
	configurationEntities, err := impl.infraProfileRepo.GetConfigurationsByProfileName(profileName)
	if err != nil {
		impl.logger.Errorw("Error in GetConfigurationsByProfileName", "profileName", profileName, "error", err)
		return err
	}
	platformMapConfigs, err := impl.infraConfigClient.GetConfigurationBeansForProfile(configurationEntities, profileName)
	if err != nil {
		impl.logger.Errorw("error in converting configurations into platformMap", "profileName", profileName, "error", err)
		return err
	}
	adapter.FillMissingConfigurationsForThePayloadV0(profileToUpdate, platformMapConfigs)
	err = impl.UpdateProfile(userId, profileName, profileToUpdate)
	if err != nil {
		impl.logger.Errorw("error in performing Update ", "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error) {
	return impl.infraConfigClient.GetConfigurationUnits()
}

func (impl *InfraConfigServiceImpl) GetConfigurationsByScopeAndTargetPlatforms(scope resourceQualifiers.Scope, targetPlatformsList []string) (map[string]*v1.InfraConfig, error) {
	resp := make(map[string]*v1.InfraConfig)
	// Get the configuration from scope then,
	// Parse all target Platforms and if the configuration[targetPlatform] found => set the in the res[targetPlatform]
	// If not found or exist, then set res[targetPlatform]= configuration[default]
	appliedProfileConfig, defaultProfileConfig, err := impl.getAppliedProfileForTriggerScope(infraGetters.GetInfraConfigScope(scope))
	if err != nil {
		impl.logger.Errorw("error in fetching configurations", "scope", scope, "error", err)
		return resp, err
	}
	resolvedAppliedConfig, variableSnapshot, err := impl.resolveScopeVariablesForAppliedProfile(scope, appliedProfileConfig)
	if err != nil {
		impl.logger.Errorw("error in resolving scope variables", "appliedProfileConfig", appliedProfileConfig, "error", err)
		return resp, err
	}
	appliedProfileConfig = resolvedAppliedConfig
	platformToInfraConfigMap, err := impl.getAppliedInfraConfigForProfile(appliedProfileConfig, defaultProfileConfig, variableSnapshot, targetPlatformsList)
	if err != nil {
		impl.logger.Errorw("error in fetching infra configurations", "appliedProfileConfig", appliedProfileConfig, "error", err)
		return resp, err
	}
	return platformToInfraConfigMap, err
}

func (impl *InfraConfigServiceImpl) HandleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error {
	return impl.infraConfigClient.HandleInfraConfigTriggerAudit(workflowId, triggeredBy, infraConfigs)
}

// loadDefaultProfile: loads default configurations from environment and save them in the DB.
//   - create the default profile only once if not exists in db already.
//     (container restarts won't create a new default profile everytime)
//   - load the default configurations provided in bean.InfraConfig.
//   - if DB is out of sync with bean.InfraConfig,
//     then it will create the new entries in DB, only for the missing configurations.
//   - also handles the one-time migration for buildx K8sDriverOpts from environment,
//     then update the marker in attribute table as true (i.e. migrated).
func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {
	profile, err := impl.infraProfileRepo.GetProfileByName(v1.GLOBAL_PROFILE_NAME)
	// make sure about no rows error
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting profile GetProfileByName", "err", err)
		return err
	}
	profileCreationRequired := errors.Is(err, pg.ErrNoRows)
	migrationRequired, err := impl.isMigrationRequired()
	if err != nil {
		impl.logger.Errorw("error in checking if migration is required", "error", err)
		return err
	}

	// step 1: initiate Transaction
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error starting transaction at loadDefaultProfile", "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)

	// step 2: creating Global Profile in case when required
	if profileCreationRequired {
		profile, err = impl.createGlobalProfile(tx)
		if err != nil {
			impl.logger.Errorw("error in creating global profile", "error", err)
			return err
		}
	}

	// step 3: get creatableConfigurations and creatablePlatformMappings
	creatableConfigurations, creatablePlatformMappings, err := impl.getCreatableConfigurationsAndPlatformMappings(migrationRequired, profile.Id)
	if err != nil {
		impl.logger.Errorw("error in loading configuration at loadConfiguration", "err", err)
		return err
	}

	//step 4: create configurations and platform
	err = impl.createConfigurationsAndPlatforms(creatableConfigurations, creatablePlatformMappings, migrationRequired, tx)
	if err != nil {
		impl.logger.Errorw("error creating configurations and platforms", "error", err)
		return err
	}

	// step 5: mark the key if the migration was required
	if migrationRequired {
		err = impl.updateBuildxDriverTypeInExistingProfiles(tx)
		if err != nil {
			impl.logger.Errorw("error in updating buildx driver type in existing profiles", "error", err)
			return err
		}
		err = impl.markMigrationComplete(tx)
		if err != nil {
			impl.logger.Errorw("error in marking migration complete markMigrationComplete", "err", err)
			return err
		}
	}

	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save default configurations", "error", err)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) createConfigurationsAndPlatforms(creatableConfigurations []*repository.InfraProfileConfigurationEntity,
	platformMappings []*repository.ProfilePlatformMapping, migrationRequired bool, tx *pg.Tx) error {
	if migrationRequired && len(platformMappings) > 0 {
		err := impl.infraProfileRepo.CreatePlatformProfileMapping(tx, platformMappings)
		if err != nil {
			impl.logger.Errorw("error saving platform mappings", "error", err)
			return err
		}
	}
	creatableConfigurations = adapter.UpdatePlatformMappingInConfigEntities(creatableConfigurations, platformMappings)
	if len(creatableConfigurations) > 0 {
		err := impl.infraProfileRepo.CreateConfigurations(tx, creatableConfigurations)
		if err != nil {
			impl.logger.Errorw("error saving configurations", "error", err)
			return err
		}
	}
	return nil
}

func (impl *InfraConfigServiceImpl) getCreatableConfigurationsAndPlatformMappings(migrationRequired bool, profileId int) ([]*repository.InfraProfileConfigurationEntity, []*repository.ProfilePlatformMapping, error) {
	envConfigs, err := impl.loadConfiguration(profileId)
	if err != nil {
		impl.logger.Errorw("error in loading configuration at loadConfiguration", "err", err)
		return nil, nil, err
	}

	existingPlatformMappings, err := impl.infraProfileRepo.GetPlatformsByProfileById(profileId)
	if err != nil {
		impl.logger.Errorw("error in fetching ciRunnerProfilePlatform", "error", err)
		return nil, nil, err
	}

	creatableConfigs := make([]*repository.InfraProfileConfigurationEntity, 0)

	dbConfigs, err := impl.infraProfileRepo.GetConfigurationsByProfileId(profileId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error fetching default configurations from database", "error", err)
		return nil, nil, err
	}

	// filterOut the creatable Configs found in the db
	creatableConfigsForDefault, envConfigs := filterCreatableConfigForDefaultAndOverrideEnvConfigs(envConfigs, dbConfigs)

	if len(creatableConfigsForDefault) > 0 {
		creatableConfigs = append(creatableConfigs, creatableConfigsForDefault...)
	}

	// check for the migration required and BuildxK8sDriverOptions present to append in the creatableConfigs
	if migrationRequired && impl.ciConfig != nil && impl.ciConfig.BuildxK8sDriverOptions != "" {
		cmCreatableConfigs, err := impl.getCreatableK8sDriverConfigs(profileId, envConfigs)
		if err != nil {
			impl.logger.Errorw("error in getting creatable k8s driver configs", "error", err, "profileId", profileId)
			return nil, nil, err
		}
		if len(cmCreatableConfigs) > 0 {
			creatableConfigs = append(creatableConfigs, cmCreatableConfigs...)
		}
	}

	existingPlatforms := sliceUtil.NewSliceFromFuncExec(existingPlatformMappings,
		func(existingPlatformMapping *repository.ProfilePlatformMapping) string {
			return existingPlatformMapping.Platform
		})

	creatableConfigs, creatableProfilePlatformMappings := prepareConfigurationsAndMappings(creatableConfigs, dbConfigs, existingPlatforms, profileId, migrationRequired)

	creatableConfigs = adapter.UpdatePlatformMappingInConfigEntities(creatableConfigs, existingPlatformMappings)

	return creatableConfigs, creatableProfilePlatformMappings, nil
}

func (impl *InfraConfigServiceImpl) loadConfiguration(profileId int) ([]*repository.InfraProfileConfigurationEntity, error) {
	envConfigs, err := impl.infraConfigClient.GetInfraConfigEntities(profileId, impl.infraConfig)
	if err != nil {
		impl.logger.Errorw("error loading default configurations from environment", "error", err)
		return nil, err
	}
	return envConfigs, nil
}

func (impl *InfraConfigServiceImpl) validateUpdateRequest(profileToUpdate *v1.ProfileBeanDto, profileName string) error {
	globalProfile, err := impl.GetProfileByName(v1.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileName", v1.GLOBAL_PROFILE_NAME, "profileToUpdate", profileToUpdate, "error", err)
		return err
	}
	exist, err := impl.infraProfileRepo.CheckIfProfileExistsByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching profile by name", "profileName", profileName, "error", err)
		return err
	}
	if !exist {
		impl.logger.Errorw("profile does not exist", "profileName", profileName, "error", infraErrors.ProfileDoNotExists)
		return globalUtil.NewApiError(http.StatusBadRequest, infraErrors.ProfileDoNotExists, infraErrors.ProfileDoNotExists)
	}
	exist, err = impl.infraProfileRepo.CheckIfProfileExistsByName(profileToUpdate.GetName())
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching profile by name", "profileName", profileName, "error", err)
		return err
	}
	if exist && profileName != profileToUpdate.GetName() {
		impl.logger.Errorw("profile already exist", "profileNameRequested", profileToUpdate.GetName(), "err", infraErrors.ProfileAlreadyExistsErr)
		return globalUtil.NewApiError(http.StatusBadRequest, infraErrors.ProfileAlreadyExistsErr, infraErrors.ProfileAlreadyExistsErr)
	}

	err = impl.validate(profileToUpdate, globalProfile)
	if err != nil {
		impl.logger.Errorw("error in validating payload", "profileName", profileName, "error", err)
		return globalUtil.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validate(profileToUpdate, defaultProfile *v1.ProfileBeanDto) error {
	err := util.ValidatePayloadConfig(profileToUpdate)
	if err != nil {
		return err
	}

	_, err = impl.infraConfigClient.Validate(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, infraErrors.PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) getCreatableAndUpdatableProfilePlatforms(userId int32, profileToUpdate *v1.ProfileBeanDto,
	profileName string, existingPlatforms []*repository.ProfilePlatformMapping) ([]*repository.ProfilePlatformMapping, []*repository.ProfilePlatformMapping, error) {
	var err error
	// Create a map for updated platforms
	updatedMap := make(map[string]bool)
	updatedMap[v1.RUNNER_PLATFORM] = false
	for platform := range profileToUpdate.GetConfigurations() {
		updatedMap[platform] = true
	}
	if !updatedMap[v1.RUNNER_PLATFORM] {
		return nil, nil,
			globalUtil.NewApiError(http.StatusBadRequest, infraErrors.DeletionBlockedForDefaultPlatform, infraErrors.DeletionBlockedForDefaultPlatform)
	}

	infraProfileEntity, err := impl.infraProfileRepo.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching list of the platforms for the profile,")
		return nil, nil, err
	}

	// Create a map for existing platforms
	existingPlatformListMap := make(map[string]*repository.ProfilePlatformMapping, len(existingPlatforms))
	for _, platform := range existingPlatforms {
		existingPlatformListMap[platform.Platform] = platform
	}

	// Prepare platform mappings for update and creation
	updatableProfilePlatformMap := make([]*repository.ProfilePlatformMapping, 0)
	creatableProfilePlatformMap := make([]*repository.ProfilePlatformMapping, 0)
	// Handle platforms that are in existingPlatformListMap
	for platformName, platform := range existingPlatformListMap {
		if !updatedMap[platformName] {
			platform.Active = false
			platform.UpdateAuditLog(userId)
			platform.UniqueId = repository.GetUniqueId(platform.Id, platformName)
			// Platform exists in existing but not in updated, mark as inactive for deletion
			updatableProfilePlatformMap = append(updatableProfilePlatformMap, platform)
		}
	}

	// Handle platforms that are only in updatedMap
	for platform := range updatedMap {
		if _, exist := existingPlatformListMap[platform]; !exist {
			// Platform is new, mark for creation
			creatableProfilePlatformMap = append(creatableProfilePlatformMap, &repository.ProfilePlatformMapping{
				ProfileId: infraProfileEntity.Id,
				Platform:  platform,
				Active:    true,
				AuditLog:  sql.NewDefaultAuditLog(userId),
				UniqueId:  repository.GetUniqueId(infraProfileEntity.Id, platform),
			})
		}
	}
	return updatableProfilePlatformMap, creatableProfilePlatformMap, nil
}

func (impl *InfraConfigServiceImpl) sanitizeAndGetUpdatableAndCreatableConfigurationEntities(userId int32, profileName string,
	profileToUpdate *v1.ProfileBeanDto, existingPlatforms []*repository.ProfilePlatformMapping) (updatableInfraConfigurations []*repository.InfraProfileConfigurationEntity, creatableInfraConfigurations []*repository.InfraProfileConfigurationEntity, err error) {
	// Convert profile configurations into infraConfigurationEntities
	infraConfigurationEntities, err := impl.infraConfigClient.GetInfraProfileConfigurationEntities(profileToUpdate, userId)
	if err != nil {
		impl.logger.Errorw("error in converting profile configurations to infraConfigurationEntities", "profileToUpdate", profileToUpdate, "error", err)
		return nil, nil, err
	}
	// Fetch existing configurations and platforms for the given profile
	existingConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(profileName)
	if err != nil && !errors.Is(err, infraErrors.NoPropertiesFoundError) {
		impl.logger.Errorw("error in fetching existing configuration ids", "profileName", profileName, "error", err)
		return updatableInfraConfigurations, creatableInfraConfigurations, err
	}
	// mandatory config validations for profile
	for platform, configurations := range profileToUpdate.GetConfigurations() {
		configuredInfraConfigKeys := getConfiguredInfraConfigKeys(platform, configurations)
		// Check if any default keys are still true (missing)
		missingKeys := util.GetMissingRequiredConfigKeys(profileName, platform, configuredInfraConfigKeys)
		for _, missingKey := range missingKeys {
			index, found := sliceUtil.Find(existingConfigurations, func(entity *repository.InfraProfileConfigurationEntity) bool {
				if entity == nil || entity.ProfilePlatformMapping == nil {
					return false
				}
				return entity.Key == util.GetConfigKey(missingKey) && entity.ProfilePlatformMapping.Platform == platform
			})
			if found {
				infraConfigurationEntities = append(infraConfigurationEntities, existingConfigurations[index])
			} else {
				impl.logger.Errorw("missing default configuration keys for platform", "platform", platform, "profileName", profileName, "missingKey", missingKey)
				errMsg := infraErrors.ConfigurationMissingError(missingKey, profileName, platform)
				return nil, nil,
					globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
			}
		}
	}
	// Separate creatable and updatable configurations
	creatableInfraConfigurations = make([]*repository.InfraProfileConfigurationEntity, 0, len(infraConfigurationEntities))
	updatableInfraConfigurations = make([]*repository.InfraProfileConfigurationEntity, 0)
	for _, configuration := range infraConfigurationEntities {
		if configuration.Id == 0 {
			creatableInfraConfigurations = append(creatableInfraConfigurations, configuration)
		} else {
			updatableInfraConfigurations = append(updatableInfraConfigurations, configuration)
		}
	}
	// case where existingConfig is not found
	if existingConfigurations == nil && len(updatableInfraConfigurations) > 0 {
		return nil, nil,
			globalUtil.NewApiError(http.StatusBadRequest, infraErrors.UpdatableConfigurationFoundErr, infraErrors.UpdatableConfigurationFoundErr)
	}

	if len(creatableInfraConfigurations) > 0 {
		creatableInfraConfigurations, err = impl.sanitizeCreatableConfigurations(userId, profileName, creatableInfraConfigurations, existingConfigurations)
		if err != nil {
			return updatableInfraConfigurations, nil, globalUtil.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
		}
	}

	if len(updatableInfraConfigurations) > 0 {
		updatableInfraConfigurations, err = impl.sanitizeUpdatableConfigurations(updatableInfraConfigurations, existingConfigurations, profileName)
		if err != nil {
			return updatableInfraConfigurations, nil, globalUtil.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
		}
	}
	if len(existingConfigurations) > 0 {
		deletableInfraConfigurations, err := impl.sanitizeDeletableConfigurations(userId, updatableInfraConfigurations, existingConfigurations)
		if err != nil {
			return updatableInfraConfigurations, nil, err
		}
		updatableInfraConfigurations = append(updatableInfraConfigurations, deletableInfraConfigurations...)
	}

	updatableInfraConfigurations = adapter.UpdatePlatformMappingInConfigEntities(updatableInfraConfigurations, existingPlatforms)

	creatableInfraConfigurations = adapter.UpdatePlatformMappingInConfigEntities(creatableInfraConfigurations, existingPlatforms)

	return updatableInfraConfigurations, creatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) sanitizeDeletableConfigurations(userId int32, updatableInfraConfigurations, existingConfigurations []*repository.InfraProfileConfigurationEntity) ([]*repository.InfraProfileConfigurationEntity, error) {
	deletableInfraConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0)
	if len(updatableInfraConfigurations) > 0 {
		// Create a map for existing configuration IDs for O(1) lookup
		updatableConfigMap := make(map[int]*repository.InfraProfileConfigurationEntity, len(updatableInfraConfigurations))
		for _, updatableConfig := range updatableInfraConfigurations {
			updatableConfigMap[updatableConfig.Id] = updatableConfig
		}
		for _, existingConfig := range existingConfigurations {
			if _, exists := updatableConfigMap[existingConfig.Id]; !exists {
				existingConfig.Active = false
				existingConfig.AuditLog = sql.AuditLog{
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				}
				deletableInfraConfigurations = append(deletableInfraConfigurations, existingConfig)
			}
		}
	}
	//when no any updatableInfraConfigurations found means we need to delete the existingConfigurations from db
	if len(existingConfigurations) > 0 && len(updatableInfraConfigurations) == 0 {
		for _, existingConfig := range existingConfigurations {
			existingConfig.Active = false
			existingConfig.AuditLog = sql.AuditLog{
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			}
			deletableInfraConfigurations = append(deletableInfraConfigurations, existingConfig)
		}
	}
	return deletableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) sanitizeCreatableConfigurations(userId int32, profileName string, creatableInfraConfigurations []*repository.InfraProfileConfigurationEntity,
	existingConfigurations []*repository.InfraProfileConfigurationEntity) ([]*repository.InfraProfileConfigurationEntity, error) {
	// Create a map with composite keys (Key|Platform) for existing configurations
	existingConfigMap := make(map[string]*repository.InfraProfileConfigurationEntity, len(existingConfigurations))
	for _, existingConfig := range existingConfigurations {
		compositeKey := util.GetConfigCompositeKey(existingConfig)
		existingConfigMap[compositeKey] = existingConfig
	}

	// Collect duplicate configurations
	var duplicateConfigs []string
	for _, creatableConfig := range creatableInfraConfigurations {
		compositeKey := util.GetConfigCompositeKey(creatableConfig)
		if _, exists := existingConfigMap[compositeKey]; exists {
			duplicateConfigs = append(duplicateConfigs, fmt.Sprintf("Key: %s, UniqueId: %s", util.GetConfigKeyStr(creatableConfig.Key), creatableConfig.UniqueId))
		}
	}
	// If duplicates are found, log and return an error
	if len(duplicateConfigs) > 0 {
		impl.logger.Errorw("Duplicate creatable configurations detected", "duplicateConfigs", duplicateConfigs, "profileName", profileName)
		errMsg := fmt.Sprintf("cannot create configurations as the following Key and Platform combinations already exist in the %q profile: %v", profileName, duplicateConfigs)
		return nil, globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}

	// Update configurations with ProfileId, Active status, and AuditLog
	for _, config := range creatableInfraConfigurations {
		config.Active = true
		config.AuditLog = sql.NewDefaultAuditLog(userId)
	}
	return creatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) sanitizeUpdatableConfigurations(updatableInfraConfigurations []*repository.InfraProfileConfigurationEntity, existingConfigurations []*repository.InfraProfileConfigurationEntity, profileName string) ([]*repository.InfraProfileConfigurationEntity, error) {
	// Create a map for existing configuration IDs for O(1) lookup
	existingConfigMap := make(map[int]*repository.InfraProfileConfigurationEntity, len(existingConfigurations))
	for _, existingConfig := range existingConfigurations {
		existingConfigMap[existingConfig.Id] = existingConfig
	}
	// Collect invalid configuration IDs
	var invalidConfigs []int
	for _, updatableConfig := range updatableInfraConfigurations {
		if _, exists := existingConfigMap[updatableConfig.Id]; !exists {
			invalidConfigs = append(invalidConfigs, updatableConfig.Id)
		}
	}
	// If any invalid configurations found, log and return an error
	if len(invalidConfigs) > 0 {
		impl.logger.Errorw("Invalid updatable configurations detected", "invalidConfigIds", invalidConfigs, "profileName", profileName)
		errMsg := fmt.Sprintf("cannot update configurations with ids %v as they do not belong to %q profile", invalidConfigs, profileName)
		return nil, globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return updatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) getAppliedInfraConfigForProfile(appliedProfileConfig, defaultProfileConfig *v1.ProfileBeanDto, variableSnapshots map[string]map[string]string, targetPlatformsList []string) (map[string]*v1.InfraConfig, error) {
	resp := make(map[string]*v1.InfraConfig)
	for _, targetPlatform := range targetPlatformsList {
		appliedConfiguration := impl.getAppliedConfigurationForTargetPlatform(appliedProfileConfig, defaultProfileConfig, targetPlatform)
		infraConfigForTrigger, err := impl.getInfraConfigurationForTrigger(appliedConfiguration)
		if err != nil {
			impl.logger.Errorw("error in updating configurations", "appliedConfiguration", appliedConfiguration, "error", err)
			return resp, err
		}
		if variableSnapshot, ok := variableSnapshots[targetPlatform]; ok && variableSnapshot != nil {
			infraConfigForTrigger = infraConfigForTrigger.SetCiVariableSnapshot(variableSnapshot)
		}
		resp[targetPlatform] = infraConfigForTrigger
	}
	return resp, nil
}

// getAppliedProfileForTriggerScope fetches the applied configuration for the given target platform
//   - CASE 1: target platform exists in the applied configuration,
//     return the applied configuration for the target platform
//   - CASE 2: target platform does not exist in the applied configuration, but exists in the default configuration,
//     return the default configuration for the target platform
//   - CASE 3: target platform does not exist in the applied configuration and default configuration,
//     return the runner configuration of the default.
//     --------------------------------- Flow Chart ------------------------------------
//     |		Applied Config								   Default Config		   |
//     |	 -------------------- 							-------------------- 	   |
//     |	|    linux/arm64	 |						   |    linux/arm64	 	|	   |
//     |	| 	 runner			 |						   |    linux/amd64	 	|	   |
//     |	| 	 				 |						   |    runner			|	   |
//     |	 -------------------- 							-------------------- 	   |
//     |	  ------------------------------------------------------------------  	   |
//     |	 | Target Platform   | 				Returns 						|  	   |
//     |	  ------------------------------------------------------------------  	   |
//     |	 | linux/arm64		 |				Applied Config[linux/arm64]	 	|  	   |
//     |	 | linux/amd64		 |				Default Config[linux/amd64]	 	|  	   |
//     |	 | linux/arm/v7		 |				Default Config[runner]	 		|  	   |
//     |	  ------------------------------------------------------------------  	   |
//     ---------------------------------------------------------------------------------
func (impl *InfraConfigServiceImpl) getAppliedConfigurationForTargetPlatform(appliedProfileConfig, defaultProfileConfig *v1.ProfileBeanDto, targetPlatform string) []*v1.ConfigurationBean {
	var appliedConfiguration []*v1.ConfigurationBean
	if targetConfiguration, ok := appliedProfileConfig.GetConfigurations()[targetPlatform]; ok && targetConfiguration != nil {
		appliedConfiguration = appliedProfileConfig.GetConfigurations()[targetPlatform]
	} else {
		if defaultConfiguration, ok := defaultProfileConfig.GetConfigurations()[targetPlatform]; ok && defaultConfiguration != nil {
			appliedConfiguration = defaultConfiguration
		} else {
			appliedConfiguration = defaultProfileConfig.GetConfigurations()[v1.RUNNER_PLATFORM]
		}
	}
	return appliedConfiguration
}

func (impl *InfraConfigServiceImpl) getInfraConfigurationForTrigger(configBeans []*v1.ConfigurationBean) (*v1.InfraConfig, error) {
	var err error
	infraConfiguration := &v1.InfraConfig{}
	for _, configBean := range configBeans {
		infraConfiguration, err = impl.infraConfigClient.OverrideInfraConfig(infraConfiguration, configBean)
		if err != nil {
			impl.logger.Errorw("error in overriding config", "configBean", configBean, "error", err)
			return infraConfiguration, err
		}
	}
	return infraConfiguration, nil
}

// getAppliedProfileForTriggerScope here we are returning the default platform configuration in case no other platform config found
func (impl *InfraConfigServiceImpl) getAppliedProfileForTriggerScope(scope *v1.Scope) (*v1.ProfileBeanDto, *v1.ProfileBeanDto, error) {
	// Fetching scope-specific configurations and associated profile IDs
	infraProfilesEntities, appliedProfileIds, err := impl.getInfraProfilesByScope(scope, true)
	// If we got an error other than NO_PROPERTIES_FOUND_ERROR, return it
	if err != nil && !errors.Is(err, infraErrors.NoPropertiesFoundError) {
		impl.logger.Errorw("error in fetching configurations", "scope", scope, "error", err)
		return nil, nil, err
	}
	profilesMap, defaultProfileId, err := impl.fillConfigurationsInProfiles(infraProfilesEntities)
	if err != nil {
		impl.logger.Errorw("error in filling configuration in profile objects", "infraProfilesEntities", infraProfilesEntities, "error", err)
		return nil, nil, err
	}
	defaultProfile := profilesMap[defaultProfileId]
	// CASE 1: If no profileIds => we are in the "global only" scenario
	if len(appliedProfileIds) == 0 {
		// Return all platforms from the global config (or whichever subset you want).
		return defaultProfile, defaultProfile, nil
	}
	// Get the profile (assuming one profileId, or adapt if multiple)
	profileId := appliedProfileIds[0]
	return profilesMap[profileId], defaultProfile, nil
}

func (impl *InfraConfigServiceImpl) fillConfigurationsInProfiles(profiles []*repository.InfraProfileEntity) (map[int]*v1.ProfileBeanDto, int, error) {
	// override profileIds with the profiles fetched from db
	profileIds := make([]int, 0, len(profiles))
	for _, profile := range profiles {
		profileIds = append(profileIds, profile.Id)
	}
	profilesMap := make(map[int]*v1.ProfileBeanDto)
	defaultProfileId := 0
	for _, profile := range profiles {
		profilesMap[profile.Id] = adapter.ConvertToProfileBean(profile)
		if profile.Name == v1.GLOBAL_PROFILE_NAME {
			defaultProfileId = profile.Id
		}
	}

	// find the configurations for the profileIds
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileIds(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "profileIds", profileIds, "error", err)
		return nil, 0, err
	}
	profilePlatforms, err := impl.infraProfileRepo.GetPlatformsByProfileIds(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "profileIds", profileIds, "error", err)
		return nil, 0, err
	}

	profilePlatformMap, err := impl.infraConfigClient.ConvertToProfilePlatformMap(infraConfigurations, profilesMap, profilePlatforms)
	if err != nil {
		impl.logger.Errorw("error in converting profile platform map  from infra configurations", "ConvertToProfilePlatformMap", infraConfigurations, "profilePlatforms", profilePlatforms, "error", err)
		return nil, 0, err
	}

	// map the configurations to their respective profiles
	for _, profile := range profiles {
		profileBean := profilesMap[profile.Id]
		if profileBean.GetBuildxDriverType().IsKubernetes() {
			profileBean.Configurations = profilePlatformMap[profile.Id]
		} else {
			configurations := profilePlatformMap[profile.Id][v1.RUNNER_PLATFORM]
			profileBean.Configurations = map[string][]*v1.ConfigurationBean{v1.RUNNER_PLATFORM: configurations}
		}
		profilesMap[profile.Id] = profileBean
	}

	// fill the default configurations for each profile if any of the default configuration is missing
	defaultProfile, ok := profilesMap[defaultProfileId]
	if !ok {
		impl.logger.Errorw("global profile not found", "defaultProfileId", defaultProfileId)
		return profilesMap, defaultProfileId, errors.New("global profile not found")
	}
	defaultProfile, err = impl.getAppliedConfigurationsForProfile(defaultProfile, defaultProfile.GetConfigurations())
	if err != nil {
		impl.logger.Errorw("error in getting applied config for defaultProfile", "defaultProfile", defaultProfile, "error", err)
		return profilesMap, defaultProfileId, err
	}
	profilesMap[defaultProfileId] = defaultProfile
	for profileId, profile := range profilesMap {
		if profile.GetName() == v1.GLOBAL_PROFILE_NAME {
			// skip global profile, as it is already processed
			continue
		}
		profile, err = impl.getAppliedConfigurationsForProfile(profile, defaultProfile.GetConfigurations())
		if err != nil {
			impl.logger.Errorw("error in getting applied config for profile", "profile", profile, "defaultConfigurations", defaultProfile.GetConfigurations(), "error", err)
			return profilesMap, defaultProfileId, err
		}
		// update map with updated profile
		profilesMap[profileId] = profile
	}
	return profilesMap, defaultProfileId, nil
}

func (impl *InfraConfigServiceImpl) getAppliedConfigurationsForProfile(profile *v1.ProfileBeanDto, defaultConfigurationsMap map[string][]*v1.ConfigurationBean) (*v1.ProfileBeanDto, error) {
	if len(profile.GetConfigurations()) == 0 {
		if profile.Configurations == nil {
			profile.Configurations = make(map[string][]*v1.ConfigurationBean)
		}
		profile.Configurations[v1.RUNNER_PLATFORM] = []*v1.ConfigurationBean{}
	}
	for platform := range profile.GetConfigurations() {
		defaultConfigurations := impl.infraConfigClient.GetDefaultConfigurationForPlatform(platform, defaultConfigurationsMap)
		updatedProfileConfigurations, err := impl.getAppliedConfigurationsForPlatform(platform, profile.GetConfigurations()[platform], defaultConfigurations)
		if err != nil {
			impl.logger.Errorw("error in getting missing configurations from default", "platform", platform, "profileConfigurations", profile.GetConfigurations()[platform], "defaultConfigurations", defaultConfigurations, "error", err)
			return profile, err
		}
		profile = profile.SetPlatformConfigurations(platform, updatedProfileConfigurations)
	}
	return profile, nil
}

func (impl *InfraConfigServiceImpl) getAppliedConfigurationsForPlatform(platform string, profileConfigurations, defaultConfigurations []*v1.ConfigurationBean) ([]*v1.ConfigurationBean, error) {
	updatedProfileConfigurations := make([]*v1.ConfigurationBean, 0)
	// If the profile has configurations, then check for missing configurations,
	// Add the missing configurations to the profile
	for supportedConfigKey := range util.GetConfigKeysMapForPlatform(platform) {
		index, found := sliceUtil.Find(profileConfigurations, func(profileConfiguration *v1.ConfigurationBean) bool {
			return profileConfiguration.Key == supportedConfigKey
		})
		if !found {
			// configuration not found, process default configuration
			updatedProfileConfiguration, err := impl.infraConfigClient.MergeInfraConfigurations(supportedConfigKey, nil, defaultConfigurations)
			if err != nil {
				impl.logger.Errorw("error in merging infra configurations", "defaultConfigurations", defaultConfigurations, "error", err)
				return profileConfigurations, err
			}
			if updatedProfileConfiguration != nil {
				updatedProfileConfigurations = append(updatedProfileConfigurations, updatedProfileConfiguration)
			}
		} else {
			// configuration found, merge it with default configuration
			updatedProfileConfiguration, err := impl.infraConfigClient.MergeInfraConfigurations(supportedConfigKey, profileConfigurations[index], defaultConfigurations)
			if err != nil {
				impl.logger.Errorw("error in merging infra configurations", "profileConfiguration", profileConfigurations[index], "defaultConfigurations", defaultConfigurations, "error", err)
				return profileConfigurations, err
			}
			if updatedProfileConfiguration != nil {
				updatedProfileConfigurations = append(updatedProfileConfigurations, updatedProfileConfiguration)
			}
		}
	}
	return updatedProfileConfigurations, nil
}

func (impl *InfraConfigServiceImpl) createGlobalProfile(tx *pg.Tx) (*repository.InfraProfileEntity, error) {
	// if default profiles not found then create default profile
	defaultProfile := &repository.InfraProfileEntity{
		Name:             v1.GLOBAL_PROFILE_NAME,
		Description:      "",
		BuildxDriverType: impl.getDefaultBuildxDriverType(),
		Active:           true,
		AuditLog:         sql.NewDefaultAuditLog(1),
	}
	err := impl.infraProfileRepo.CreateProfile(tx, defaultProfile)
	if err != nil {
		impl.logger.Errorw("error in creation of global profile", "error", err)
		return nil, err
	}
	return defaultProfile, nil
}

func (impl *InfraConfigServiceImpl) getInfraProfilesByScope(scope *v1.Scope, includeDefault bool) ([]*repository.InfraProfileEntity, []int, error) {
	profileIds, err := impl.getInfraProfileIdsByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in fetching profile ids by scope", "scope", scope, "error", err)
		return nil, profileIds, err
	}
	infraProfilesEntities, err := impl.infraProfileRepo.GetProfileListByIds(profileIds, includeDefault)
	if err != nil {
		impl.logger.Errorw("error in fetching profile entities by ids", "scope", scope, "profileIds", profileIds, "error", err)
		return nil, profileIds, err
	}
	return infraProfilesEntities, profileIds, err
}
