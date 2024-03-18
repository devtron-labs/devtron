package deploymentWindow

import (
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/go-pg/pg"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
	"time"
)

func (impl DeploymentWindowServiceImpl) GetStateForAppEnv(targetTime time.Time, appId int, envId int, userId int32) (UserActionState, *EnvironmentState, error) {

	stateResponse, err := impl.GetDeploymentWindowProfileState(targetTime, appId, []int{envId}, userId)
	if err != nil {
		impl.logger.Errorw("error fetching deployment window profile state", "err", err, "appId", appId, "envId", envId, "user", userId)
		return Allowed, nil, err
	}

	var envState *EnvironmentState
	actionState := Allowed
	if state, ok := stateResponse.EnvironmentStateMap[envId]; ok {
		actionState = state.UserActionState
		envState = &state
	}
	return actionState, envState, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileStateAppGroup(targetTime time.Time, selectors []AppEnvSelector, userId int32) (*DeploymentWindowAppGroupResponse, error) {

	appIdsToOverview, err := impl.GetDeploymentWindowProfileOverviewBulk(selectors)
	if err != nil {
		impl.logger.Errorw("error fetching deployment window profile overview bulk", "err", err, "selectors", selectors)
		return nil, err
	}

	// fetching user and super-admin data once for all calculations
	superAdmins, userEmailMap, err := impl.getUserInfoMap(err, appIdsToOverview)
	if err != nil {
		impl.logger.Errorw("error fetching userInfoMap", "err", err)
		return nil, err
	}

	profiles := make([]ProfileWrapper, 0)
	for _, overview := range appIdsToOverview {
		profiles = append(profiles, overview.Profiles...)
	}
	profileIdToProfileState, err := impl.calculateStateForProfiles(targetTime, profiles, superAdmins, userEmailMap)
	if err != nil {
		impl.logger.Errorw("error in calculating profile state", "err", err)
		return nil, err
	}

	appIdToProfiles := make(map[int][]ProfileWrapper)
	for appId, overview := range appIdsToOverview {
		for _, profile := range overview.Profiles {
			if calculatedProfile, ok := profileIdToProfileState[profile.DeploymentWindowProfile.Id]; ok {
				appIdToProfiles[appId] = append(appIdToProfiles[appId], calculatedProfile)
			}
		}
	}

	appGroupData := make([]AppData, 0)
	for appId, overview := range appIdsToOverview {

		envIdToProfileStates, err := impl.evaluateStateForEnvironments(overview.Profiles, profileIdToProfileState, targetTime, superAdmins, userEmailMap, userId)
		if err != nil {
			impl.logger.Errorw("error in calculating state for environments", "err", err)
			return nil, err
		}
		envResponse := &DeploymentWindowResponse{
			EnvironmentStateMap: envIdToProfileStates,
			Profiles:            appIdToProfiles[appId],
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
		return nil, nil, fmt.Errorf("error fetching superadmins %v", err)
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
		return nil, nil, fmt.Errorf("error in getting user infor for emailIds %v", err)
	}
	userInfoMap := make(map[int32]string, 0)
	for _, user := range allUserInfo {
		if strings.Contains(user.EmailId, "@") {
			userInfoMap[user.Id] = user.EmailId
		}
	}

	return superAdmins, userInfoMap, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileState(targetTime time.Time, appId int, envIds []int, userId int32) (*DeploymentWindowResponse, error) {
	overview, err := impl.GetDeploymentWindowProfileOverview(appId, envIds)
	if err != nil {
		impl.logger.Errorw("error in getting deployment window profile overview", "err", err, "appId", appId, "envs", envIds)
		return nil, err
	}

	superAdmins, userEmailMap, err := impl.getUserInfoMap(err, map[int]*DeploymentWindowResponse{0: overview})
	if err != nil {
		impl.logger.Errorw("error in fetching user data", "err", err)
		return nil, err
	}

	profileIdToProfileState, err := impl.calculateStateForProfiles(targetTime, overview.Profiles, superAdmins, userEmailMap)
	if err != nil {
		impl.logger.Errorw("error in calculating profile state", "err", err)
		return nil, err
	}

	envIdToProfileStates, err := impl.evaluateStateForEnvironments(overview.Profiles, profileIdToProfileState, targetTime, superAdmins, userEmailMap, userId)
	if err != nil {
		impl.logger.Errorw("error in calculating state", "err", err)
		return nil, err
	}

	response := &DeploymentWindowResponse{
		EnvironmentStateMap: envIdToProfileStates,
		Profiles:            maps.Values(profileIdToProfileState),
	}

	return response, nil
}

func (impl DeploymentWindowServiceImpl) evaluateStateForEnvironments(profiles []ProfileWrapper, idToProfileState map[int]ProfileWrapper, targetTime time.Time, superAdmins []int32, userEmailMap map[int32]string, userId int32) (map[int]EnvironmentState, error) {

	envIdToProfileStates := make(map[int][]ProfileWrapper)
	for _, profile := range profiles {
		if calculatedProfile, ok := idToProfileState[profile.DeploymentWindowProfile.Id]; ok {
			envIdToProfileStates[profile.EnvId] = append(envIdToProfileStates[profile.EnvId], calculatedProfile)
		}
	}

	envIdToEnvironmentState := make(map[int]EnvironmentState)
	for envId, profileStates := range envIdToProfileStates {

		filteredProfileStates, appliedProfile, isAllowed, err := impl.evaluateProfileStates(profileStates)
		if err != nil {
			return nil, fmt.Errorf("error in evaluating profile state %v", err)
		}
		excludedUsers, excludedUsersEmail := impl.evaluateExcludedUsers(filteredProfileStates, superAdmins, userEmailMap)

		// sorting to keep active profiles first
		sort.SliceStable(filteredProfileStates, func(i, j int) bool {
			return filteredProfileStates[i].IsActive
		})

		envState := EnvironmentState{
			ExcludedUsers:      excludedUsers,
			ExcludedUserEmails: excludedUsersEmail,
			AppliedProfile:     appliedProfile,
			UserActionState:    getUserActionStateForUser(isAllowed, excludedUsers, userId),
			CalculatedAt:       targetTime,
		}
		envIdToEnvironmentState[envId] = envState
	}
	return envIdToEnvironmentState, nil
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

func (impl DeploymentWindowServiceImpl) evaluateProfileStates(profileStates []ProfileWrapper) ([]ProfileWrapper, *ProfileWrapper, bool, error) {
	var appliedProfile *ProfileWrapper
	filteredBlackoutProfiles, _, isBlackoutActive, err := impl.evaluateCombinedProfiles(profileStates, Blackout)
	if err != nil {
		return nil, appliedProfile, false, fmt.Errorf("error in calculating state for blackout windows %v", err)
	}

	filteredMaintenanceProfiles, isMaintenanceActive, _, err := impl.evaluateCombinedProfiles(profileStates, Maintenance)
	if err != nil {
		return nil, appliedProfile, false, fmt.Errorf("error in calculating state for maintenance windows %v", err)
	}

	if len(filteredBlackoutProfiles) == 0 && len(filteredMaintenanceProfiles) == 0 {
		return nil, appliedProfile, true, nil
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

	return allProfiles, appliedProfile, isAllowed, nil
}

func (impl DeploymentWindowServiceImpl) fillExcludedUserData(profile ProfileWrapper, superAdmins []int32, userEmailMap map[int32]string) ProfileWrapper {

	excludedIds := make([]int32, 0)
	if profile.DeploymentWindowProfile.IsUserExcluded && len(profile.DeploymentWindowProfile.ExcludedUsersList) > 0 {
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
	profile.ExcludedUserEmails = emails
	profile.DeploymentWindowProfile.ExcludedUsersList = excludedIds
	return profile
}

func (impl DeploymentWindowServiceImpl) evaluateExcludedUsers(profiles []ProfileWrapper, superAdmins []int32, userEmailMap map[int32]string) ([]int32, []string) {
	combinedExcludedUsers, isSuperAdminExcluded := impl.getCombinedUserIds(profiles)

	if isSuperAdminExcluded {
		combinedExcludedUsers = utils.FilterDuplicates(append(combinedExcludedUsers, superAdmins...))
	}

	emails := make([]string, 0)
	for _, userId := range combinedExcludedUsers {
		if email, ok := userEmailMap[userId]; ok {
			emails = append(emails, email)
		}
	}

	return combinedExcludedUsers, emails
}

func (impl DeploymentWindowServiceImpl) getCombinedUserIds(profiles []ProfileWrapper) ([]int32, bool) {

	if len(profiles) == 0 {
		return []int32{}, false
	}
	userSet := mapset.NewSet()

	profile := profiles[0]
	excludedUsers := profile.DeploymentWindowProfile.ExcludedUsersList
	if profile.isRestricted() && profile.DeploymentWindowProfile.IsUserExcluded && len(excludedUsers) > 0 {
		userSet = mapset.NewSetFromSlice(utils.ToInterfaceArrayAny(excludedUsers))
	}

	isSuperAdminExcluded := profiles[0].DeploymentWindowProfile.IsSuperAdminExcluded
	for _, profile := range profiles {

		if !profile.isRestricted() {
			continue
		}
		var users []int32
		if profile.DeploymentWindowProfile.IsUserExcluded {
			users = profile.DeploymentWindowProfile.ExcludedUsersList
		}

		if !profile.DeploymentWindowProfile.IsSuperAdminExcluded {
			isSuperAdminExcluded = false
		}

		profileUserSet := mapset.NewSetFromSlice(utils.ToInterfaceArrayAny(users))
		userSet = userSet.Intersect(profileUserSet)
	}

	return utils.ToInt32Array(userSet.ToSlice()), isSuperAdminExcluded
}

func (impl DeploymentWindowServiceImpl) getLongestEndingProfile(profiles []ProfileWrapper, filterRestricted bool) *ProfileWrapper {

	if len(profiles) == 0 {
		return nil
	}

	var selectedProfile *ProfileWrapper
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

func (impl DeploymentWindowServiceImpl) getEarliestStartingProfile(profiles []ProfileWrapper, filterRestricted bool) *ProfileWrapper {
	if len(profiles) == 0 {
		return nil
	}

	var selectedProfile *ProfileWrapper
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

func (impl DeploymentWindowServiceImpl) calculateStateForProfiles(targetTime time.Time, profileStates []ProfileWrapper, superAdmins []int32, userEmailMap map[int32]string) (map[int]ProfileWrapper, error) {

	calculatedProfiles := make(map[int]ProfileWrapper)
	for _, profile := range profileStates {

		_, exists := calculatedProfiles[profile.DeploymentWindowProfile.Id]
		if profile.DeploymentWindowProfile.isExpired || exists {
			continue
		}

		zone := profile.DeploymentWindowProfile.TimeZone
		isActive, windowTimeStamp, window, err := impl.timeWindowService.GetActiveWindow(targetTime, zone, profile.DeploymentWindowProfile.DeploymentWindowList)
		if err != nil {
			impl.logger.Debugw("error in getting active window", "err", err, "targetTime", targetTime, "zone", zone, "profile", profile)
			return calculatedProfiles, fmt.Errorf("error in getting active window %v %v", err, profile.DeploymentWindowProfile.Id)
		}
		if window != nil {
			profile.IsActive = isActive
			profile.CalculatedTimestamp = windowTimeStamp
			profile.DeploymentWindowProfile.DeploymentWindowList = []*timeoutWindow.TimeWindow{window}

			//hiding profiles which start after the configured limit
			if impl.isAfterDayLimit(targetTime, profile) {
				continue
			}

			calculatedProfiles[profile.DeploymentWindowProfile.Id] = impl.fillExcludedUserData(profile, superAdmins, userEmailMap)
		} else {
			//this means all windows in this profile are expired
			//therefore we're updating the isExpired flag in the policy so that expired profiles are filtered
			//out for further evaluations, until an update operation happens on this profile
			profile.DeploymentWindowProfile.isExpired = true
			impl.updatePolicy(profile.DeploymentWindowProfile, 1, nil)
		}
	}
	return calculatedProfiles, nil
}

func (impl DeploymentWindowServiceImpl) evaluateCombinedProfiles(profileStates []ProfileWrapper, profileType DeploymentWindowType) ([]ProfileWrapper, bool, bool, error) {

	filteredProfiles := impl.filterForType(profileStates, profileType)
	allActive := true
	oneActive := false
	for _, profile := range filteredProfiles {
		isActive := profile.IsActive
		if !oneActive && isActive {
			oneActive = true
		}
		if allActive && !isActive {
			allActive = false
		}
	}
	return filteredProfiles, allActive, oneActive, nil
}

func (impl DeploymentWindowServiceImpl) filterForType(profiles []ProfileWrapper, profileType DeploymentWindowType) []ProfileWrapper {
	filteredProfiles := make([]ProfileWrapper, 0)
	for _, profileWrapper := range profiles {
		if profileWrapper.DeploymentWindowProfile.Type == profileType {
			filteredProfiles = append(filteredProfiles, profileWrapper)
		}
	}
	return filteredProfiles
}

func (impl DeploymentWindowServiceImpl) CreateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error) {
	tx, err := impl.tx.StartTx()
	if err != nil {
		return nil, err
	}
	defer func(tx *pg.Tx) {
		err := tx.Rollback()
		if err != nil {
			impl.logger.Errorw("error in rollback CreateDeploymentWindowProfile", "err", err)
		}
	}(tx)

	// create policy
	policy := profile.convertToPolicyDataModel(userId, false)

	policy, err = impl.globalPolicyManager.CreatePolicy(policy, tx)
	if err != nil {
		impl.logger.Errorw("error in CreatePolicy", "err", err)
		return nil, err
	}
	profile.Id = policy.Id

	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, profile.TimeZone, userId, err, tx, policy.Id)
	if err != nil {
		impl.logger.Errorw("error in UpdateWindowMappings", "err", err, "profileId", policy.Id)
		return nil, err
	}
	err = impl.tx.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx CreateDeploymentWindowProfile", "err", err)
		return nil, err
	}

	return profile, err
}

func (impl DeploymentWindowServiceImpl) UpdateDeploymentWindowProfile(profile *DeploymentWindowProfile, userId int32) (*DeploymentWindowProfile, error) {
	tx, err := impl.tx.StartTx()
	if err != nil {
		return nil, err
	}
	defer func(tx *pg.Tx) {
		err := tx.Rollback()
		if err != nil {
			impl.logger.Errorw("error in rollback UpdateDeploymentWindowProfile", "err", err)
		}
	}(tx)

	// update policy
	policy, err := impl.updatePolicy(profile, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in updatePolicy", "err", err)
		return nil, err
	}
	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, profile.TimeZone, userId, err, tx, policy.Id)
	if err != nil {
		impl.logger.Errorw("error in UpdateWindowMappings", "err", err, "profileId", policy.Id)
		return nil, err
	}
	err = impl.tx.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx UpdateDeploymentWindowProfile", "err", err)
		return nil, err
	}
	return profile, err
}

func (impl DeploymentWindowServiceImpl) updatePolicy(profile *DeploymentWindowProfile, userId int32, tx *pg.Tx) (*bean2.GlobalPolicyDataModel, error) {
	policy := profile.convertToPolicyDataModel(userId, false)

	policy, err := impl.globalPolicyManager.UpdatePolicy(policy, tx)
	if err != nil {
		return policy, fmt.Errorf("updatePolicy globalPolicyManager.UpdatePolicy %v", err)
	}
	return policy, nil
}

func (impl DeploymentWindowServiceImpl) DeleteDeploymentWindowProfileForName(profileName string, userId int32) error {
	policyModel, err := impl.globalPolicyManager.GetPolicyByName(profileName, bean2.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW)
	if err != nil {
		impl.logger.Errorw("error in getting policy model", "err", err, "profileName", profileName)
		return err
	}
	return impl.DeleteDeploymentWindowProfileForId(policyModel.Id, userId)
}

func (impl DeploymentWindowServiceImpl) DeleteDeploymentWindowProfileForId(profileId int, userId int32) error {
	tx, err := impl.tx.StartTx()
	if err != nil {
		return err
	}
	defer func(tx *pg.Tx) {
		err := tx.Rollback()
		if err != nil {
			impl.logger.Errorw("error in rollback DeleteDeploymentWindowProfileForId", "err", err)
		}
	}(tx)

	err = impl.globalPolicyManager.DeletePolicyById(tx, profileId, userId)
	if err != nil {
		impl.logger.Errorw("error in DeletePolicyById", "err", err, "profileId", profileId)
		return err
	}
	err = impl.timeWindowService.UpdateWindowMappings([]*timeoutWindow.TimeWindow{}, "", userId, err, tx, profileId)
	if err != nil {
		impl.logger.Errorw("error in UpdateWindowMappings", "err", err, "profileId", profileId)
		return err
	}
	err = impl.tx.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx DeleteDeploymentWindowProfileForId", "err", err)

		return err
	}

	return err
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileForName(profileName string) (*DeploymentWindowProfile, error) {
	//get policy
	policyModel, err := impl.globalPolicyManager.GetPolicyByName(profileName, bean2.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW)
	if err != nil {
		impl.logger.Errorw("error in getting policy model", "err", err, "profileName", profileName)
		return nil, err
	}
	return impl.getProfileWithWindows(policyModel)
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileForId(profileId int) (*DeploymentWindowProfile, error) {
	//get policy
	policyModel, err := impl.globalPolicyManager.GetPolicyById(profileId)
	if err != nil {
		impl.logger.Errorw("error in getting policy model", "err", err, "profileId", profileId)
		return nil, err
	}

	return impl.getProfileWithWindows(policyModel)
}

func (impl DeploymentWindowServiceImpl) getProfileWithWindows(policyModel *bean2.GlobalPolicyBaseModel) (*DeploymentWindowProfile, error) {
	idToWindows, err := impl.timeWindowService.GetWindowsForResources([]int{policyModel.Id}, repository.DeploymentWindowProfile)
	if err != nil {
		impl.logger.Errorw("error in getting GetWindowsForResources", "err", err, "profileId", policyModel.Id)
		return nil, err
	}

	windows, ok := idToWindows[policyModel.Id]
	if !ok {
		return nil, nil
	}
	profilePolicy := impl.getPolicyFromModel(policyModel)
	return profilePolicy.toDeploymentWindowProfile(policyModel, windows), nil
}

func (impl DeploymentWindowServiceImpl) getPolicyFromModel(policyModel *bean2.GlobalPolicyBaseModel) *DeploymentWindowProfilePolicy {
	profilePolicy := &DeploymentWindowProfilePolicy{}
	json.Unmarshal([]byte(policyModel.JsonData), &profilePolicy)
	return profilePolicy
}

func (impl DeploymentWindowServiceImpl) ListDeploymentWindowProfiles() ([]*DeploymentWindowProfileMetadata, error) {
	//get policy
	policyModels, err := impl.globalPolicyManager.GetAllActivePoliciesByType(bean2.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW)
	if err != nil {
		impl.logger.Errorw("error in GetAllActiveByType", "err", err)
		return nil, err
	}

	allProfiles := make([]*DeploymentWindowProfileMetadata, 0)
	for _, model := range policyModels {
		policy := impl.getPolicyFromModel(model)
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
		impl.logger.Errorw("error in getProfileMappingsForApp", "err", err, "appId", appId, "envIds", envIds)
		return nil, err
	}

	envIdToMappings := make(map[int][]ProfileMapping)
	for _, resource := range resources {
		envIdToMappings[resource.EnvId] = append(envIdToMappings[resource.EnvId], resource)
	}

	profileStates := impl.flattenProfiles(envIdToMappings, profileIdToProfile)

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
		return nil, nil, fmt.Errorf("getResourcesAndProfilesForSelections %v", err)
	}
	return resources, profileIdToProfile, nil
}

func (impl DeploymentWindowServiceImpl) flattenProfiles(envIdToMappings map[int][]ProfileMapping, profileIdToProfile map[int]*DeploymentWindowProfile) []ProfileWrapper {
	profiles := make([]ProfileWrapper, 0)
	for envId, mappings := range envIdToMappings {
		for _, mapping := range mappings {
			profile := profileIdToProfile[mapping.ProfileId]
			if !profile.Enabled {
				continue
			}
			profiles = append(profiles, ProfileWrapper{
				DeploymentWindowProfile: profile,
				EnvId:                   envId,
			})
		}
	}
	return profiles
}

func (impl DeploymentWindowServiceImpl) getProfileIdToProfile(profileIds []int) (map[int]*DeploymentWindowProfile, error) {

	models, err := impl.globalPolicyManager.GetPolicyByIds(profileIds)
	if err != nil {
		return nil, fmt.Errorf("getProfileIdToProfile GetPolicyByIds %v %v", err, profileIds)
	}
	profileIdToModel := make(map[int]*bean2.GlobalPolicyBaseModel)
	for _, model := range models {
		profileIdToModel[model.Id] = model
	}

	profileIds = maps.Keys(profileIdToModel)

	profileIdToWindows, err := impl.timeWindowService.GetWindowsForResources(profileIds, repository.DeploymentWindowProfile)
	if err != nil {
		return nil, fmt.Errorf("getProfileIdToProfile GetWindowsForResources %v %v", err, profileIds)
	}

	profileIdToProfile := make(map[int]*DeploymentWindowProfile)
	for _, profileId := range profileIds {

		windows := profileIdToWindows[profileId]

		profilePolicy := impl.getPolicyFromModel(profileIdToModel[profileId])
		deploymentProfile := profilePolicy.toDeploymentWindowProfile(profileIdToModel[profileId], windows)
		profileIdToProfile[profileId] = deploymentProfile
	}
	return profileIdToProfile, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileOverviewBulk(appEnvSelectors []AppEnvSelector) (map[int]*DeploymentWindowResponse, error) {

	profileIdToProfile, appIdToMappings, err := impl.getMappedResourcesForAppgroups(appEnvSelectors)
	if err != nil {
		impl.logger.Errorw("error in getMappedResourcesForAppgroups", "err", err)

		return nil, err
	}

	appIdToResponse := make(map[int]*DeploymentWindowResponse)
	for appId, mappings := range appIdToMappings {

		envIdToMappings := make(map[int][]ProfileMapping)
		for _, resource := range mappings {
			envIdToMappings[resource.EnvId] = append(envIdToMappings[resource.EnvId], resource)
		}

		profileStates := impl.flattenProfiles(envIdToMappings, profileIdToProfile)
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
		return nil, nil, fmt.Errorf("getResourcesAndProfilesForSelections %v", err)
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
		return nil, nil, fmt.Errorf("GetResourceMappingsForSelections %v", err)
	}

	profileIds := make([]int, 0)
	for _, resource := range resources {
		profileIds = append(profileIds, resource.ResourceId)
	}

	profileIdToProfile, err := impl.getProfileIdToProfile(profileIds)
	if err != nil {
		return nil, nil, fmt.Errorf("getProfileIdToProfile %v %v", err, profileIds)
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

func (impl DeploymentWindowServiceImpl) isAfterDayLimit(targetTime time.Time, profile ProfileWrapper) bool {
	if profile.IsActive {
		return false
	}

	if profile.DeploymentWindowProfile.Type == Blackout && profile.CalculatedTimestamp.Sub(targetTime) > time.Duration(impl.cfg.DeploymentWindowFetchDaysBlackout)*time.Hour*24 {
		return true
	}
	if profile.DeploymentWindowProfile.Type == Maintenance && profile.CalculatedTimestamp.Sub(targetTime) > time.Duration(impl.cfg.DeploymentWindowFetchDaysMaintenance)*time.Hour*24 {
		return true
	}
	return false
}
