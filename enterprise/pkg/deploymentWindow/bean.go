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
	return action == Allowed
}

func (action UserActionState) IsActionAllowedWithBypass() bool {
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

func (state *EnvironmentState) GetSerializedAuditData(triggerMessage string) string {
	if state == nil {
		return "{}"
	}
	profile := state.AppliedProfile

	audit := DeploymentWindowAuditData{
		Audit:          state,
		TriggeredAt:    state.CalculatedAt,
		TriggerMessage: triggerMessage,
	}
	if profile != nil {
		audit.Name = profile.DeploymentWindowProfile.Name
		audit.Id = profile.DeploymentWindowProfile.Id
		audit.Type = string(profile.DeploymentWindowProfile.Type)
	}
	dataJson, _ := json.Marshal(audit)
	return string(dataJson)
}

func GetAuditDataFromSerializedValue(data string) DeploymentWindowAuditData {
	audit := DeploymentWindowAuditData{}
	json.Unmarshal([]byte(data), &audit)
	return audit
}

type DeploymentWindowAuditData struct {
	Name           string            `json:"name"`
	Id             int               `json:"id"`
	Type           string            `json:"type"`
	Audit          *EnvironmentState `json:"audit"`
	TriggeredAt    time.Time         `json:"triggeredAt"`
	TriggerMessage string            `json:"triggerMessage"`
}

// DeploymentWindowProfileMetadata defines model for DeploymentWindowProfileMetadata.
type DeploymentWindowProfileMetadata struct {
	Description string               `json:"description"`
	Id          int                  `json:"id"`
	Name        string               `json:"name"`
	Type        DeploymentWindowType `json:"type"`
	isExpired   bool
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
	CalculatedAt    time.Time       `json:"calculatedAt"`
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
}

func (state UserActionState) GetBypassActionMessageForProfileAndState(envState *EnvironmentState) string {
	if state == Allowed {
		return ""
	}
	var profile *DeploymentWindowProfile
	if envState != nil && envState.AppliedProfile != nil {
		profile = envState.AppliedProfile.DeploymentWindowProfile
	}

	if profile != nil && profile.Type == Blackout {
		return "Initiated during blackout window " + strconv.Quote(profile.Name)
	} else if profile != nil && profile.Type == Maintenance {
		return "Initiated outside maintenance window"
	}
	return ""
}

func (item ProfileState) isRestricted() bool {
	return (item.DeploymentWindowProfile.Type == Blackout && item.IsActive) || (item.DeploymentWindowProfile.Type == Maintenance && !item.IsActive)
}
