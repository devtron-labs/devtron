package infraConfig

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type InfraConfigService interface {

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*ProfileBean, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error

	// GetProfileList fetches all the profile and their configurations matching the given profileNameLike string.
	// If profileNameLike is empty, it will fetch all the active profiles.
	GetProfileList(profileNameLike string) ([]ProfileBean, []ConfigurationBean, error)

	// GetProfileListMin is a lite weight method which fetches all the profile names.
	GetProfileListMin() ([]string, error)

	CreateProfile(userId int32, profileBean *ProfileBean) error

	// DeleteProfile deletes the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	DeleteProfile(userId int32, profileName string) error

	GetIdentifierList(listFilter *IdentifierListFilter) (*IdentifierProfileResponse, error)

	ApplyProfileToIdentifiers(userId int32, applyIdentifiersRequest InfraProfileApplyRequest) error
}

type InfraConfigServiceImpl struct {
	logger                              *zap.SugaredLogger
	infraProfileRepo                    InfraConfigRepository
	units                               *units.Units
	infraConfig                         *InfraConfig
	appService                          app.AppService
	configurationValidator              Validator
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService
	qualifierMappingService             resourceQualifiers.QualifierMappingService
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo InfraConfigRepository,
	units *units.Units,
	appService app.AppService,
	configurationValidator Validator,
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService,
	qualifierMappingService resourceQualifiers.QualifierMappingService) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return nil, err
	}
	infraProfileService := &InfraConfigServiceImpl{
		logger:                              logger,
		infraProfileRepo:                    infraProfileRepo,
		units:                               units,
		infraConfig:                         infraConfiguration,
		appService:                          appService,
		configurationValidator:              configurationValidator,
		devtronResourceSearchableKeyService: devtronResourceSearchableKeyService,
		qualifierMappingService:             qualifierMappingService,
	}
	err = infraProfileService.loadDefaultProfile()
	return infraProfileService, err
}

