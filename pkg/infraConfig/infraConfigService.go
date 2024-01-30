package infraConfig

import (
	"fmt"
	"github.com/caarlos0/env"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"sort"
	"strings"
	"time"
)

const CannotDeleteDefaultProfile = "cannot delete default profile"
const InvalidProfileName = "profile name is invalid"
const PayloadValidationError = "payload validation failed"

type InfraConfigService interface {

	// todo: @gireesh for all get apis, check if we can get profile and configurations in one db call

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit

	// GetDefaultProfile fetches the default profile and its configurations.
	GetDefaultProfile(setAppCount bool) (*ProfileBean, error)

	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(profileName string) (*ProfileBean, error)

	// GetProfileList fetches all the profile and their configurations matching the given profileNameLike string.
	// If profileNameLike is empty, it will fetch all the active profiles.
	GetProfileList(profileNameLike string) (*ProfilesResponse, error)

	CreateProfile(userId int32, profileBean *ProfileBean) error

	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error

	// DeleteProfile deletes the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	DeleteProfile(profileName string) error

	GetIdentifierList(listFilter *IdentifierListFilter) (*IdentifierProfileResponse, error)

	ApplyProfileToIdentifiers(userId int32, applyIdentifiersRequest InfraProfileApplyRequest) error
}

type InfraConfigServiceImpl struct {
	logger                              *zap.SugaredLogger
	infraProfileRepo                    InfraConfigRepository
	qualifiersMappingRepository         resourceQualifiers.QualifiersMappingRepository
	appRepository                       appRepository.AppRepository
	units                               *units.Units
	infraConfig                         *InfraConfig
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService
	validator                           *validator.Validate
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo InfraConfigRepository,
	qualifiersMappingRepository resourceQualifiers.QualifiersMappingRepository,
	appRepository appRepository.AppRepository,
	units *units.Units,
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService,
	validator *validator.Validate) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return nil, err
	}
	infraProfileService := &InfraConfigServiceImpl{
		logger:                              logger,
		infraProfileRepo:                    infraProfileRepo,
		qualifiersMappingRepository:         qualifiersMappingRepository,
		appRepository:                       appRepository,
		units:                               units,
		devtronResourceSearchableKeyService: devtronResourceSearchableKeyService,
		infraConfig:                         infraConfiguration,
		validator:                           validator,
	}
	err = infraProfileService.loadDefaultProfile()
	return infraProfileService, err
}

func (impl *InfraConfigServiceImpl) GetDefaultProfile(setAppCount bool) (*ProfileBean, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "error", err)
		return nil, err
	}

	profileBean := infraProfile.ConvertToProfileBean()
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId([]int{infraProfile.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}
	if setAppCount {
		identifierCount, err := impl.infraProfileRepo.GetIdentifierCountForDefaultProfile(infraProfile.Id, GetIdentifierKey(APPLICATION, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()))
		if err != nil {
			impl.logger.Errorw("error in fetching app count for default profile", "error", err)
			return nil, err
		}
		profileBean.AppCount = identifierCount
	}
	configurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfiguration) ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})
	profileBean.Configurations = configurationBeans
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) GetProfileByName(profileName string) (*ProfileBean, error) {
	if profileName == "" {
		return nil, errors.New(InvalidProfileName)
	}

	infraProfile, err := impl.infraProfileRepo.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "error", err)
		return nil, err
	}

	profileBean := infraProfile.ConvertToProfileBean()
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId([]int{infraProfile.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfiguration) ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})
	profileBean.Configurations = configurationBeans
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) GetProfileList(profileNameLike string) (*ProfilesResponse, error) {
	// fetch all the profiles matching the given profileNameLike filter
	infraProfiles, err := impl.infraProfileRepo.GetProfileList(profileNameLike)
	if err != nil {
		impl.logger.Errorw("error in fetching default profiles", "profileNameLike", profileNameLike, "error", err)
		return nil, err
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
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "profileIds", profileIds, "error", err)
		return nil, err
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
	profileAppCount, err := impl.infraProfileRepo.GetIdentifierCountForNonDefaultProfiles(profileIds, GetIdentifierKey(APPLICATION, searchableKeyNameIdMap))
	if err != nil {
		impl.logger.Errorw("error in fetching app count for non default profiles", "error", err)
		return nil, err
	}

	defaultProfileIdentifierCount, err := impl.infraProfileRepo.GetIdentifierCountForDefaultProfile(defaultProfileId, GetIdentifierKey(APPLICATION, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()))
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "", defaultProfileId, defaultProfileId, "error", err)
		return nil, err
	}

	// set app count for each profile
	for _, profileAppCnt := range profileAppCount {
		profileBean := profilesMap[profileAppCnt.ProfileId]
		profileBean.AppCount = profileAppCnt.IdentifierCount
		profilesMap[profileAppCnt.ProfileId] = profileBean
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

	// order profiles by name
	sort.Slice(profiles, func(i, j int) bool {
		if strings.Compare(profiles[i].Name, profiles[j].Name) <= 0 {
			return true
		}
		return false
	})

	resp := &ProfilesResponse{
		Profiles: profiles,
	}
	resp.DefaultConfigurations = defaultConfigurations
	resp.ConfigurationUnits = impl.GetConfigurationUnits()
	return resp, nil
}

