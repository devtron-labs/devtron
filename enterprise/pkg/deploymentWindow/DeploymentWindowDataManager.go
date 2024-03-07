package deploymentWindow

import (
	"encoding/json"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/samber/lo"
)

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

	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, userId, err, tx, policy.Id)
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
	err = impl.timeWindowService.UpdateWindowMappings(profile.DeploymentWindowList, userId, err, tx, policy.Id)
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

	err = impl.globalPolicyManager.DeletePolicyById(tx, profileId, userId)
	if err != nil {
		return err
	}
	err = impl.timeWindowService.UpdateWindowMappings([]*timeoutWindow.TimeWindow{}, userId, err, tx, profileId)
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

	return lo.Map(policyModels, func(model *bean2.GlobalPolicyBaseModel, index int) *DeploymentWindowProfileMetadata {
		policy, err := impl.getPolicyFromModel(model)
		if err != nil {
			return nil
		}
		return &DeploymentWindowProfileMetadata{
			Description: model.Description,
			Id:          model.Id,
			Name:        model.Name,
			Type:        policy.Type,
		}
	}), nil
}

func (impl DeploymentWindowServiceImpl) GetDeploymentWindowProfileOverview(appId int, envIds []int) (*DeploymentWindowResponse, error) {

	selections := lo.Map(envIds, func(envId int, index int) *resourceQualifiers.SelectionIdentifier {
		return &resourceQualifiers.SelectionIdentifier{
			AppId: appId,
			EnvId: envId,
		}
	})

	resources, profileIdToProfile, err := impl.getResourcesAndProfilesForSelections(selections)
	if err != nil {
		return nil, err
	}

	envIdToQualifierMappings := lo.GroupBy(resources, func(item resourceQualifiers.ResourceQualifierMappings) int {
		return item.SelectionIdentifier.EnvId
	})
	profileStates := impl.getProfileStates(envIdToQualifierMappings, profileIdToProfile)

	return &DeploymentWindowResponse{
		Profiles: profileStates,
	}, nil
}

func (impl DeploymentWindowServiceImpl) getProfileStates(envIdToQualifierMappings map[int][]resourceQualifiers.ResourceQualifierMappings, profileIdToProfile map[int]*DeploymentWindowProfile) []ProfileState {
	profileStates := make([]ProfileState, 0)
	for envId, qualifierMappings := range envIdToQualifierMappings {
		for _, qualifierMapping := range qualifierMappings {
			profile := profileIdToProfile[qualifierMapping.ResourceId]
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
	profileIdToModel := make(map[int]*bean2.GlobalPolicyBaseModel)
	for _, model := range models {
		profileIdToModel[model.Id] = model
	}

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

	selections := lo.Map(appEnvSelectors, func(appEnv AppEnvSelector, index int) *resourceQualifiers.SelectionIdentifier {
		return &resourceQualifiers.SelectionIdentifier{
			AppId: appEnv.AppId,
			EnvId: appEnv.EnvId,
		}
	})
	resources, profileIdToProfile, err := impl.getResourcesAndProfilesForSelections(selections)
	if err != nil {
		return nil, err
	}
	appIdToQualifierMappings := lo.GroupBy(resources, func(item resourceQualifiers.ResourceQualifierMappings) int {
		return item.SelectionIdentifier.AppId
	})

	appIdToResponse := make(map[int]*DeploymentWindowResponse)
	for appId, mappings := range appIdToQualifierMappings {
		envIdToQualifierMappings := lo.GroupBy(mappings, func(item resourceQualifiers.ResourceQualifierMappings) int {
			return item.SelectionIdentifier.EnvId
		})
		profileStates := impl.getProfileStates(envIdToQualifierMappings, profileIdToProfile)
		appIdToResponse[appId] = &DeploymentWindowResponse{
			Profiles: profileStates,
		}

	}
	return appIdToResponse, nil
}

func (impl DeploymentWindowServiceImpl) getResourcesAndProfilesForSelections(selections []*resourceQualifiers.SelectionIdentifier) ([]resourceQualifiers.ResourceQualifierMappings, map[int]*DeploymentWindowProfile, error) {
	resources, err := impl.resourceMappingService.GetResourceMappingsForSelections(resourceQualifiers.DeploymentWindowProfile, resourceQualifiers.ApplicationEnvironmentSelector, selections)
	if err != nil {
		return nil, nil, err
	}

	profileIds := lo.Map(resources, func(mapping resourceQualifiers.ResourceQualifierMappings, index int) int {
		return mapping.ResourceId
	})
	profileIdToProfile, err := impl.getProfileIdToProfile(profileIds)
	if err != nil {
		return nil, nil, err
	}
	return resources, profileIdToProfile, nil
}