func (impl *InfraConfigServiceImpl) GetProfileByName(name string) (*ProfileBean, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(name)
	if err != nil {
		impl.logger.Errorw("error in fetching profile", "profileName", name, "error", err)
		return nil, err
	}

	profileBean := infraProfile.ConvertToProfileBean()
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileIds([]int{infraProfile.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching configurations using profileId", "profileId", infraProfile.Id, "error", err)
		return nil, err
	}

	configurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfigurationEntity) ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})

	profileBean.Configurations = configurationBeans
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileToUpdate *ProfileBean) error {

	// validation
	if err := impl.validate(profileToUpdate); err != nil {
		impl.logger.Errorw("error occurred while validating the profile update request", "profileName", profileName, "profileToUpdate", profileToUpdate, "error", err)
		return err
	}
	// validations end

	infraProfileEntity := profileToUpdate.ConvertToInfraProfileEntity()
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true

	sanitizedUpdatableInfraConfigurations, sanitizedCreatableInfraConfigurations, err := impl.sanitizeAndGetUpdatableAndCreatableConfigurationEntities(userId, profileName, profileToUpdate.Configurations)
	// if sanity failed thow the error
	if err != nil {
		return err
	}

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileBean", profileToUpdate, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfileEntity.UpdatedOn = time.Now()
	infraProfileEntity.UpdatedBy = userId
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfileEntity)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "infraProfile", infraProfileEntity, "error", err)
		return err
	}

	err = impl.infraProfileRepo.UpdateConfigurations(tx, sanitizedUpdatableInfraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "updatableInfraConfigurations", sanitizedUpdatableInfraConfigurations, "error", err)
		return err
	}

	err = impl.infraProfileRepo.CreateConfigurations(tx, sanitizedCreatableInfraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "creatableInfraConfigurations", sanitizedCreatableInfraConfigurations, "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileCreateRequest", profileToUpdate, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) GetProfileList(profileNameLike string) ([]ProfileBean, []ConfigurationBean, error) {
	// fetch all the profiles matching the given profileNameLike filter
	infraProfiles, err := impl.infraProfileRepo.GetProfileList(profileNameLike)
	if err != nil {
		impl.logger.Errorw("error in fetching profiles", "profileNameLike", profileNameLike, "error", err)
		return nil, nil, err
	}
	defaultProfileId := 0
	// extract out profileIds from the profiles
	profileIds := make([]int, len(infraProfiles))
	profilesMap := make(map[int]ProfileBean)
	for i, _ := range infraProfiles {
		profileBean := infraProfiles[i].ConvertToProfileBean()
		if profileBean.Name == DEFAULT_PROFILE_NAME {
			defaultProfileId = profileBean.Id
		}
		profileIds[i] = profileBean.Id
		profilesMap[profileBean.Id] = profileBean
	}

	// fetch all the configurations matching the given profileIds
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileIds(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching configurations of profileIds", "profileIds", profileIds, "error", err)
		return nil, nil, err
	}

	// map the configurations to their respective profiles
	for _, configuration := range infraConfigurations {
		profileBean := profilesMap[configuration.ProfileId]
		configurationBean := configuration.ConvertToConfigurationBean()
		configurationBean.ProfileName = profileBean.Name
		profileBean.Configurations = append(profileBean.Configurations, configurationBean)
		profilesMap[configuration.ProfileId] = profileBean
	}

	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	profileAppCount, err := impl.qualifierMappingService.GetActiveIdentifierCountPerResource(resourceQualifiers.InfraProfile, profileIds, GetIdentifierKey(APPLICATION, searchableKeyNameIdMap))
	if err != nil {
		impl.logger.Errorw("error in fetching app count for non default profiles", "profileIds", profileIds, "error", err)
		return nil, nil, err
	}

	defaultProfileIdentifierCount, err := impl.getIdentifierCountForDefaultProfile(defaultProfileId)
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "", defaultProfileId, defaultProfileId, "error", err)
		return nil, nil, err
	}

	// set app count for each profile
	for _, profileAppCnt := range profileAppCount {
		profileBean := profilesMap[profileAppCnt.ResourceId]
		profileBean.AppCount = profileAppCnt.IdentifierCount
		profilesMap[profileAppCnt.ResourceId] = profileBean
	}

	// fill the default configurations for each profile if any of the default configuration is missing
	defaultProfile := profilesMap[defaultProfileId]
	defaultProfile.AppCount = defaultProfileIdentifierCount
	profilesMap[defaultProfileId] = defaultProfile

	defaultConfigurations := defaultProfile.Configurations
	profiles := make([]ProfileBean, 0, len(profilesMap))
	for profileId, profile := range profilesMap {
		if profile.Name == DEFAULT_PROFILE_NAME {
			profiles = append(profiles, profile)
			// update map with updated profile
			profilesMap[profileId] = profile
			continue
		}
		profile = UpdateProfileMissingConfigurationsWithDefault(profile, defaultConfigurations)
		profiles = append(profiles, profile)
		// update map with updated profile
		profilesMap[profileId] = profile
	}
	return profiles, defaultConfigurations, nil
}

func (impl *InfraConfigServiceImpl) getIdentifierCountForDefaultProfile(defaultProfileId int) (int, error) {
	// default configurations will be inherited to profiles which doesn't have atleast one default configuration key property on its own
	// so, we will find out all profileIds which doesn't inherit any property from default profile.
	// then find the different identifier ids for above profileIds.
	// then find the count of identifiers excluding above identifier Ids.
	profileIds, err := impl.infraProfileRepo.GetProfilesWhichContainsAllDefaultConfigurationKeysWithProfileId(defaultProfileId)
	if err != nil {
		impl.logger.Errorw("error in fetching profileIds which contains all default configuration keys", "error", err)
		return 0, err
	}
	excludeIdentifierIds := make([]int, 0)
	// if such profiles are found , find the appIds on which these profiles are applied
	// else we can just go and get all the active ci/cd appIds
	if len(profileIds) != 0 {
		excludeIdentifierIds, err = impl.qualifierMappingService.GetIdentifierIdsByResourceTypeAndIds(resourceQualifiers.InfraProfile, profileIds, GetIdentifierKey(APPLICATION, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()))
		if err != nil {
			impl.logger.Errorw("error in fetching identifierIds for profileIds", "profileIds", profileIds, "error", err)
			return 0, err
		}
	}
	return impl.appService.GetActiveCiCdAppsCount(excludeIdentifierIds)
}