func (impl *InfraConfigServiceImpl) CreateProfile(userId int32, profileBean *ProfileBean) error {
	defaultProfile, err := impl.GetDefaultProfile(false)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileCreateRequest", profileBean, "error", err)
		return err
	}
	if err := impl.Validate(profileBean, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileCreateRequest", profileBean, "error", err)
		return err
	}

	infraProfile := profileBean.ConvertToInfraProfile()
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

	infraConfigurations := make([]*InfraProfileConfiguration, 0, len(profileBean.Configurations))
	for _, configuration := range profileBean.Configurations {
		infraConfiguration := configuration.ConvertToInfraProfileConfiguration()
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

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error {
	if profileName == "" || (profileName == DEFAULT_PROFILE_NAME && profileBean.Name != DEFAULT_PROFILE_NAME) {
		return errors.New(InvalidProfileName)
	}

	// validation
	defaultProfile, err := impl.GetDefaultProfile(false)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileCreateRequest", profileBean, "error", err)
		return err
	}
	if err := impl.Validate(profileBean, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileCreateRequest", profileBean, "error", err)
		return err
	}

	// validations end

	infraProfile := profileBean.ConvertToInfraProfile()
	// user couldn't delete the profile, always set this to active
	infraProfile.Active = true
	infraConfigurations := util.Transform(profileBean.Configurations, func(config ConfigurationBean) *InfraProfileConfiguration {
		// user couldn't delete the configuration for default profile, always set this to active
		if profileName == DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		return config.ConvertToInfraProfileConfiguration()
	})

	// filter out creatable and updatable configurations
	creatableInfraConfigurations := make([]*InfraProfileConfiguration, 0, len(infraConfigurations))
	updatableInfraConfigurations := make([]*InfraProfileConfiguration, 0, len(infraConfigurations))
	for _, configuration := range infraConfigurations {
		if configuration.Id == 0 {
			if profileName == DEFAULT_PROFILE_NAME {
				return errors.New("cannot create new configuration for default profile")
			}
			creatableInfraConfigurations = append(creatableInfraConfigurations, configuration)
		} else {
			updatableInfraConfigurations = append(updatableInfraConfigurations, configuration)
		}
	}

	existingConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching existing configuration ids", "profileName", profileName, "error", err)
		if !errors.Is(err, pg.ErrNoRows) {
			return err
		}
	}

	if len(updatableInfraConfigurations) > 0 {

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
				return errors.New(fmt.Sprintf("cannot update configuration with id %d as it does not belong to %s profile", updatableInfraConfiguration.Id, profileName))
			}
		}

	}

	if len(creatableInfraConfigurations) > 0 {

		// at max there can be 5 default configurations (even in future these can be in the order of 10x)
		// so below double loop won't be problematic
		for _, configuration := range creatableInfraConfigurations {
			for _, existingConfiguration := range existingConfigurations {
				if configuration.Key == existingConfiguration.Key {
					return errors.New(fmt.Sprintf("cannot create configuration with key %s as it already exists in %s profile", configuration.Key, profileName))
				}
			}
		}

		pId, err := impl.infraProfileRepo.GetProfileIdByName(profileName)
		if err != nil {
			impl.logger.Errorw("error in fetching profile", "profileName", profileName, "error", err)
			return err
		}
		for _, configuration := range creatableInfraConfigurations {
			configuration.ProfileId = pId
			configuration.Active = true
			configuration.AuditLog = sql.NewDefaultAuditLog(userId)
		}
	}

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileBean", profileBean, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfile.UpdatedOn = time.Now()
	infraProfile.UpdatedBy = userId
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfile)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "infraProfile", infraProfile, "error", err)
		return err
	}

	err = impl.infraProfileRepo.UpdateConfigurations(tx, updatableInfraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "updatableInfraConfigurations", updatableInfraConfigurations, "error", err)
		return err
	}

	err = impl.infraProfileRepo.CreateConfigurations(tx, creatableInfraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "creatableInfraConfigurations", creatableInfraConfigurations, "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileCreateRequest", profileBean, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) DeleteProfile(profileName string) error {
	if profileName == "" {
		return errors.New(InvalidProfileName)
	}

	if profileName == DEFAULT_PROFILE_NAME {
		return errors.New(CannotDeleteDefaultProfile)
	}

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to delete profile", "profileName", profileName, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)

	// step1: delete profile
	err = impl.infraProfileRepo.DeleteProfile(tx, profileName)
	if err != nil {
		impl.logger.Errorw("error in deleting profile", "profileName", profileName, "error", err)
		return err
	}

	// step2: delete configurations
	err = impl.infraProfileRepo.DeleteConfigurations(tx, profileName)
	if err != nil {
		impl.logger.Errorw("error in deleting configurations", "profileName", profileName, "error", err)
	}

	// step3: delete profile identifier mappings
	err = impl.infraProfileRepo.DeleteProfileIdentifierMappings(tx, profileName)
	if err != nil {
		impl.logger.Errorw("error in deleting profile identifier mappings", "profileName", profileName, "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to delete profile", "profileName", profileName, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if profile.Id != 0 {
		// return here because entry already exists and we dont need to create it again
		return nil
	}

	infraConfiguration := impl.infraConfig
	cpuLimit, err := infraConfiguration.LoadCiLimitCpu()
	if err != nil {
		return err
	}
	memLimit, err := infraConfiguration.LoadCiLimitMem()
	if err != nil {
		return err
	}
	cpuReq, err := infraConfiguration.LoadCiReqCpu()
	if err != nil {
		return err
	}
	memReq, err := infraConfiguration.LoadCiReqMem()
	if err != nil {
		return err
	}
	timeout, err := infraConfiguration.LoadDefaultTimeout()
	if err != nil {
		return err
	}

	defaultConfigurations := []*InfraProfileConfiguration{cpuLimit, memLimit, cpuReq, memReq, timeout}
	defaultProfile := &InfraProfile{
		Name:        DEFAULT_PROFILE_NAME,
		Description: "",
		Active:      true,
		AuditLog:    sql.NewDefaultAuditLog(1),
	}
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)
	err = impl.infraProfileRepo.CreateProfile(tx, defaultProfile)
	if err != nil {
		impl.logger.Errorw("error in saving default profile", "error", err)
		return err
	}

	util.Transform(defaultConfigurations, func(config *InfraProfileConfiguration) *InfraProfileConfiguration {
		config.ProfileId = defaultProfile.Id
		config.Active = true
		config.AuditLog = sql.NewDefaultAuditLog(1)
		return config
	})
	err = impl.infraProfileRepo.CreateConfigurations(tx, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in saving default configurations", "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save default configurations", "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[ConfigKeyStr]map[string]units.Unit)
	configurationUnits[CPU_REQUEST] = impl.units.GetCpuUnits()
	configurationUnits[CPU_LIMIT] = impl.units.GetCpuUnits()

	configurationUnits[MEMORY_REQUEST] = impl.units.GetMemoryUnits()
	configurationUnits[MEMORY_LIMIT] = impl.units.GetMemoryUnits()

	configurationUnits[TIME_OUT] = impl.units.GetTimeUnits()

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileBean *ProfileBean, defaultProfile *ProfileBean) error {
	err := impl.validator.Struct(profileBean)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	defaultConfigurationsKeyMap := GetDefaultConfigKeysMap()
	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileBean.Configurations {
		if _, ok := defaultConfigurationsKeyMap[propertyConfig.Key]; !ok {
			if err == nil {
				err = errors.New(fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key))
			}
			err = errors.Wrap(err, fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key))
		}
	}

	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	err = impl.validateCpuMem(profileBean, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *ProfileBean, defaultProfile *ProfileBean) error {

	configurationUnits := impl.units
	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *ConfigurationBean
		cpuReq   *ConfigurationBean
		memLimit *ConfigurationBean
		memReq   *ConfigurationBean
	)

	for i, _ := range profileBean.Configurations {
		// get cpu limit and req
		switch profileBean.Configurations[i].Key {
		case CPU_LIMIT:
			cpuLimit = &profileBean.Configurations[i]
		case CPU_REQUEST:
			cpuReq = &profileBean.Configurations[i]
		case MEMORY_LIMIT:
			memLimit = &profileBean.Configurations[i]
		case MEMORY_REQUEST:
			memReq = &profileBean.Configurations[i]
		}
	}

	for i, _ := range defaultProfile.Configurations {
		// get cpu limit and req
		switch defaultProfile.Configurations[i].Key {
		case CPU_LIMIT:
			if cpuLimit == nil {
				cpuLimit = &defaultProfile.Configurations[i]
			}
		case CPU_REQUEST:
			if cpuReq == nil {
				cpuReq = &defaultProfile.Configurations[i]
			}
		case MEMORY_LIMIT:
			if memLimit == nil {
				memLimit = &defaultProfile.Configurations[i]
			}
		case MEMORY_REQUEST:
			if memReq == nil {
				memReq = &defaultProfile.Configurations[i]
			}
		}
	}

	// validate cpu
	cpuLimitUnitSuffix := units.GetCPUUnit(units.CPUUnitStr(cpuLimit.Unit))
	cpuReqUnitSuffix := units.GetCPUUnit(units.CPUUnitStr(cpuReq.Unit))
	var cpuLimitUnit units.Unit
	var cpuReqUnit units.Unit
	for cpuUnitSuffix, cpuUnit := range configurationUnits.GetCpuUnits() {
		if string(units.GetCPUUnitStr(cpuLimitUnitSuffix)) == cpuUnitSuffix {
			cpuLimitUnit = cpuUnit
		}

		if string(units.GetCPUUnitStr(cpuReqUnitSuffix)) == cpuUnitSuffix {
			cpuReqUnit = cpuUnit
		}

	}

	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRationCPU := cpuLimit.Value / cpuReq.Value
	convFactorCPU := cpuLimitUnit.ConversionFactor / cpuReqUnit.ConversionFactor

	if limitToReqRationCPU*convFactorCPU < 1 {
		return errors.New("cpu limit should not be less than cpu request")
	}

	// validate mem

	memLimitUnitSuffix := units.GetMemoryUnit(units.MemoryUnitStr(memLimit.Unit))
	memReqUnitSuffix := units.GetMemoryUnit(units.MemoryUnitStr(memReq.Unit))
	var memLimitUnit units.Unit
	var memReqUnit units.Unit

	for memUnitSuffix, memUnit := range configurationUnits.GetMemoryUnits() {
		if string(units.GetMemoryUnitStr(memLimitUnitSuffix)) == memUnitSuffix {
			memLimitUnit = memUnit
		}

		if string(units.GetMemoryUnitStr(memReqUnitSuffix)) == memUnitSuffix {
			memReqUnit = memUnit
		}
	}

	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	// not directly comparing lim*lf with req*rf because this multiplication can overflow float64 limit
	// so, we are dividing lim/req and lf/rf and then comparing.
	limitToReqRationMem := memLimit.Value / memReq.Value
	convFactorMem := memLimitUnit.ConversionFactor / memReqUnit.ConversionFactor

	if limitToReqRationMem*convFactorMem < 1 {
		return errors.New("memory limit should not be less than memory request")
	}

	return nil
}

