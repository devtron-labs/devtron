package deploymentWindow

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	"strconv"
	"time"
)

type DeploymentWindowType string

const (
	Blackout    DeploymentWindowType = "BLACKOUT"
	Maintenance DeploymentWindowType = "MAINTENANCE"
)

type UserActionState string

const (
	Allowed UserActionState = "ALLOWED"
	Blocked UserActionState = "BLOCKED"
	Partial UserActionState = "PARTIAL"
)

func (action UserActionState) IsActionAllowed() bool {
	return action == Allowed || action == Partial
}

type AppEnvSelector struct {
	AppId int `json:"appId"`
	EnvId int `json:"envId"`
}

type DeploymentWindowAppGroupResponse struct {
	AppData []AppData `json:"appData,omitempty"`
}

type AppData struct {
	AppId                 int                       `json:"appId"`
	DeploymentProfileList *DeploymentWindowResponse `json:"deploymentProfileList,omitempty"`
	//DeploymentWindowResponse
}

// DeploymentWindowProfile defines model for DeploymentWindowProfile.
type DeploymentWindowProfile struct {
	DeploymentWindowList []*timeoutWindow.TimeWindow `json:"deploymentWindowList,omitempty"`
	Enabled              bool                        `json:"enabled"`
	TimeZone             string                      `json:"timeZone"`
	DisplayMessage       string                      `json:"displayMessage"`
	ExcludedUsersList    []int32                     `json:"excludedUsersList"`
	IsSuperAdminExcluded bool                        `json:"isSuperAdminExcluded"`
	IsUserExcluded       bool                        `json:"isUserExcluded"`
	DeploymentWindowProfileMetadata
}

func (profile *DeploymentWindowProfile) GetSerializedAuditData() string {
	if profile == nil {
		return ""
	}
	data := make(map[string]interface{})
	data["name"] = profile.Name
	data["id"] = profile.Id
	data["type"] = profile.Type

	dataJson, _ := json.Marshal(data)
	return string(dataJson)
}

// DeploymentWindowProfileMetadata defines model for DeploymentWindowProfileMetadata.
type DeploymentWindowProfileMetadata struct {
	Description string               `json:"description"`
	Id          int                  `json:"id"`
	Name        string               `json:"name"`
	Type        DeploymentWindowType `json:"type"`
}

// DeploymentWindowProfileRequest defines model for DeploymentWindowProfileRequest.
type DeploymentWindowProfileRequest struct {
	DeploymentWindowProfile *DeploymentWindowProfile `json:"deploymentWindowProfile,omitempty"`
}

type EnvironmentState struct {
	// ExcludedUsers final calculated list of user ids including superadmins who are excluded.
	ExcludedUsers      []int32  `json:"excludedUsers"`
	ExcludedUserEmails []string `json:"excludedUserEmails"`

	//// Timestamp indicating the window end or next window start timestamp based on current time and
	//Timestamp time.Time `json:"timestamp"`
	AppliedProfile *ProfileState `json:"appliedProfile"`

	// UserActionState describes the  eventual action state for the user
	UserActionState UserActionState `json:"userActionState"`
}

type ProfileState struct {
	CalculatedTimestamp     time.Time                `json:"calculatedTimestamp"`
	DeploymentWindowProfile *DeploymentWindowProfile `json:"deploymentWindowProfile,omitempty"`
	EnvId                   int                      `json:"envId"`
	IsActive                bool                     `json:"isActive"`
	ExcludedUserEmails      []string                 `json:"excludedUserEmails"`
}

// DeploymentWindowResponse defines model for DeploymentWindowResponse.
type DeploymentWindowResponse struct {
	EnvironmentStateMap map[int]EnvironmentState `json:"environmentStateMap,omitempty"`
	Profiles            []ProfileState           `json:"profiles,omitempty"`
	//SuperAdmins         []string                 `json:"superAdmins,omitempty"`
}

func (state UserActionState) GetBypassActionMessageForProfileAndState(profile *DeploymentWindowProfile) string {
	if state == Allowed {
		return ""
	}
	if profile != nil && profile.Type == Blackout {
		return "Initiated during blackout window " + strconv.Quote(profile.Name)
	}
	return "Initiated outside maintenance window"
}