func (impl *InfraConfigServiceImpl) CreateProfile(userId int32, profileBean *ProfileBean) error {

	if err := impl.validate(profileBean); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileCreateRequest", profileBean, "error", err)
		return err
	}

	infraProfile := profileBean.ConvertToInfraProfileEntity()
	infraProfile.Active = true
	infraProfile.AuditLog = sql.NewDefaultAuditLog(userId)

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to create profile", "profileCreateRequest", profileBean, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	err = impl.infraProfileRepo.CreateProfile(tx, infraProfile)
	if err != nil {
		impl.logger.Errorw("error in creating infra config profile", "infraProfile", infraProfile, "err", err)
		return err
	}

	infraConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(profileBean.Configurations))
	for _, configuration := range profileBean.Configurations {
		infraConfiguration := configuration.ConvertToInfraProfileConfigurationEntity()
		infraConfiguration.Id = 0
		infraConfiguration.Active = true
		infraConfiguration.AuditLog = sql.NewDefaultAuditLog(userId)
		infraConfiguration.ProfileId = infraProfile.Id
		infraConfigurations = append(infraConfigurations, infraConfiguration)
	}

	err = impl.infraProfileRepo.CreateConfigurations(tx, infraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating infra configurations", "infraConfigurations", infraConfigurations, "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction while creating infra configuration profile", "profileCreateRequest", profileBean, "err", err)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) DeleteProfile(userId int32, profileName string) error {
	if profileName == DEFAULT_PROFILE_NAME {
		return errors.New(CannotDeleteDefaultProfile)
	}

	profileToBeDeleted, err := impl.infraProfileRepo.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching profile", "profileName", profileName, "error", err)
		if errors.Is(err, pg.ErrNoRows) {
			return nil
		}
		return err
	}

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to delete profile", "profileName", profileName, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)

	// step1: delete configurations
	err = impl.infraProfileRepo.DeleteConfigurations(tx, profileToBeDeleted.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting configurations", "profileName", profileName, "error", err)
		return err
	}

	// step2: delete profile identifier mappings
	err = impl.qualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.InfraProfile, profileToBeDeleted.Id, sql.AuditLog{UpdatedOn: time.Now(), UpdatedBy: userId}, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting profile identifier mappings", "profileName", profileName, "error", err)
		return err
	}

	// step3: delete profile
	err = impl.infraProfileRepo.DeleteProfile(tx, profileToBeDeleted.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting profile", "profileName", profileName, "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to delete profile", "profileName", profileName, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) GetIdentifierList(listFilter *IdentifierListFilter) (*IdentifierProfileResponse, error) {
	// case-1 : if no profile name is provided get all the identifiers for the first page and return first page results
	// steps: get all the active apps using limit and offset and then fetch profiles for those apps.

	// case-2 : if profile name is provided get those apps which are found in resource_qualifier_mapping table.

	identifierListResponse := &IdentifierProfileResponse{}
	identifiers, err := impl.getIdentifierList(*listFilter, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap())
	if err != nil {
		impl.logger.Errorw("error in fetching identifiers", "listFilter", listFilter, "error", err)
		return nil, err
	}
	profileIds := make([]int, 0)
	totalIdentifiersCount := 0
	for _, identifier := range identifiers {
		if identifier.ProfileId != 0 {
			profileIds = append(profileIds, identifier.ProfileId)
		}
		totalIdentifiersCount = identifier.TotalIdentifierCount
	}

	profilesMap, defaultProfileId, err := impl.getProfilesWithConfigurations(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching profiles with configurations", "profileIds", profileIds, "error", err)
		return nil, err
	}

	for _, identifier := range identifiers {
		// if profile id is 0 or default profile id then set default profile in the identifier
		if identifier.ProfileId == defaultProfileId || identifier.ProfileId == 0 {
			profile := profilesMap[defaultProfileId]
			identifier.Profile = &profile
		} else {
			profile := profilesMap[identifier.ProfileId]
			identifier.Profile = &profile
		}
	}

	overriddenIdentifiersCount, err := impl.qualifierMappingService.GetActiveMappingsCount(resourceQualifiers.InfraProfile)
	if err != nil {
		impl.logger.Errorw("error in fetching total overridden count", "listFilter", listFilter, "error", err)
		return nil, err
	}

	if len(identifiers) == 0 {
		identifiers = []*Identifier{}
	}
	identifierListResponse.Identifiers = identifiers
	identifierListResponse.TotalIdentifierCount = totalIdentifiersCount
	identifierListResponse.OverriddenIdentifierCount = overriddenIdentifiersCount
	return identifierListResponse, nil
}

