package deploymentWindow

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
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

func (action UserActionState) IsActionBypass() bool {
	return action == Partial
}

func GetNotFoundError(err error) error {
	msg := "not found"
	return &util.ApiError{
		HttpStatusCode:    404,
		Code:              "404",
		InternalMessage:   err.Error(),
		UserMessage:       msg,
		UserDetailMessage: msg,
	}
}

func GetActionBlockedError(triggerMessage string, internalCode string) error {
	return &util.ApiError{
		HttpStatusCode:    422,
		Code:              internalCode,
		InternalMessage:   triggerMessage,
		UserMessage:       triggerMessage,
		UserDetailMessage: "action blocked",
	}
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
}

// DeploymentWindowProfile defines model for DeploymentWindowProfile.
type DeploymentWindowProfile struct {
	DeploymentWindowList []*timeoutWindow.TimeWindow `json:"deploymentWindowList,omitempty" validate:"required,min=1"`
	Enabled              bool                        `json:"enabled"`
	TimeZone             string                      `json:"timeZone"`
	DisplayMessage       string                      `json:"displayMessage"`
	ExcludedUsersEmails  []string                    `json:"excludedUsersEmails"`
	IsSuperAdminExcluded bool                        `json:"isSuperAdminExcluded"`
	IsUserExcluded       bool                        `json:"isUserExcluded"`
	DeploymentWindowProfileMetadata
}

func (state *EnvironmentState) GetSerializedAuditData(triggerType string, triggerMessage string) string {
	if state == nil {
		return "{}"
	}
	profile := state.AppliedProfile

	audit := DeploymentWindowAuditData{
		Audit:          state,
		TriggeredAt:    state.CalculatedAt,
		TriggerMessage: triggerMessage,
		TriggerType:    triggerType,
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
	TriggerType    string            `json:"triggerType"`
}

// DeploymentWindowProfileMetadata defines model for DeploymentWindowProfileMetadata.
type DeploymentWindowProfileMetadata struct {
	Description string               `json:"description"`
	Id          int                  `json:"id"`
	Name        string               `json:"name"`
	Type        DeploymentWindowType `json:"type" validate:"oneof=BLACKOUT MAINTENANCE"`
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

	AppliedProfile *ProfileWrapper `json:"appliedProfile"`

	// UserActionState describes the  eventual action state for the user
	UserActionState UserActionState `json:"userActionState"`
	CalculatedAt    time.Time       `json:"calculatedAt"`
}

type ProfileWrapper struct {
	DeploymentWindowProfile *DeploymentWindowProfile `json:"deploymentWindowProfile,omitempty"`
	ProfileStateData
	EnvId             int `json:"envId"`
	ExcludedUsersList []int32
}

type ProfileStateData struct {
	CalculatedTimestamp time.Time `json:"calculatedTimestamp"`
	IsActive            bool      `json:"isActive"`
	ExcludedUserEmails  []string  `json:"excludedUserEmails"`
}

// DeploymentWindowResponse defines model for DeploymentWindowResponse.
type DeploymentWindowResponse struct {
	EnvironmentStateMap map[int]EnvironmentState `json:"environmentStateMap,omitempty"`
	Profiles            []ProfileWrapper         `json:"profiles,omitempty"`
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

func (state UserActionState) GetErrorMessageForProfileAndState(envState *EnvironmentState) string {
	if state == Allowed {
		return ""
	}
	var profile *DeploymentWindowProfile
	if envState != nil && envState.AppliedProfile != nil {
		profile = envState.AppliedProfile.DeploymentWindowProfile
	}

	if profile != nil && profile.Type == Blackout {
		return "You are not authorized to deploy during blackout window " + strconv.Quote(profile.Name)
	} else if profile != nil && profile.Type == Maintenance {
		return "You are not authorized to deploy outside maintenance window"
	}
	return ""
}

func (item ProfileWrapper) isRestricted() bool {
	return (item.DeploymentWindowProfile.Type == Blackout && item.IsActive) || (item.DeploymentWindowProfile.Type == Maintenance && !item.IsActive)
}

func (a ProfileWrapper) compareProfile(b ProfileWrapper) bool {
	if a.DeploymentWindowProfile.Type != b.DeploymentWindowProfile.Type {
		return a.DeploymentWindowProfile.Type == Blackout
	}
	if a.IsActive != b.IsActive {
		return a.IsActive
	}
	return a.CalculatedTimestamp.Before(b.CalculatedTimestamp)
}
