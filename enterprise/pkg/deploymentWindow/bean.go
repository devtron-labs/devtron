package deploymentWindow

import (
	"encoding/json"
	"strconv"
	"time"
)

type DeploymentWindowType string

const (
	Blackout    DeploymentWindowType = "BLACKOUT"
	Maintenance DeploymentWindowType = "MAINTENANCE"
)

type Frequency string

const (
	Fixed       Frequency = "FIXED"
	Daily       Frequency = "DAILY"
	Monthly     Frequency = "MONTHLY"
	Weekly      Frequency = "WEEKLY"
	WeeklyRange Frequency = "WEEKLY_RANGE"
)

type DayOfWeek string

const (
	Sunday    DayOfWeek = "Sunday"
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"
)

func (day DayOfWeek) toWeekday() time.Weekday {
	switch day {
	case Sunday:
		return time.Weekday(0)
	case Monday:
		return time.Weekday(0)
	case Tuesday:
		return time.Weekday(0)
	case Wednesday:
		return time.Weekday(0)
	case Thursday:
		return time.Weekday(0)
	case Friday:
		return time.Weekday(0)
	case Saturday:
		return time.Weekday(0)
	}
	return time.Weekday(-1)
}

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
	DeploymentWindowList []*TimeWindow `json:"deploymentWindowList,omitempty"`
	Enabled              bool          `json:"enabled"`
	TimeZone             string        `json:"timeZone"`
	DisplayMessage       string        `json:"displayMessage"`
	ExcludedUsersList    []int32       `json:"excludedUsersList"`
	IsSuperAdminExcluded bool          `json:"isSuperAdminExcluded"`
	IsUserExcluded       bool          `json:"isUserExcluded"`
	DeploymentWindowProfileMetadata
}

func (profile DeploymentWindowProfile) GetSerializedAuditData() string {
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
	AllExcludedUsers        []string                 `json:"allExcludedUsers"`
}

// DeploymentWindowResponse defines model for DeploymentWindowResponse.
type DeploymentWindowResponse struct {
	EnvironmentStateMap map[int]EnvironmentState `json:"environmentStateMap,omitempty"`
	Profiles            []ProfileState           `json:"profiles,omitempty"`
	//SuperAdmins         []string                 `json:"superAdmins,omitempty"`
}

// TimeWindow defines model for TimeWindow.
type TimeWindow struct {
	//Id        int       `json:"id"`
	Frequency Frequency `json:"frequency"`

	// relevant for daily and monthly
	DayFrom int `json:"dayFrom"`
	DayTo   int `json:"dayTo"`

	// relevant for
	HourMinuteFrom string `json:"hourMinuteFrom"`
	HourMinuteTo   string `json:"hourMinuteTo"`

	// optional for frequencies other than FIXED, otherwise required
	TimeFrom time.Time `json:"timeFrom"`
	TimeTo   time.Time `json:"timeTo"`

	// relevant for weekly range
	WeekdayFrom DayOfWeek `json:"weekdayFrom"`
	WeekdayTo   DayOfWeek `json:"weekdayTo"`

	// relevant for weekly
	Weekdays []DayOfWeek `json:"weekdays"`
}

func (window *TimeWindow) toJsonString() string {
	marshal, err := json.Marshal(window)
	if err != nil {
		return ""
	}
	return string(marshal)
}
func (window *TimeWindow) setFromJsonString(jsonString string) {
	json.Unmarshal([]byte(jsonString), window)
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