func (impl *InfraConfigServiceImpl) ApplyProfileToIdentifiers(userId int32, applyIdentifiersRequest InfraProfileApplyRequest) error {

	// check if profile exists or not
	updateToProfile, err := impl.infraProfileRepo.GetProfileByName(applyIdentifiersRequest.UpdateToProfile)
	if err != nil {
		impl.logger.Errorw("error in checking profile exists ", "profileId", applyIdentifiersRequest.UpdateToProfile, "error", err)
		if errors.Is(err, pg.ErrNoRows) {
			return errors.New("cannot apply profile that does not exists")
		}
		return err
	}

	applyIdentifiersRequest.UpdateToProfileId = updateToProfile.Id
	identifierIdNameMap, err := impl.getFilteredIdentifiers(applyIdentifiersRequest)
	if err != nil {
		return err
	}

	if len(identifierIdNameMap) == 0 {
		// return here because there are no identifiers to apply profile
		return errors.New(profileApplyErr)
	}

	// set identifiers in the filter
	identifierIds := make([]int, 0, len(identifierIdNameMap))
	for identifierId, _ := range identifierIdNameMap {
		identifierIds = append(identifierIds, identifierId)
	}

	// fetch default profile
	defaultProfile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return errors.New("default profile does not exists")
		}
		impl.logger.Errorw("error in fetching default profile", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	return impl.applyProfile(userId, applyIdentifiersRequest, searchableKeyNameIdMap, defaultProfile.Id, identifierIdNameMap)
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[ConfigKeyStr]map[string]units.Unit)
	cpuUnits := make(map[string]units.Unit)
	memUnits := make(map[string]units.Unit)
	timeUnits := make(map[string]units.Unit)
	for key, val := range impl.units.GetCpuUnits() {
		cpuUnits[string(key)] = val
	}
	for key, val := range impl.units.GetMemoryUnits() {
		memUnits[string(key)] = val
	}
	for key, val := range impl.units.GetTimeUnits() {
		timeUnits[string(key)] = val
	}

	configurationUnits[CPU_REQUEST] = cpuUnits
	configurationUnits[CPU_LIMIT] = cpuUnits

	configurationUnits[MEMORY_REQUEST] = memUnits
	configurationUnits[MEMORY_LIMIT] = memUnits

	configurationUnits[TIME_OUT] = timeUnits

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) GetProfileListMin() ([]string, error) {
	// fetch all the profiles matching the given profileNameLike filter
	infraProfiles, err := impl.infraProfileRepo.GetActiveProfileNames()
	if err != nil {
		impl.logger.Errorw("error in fetching default profiles", "error", err)
		return nil, err
	}
	return infraProfiles, nil
}