func (impl *InfraConfigServiceImpl) GetIdentifierList(listFilter *IdentifierListFilter) (*IdentifierProfileResponse, error) {
	// case-1 : if no profile name is provided get all the identifiers for the first page and return first page results
	// steps: get all the active apps using limit and offset and then fetch profiles for those apps.

	// case-2 : if profile name is provided get those apps which are found in resource_qualifier_mapping table.

	identifierListResponse := &IdentifierProfileResponse{}
	identifiers, err := impl.infraProfileRepo.GetIdentifierList(*listFilter, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap())
	if err != nil {
		impl.logger.Errorw("error in fetching identifiers", "listFilter", listFilter, "error", err)
		return nil, err
	}
	profileIds := make([]int, 0)
	totalIdentifiersCount := 0
	overriddenIdentifiersCount := 0
	for _, identifier := range identifiers {
		profileIds = append(profileIds, identifier.ProfileId)
		totalIdentifiersCount = identifier.TotalIdentifierCount
		overriddenIdentifiersCount = identifier.OverriddenIdentifierCount
	}

	profiles, err := impl.infraProfileRepo.GetProfileListByIds(profileIds, true)
	if err != nil {
		impl.logger.Errorw("error in fetching profiles", "profileIds", profileIds, "error", err)
		return nil, err
	}

	// override profileIds with the profiles fetched from db
	profileIds = []int{}
	for _, profile := range profiles {
		profileIds = append(profileIds, profile.Id)
	}
	profilesMap := make(map[int]ProfileBean)
	defaultProfileId := 0
	for _, profile := range profiles {
		profileIds = append(profileIds, profile.Id)
		profilesMap[profile.Id] = profile.ConvertToProfileBean()
		if profile.Name == DEFAULT_PROFILE_NAME {
			defaultProfileId = profile.Id
		}
	}

	// find the configurations for the profileIds
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId(profileIds)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "profileIds", profileIds, "error", err)
		return nil, err
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

	if len(identifiers) == 0 {
		identifiers = []*Identifier{}
	}
	identifierListResponse.Identifiers = identifiers
	identifierListResponse.TotalIdentifierCount = totalIdentifiersCount
	identifierListResponse.OverriddenIdentifierCount = overriddenIdentifiersCount
	return identifierListResponse, nil
}

