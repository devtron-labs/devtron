package deploymentWindow

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"time"
)

func (impl DeploymentWindowServiceImpl) CheckTriggerAllowedState(targetTime time.Time, appId int, envId int, userId int32) (bool, error) {
	stateResponse, err := impl.GetDeploymentWindowProfileState(targetTime, appId, []int{envId}, userId)
	if err != nil {
		return false, err
	}
	isAllowed := true
	if state, ok := stateResponse.EnvironmentStateMap[envId]; ok {
		isAllowed = state.UserActionState == Allowed || state.UserActionState == Partial
	}
	return len(stateResponse.Profiles) == 0 || isAllowed, nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileState(targetTime time.Time, appId int, envIds []int, userId int32) (*DeploymentWindowResponse, error) {
	overview, err := impl.GetDeploymentWindowProfileOverview(appId, envIds)
	if err != nil {
		return nil, err
	}

	superAdmins, err := impl.userService.GetSuperAdmins()
	if err != nil {
		return nil, err
	}
	overview.SuperAdmins = superAdmins

	envIdToProfileStates := lo.GroupBy(overview.Profiles, func(item ProfileState) int {
		return item.EnvId
	})

	envIdToEnvironmentState := make(map[int]EnvironmentState)
	resultProfiles := make([]ProfileState, 0)
	for envId, profileStates := range envIdToProfileStates {
		filteredProfileStates, appliedProfile, excludedUsers, canDeploy, err := impl.getAppliedProfileAndCalculateStates(targetTime, profileStates, superAdmins)
		if err != nil {
			return nil, err
		}
		envState := EnvironmentState{
			ExcludedUsers:   excludedUsers,
			AppliedProfile:  appliedProfile,
			UserActionState: getUserActionStateForUser(canDeploy, excludedUsers, userId),
		}
		envIdToEnvironmentState[envId] = envState
		resultProfiles = append(resultProfiles, filteredProfileStates...)
	}
	response := &DeploymentWindowResponse{
		EnvironmentStateMap: envIdToEnvironmentState,
		Profiles:            resultProfiles,
		SuperAdmins:         superAdmins,
	}
	return response, nil
}

func getUserActionStateForUser(canDeploy bool, excludedUsers []int32, userId int32) UserActionState {
	userActionState := Allowed
	if !canDeploy {
		if slices.Contains(excludedUsers, userId) {
			userActionState = Partial
		} else {
			userActionState = Blocked
		}
	}
	return userActionState
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileStateAppGroup(selectors []AppEnvSelector) (*DeploymentWindowAppGroupResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (impl DeploymentWindowServiceImpl) getAppliedProfileAndCalculateStates(targetTime time.Time, profileStates []ProfileState, superAdmins []int32) ([]ProfileState, *DeploymentWindowProfile, []int32, bool, error) {

	var appliedProfile *DeploymentWindowProfile
	var combinedExcludedUsers []int32

	filteredBlackoutProfiles, _, isBlackoutActive, err := impl.calculateStateForProfiles(targetTime, profileStates, Blackout)
	if err != nil {
		return nil, appliedProfile, combinedExcludedUsers, false, err
	}

	filteredMaintenanceProfiles, isMaintenanceActive, _, err := impl.calculateStateForProfiles(targetTime, profileStates, Maintenance)
	if err != nil {
		return nil, appliedProfile, combinedExcludedUsers, false, err
	}

	if len(filteredBlackoutProfiles) == 0 && len(filteredMaintenanceProfiles) == 0 {
		return nil, appliedProfile, combinedExcludedUsers, true, nil
	}

	canDeploy := !isBlackoutActive && isMaintenanceActive
	allProfiles := append(filteredBlackoutProfiles, filteredMaintenanceProfiles...)
	if isBlackoutActive && isMaintenanceActive { //deployment is blocked restriction through blackout
		// if both are active then blackout takes precedence in overall calculation
		appliedProfile = impl.getLongestEndingProfile(filteredBlackoutProfiles)
		combinedExcludedUsers = impl.getCombinedUserIds(filteredBlackoutProfiles, superAdmins)

	} else if !isBlackoutActive && !isMaintenanceActive { //deployment is blocked restriction through maintenance
		// if nothing is active then earliest starting maintenance will be shown
		appliedProfile = impl.getEarliestStartingProfile(filteredMaintenanceProfiles)
		combinedExcludedUsers = impl.getCombinedUserIds(filteredMaintenanceProfiles, superAdmins)
	} else if isBlackoutActive && !isMaintenanceActive { //deployment is blocked restriction through both
		// longest of restrictions coming from both blackout and maintenance
		appliedProfile = impl.getLongestEndingProfile(allProfiles)
		combinedExcludedUsers = impl.getCombinedUserIds(allProfiles, superAdmins)

	} else if !isBlackoutActive && isMaintenanceActive { //deployment not blocked
		// applied profile here would be the longest running maintenance profile even if a blackout starts before that
		appliedProfile = impl.getLongestEndingProfile(filteredMaintenanceProfiles)
	}
	return allProfiles, appliedProfile, combinedExcludedUsers, canDeploy, nil
}

func (impl DeploymentWindowServiceImpl) getCombinedUserIds(profiles []ProfileState, superAdmins []int32) []int32 {

	if len(profiles) == 0 {
		return []int32{}
	}
	userSet := mapset.NewSet()

	userSet.Add(profiles[0].DeploymentWindowProfile.ExcludedUsersList)

	isSuperAdminExcluded := true
	lo.ForEach(profiles, func(profile ProfileState, index int) {
		var users []int32
		if profile.DeploymentWindowProfile.IsUserExcluded {
			users = profile.DeploymentWindowProfile.ExcludedUsersList
		}

		if !profile.DeploymentWindowProfile.IsSuperAdminExcluded {
			isSuperAdminExcluded = false
		}

		profileUserSet := mapset.NewSet(users)
		userSet = userSet.Intersect(profileUserSet)
	})

	if isSuperAdminExcluded {
		userSet = userSet.Union(mapset.NewSet(superAdmins))
	}

	return utils.ToInt32Array(userSet.ToSlice())
}

func (impl DeploymentWindowServiceImpl) getLongestEndingProfile(profiles []ProfileState) *DeploymentWindowProfile {

	if len(profiles) == 0 {
		return nil
	}

	return lo.Reduce(profiles, func(profile ProfileState, item ProfileState, index int) ProfileState {
		if item.CalculatedTimestamp.After(profile.CalculatedTimestamp) {
			return item
		}
		return profile
	}, profiles[0]).DeploymentWindowProfile
}

func (impl DeploymentWindowServiceImpl) getEarliestStartingProfile(profiles []ProfileState) *DeploymentWindowProfile {
	if len(profiles) == 0 {
		return nil
	}

	return lo.Reduce(profiles, func(profile ProfileState, item ProfileState, index int) ProfileState {
		if item.CalculatedTimestamp.Before(profile.CalculatedTimestamp) {
			return item
		}
		return profile
	}, profiles[0]).DeploymentWindowProfile
}

func (impl DeploymentWindowServiceImpl) calculateStateForProfiles(targetTime time.Time, profileStates []ProfileState, profileType DeploymentWindowType) ([]ProfileState, bool, bool, error) {

	filteredProfiles := lo.Filter(profileStates, func(item ProfileState, index int) bool {
		return item.DeploymentWindowProfile.Type == profileType
	})

	allActive := true
	oneActive := false
	finalProfileStates := make([]ProfileState, 0)

	for _, profile := range filteredProfiles {
		loc, err := impl.getTimeZoneData(profile.DeploymentWindowProfile.TimeZone)
		if err != nil {
			return nil, false, false, err
		}
		timeWithZone := targetTime.In(loc)
		isActive, windowTimeStamp, window := impl.getActiveWindow(timeWithZone, profile.DeploymentWindowProfile.DeploymentWindowList)

		if window == nil {
			// doing nothing if no window is returned
			// this means that no relevant window in the profile was found therefore skipping this profile
			continue
		}
		profile.IsActive = isActive
		profile.CalculatedTimestamp = windowTimeStamp
		profile.DeploymentWindowProfile.DeploymentWindowList = []*TimeWindow{window}

		if !oneActive && isActive {
			oneActive = true
		}
		if allActive && !isActive {
			allActive = false
		}
		finalProfileStates = append(finalProfileStates, profile)
	}
	return finalProfileStates, allActive, oneActive, nil
}

func (impl DeploymentWindowServiceImpl) getActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow) {
	isActive := false
	maxEndTimeStamp := time.Time{}
	minStartTimeStamp := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	var appliedWindow *TimeWindow
	for _, window := range windows {
		if timestamp, isInside := window.toTimeRange().GetScheduleSpec(targetTimeWithZone); isInside && !timestamp.IsZero() {
			isActive = true
			if timestamp.After(maxEndTimeStamp) {
				maxEndTimeStamp = timestamp
				appliedWindow = window
			}
		} else if !isInside && !timestamp.IsZero() {
			if timestamp.Before(minStartTimeStamp) {
				minStartTimeStamp = timestamp
				appliedWindow = window
			}
		}
	}
	if isActive {
		return true, maxEndTimeStamp, appliedWindow
	}
	return false, minStartTimeStamp, appliedWindow
}
