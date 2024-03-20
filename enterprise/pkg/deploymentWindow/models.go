package deploymentWindow

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"github.com/devtron-labs/devtron/util"
)

type DeploymentWindowProfilePolicy struct {
	TimeZone             string               `json:"timeZone"`
	DisplayMessage       string               `json:"displayMessage"`
	ExcludedUsersList    []int32              `json:"excludedUsersList"`
	IsSuperAdminExcluded bool                 `json:"isSuperAdminExcluded"`
	IsUserExcluded       bool                 `json:"isUserExcluded"`
	Type                 DeploymentWindowType `json:"type"`
	IsExpired            bool                 `json:"isExpired"`
}

func (profile DeploymentWindowProfile) toPolicy() DeploymentWindowProfilePolicy {
	return DeploymentWindowProfilePolicy{
		TimeZone:             profile.TimeZone,
		DisplayMessage:       profile.DisplayMessage,
		ExcludedUsersList:    profile.ExcludedUsersList,
		IsSuperAdminExcluded: profile.IsSuperAdminExcluded,
		IsUserExcluded:       profile.IsUserExcluded,
		Type:                 profile.Type,
		IsExpired:            profile.isExpired,
	}
}

func (profile DeploymentWindowProfile) convertToPolicyDataModel(userId int32) *bean.GlobalPolicyDataModel {

	policyBytes, _ := json.Marshal(profile.toPolicy())
	return &bean.GlobalPolicyDataModel{
		GlobalPolicyBaseModel: bean.GlobalPolicyBaseModel{
			Id:            profile.Id,
			Name:          profile.Name,
			Description:   profile.Description,
			Enabled:       profile.Enabled,
			PolicyOf:      bean.GLOBAL_POLICY_TYPE_DEPLOYMENT_WINDOW,
			PolicyVersion: bean.GLOBAL_POLICY_VERSION_V1,
			JsonData:      string(policyBytes),
			Active:        true,
			UserId:        userId,
		},
		SearchableFields: []util.SearchableField{},
	}
}

func (profilePolicy DeploymentWindowProfilePolicy) toDeploymentWindowProfile(policyModel *bean.GlobalPolicyBaseModel, windows []*timeoutWindow.TimeWindow) *DeploymentWindowProfile {
	return &DeploymentWindowProfile{
		DeploymentWindowList: windows,
		Enabled:              policyModel.Enabled,
		TimeZone:             profilePolicy.TimeZone,
		DisplayMessage:       profilePolicy.DisplayMessage,
		ExcludedUsersList:    profilePolicy.ExcludedUsersList,
		IsSuperAdminExcluded: profilePolicy.IsSuperAdminExcluded,
		IsUserExcluded:       profilePolicy.IsUserExcluded,
		DeploymentWindowProfileMetadata: DeploymentWindowProfileMetadata{
			Description: policyModel.Description,
			Id:          policyModel.Id,
			Name:        policyModel.Name,
			Type:        profilePolicy.Type,
			isExpired:   profilePolicy.IsExpired,
		},
	}
}

type ProfileMapping struct {
	ProfileId int
	AppId     int
	EnvId     int
}