func (impl *InfraConfigServiceImpl) getFilteredIdentifiers(applyIdentifiersRequest InfraProfileApplyRequest) (map[int]string, error) {
	identifierIdNameMap := make(map[int]string)
	// apply profile for those identifiers those qualified by the applyIdentifiersRequest.IdentifiersFilter
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	if applyIdentifiersRequest.IdentifiersFilter != nil {

		// get apps using filter
		identifiersList, err := impl.getIdentifierList(*applyIdentifiersRequest.IdentifiersFilter, searchableKeyNameIdMap)
		if err != nil {
			impl.logger.Errorw("error in fetching identifiers", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
			return identifierIdNameMap, err
		}

		for _, identifier := range identifiersList {
			identifierIdNameMap[identifier.Id] = identifier.Name
		}

	} else {
		// apply profile for those identifiers those are provided in the applyIdentifiersRequest.Identifiers by the user

		// get all the apps with the given identifiers, getting apps because the current supported identifier type is only apps.
		// may need to fetch respective identifier objects in future

		// here we are fetching only the active identifiers list, if user provided identifiers have any inactive at the time of this computation
		// ignore applying profile for those inactive identifiers
		ActiveIdentifiers, err := impl.appService.FindAppByNames(applyIdentifiersRequest.Identifiers)
		if err != nil {
			impl.logger.Errorw("error in fetching apps using ids", "appIds", applyIdentifiersRequest.Identifiers, "error", err)
			return identifierIdNameMap, err
		}

		for _, identifier := range ActiveIdentifiers {
			identifierIdNameMap[identifier.Id] = identifier.AppName
		}

	}
	return identifierIdNameMap, nil
}

func (impl *InfraConfigServiceImpl) applyProfile(userId int32, applyIdentifiersRequest InfraProfileApplyRequest, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, defaultProfileId int, identifierIdNameMap map[int]string) error {
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to apply profile to identifiers", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)

	// mark the old profile identifier mappings inactive as they will be overridden by the new profile
	err = impl.qualifierMappingService.DeleteGivenQualifierMappingsByResourceType(resourceQualifiers.InfraProfile, GetIdentifierKey(APPLICATION, searchableKeyNameIdMap), applyIdentifiersRequest.IdentifierIds, sql.AuditLog{UpdatedOn: time.Now(), UpdatedBy: userId}, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting profile identifier mappings", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}

	// don't store qualifier mappings for default profile
	if applyIdentifiersRequest.UpdateToProfileId != defaultProfileId {
		qualifierMappings := make([]*resourceQualifiers.QualifierMapping, 0, len(applyIdentifiersRequest.Identifiers))
		for _, identifierId := range applyIdentifiersRequest.IdentifierIds {
			qualifierMapping := &resourceQualifiers.QualifierMapping{
				ResourceId:            applyIdentifiersRequest.UpdateToProfileId,
				ResourceType:          resourceQualifiers.InfraProfile,
				Active:                true,
				AuditLog:              sql.NewDefaultAuditLog(userId),
				IdentifierValueInt:    identifierId,
				IdentifierKey:         GetIdentifierKey(APPLICATION, searchableKeyNameIdMap),
				IdentifierValueString: identifierIdNameMap[identifierId],
			}
			qualifierMappings = append(qualifierMappings, qualifierMapping)
		}
		_, err = impl.qualifierMappingService.CreateQualifierMappings(qualifierMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in creating profile identifier mappings", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
			return err
		}
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to apply profile to identifiers", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}
	return err
}

// loadDefaultProfile loads default configurations from environment and save them in db.
// this will only create the default profile only once if not exists in db.(container restarts won't create new default profile everytime)
// this will load the default configurations provided in InfraConfig. if db is in out of sync with InfraConfig then it will create new entries for those missing configurations in db.
func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	// make sure about no rows error
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	profileCreationRequired := errors.Is(err, pg.ErrNoRows)
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	if profileCreationRequired {
		// if default profiles not found then create default profile
		defaultProfile := &InfraProfileEntity{
			Name:        DEFAULT_PROFILE_NAME,
			Description: "",
			Active:      true,
			AuditLog:    sql.NewDefaultAuditLog(1),
		}

		err = impl.infraProfileRepo.CreateProfile(tx, defaultProfile)
		if err != nil {
			impl.logger.Errorw("error in saving default profile", "error", err)
			return err
		}
		profile = defaultProfile
	}

	defaultConfigurationsFromEnv, err := impl.infraConfig.LoadInfraConfigInEntities()
	if err != nil {
		impl.logger.Errorw("error in loading default configurations from environment", "error", err)
		return err
	}

	// get db configurations and create new entries if db is out of sync
	defaultConfigurationsFromDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[ConfigKey]bool)
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
	}

	creatableConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	for _, configurationFromEnv := range defaultConfigurationsFromEnv {
		if !defaultConfigurationsFromDBMap[configurationFromEnv.Key] {
			configurationFromEnv.ProfileId = profile.Id
			configurationFromEnv.Active = true
			configurationFromEnv.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, configurationFromEnv)
		}
	}

	if len(creatableConfigurations) > 0 {
		err = impl.infraProfileRepo.CreateConfigurations(tx, creatableConfigurations)
		if err != nil {
			impl.logger.Errorw("error in saving default configurations", "configurations", creatableConfigurations, "error", err)
			return err
		}
	}

	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save default configurations", "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) getInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {

	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByScope(scope, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap())
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	infraConfigurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfigurationEntity) ConfigurationBean {
		return config.ConvertToConfigurationBean()
	})

	infraConfiguration := &InfraConfig{}
	overrideInfraConfigFunc := func(config ConfigurationBean) {
		switch config.Key {
		case CPU_LIMIT:
			infraConfiguration.setCiLimitCpu(impl.getResolvedValue(config).(string))
		case CPU_REQUEST:
			infraConfiguration.setCiReqCpu(impl.getResolvedValue(config).(string))
		case MEMORY_LIMIT:
			infraConfiguration.setCiLimitMem(impl.getResolvedValue(config).(string))
		case MEMORY_REQUEST:
			infraConfiguration.setCiReqMem(impl.getResolvedValue(config).(string))
		case TIME_OUT:
			infraConfiguration.setCiDefaultTimeout(impl.getResolvedValue(config).(int64))
		}
	}

	getDefaultConfigurationKeys := GetDefaultConfigKeysMap()
	for _, infraConfigBean := range infraConfigurationBeans {
		overrideInfraConfigFunc(infraConfigBean)
		// set the key to false so that we can find the missing configurations
		getDefaultConfigurationKeys[infraConfigBean.Key] = false
	}

	// if configurations found for this scope are less than the default configurations, it means some configurations are missing
	// fill the missing configurations with default configurations
	if len(infraConfigurationBeans) < len(getDefaultConfigurationKeys) {
		defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(DEFAULT_PROFILE_NAME)
		if err != nil {
			impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
			return nil, err
		}

		for _, defaultConfiguration := range defaultConfigurations {
			defaultConfigurationBean := defaultConfiguration.ConvertToConfigurationBean()
			// if the key is found true in the map, it means the configuration is missing for the given scope (search for  hack:))
			if _, ok := getDefaultConfigurationKeys[defaultConfigurationBean.Key]; ok {
				overrideInfraConfigFunc(defaultConfigurationBean)
			}
		}
	}
	return infraConfiguration, nil
}