func (impl *InfraConfigServiceImpl) ApplyProfileToIdentifiers(userId int32, applyIdentifiersRequest InfraProfileApplyRequest) error {
	if applyIdentifiersRequest.IdentifiersFilter == nil && applyIdentifiersRequest.Identifiers == nil {
		return errors.New("invalid apply request")
	}

	if applyIdentifiersRequest.IdentifiersFilter != nil && applyIdentifiersRequest.Identifiers != nil {
		return errors.New("invalid apply request")
	}

	// check if profile exists or not
	updateToProfile, err := impl.infraProfileRepo.GetProfileByName(applyIdentifiersRequest.UpdateToProfile)
	if err != nil && errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking profile exists ", "profileId", applyIdentifiersRequest.UpdateToProfile, "error", err)
		return err
	}

	if errors.Is(err, pg.ErrNoRows) || updateToProfile == nil {
		return errors.New("cannot apply profile that does not exists")
	}

	applyIdentifiersRequest.UpdateToProfileId = updateToProfile.Id
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()

	identifierIdNameMap := make(map[int]string)

	// apply profile for those identifiers those qualified by the applyIdentifiersRequest.IdentifiersFilter
	if applyIdentifiersRequest.IdentifiersFilter != nil {
		// validate IdentifiersFilter
		err = impl.validator.Struct(applyIdentifiersRequest.IdentifiersFilter)
		if err != nil {
			err = errors.Wrap(err, PayloadValidationError)
			return err
		}

		// get apps using filter
		identifiersList, err := impl.infraProfileRepo.GetIdentifierList(*applyIdentifiersRequest.IdentifiersFilter, searchableKeyNameIdMap)
		if err != nil {
			impl.logger.Errorw("error in fetching identifiers", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
			return err
		}
		// set identifiers in the filter
		identifierIds := make([]int, 0, len(identifiersList))
		for _, identifier := range identifiersList {
			identifierIds = append(identifierIds, identifier.Id)
			identifierIdNameMap[identifier.Id] = identifier.Name
		}
		// set identifierIds in the filter
		applyIdentifiersRequest.IdentifierIds = identifierIds

	} else {
		// apply profile for those identifiers those are provided in the applyIdentifiersRequest.Identifiers by the user

		// get all the apps with the given identifiers, getting apps because the current supported identifier type is only apps.
		// may need to fetch respective identifier objects in future

		// here we are fetching only the active identifiers list, if user provided identifiers have any inactive at the time of this computation
		// ignore applying profile for those inactive identifiers
		ActiveIdentifierIds, err := impl.appRepository.FindIdsByNames(applyIdentifiersRequest.Identifiers)
		if err != nil {
			impl.logger.Errorw("error in fetching apps using ids", "appIds", applyIdentifiersRequest.Identifiers, "error", err)
			return err
		}

		// reset the identifierIds in the applyProfileRequest with active identifiers for further processing
		applyIdentifiersRequest.IdentifierIds = ActiveIdentifierIds
	}

	// fetch default profile
	defaultProfile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}
	return impl.applyProfile(userId, applyIdentifiersRequest, searchableKeyNameIdMap, defaultProfile.Id, identifierIdNameMap)
}

