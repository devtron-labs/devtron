package deploymentWindow

import (
	"encoding/json"
	mapset "github.com/deckarep/golang-set"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
	"time"
)

func (impl DeploymentWindowServiceImpl) GetStateForAppEnv(targetTime time.Time, appId int, envId int, userId int32) (UserActionState, *EnvironmentState, error) {

	stateResponse, err := impl.GetDeploymentWindowProfileState(targetTime, appId, []int{envId}, 0, userId)
	if err != nil {
		return Allowed, nil, err
	}

	var envState *EnvironmentState
	actionState := Allowed
	if state, ok := stateResponse.EnvironmentStateMap[envId]; ok {
		actionState = state.UserActionState
		envState = &state
		envState.CalculatedAt = targetTime
	}
	return actionState, envState, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileStateAppGroup(targetTime time.Time, selectors []AppEnvSelector, filterForDays int, userId int32) (*DeploymentWindowAppGroupResponse, error) {

	appIdsToOverview, err := impl.GetDeploymentWindowProfileOverviewBulk(selectors)
	if err != nil {
		return nil, err
	}

	superAdmins, userEmailMap, err := impl.getUserInfoMap(err, appIdsToOverview)
	if err != nil {
		return nil, err
	}

	cachedStates := make(map[int]ProfileState)
	appGroupData := make([]AppData, 0)
	var envResponse *DeploymentWindowResponse
	for appId, overview := range appIdsToOverview {

		envResponse, cachedStates, err = impl.calculateStateForEnvironments(targetTime, overview, filterForDays, cachedStates, superAdmins, userEmailMap, userId)
		if err != nil {
			return nil, err
		}

		appGroupData = append(appGroupData, AppData{
			AppId:                 appId,
			DeploymentProfileList: envResponse,
		})
	}
	return &DeploymentWindowAppGroupResponse{AppData: appGroupData}, nil
}

func (impl DeploymentWindowServiceImpl) getUserInfoMap(err error, appIdsToOverview map[int]*DeploymentWindowResponse) ([]int32, map[int32]string, error) {
	superAdmins, err := impl.userService.GetSuperAdminIds()
	if err != nil {
		return nil, nil, err
	}
	allExcludedUserIds := make([]int32, 0)
	for _, response := range appIdsToOverview {
		for _, profile := range response.Profiles {
			deploymentProfile := profile.DeploymentWindowProfile
			if deploymentProfile.IsUserExcluded {
				allExcludedUserIds = append(allExcludedUserIds, deploymentProfile.ExcludedUsersList...)
			}
		}
	}
	allUserIds := utils.FilterDuplicates(append(allExcludedUserIds, superAdmins...))
	allUserInfo, err := impl.userService.GetByIds(allUserIds)
	if err != nil {
		return nil, nil, err
	}
	userInfoMap := make(map[int32]string, 0)
	for _, user := range allUserInfo {
		if strings.Contains(user.EmailId, "@") {
			userInfoMap[user.Id] = user.EmailId
		}
	}

	return superAdmins, userInfoMap, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileState(targetTime time.Time, appId int, envIds []int, filterForDays int, userId int32) (*DeploymentWindowResponse, error) {
	overview, err := impl.GetDeploymentWindowProfileOverview(appId, envIds)
	if err != nil {
		return nil, err
	}

	superAdmins, userEmailMap, err := impl.getUserInfoMap(err, map[int]*DeploymentWindowResponse{0: overview})
	if err != nil {
		return nil, err
	}

	response, _, err := impl.calculateStateForEnvironments(targetTime, overview, filterForDays, make(map[int]ProfileState), superAdmins, userEmailMap, userId)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (impl DeploymentWindowServiceImpl) calculateStateForEnvironments(targetTime time.Time, overview *DeploymentWindowResponse, filterForDays int, calculatedStates map[int]ProfileState, superAdmins []int32, userEmailMap map[int32]string, userId int32) (*DeploymentWindowResponse, map[int]ProfileState, error) {

	envIdToProfileStates := make(map[int][]ProfileState)
	for _, profile := range overview.Profiles {
		envIdToProfileStates[profile.EnvId] = append(envIdToProfileStates[profile.EnvId], profile)
	}

	envIdToEnvironmentState := make(map[int]EnvironmentState)

	var filteredProfileStates []ProfileState
	var appliedProfile *ProfileState
	var excludedUsers []int32
	var excludedUsersEmail []string
	var isAllowed bool
	var err error
	for envId, profileStates := range envIdToProfileStates {

		filteredProfileStates, calculatedStates, appliedProfile, isAllowed, err = impl.evaluateProfileStates(targetTime, profileStates, calculatedStates, filterForDays)
		if err != nil {
			return nil, calculatedStates, err
		}

		filteredProfileStates, excludedUsers, excludedUsersEmail, err = impl.evaluateExcludedUsers(filteredProfileStates, appliedProfile, superAdmins, userEmailMap)
		if err != nil {
			return nil, calculatedStates, err
		}

		// sorting to keep active profiles first
		sort.SliceStable(filteredProfileStates, func(i, j int) bool {
			return filteredProfileStates[i].IsActive
		})

		envState := EnvironmentState{
			ExcludedUsers:      excludedUsers,
			ExcludedUserEmails: excludedUsersEmail,
			AppliedProfile:     appliedProfile,
			UserActionState:    getUserActionStateForUser(isAllowed, excludedUsers, userId),
		}
		envIdToEnvironmentState[envId] = envState
	}
	response := &DeploymentWindowResponse{
		EnvironmentStateMap: envIdToEnvironmentState,
		Profiles:            filteredProfileStates,
	}
	return response, calculatedStates, nil
}

func getUserActionStateForUser(isAllowed bool, excludedUsers []int32, userId int32) UserActionState {
	userActionState := Allowed
	if !isAllowed {
		if slices.Contains(excludedUsers, userId) {
			userActionState = Partial
		} else {
			userActionState = Blocked
		}
	}
	return userActionState
}

func (impl DeploymentWindowServiceImpl) evaluateProfileStates(targetTime time.Time, profileStates []ProfileState, calculatedStates map[int]ProfileState, filterForDays int) ([]ProfileState, map[int]ProfileState, *ProfileState, bool, error) {
	var appliedProfile *ProfileState
	filteredBlackoutProfiles, calculatedStates, _, isBlackoutActive, err := impl.calculateStateForProfiles(targetTime, profileStates, Blackout, calculatedStates, filterForDays)
	if err != nil {
		return nil, calculatedStates, appliedProfile, false, err
	}

	filteredMaintenanceProfiles, calculatedStates, isMaintenanceActive, _, err := impl.calculateStateForProfiles(targetTime, profileStates, Maintenance, calculatedStates, filterForDays)
	if err != nil {
		return nil, calculatedStates, appliedProfile, false, err
	}

	if len(filteredBlackoutProfiles) == 0 && len(filteredMaintenanceProfiles) == 0 {
		return nil, calculatedStates, appliedProfile, true, nil
	}

	isAllowed := !isBlackoutActive && isMaintenanceActive
	allProfiles := append(filteredBlackoutProfiles, filteredMaintenanceProfiles...)
	if isBlackoutActive && isMaintenanceActive { //action is blocked, restriction through blackout
		// if both are active then blackout takes precedence in overall calculation
		appliedProfile = impl.getLongestEndingProfile(filteredBlackoutProfiles, true)
	} else if !isBlackoutActive && !isMaintenanceActive { //action is blocked, restriction through maintenance
		// if nothing is active then earliest starting maintenance will be shown
		appliedProfile = impl.getEarliestStartingProfile(filteredMaintenanceProfiles, true)
	} else if isBlackoutActive && !isMaintenanceActive { //action is blocked, restriction through both
		// longest of restrictions coming from both blackout and maintenance
		appliedProfile = impl.getLongestEndingProfile(allProfiles, true)
	} else if !isBlackoutActive && isMaintenanceActive { //action not blocked
		// applied profile here would be the longest running maintenance profile even if a blackout starts before that
		appliedProfile = impl.getLongestEndingProfile(filteredMaintenanceProfiles, false)
		if appliedProfile == nil {
			appliedProfile = impl.getEarliestStartingProfile(filteredBlackoutProfiles, false)
		}
	}

	return allProfiles, calculatedStates, appliedProfile, isAllowed, nil
}

func (impl DeploymentWindowServiceImpl) evaluateExcludedUsers(allProfiles []ProfileState, appliedProfile *ProfileState, superAdmins []int32, userEmailMap map[int32]string) ([]ProfileState, []int32, []string, error) {
	combinedExcludedUsers, isSuperAdminExcluded := impl.getCombinedUserIds(allProfiles)

	if isSuperAdminExcluded {
		combinedExcludedUsers = utils.FilterDuplicates(append(combinedExcludedUsers, superAdmins...))
	}

	for i, profile := range allProfiles {

		excludedIds := make([]int32, 0)
		if len(profile.DeploymentWindowProfile.ExcludedUsersList) > 0 {
			excludedIds = profile.DeploymentWindowProfile.ExcludedUsersList
		}

		if profile.DeploymentWindowProfile.IsSuperAdminExcluded {
			excludedIds = utils.FilterDuplicates(append(excludedIds, superAdmins...))
		}
		emails := make([]string, 0)
		for _, id := range excludedIds {
			if email, ok := userEmailMap[id]; ok {
				emails = append(emails, email)
			}
		}
		allProfiles[i].ExcludedUserEmails = emails
		allProfiles[i].DeploymentWindowProfile.ExcludedUsersList = excludedIds

		if appliedProfile != nil && profile.DeploymentWindowProfile.Id == appliedProfile.DeploymentWindowProfile.Id {
			appliedProfile.ExcludedUserEmails = emails
		}
	}

	emails := make([]string, 0)
	for _, userId := range combinedExcludedUsers {
		if email, ok := userEmailMap[userId]; ok {
			emails = append(emails, email)
		}
	}

	return allProfiles, combinedExcludedUsers, emails, nil
}

func (impl DeploymentWindowServiceImpl) getCombinedUserIds(profiles []ProfileState) ([]int32, bool) {

	if len(profiles) == 0 {
		return []int32{}, false
	}
	userSet := mapset.NewSet()

	profile := profiles[0]
	excludedUsers := profile.DeploymentWindowProfile.ExcludedUsersList
	if profile.isRestricted() && len(excludedUsers) > 0 {
		userSet = mapset.NewSetFromSlice(utils.ToInterfaceArrayAny(excludedUsers))
	}

	isSuperAdminExcluded := true
	for _, profile := range profiles {

		var users []int32
		if profile.DeploymentWindowProfile.IsUserExcluded {
			users = profile.DeploymentWindowProfile.ExcludedUsersList
		}

		if !profile.DeploymentWindowProfile.IsSuperAdminExcluded {
			isSuperAdminExcluded = false
		}

		profileUserSet := mapset.NewSetFromSlice(utils.ToInterfaceArrayAny(users))
		if profile.isRestricted() {
			userSet = userSet.Intersect(profileUserSet)
		}
	}

	return utils.ToInt32Array(userSet.ToSlice()), isSuperAdminExcluded
}

func (impl DeploymentWindowServiceImpl) getLongestEndingProfile(profiles []ProfileState, filterRestricted bool) *ProfileState {

	if len(profiles) == 0 {
		return nil
	}

	var selectedProfile *ProfileState
	for _, profile := range profiles {
		if filterRestricted != profile.isRestricted() {
			continue
		}
		if selectedProfile == nil || profile.CalculatedTimestamp.After(selectedProfile.CalculatedTimestamp) {
			selectedProfile = &profile
		}
	}
	return selectedProfile
}

func (impl DeploymentWindowServiceImpl) getEarliestStartingProfile(profiles []ProfileState, filterRestricted bool) *ProfileState {
	if len(profiles) == 0 {
		return nil
	}

	var selectedProfile *ProfileState
	for _, profile := range profiles {
		if filterRestricted != profile.isRestricted() {
			continue
		}
		if selectedProfile == nil || profile.CalculatedTimestamp.Before(selectedProfile.CalculatedTimestamp) {
			selectedProfile = &profile
		}
	}

	return selectedProfile
}

func (impl DeploymentWindowServiceImpl) calculateStateForProfiles(targetTime time.Time, profileStates []ProfileState, profileType DeploymentWindowType, calculatedStates map[int]ProfileState, filterForDays int) ([]ProfileState, map[int]ProfileState, bool, bool, error) {

	filteredProfiles := impl.filterForType(profileStates, profileType)

	calculatedProfiles := make([]ProfileState, 0)
	for _, profile := range filteredProfiles {
		// to avoid recomputing states for profiles which have been already computed
		if profileState, ok := calculatedStates[profile.DeploymentWindowProfile.Id]; ok {
			calculatedProfiles = append(calculatedProfiles, profileState)
			continue
		}

		zone := profile.DeploymentWindowProfile.TimeZone
		isActive, windowTimeStamp, window, err := impl.timeWindowService.GetActiveWindow(targetTime, zone, profile.DeploymentWindowProfile.DeploymentWindowList)
		if err != nil {
			impl.logger.Errorw("error in getting active window", "err", err, "targetTime", targetTime, "zone", zone, "profile", profile)
		}
		if window != nil {
			profile.IsActive = isActive
			profile.CalculatedTimestamp = windowTimeStamp
			profile.DeploymentWindowProfile.DeploymentWindowList = []*timeoutWindow.TimeWindow{window}
		}

		calculatedProfiles = append(calculatedProfiles, profile)
		calculatedStates[profile.DeploymentWindowProfile.Id] = profile
	}

	allActive := true
	oneActive := false
	finalProfileStates := make([]ProfileState, 0)
	for _, profile := range filteredProfiles {

		if len(profile.DeploymentWindowProfile.DeploymentWindowList) == 0 {
			// doing nothing if no window is returned
			// this means that no relevant window in the profile was found therefore skipping this profile
			continue
		}
		isActive := profile.IsActive
		if filterForDays > 0 && !isActive && profile.CalculatedTimestamp.Sub(targetTime) > time.Duration(filterForDays)*time.Hour*24 {
			continue
		}

		if !oneActive && isActive {
			oneActive = true
		}
		if allActive && !isActive {
			allActive = false
		}
		finalProfileStates = append(finalProfileStates, profile)
	}
	return finalProfileStates, calculatedStates, allActive, oneActive, nil
}

func (impl DeploymentWindowServiceImpl) filterForType(profileStates []ProfileState, profileType DeploymentWindowType) []ProfileState {
	filteredProfiles := make([]ProfileState, 0)
	for _, state := range profileStates {
		if state.DeploymentWindowProfile.Type == profileType {
			filteredProfiles = append(filteredProfiles, state)
		}
	}
	return filteredProfiles
}

func (impl DeploymentWindowServiceImpl) CreateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error) {
	tx, err := impl.StartATransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// create policy
	policy, err := profile.convertToPolicyDataModel(userId)
	if err != nil {
		return nil, err
	}
	policy, err = impl.globalPolicyManager.CreatePolicy(policy, tx)
	if err != nil {
		return nil, err
	}
	profile.Id = policy.Id

	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, profile.TimeZone, userId, err, tx, policy.Id)
	if err != nil {
		return nil, err
	}
	err = impl.CommitATransaction(tx)
	if err != nil {
		return nil, err
	}

	return profile, err
}

func (impl DeploymentWindowServiceImpl) UpdateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error) {
	tx, err := impl.StartATransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// create policy
	policy, err := profile.convertToPolicyDataModel(userId)
	if err != nil {
		return nil, err
	}
	policy, err = impl.globalPolicyManager.UpdatePolicy(policy, tx)
	if err != nil {
		return nil, err
	}
	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, profile.TimeZone, userId, err, tx, policy.Id)
	if err != nil {
		return nil, err
	}
	err = impl.CommitATransaction(tx)
	if err != nil {
		return nil, err
	}
	return profile, err
}