func (impl *InfraConfigServiceImpl) getResolvedValue(configurationBean ConfigurationBean) interface{} {
	// for timeout we need to get the value in seconds
	if configurationBean.Key == GetConfigKeyStr(TimeOut) {
		// if user ever gives the timeout in float, after conversion to int64 it will be rounded off
		timeUnit := units.TimeUnitStr(configurationBean.Unit)
		return int64(configurationBean.Value * impl.units.GetTimeUnits()[timeUnit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value)
	}
	return fmt.Sprintf("%v%v", configurationBean.Value, configurationBean.Unit)
}

func (impl *InfraConfigServiceImpl) validate(profileToUpdate *ProfileBean) error {
	defaultProfile, err := impl.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileName", DEFAULT_PROFILE_NAME, "profileToUpdate", profileToUpdate, "error", err)
		return err
	}
	defaultConfigurationsKeyMap := GetDefaultConfigKeysMap()
	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileToUpdate.Configurations {
		if _, ok := defaultConfigurationsKeyMap[propertyConfig.Key]; !ok {
			errorMsg := fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key)
			if err == nil {
				err = errors.New(errorMsg)
			}
			err = errors.Wrap(err, errorMsg)
		}
	}

	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	err = impl.configurationValidator.ValidateCpuMem(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) getIdentifierList(listFilter IdentifierListFilter, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*Identifier, error) {
	identifierType := GetIdentifierKey(listFilter.IdentifierType, searchableKeyNameIdMap)
	// for empty profile name we have to get identifiers
	if listFilter.ProfileName == ALL_PROFILES {
		return impl.getIdentifiersListForMiscProfiles(listFilter, identifierType, nil)
	}

	// for default profile
	if listFilter.ProfileName == DEFAULT_PROFILE_NAME {
		identifiers, err := impl.getIdentifiersListForDefaultProfile(listFilter, identifierType)
		return identifiers, err
	}

	// for any other profile
	identifiers, err := impl.getIdentifiersListForNonDefaultProfile(listFilter, identifierType)
	return identifiers, err

}

func (impl *InfraConfigServiceImpl) getIdentifiersListForNonDefaultProfile(listFilter IdentifierListFilter, identifierType int) ([]*Identifier, error) {
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappingsWithIdentifierFilter(resourceQualifiers.InfraProfile, identifierType, listFilter.IdentifierNameLike, listFilter.SortOrder, listFilter.Limit, listFilter.Offset, true)
	if err != nil {
		impl.logger.Errorw("error in fetching identifier mappings", "listFilter", listFilter, "error", err)
		return nil, err
	}
	identifiers := make([]*Identifier, 0)
	for _, mapping := range qualifierMappings {
		identifier := &Identifier{
			Id:                   mapping.IdentifierValueInt,
			Name:                 mapping.IdentifierValueString,
			ProfileId:            mapping.ResourceId,
			TotalIdentifierCount: mapping.TotalCount,
		}
		identifiers = append(identifiers, identifier)
	}
	return identifiers, err
}

