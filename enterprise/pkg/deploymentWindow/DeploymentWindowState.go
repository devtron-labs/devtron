package deploymentWindow

import (
	"github.com/samber/lo"
	"time"
)

func (impl DeploymentWindowServiceImpl) CheckTriggerAllowedState(target time.Time, appId int, envId int, userId int32) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileState(target time.Time, appId int, envIds []int, userId int32) (*DeploymentWindowResponse, error) {
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
	for envId, profileStates := range envIdToProfileStates {

		blackoutProfiles := lo.Filter(profileStates, func(item ProfileState, index int) bool {
			return item.DeploymentWindowProfile.Type == Blackout
		})

		//maintenanceProfiles := lo.Filter(profileStates, func(item ProfileState, index int) bool {
		//	return item.DeploymentWindowProfile.Type == Maintenance
		//})

		for _, profile := range blackoutProfiles {
			loc, err := impl.getTimeZoneData(profile.DeploymentWindowProfile.TimeZone)
			if err != nil {
				return nil, err
			}
			timeWithZone := target.In(loc)
			isActive := false
			maxEndTimeStamp := time.Time{}
			minStartTimeStamp := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
			for _, window := range profile.DeploymentWindowProfile.DeploymentWindowList {
				if timestamp, isInside := window.toTimeRange().GetScheduleSpec(timeWithZone); isInside && !timestamp.IsZero() {
					isActive = true
					if timestamp.After(maxEndTimeStamp) {
						maxEndTimeStamp = timestamp
					}
				} else if !isInside && !timestamp.IsZero() {
					if timestamp.Before(minStartTimeStamp) {
						minStartTimeStamp = timestamp
					}
				}
			}
			if isActive {
				profile.CalculatedTimestamp = maxEndTimeStamp
			} else {
				profile.CalculatedTimestamp = minStartTimeStamp
			}
			profile.IsActive = isActive
		}

		envState := EnvironmentState{
			ExcludedUsers:   nil,
			Timestamp:       time.Time{},
			UserActionState: "",
		}
		envIdToEnvironmentState[envId] = envState
	}

}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileStateAppGroup(selectors []AppEnvSelector) (*DeploymentWindowAppGroupResponse, error) {
	//TODO implement me
	panic("implement me")
}