func (impl DeploymentWindowServiceImpl) DeleteDeploymentWindowProfileForId(profileId int, userId int32) error {
	tx, err := impl.StartATransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = impl.globalPolicyManager.DeletePolicyById(profileId, userId)
	if err != nil {
		return err
	}
	err = impl.timeWindowService.UpdateWindowMappings([]*timeoutWindow.TimeWindow{}, "", userId, err, tx, profileId)
	if err != nil {
		return err
	}
	err = impl.CommitATransaction(tx)
	if err != nil {
		return err
	}

	return err
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileForId(profileId int) (*DeploymentWindowProfile, error) {
	//get policy
	policyModel, err := impl.globalPolicyManager.GetPolicyById(profileId)
	if err != nil {
		return nil, err
	}

	idToWindows, err := impl.timeWindowService.GetWindowsForResources([]int{profileId}, repository.DeploymentWindowProfile)
	if err != nil {
		return nil, err
	}

	windows, ok := idToWindows[profileId]
	if !ok {
		return nil, nil
	}
	profilePolicy, err := impl.getPolicyFromModel(policyModel)
	if err != nil {
		return nil, err
	}

	return profilePolicy.toDeploymentWindowProfile(policyModel, windows), nil
}

func (impl DeploymentWindowServiceImpl) getPolicyFromModel(policyModel *bean2.GlobalPolicyBaseModel) (*DeploymentWindowProfilePolicy, error) {
	profilePolicy := &DeploymentWindowProfilePolicy{}
	err := json.Unmarshal([]byte(policyModel.JsonData), &profilePolicy)
	if err != nil {
		return nil, err
	}
	return profilePolicy, nil
}

func (impl DeploymentWindowServiceImpl) ListDeploymentWindowProfiles() ([]*DeploymentWindowProfileMetadata, error) {
	//get policy
	policyModels, err := impl.globalPolicyManager.GetAllActiveByType(bean2.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW)
	if err != nil {
		return nil, err
	}

	allProfiles := make([]*DeploymentWindowProfileMetadata, 0)
	for _, model := range policyModels {
		policy, err := impl.getPolicyFromModel(model)
		if err != nil {
			return nil, err
		}
		allProfiles = append(allProfiles, &DeploymentWindowProfileMetadata{
			Description: model.Description,
			Id:          model.Id,
			Name:        model.Name,
			Type:        policy.Type,
		})

	}
	return allProfiles, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileOverview(appId int, envIds []int) (*DeploymentWindowResponse, error) {

	resources, profileIdToProfile, err := impl.getProfileMappingsForApp(appId, envIds)
	if err != nil {
		return nil, err
	}

	envIdToMappings := make(map[int][]ProfileMapping)
	for _, resource := range resources {
		envIdToMappings[resource.EnvId] = append(envIdToMappings[resource.EnvId], resource)
	}

	profileStates := impl.getProfileStates(envIdToMappings, profileIdToProfile)

	return &DeploymentWindowResponse{
		Profiles: profileStates,
	}, nil
}

func (impl DeploymentWindowServiceImpl) getProfileMappingsForApp(appId int, envIds []int) ([]ProfileMapping, map[int]*DeploymentWindowProfile, error) {

	selections := make([]*resourceQualifiers.SelectionIdentifier, 0)
	for _, envId := range envIds {
		selections = append(selections, &resourceQualifiers.SelectionIdentifier{
			AppId: appId,
			EnvId: envId,
		})
	}

	resources, profileIdToProfile, err := impl.getResourcesAndProfilesForSelections(selections)
	if err != nil {
		return nil, nil, err
	}
	return resources, profileIdToProfile, nil
}

func (impl DeploymentWindowServiceImpl) getProfileStates(envIdToMappings map[int][]ProfileMapping, profileIdToProfile map[int]*DeploymentWindowProfile) []ProfileState {
	profileStates := make([]ProfileState, 0)
	for envId, mappings := range envIdToMappings {
		for _, mapping := range mappings {
			profile := profileIdToProfile[mapping.ProfileId]
			if !profile.Enabled {
				continue
			}
			profileStates = append(profileStates, ProfileState{
				DeploymentWindowProfile: profile,
				EnvId:                   envId,
			})
		}
	}
	return profileStates
}

func (impl DeploymentWindowServiceImpl) getProfileIdToProfile(profileIds []int) (map[int]*DeploymentWindowProfile, error) {

	models, err := impl.globalPolicyManager.GetPolicyByIds(profileIds)
	if err != nil {
		return nil, err
	}
	//	}
	profileIdToModel := make(map[int]*bean2.GlobalPolicyBaseModel)
	for _, model := range models {
		profileIdToModel[model.Id] = model
	}
	//	}

	profileIds = maps.Keys(profileIdToModel)

	profileIdToWindows, err := impl.timeWindowService.GetWindowsForResources(profileIds, repository.DeploymentWindowProfile)
	if err != nil {
		return nil, err
	}

	profileIdToProfile := make(map[int]*DeploymentWindowProfile)
	for _, profileId := range profileIds {

		windows := profileIdToWindows[profileId]

		profilePolicy, err := impl.getPolicyFromModel(profileIdToModel[profileId])
		if err != nil {
			return nil, err
		}
		deploymentProfile := profilePolicy.toDeploymentWindowProfile(profileIdToModel[profileId], windows)
		profileIdToProfile[profileId] = deploymentProfile
	}
	return profileIdToProfile, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileOverviewBulk(appEnvSelectors []AppEnvSelector) (map[int]*DeploymentWindowResponse, error) {

	profileIdToProfile, appIdToMappings, err := impl.getMappedResourcesForAppgroups(appEnvSelectors)
	if err != nil {
		return nil, err
	}

	appIdToResponse := make(map[int]*DeploymentWindowResponse)
	for appId, mappings := range appIdToMappings {
		//envIdToMappings := lo.GroupBy(mappings, func(item ProfileMapping) int {
		//	return item.EnvId
		//})
		envIdToMappings := make(map[int][]ProfileMapping)
		for _, resource := range mappings {
			envIdToMappings[resource.EnvId] = append(envIdToMappings[resource.EnvId], resource)
		}

		profileStates := impl.getProfileStates(envIdToMappings, profileIdToProfile)
		appIdToResponse[appId] = &DeploymentWindowResponse{
			Profiles: profileStates,
		}

	}
	return appIdToResponse, nil
}

func (impl DeploymentWindowServiceImpl) getMappedResourcesForAppgroups(appEnvSelectors []AppEnvSelector) (map[int]*DeploymentWindowProfile, map[int][]ProfileMapping, error) {
	selections := make([]*resourceQualifiers.SelectionIdentifier, 0)
	for _, selector := range appEnvSelectors {
		selections = append(selections, &resourceQualifiers.SelectionIdentifier{
			AppId: selector.AppId,
			EnvId: selector.EnvId,
		})
	}

	mappings, profileIdToProfile, err := impl.getResourcesAndProfilesForSelections(selections)
	if err != nil {
		return nil, nil, err
	}

	appIdToMappings := make(map[int][]ProfileMapping)
	for _, mapping := range mappings {
		appIdToMappings[mapping.AppId] = append(appIdToMappings[mapping.AppId], mapping)
	}

	return profileIdToProfile, appIdToMappings, nil
}

func (impl DeploymentWindowServiceImpl) getResourcesAndProfilesForSelections(selections []*resourceQualifiers.SelectionIdentifier) ([]ProfileMapping, map[int]*DeploymentWindowProfile, error) {
	resources, err := impl.resourceMappingService.GetResourceMappingsForSelections(resourceQualifiers.DeploymentWindow, resourceQualifiers.ApplicationEnvironmentSelector, selections)
	if err != nil {
		return nil, nil, err
	}

	profileIds := make([]int, 0)
	for _, resource := range resources {
		profileIds = append(profileIds, resource.ResourceId)
	}

	profileIdToProfile, err := impl.getProfileIdToProfile(profileIds)
	if err != nil {
		return nil, nil, err
	}

	mappings := make([]ProfileMapping, 0)
	for _, resource := range resources {
		if _, ok := profileIdToProfile[resource.ResourceId]; ok {
			mappings = append(mappings, ProfileMapping{
				ProfileId: resource.ResourceId,
				AppId:     resource.SelectionIdentifier.AppId,
				EnvId:     resource.SelectionIdentifier.EnvId,
			})
		}
	}

	return mappings, profileIdToProfile, nil
}