func (impl *InfraConfigServiceImpl) getIdentifiersListForMiscProfiles(listFilter IdentifierListFilter, identifierType int, excludeAppIds []int) ([]*Identifier, error) {
	// get apps first and then get their respective profile Ids
	// get apps using filters
	apps, err := impl.appService.FindAppsWithFilter(listFilter.IdentifierNameLike, listFilter.SortOrder, listFilter.Limit, listFilter.Offset, excludeAppIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps using filters", "listFilter", listFilter, "excludeAppIds", excludeAppIds, "error", err)
		return nil, err
	}
	// get profileVsIdentifierMappings
	identifierProfileIdMap, err := impl.fetchIdentifiersWithProfileId(identifierType)
	if err != nil {
		impl.logger.Errorw("error in fetching identifierId vs profileId map", "identifierType", identifierType, "error", err)
		return nil, err
	}

	identifiers := make([]*Identifier, 0)
	for _, app := range apps {
		identifier := &Identifier{
			Id:                   app.Id,
			Name:                 app.AppName,
			TotalIdentifierCount: app.TotalCount,
			ProfileId:            identifierProfileIdMap[app.Id],
		}
		identifiers = append(identifiers, identifier)
	}
	return identifiers, err
}

func (impl *InfraConfigServiceImpl) getIdentifiersListForDefaultProfile(listFilter IdentifierListFilter, identifierType int) ([]*Identifier, error) {

	profileIds, err := impl.infraProfileRepo.GetProfilesWhichContainsAllDefaultConfigurationKeysUsingProfileName()
	if err != nil {
		impl.logger.Errorw("error in fetching profileIds which contains all default configuration keys", "error", err)
		return nil, err
	}

	excludeIdentifierIds := make([]int, 0)
	if len(profileIds) == 0 {
		excludeIdentifierIds, err = impl.qualifierMappingService.GetIdentifierIdsByResourceTypeAndIds(resourceQualifiers.InfraProfile, profileIds, GetIdentifierKey(APPLICATION, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()))
		if err != nil {
			impl.logger.Errorw("error in fetching identifierIds for profileIds", "profileIds", profileIds, "error", err)
			return nil, err
		}
	}

	return impl.getIdentifiersListForMiscProfiles(listFilter, identifierType, excludeIdentifierIds)
}

func (impl *InfraConfigServiceImpl) fetchIdentifiersWithProfileId(identifierType int) (map[int]int, error) {
	qualifierMappings, err := impl.qualifierMappingService.GetQualifierMappingsWithIdentifierFilter(resourceQualifiers.InfraProfile, identifierType, "", "", 0, 0, false)
	if err != nil {
		impl.logger.Errorw("error in fetching identifier mappings for infra profile resource", "error", err)
		return nil, err
	}
	identifierProfileIdMap := make(map[int]int)
	for _, mapping := range qualifierMappings {
		identifierProfileIdMap[mapping.IdentifierValueInt] = mapping.ResourceId
	}
	return identifierProfileIdMap, err
}