func (impl *InfraConfigServiceImpl) applyProfile(userId int32, applyIdentifiersRequest InfraProfileApplyRequest, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, defaultProfileId int, identifierIdNameMap map[int]string) error {
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to apply profile to identifiers", "applyIdentifiersRequest", applyIdentifiersRequest, "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)

	// mark the old profile identifier mappings inactive as they will be overridden by the new profile
	err = impl.infraProfileRepo.DeleteProfileIdentifierMappingsByIds(tx, userId, applyIdentifiersRequest.IdentifierIds, APPLICATION, searchableKeyNameIdMap)
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
		_, err = impl.qualifiersMappingRepository.CreateQualifierMappings(qualifierMappings, tx)
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

func (impl *InfraConfigServiceImpl) getInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {

	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByScope(scope, impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap())
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	infraConfigurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfiguration) ConfigurationBean {
		return config.ConvertToConfigurationBean()
	})

	infraConfiguration := &InfraConfig{}
	updateInfraConfig := func(config ConfigurationBean) {
		switch config.Key {
		case CPU_LIMIT:
			infraConfiguration.setCiLimitCpu(impl.GetResolvedValue(config).(string))
		case CPU_REQUEST:
			infraConfiguration.setCiReqCpu(impl.GetResolvedValue(config).(string))
		case MEMORY_LIMIT:
			infraConfiguration.setCiLimitMem(impl.GetResolvedValue(config).(string))
		case MEMORY_REQUEST:
			infraConfiguration.setCiReqMem(impl.GetResolvedValue(config).(string))
		case TIME_OUT:
			infraConfiguration.setCiDefaultTimeout(impl.GetResolvedValue(config).(int64))
		}
	}

	getDefaultConfigurationKeys := GetDefaultConfigKeysMap()
	for _, infraConfigBean := range infraConfigurationBeans {
		updateInfraConfig(infraConfigBean)
		// hack:) , set the key to false so that we can find the missing configurations
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
				updateInfraConfig(defaultConfigurationBean)
			}
		}
	}
	return infraConfiguration, nil
}

func (impl *InfraConfigServiceImpl) GetResolvedValue(configurationBean ConfigurationBean) interface{} {
	// for timeout we need to get the value in seconds
	if configurationBean.Key == GetConfigKeyStr(TimeOut) {
		// if user ever gives the timeout in float, after conversion to int64 it will be rounded off
		return int64(configurationBean.Value * impl.units.GetTimeUnits()[configurationBean.Unit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value)
	}
	return fmt.Sprintf("%v%v", configurationBean.Value, configurationBean.Unit)
}