func (impl *InfraConfigServiceImpl) sanitizeAndGetUpdatableAndCreatableConfigurationEntities(userId int32, profileName string, configurationBeans []ConfigurationBean) ([]*InfraProfileConfigurationEntity, []*InfraProfileConfigurationEntity, error) {
	infraConfigurationEntities := util.Transform(configurationBeans, func(config ConfigurationBean) *InfraProfileConfigurationEntity {
		// user couldn't delete the configuration for default profile, always set this to active
		if profileName == DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		return config.ConvertToInfraProfileConfigurationEntity()
	})

	// filter out creatable and updatable configurations
	creatableInfraConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(infraConfigurationEntities))
	updatableInfraConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(infraConfigurationEntities))
	for _, configuration := range infraConfigurationEntities {
		if configuration.Id == 0 {
			creatableInfraConfigurations = append(creatableInfraConfigurations, configuration)
		} else {
			updatableInfraConfigurations = append(updatableInfraConfigurations, configuration)
		}
	}

	// return error if creatable configurations exists and the profile is default, as user cannot create configurations for default profile
	if profileName == DEFAULT_PROFILE_NAME && len(creatableInfraConfigurations) > 0 {
		return updatableInfraConfigurations, creatableInfraConfigurations, errors.New(CREATION_BLOCKED_FOR_DEFAULT_PROFILE_CONFIGURATIONS)
	}
	existingConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching existing configuration ids", "profileName", profileName, "error", err)

	}

	if len(updatableInfraConfigurations) > 0 {

		updatableInfraConfigurations, err = impl.sanitizeUpdatableConfigurations(updatableInfraConfigurations, existingConfigurations, profileName)
		if err != nil {
			return updatableInfraConfigurations, nil, err
		}

	}

	if len(creatableInfraConfigurations) > 0 {

		creatableInfraConfigurations, err = impl.sanitizeCreatableConfigurations(userId, profileName, creatableInfraConfigurations, existingConfigurations)
		if err != nil {
			return updatableInfraConfigurations, creatableInfraConfigurations, err
		}
	}

	return updatableInfraConfigurations, creatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) sanitizeCreatableConfigurations(userId int32, profileName string, creatableInfraConfigurations []*InfraProfileConfigurationEntity, existingConfigurations []*InfraProfileConfigurationEntity) ([]*InfraProfileConfigurationEntity, error) {
	// at max there can be 5 default configurations (even in future these can be in the order of 10x)
	// so below double loop won't be problematic
	for _, configuration := range creatableInfraConfigurations {
		for _, existingConfiguration := range existingConfigurations {
			if configuration.Key == existingConfiguration.Key {
				return creatableInfraConfigurations, errors.New(fmt.Sprintf("cannot create configuration with key %s as it already exists in %s profile", configuration.Key, profileName))
			}
		}
	}

	pId, err := impl.infraProfileRepo.GetProfileIdByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching profile", "profileName", profileName, "error", err)
		return creatableInfraConfigurations, err
	}
	for _, configuration := range creatableInfraConfigurations {
		configuration.ProfileId = pId
		configuration.Active = true
		configuration.AuditLog = sql.NewDefaultAuditLog(userId)
	}
	return creatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) sanitizeUpdatableConfigurations(updatableInfraConfigurations []*InfraProfileConfigurationEntity, existingConfigurations []*InfraProfileConfigurationEntity, profileName string) ([]*InfraProfileConfigurationEntity, error) {
	// at max there can be 5 default configurations (even in future these can be in the order of 10x)
	// so below double loop won't be problematic
	for _, updatableInfraConfiguration := range updatableInfraConfigurations {
		profileContainsThisConfiguration := false
		for _, existingConfiguration := range existingConfigurations {
			if updatableInfraConfiguration.Id == existingConfiguration.Id {
				profileContainsThisConfiguration = true
				break
			}
		}
		if !profileContainsThisConfiguration {
			return nil, errors.New(fmt.Sprintf("cannot update configuration with id %d as it does not belong to %s profile", updatableInfraConfiguration.Id, profileName))
		}
	}
	return updatableInfraConfigurations, nil
}

func (impl *InfraConfigServiceImpl) getProfilesWithConfigurations(profileIds []int) (map[int]ProfileBean, int, error) {
	profiles, err := impl.infraProfileRepo.GetProfileListByIds(profileIds, true)
	if err != nil {
		impl.logger.Errorw("error in fetching profiles", "profileIds", profileIds, "error", err)
		return nil, 0, err
	}

	// override profileIds with the profiles fetched from db
	profileIds = []int{}
	for _, profile := range profiles {
		profileIds = append(profileIds, profile.Id)
	}
	profilesMap := make(map[int]ProfileBean)
	defaultProfileId := 0
	for _, profile := range profiles {
		profilesMap[profile.Id] = profile.ConvertToProfileBean()
		if profile.Name == DEFAULT_PROFILE_NAME {
			defaultProfileId = profile.Id
		}
	}

	// find the configurations for the profileIds
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileIds(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "profileIds", profileIds, "error", err)
		return nil, 0, err
	}

	// map the configurations to their respective profiles
	for _, configuration := range infraConfigurations {
		profileBean := profilesMap[configuration.ProfileId]
		configurationBean := configuration.ConvertToConfigurationBean()
		configurationBean.ProfileName = profileBean.Name
		profileBean.Configurations = append(profileBean.Configurations, configurationBean)
		profilesMap[configuration.ProfileId] = profileBean
	}

	// fill the default configurations for each profile if any of the default configuration is missing
	defaultProfile := profilesMap[defaultProfileId]
	defaultConfigurations := defaultProfile.Configurations

	for profileId, profile := range profilesMap {
		if profile.Name == DEFAULT_PROFILE_NAME {
			profilesMap[profileId] = profile
			continue
		}
		profile = UpdateProfileMissingConfigurationsWithDefault(profile, defaultConfigurations)
		// update map with updated profile
		profilesMap[profileId] = profile
	}
	return profilesMap, defaultProfileId, nil
}
