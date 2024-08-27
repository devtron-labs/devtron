//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Ptr is a helper that returns a pointer to v.
func Ptr[T any](v T) *T {
	return &v
}

// AccessControlValue represents an access control value within GitLab,
// used for managing access to certain project features.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html
type AccessControlValue string

// List of available access control values.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html
const (
	DisabledAccessControl AccessControlValue = "disabled"
	EnabledAccessControl  AccessControlValue = "enabled"
	PrivateAccessControl  AccessControlValue = "private"
	PublicAccessControl   AccessControlValue = "public"
)

// AccessControl is a helper routine that allocates a new AccessControlValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func AccessControl(v AccessControlValue) *AccessControlValue {
	return Ptr(v)
}

// AccessLevelValue represents a permission level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/user/permissions.html
type AccessLevelValue int

// List of available access levels.
//
// GitLab API docs: https://docs.gitlab.com/ee/user/permissions.html
const (
	NoPermissions            AccessLevelValue = 0
	MinimalAccessPermissions AccessLevelValue = 5
	GuestPermissions         AccessLevelValue = 10
	ReporterPermissions      AccessLevelValue = 20
	DeveloperPermissions     AccessLevelValue = 30
	MaintainerPermissions    AccessLevelValue = 40
	OwnerPermissions         AccessLevelValue = 50
	AdminPermissions         AccessLevelValue = 60

	// Deprecated: Renamed to MaintainerPermissions in GitLab 11.0.
	MasterPermissions AccessLevelValue = 40
	// Deprecated: Renamed to OwnerPermissions.
	OwnerPermission AccessLevelValue = 50
)

// AccessLevel is a helper routine that allocates a new AccessLevelValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func AccessLevel(v AccessLevelValue) *AccessLevelValue {
	return Ptr(v)
}

// UserIDValue represents a user ID value within GitLab.
type UserIDValue string

// List of available user ID values.
const (
	UserIDAny  UserIDValue = "Any"
	UserIDNone UserIDValue = "None"
)

// ApproverIDsValue represents an approver ID value within GitLab.
type ApproverIDsValue struct {
	value interface{}
}

// ApproverIDs is a helper routine that creates a new ApproverIDsValue.
func ApproverIDs(v interface{}) *ApproverIDsValue {
	switch v.(type) {
	case UserIDValue, []int:
		return &ApproverIDsValue{value: v}
	default:
		panic("Unsupported value passed as approver ID")
	}
}

// EncodeValues implements the query.Encoder interface.
func (a *ApproverIDsValue) EncodeValues(key string, v *url.Values) error {
	switch value := a.value.(type) {
	case UserIDValue:
		v.Set(key, string(value))
	case []int:
		v.Del(key)
		v.Del(key + "[]")
		for _, id := range value {
			v.Add(key+"[]", strconv.Itoa(id))
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (a ApproverIDsValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *ApproverIDsValue) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, a.value)
}

// AssigneeIDValue represents an assignee ID value within GitLab.
type AssigneeIDValue struct {
	value interface{}
}

// AssigneeID is a helper routine that creates a new AssigneeIDValue.
func AssigneeID(v interface{}) *AssigneeIDValue {
	switch v.(type) {
	case UserIDValue, int:
		return &AssigneeIDValue{value: v}
	default:
		panic("Unsupported value passed as assignee ID")
	}
}

// EncodeValues implements the query.Encoder interface.
func (a *AssigneeIDValue) EncodeValues(key string, v *url.Values) error {
	switch value := a.value.(type) {
	case UserIDValue:
		v.Set(key, string(value))
	case int:
		v.Set(key, strconv.Itoa(value))
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (a AssigneeIDValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *AssigneeIDValue) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, a.value)
}

// ReviewerIDValue represents a reviewer ID value within GitLab.
type ReviewerIDValue struct {
	value interface{}
}

// ReviewerID is a helper routine that creates a new ReviewerIDValue.
func ReviewerID(v interface{}) *ReviewerIDValue {
	switch v.(type) {
	case UserIDValue, int:
		return &ReviewerIDValue{value: v}
	default:
		panic("Unsupported value passed as reviewer ID")
	}
}

// EncodeValues implements the query.Encoder interface.
func (a *ReviewerIDValue) EncodeValues(key string, v *url.Values) error {
	switch value := a.value.(type) {
	case UserIDValue:
		v.Set(key, string(value))
	case int:
		v.Set(key, strconv.Itoa(value))
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (a ReviewerIDValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *ReviewerIDValue) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, a.value)
}

// AvailabilityValue represents an availability value within GitLab.
type AvailabilityValue string

// List of available availability values.
//
// Undocummented, see code at:
// https://gitlab.com/gitlab-org/gitlab-foss/-/blob/master/app/models/user_status.rb#L22
const (
	NotSet AvailabilityValue = "not_set"
	Busy   AvailabilityValue = "busy"
)

// Availability is a helper routine that allocates a new AvailabilityValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func Availability(v AvailabilityValue) *AvailabilityValue {
	return Ptr(v)
}

// BuildStateValue represents a GitLab build state.
type BuildStateValue string

// These constants represent all valid build states.
const (
	Created            BuildStateValue = "created"
	WaitingForResource BuildStateValue = "waiting_for_resource"
	Preparing          BuildStateValue = "preparing"
	Pending            BuildStateValue = "pending"
	Running            BuildStateValue = "running"
	Success            BuildStateValue = "success"
	Failed             BuildStateValue = "failed"
	Canceled           BuildStateValue = "canceled"
	Skipped            BuildStateValue = "skipped"
	Manual             BuildStateValue = "manual"
	Scheduled          BuildStateValue = "scheduled"
)

// BuildState is a helper routine that allocates a new BuildStateValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func BuildState(v BuildStateValue) *BuildStateValue {
	return Ptr(v)
}

// CommentEventAction identifies if a comment has been newly created or updated.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#comment-events
type CommentEventAction string

const (
	CommentEventActionCreate CommentEventAction = "create"
	CommentEventActionUpdate CommentEventAction = "update"
)

// ContainerRegistryStatus represents the status of a Container Registry.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#list-registry-repositories
type ContainerRegistryStatus string

// ContainerRegistryStatus represents all valid statuses of a Container Registry.
//
// Undocumented, see code at:
// https://gitlab.com/gitlab-org/gitlab/-/blob/master/app/models/container_repository.rb?ref_type=heads#L35
const (
	ContainerRegistryStatusDeleteScheduled ContainerRegistryStatus = "delete_scheduled"
	ContainerRegistryStatusDeleteFailed    ContainerRegistryStatus = "delete_failed"
	ContainerRegistryStatusDeleteOngoing   ContainerRegistryStatus = "delete_ongoing"
)

// DeploymentApprovalStatus represents a Gitlab deployment approval status.
type DeploymentApprovalStatus string

// These constants represent all valid deployment approval statuses.
const (
	DeploymentApprovalStatusApproved DeploymentApprovalStatus = "approved"
	DeploymentApprovalStatusRejected DeploymentApprovalStatus = "rejected"
)

// DeploymentStatusValue represents a Gitlab deployment status.
type DeploymentStatusValue string

// These constants represent all valid deployment statuses.
const (
	DeploymentStatusCreated  DeploymentStatusValue = "created"
	DeploymentStatusRunning  DeploymentStatusValue = "running"
	DeploymentStatusSuccess  DeploymentStatusValue = "success"
	DeploymentStatusFailed   DeploymentStatusValue = "failed"
	DeploymentStatusCanceled DeploymentStatusValue = "canceled"
)

// DeploymentStatus is a helper routine that allocates a new
// DeploymentStatusValue to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func DeploymentStatus(v DeploymentStatusValue) *DeploymentStatusValue {
	return Ptr(v)
}

// DORAMetricType represents all valid DORA metrics types.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/dora/metrics.html
type DORAMetricType string

// List of available DORA metric type names.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/dora/metrics.html
const (
	DORAMetricDeploymentFrequency  DORAMetricType = "deployment_frequency"
	DORAMetricLeadTimeForChanges   DORAMetricType = "lead_time_for_changes"
	DORAMetricTimeToRestoreService DORAMetricType = "time_to_restore_service"
	DORAMetricChangeFailureRate    DORAMetricType = "change_failure_rate"
)

// DORAMetricInterval represents the time period over which the
// metrics are aggregated.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/dora/metrics.html
type DORAMetricInterval string

// List of available DORA metric interval types.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/dora/metrics.html
const (
	DORAMetricIntervalDaily   DORAMetricInterval = "daily"
	DORAMetricIntervalMonthly DORAMetricInterval = "monthly"
	DORAMetricIntervalAll     DORAMetricInterval = "all"
)

// EventTypeValue represents actions type for contribution events.
type EventTypeValue string

// List of available action type.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/profile/contributions_calendar.html#user-contribution-events
const (
	CreatedEventType   EventTypeValue = "created"
	UpdatedEventType   EventTypeValue = "updated"
	ClosedEventType    EventTypeValue = "closed"
	ReopenedEventType  EventTypeValue = "reopened"
	PushedEventType    EventTypeValue = "pushed"
	CommentedEventType EventTypeValue = "commented"
	MergedEventType    EventTypeValue = "merged"
	JoinedEventType    EventTypeValue = "joined"
	LeftEventType      EventTypeValue = "left"
	DestroyedEventType EventTypeValue = "destroyed"
	ExpiredEventType   EventTypeValue = "expired"
)

// EventTargetTypeValue represents actions type value for contribution events.
type EventTargetTypeValue string

// List of available action type.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/events.html#target-types
const (
	IssueEventTargetType        EventTargetTypeValue = "issue"
	MilestoneEventTargetType    EventTargetTypeValue = "milestone"
	MergeRequestEventTargetType EventTargetTypeValue = "merge_request"
	NoteEventTargetType         EventTargetTypeValue = "note"
	ProjectEventTargetType      EventTargetTypeValue = "project"
	SnippetEventTargetType      EventTargetTypeValue = "snippet"
	UserEventTargetType         EventTargetTypeValue = "user"
)

// FileActionValue represents the available actions that can be performed on a file.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/commits.html#create-a-commit-with-multiple-files-and-actions
type FileActionValue string

// The available file actions.
const (
	FileCreate FileActionValue = "create"
	FileDelete FileActionValue = "delete"
	FileMove   FileActionValue = "move"
	FileUpdate FileActionValue = "update"
	FileChmod  FileActionValue = "chmod"
)

// FileAction is a helper routine that allocates a new FileActionValue value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func FileAction(v FileActionValue) *FileActionValue {
	return Ptr(v)
}

// GenericPackageSelectValue represents a generic package select value.
type GenericPackageSelectValue string

// The available generic package select values.
const (
	SelectPackageFile GenericPackageSelectValue = "package_file"
)

// GenericPackageSelect is a helper routine that allocates a new
// GenericPackageSelectValue value to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func GenericPackageSelect(v GenericPackageSelectValue) *GenericPackageSelectValue {
	return Ptr(v)
}

// GenericPackageStatusValue represents a generic package status.
type GenericPackageStatusValue string

// The available generic package statuses.
const (
	PackageDefault GenericPackageStatusValue = "default"
	PackageHidden  GenericPackageStatusValue = "hidden"
)

// GenericPackageStatus is a helper routine that allocates a new
// GenericPackageStatusValue value to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func GenericPackageStatus(v GenericPackageStatusValue) *GenericPackageStatusValue {
	return Ptr(v)
}

// ISOTime represents an ISO 8601 formatted date.
type ISOTime time.Time

// ISO 8601 date format.
const iso8601 = "2006-01-02"

// ParseISOTime parses an ISO 8601 formatted date.
func ParseISOTime(s string) (ISOTime, error) {
	t, err := time.Parse(iso8601, s)
	return ISOTime(t), err
}

// MarshalJSON implements the json.Marshaler interface.
func (t ISOTime) MarshalJSON() ([]byte, error) {
	if reflect.ValueOf(t).IsZero() {
		return []byte(`null`), nil
	}

	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		// ISO 8901 uses 4 digits for the years.
		return nil, errors.New("json: ISOTime year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(iso8601)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, iso8601)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *ISOTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	isotime, err := time.Parse(`"`+iso8601+`"`, string(data))
	*t = ISOTime(isotime)

	return err
}

// EncodeValues implements the query.Encoder interface.
func (t *ISOTime) EncodeValues(key string, v *url.Values) error {
	if t == nil || (time.Time(*t)).IsZero() {
		return nil
	}
	v.Add(key, t.String())
	return nil
}

// String implements the Stringer interface.
func (t ISOTime) String() string {
	return time.Time(t).Format(iso8601)
}

// Labels represents a list of labels.
type Labels []string

// LabelOptions is a custom type with specific marshaling characteristics.
type LabelOptions []string

// MarshalJSON implements the json.Marshaler interface.
func (l *LabelOptions) MarshalJSON() ([]byte, error) {
	if *l == nil {
		return []byte(`null`), nil
	}
	return json.Marshal(strings.Join(*l, ","))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (l *LabelOptions) UnmarshalJSON(data []byte) error {
	type alias LabelOptions
	if !bytes.HasPrefix(data, []byte("[")) {
		data = []byte(fmt.Sprintf("[%s]", string(data)))
	}
	return json.Unmarshal(data, (*alias)(l))
}

// EncodeValues implements the query.EncodeValues interface.
func (l *LabelOptions) EncodeValues(key string, v *url.Values) error {
	v.Set(key, strings.Join(*l, ","))
	return nil
}

// LinkTypeValue represents a release link type.
type LinkTypeValue string

// List of available release link types.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/releases/links.html#create-a-release-link
const (
	ImageLinkType   LinkTypeValue = "image"
	OtherLinkType   LinkTypeValue = "other"
	PackageLinkType LinkTypeValue = "package"
	RunbookLinkType LinkTypeValue = "runbook"
)

// LinkType is a helper routine that allocates a new LinkType value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func LinkType(v LinkTypeValue) *LinkTypeValue {
	return Ptr(v)
}

// LicenseApprovalStatusValue describe the approval statuses of a license.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/managed_licenses.html
type LicenseApprovalStatusValue string

// List of available license approval statuses.
const (
	LicenseApproved    LicenseApprovalStatusValue = "approved"
	LicenseBlacklisted LicenseApprovalStatusValue = "blacklisted"
	LicenseAllowed     LicenseApprovalStatusValue = "allowed"
	LicenseDenied      LicenseApprovalStatusValue = "denied"
)

// LicenseApprovalStatus is a helper routine that allocates a new license
// approval status value to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func LicenseApprovalStatus(v LicenseApprovalStatusValue) *LicenseApprovalStatusValue {
	return Ptr(v)
}

// MergeMethodValue represents a project merge type within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#project-merge-method
type MergeMethodValue string

// List of available merge type
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#project-merge-method
const (
	NoFastForwardMerge MergeMethodValue = "merge"
	FastForwardMerge   MergeMethodValue = "ff"
	RebaseMerge        MergeMethodValue = "rebase_merge"
)

// MergeMethod is a helper routine that allocates a new MergeMethod
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func MergeMethod(v MergeMethodValue) *MergeMethodValue {
	return Ptr(v)
}

// NoteTypeValue represents the type of a Note.
type NoteTypeValue string

// List of available note types.
const (
	DiffNote       NoteTypeValue = "DiffNote"
	DiscussionNote NoteTypeValue = "DiscussionNote"
	GenericNote    NoteTypeValue = "Note"
	LegacyDiffNote NoteTypeValue = "LegacyDiffNote"
)

// NoteType is a helper routine that allocates a new NoteTypeValue to
// store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func NoteType(v NoteTypeValue) *NoteTypeValue {
	return Ptr(v)
}

// NotificationLevelValue represents a notification level.
type NotificationLevelValue int

// String implements the fmt.Stringer interface.
func (l NotificationLevelValue) String() string {
	return notificationLevelNames[l]
}

// MarshalJSON implements the json.Marshaler interface.
func (l NotificationLevelValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (l *NotificationLevelValue) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw := raw.(type) {
	case float64:
		*l = NotificationLevelValue(raw)
	case string:
		*l = notificationLevelTypes[raw]
	case nil:
		// No action needed.
	default:
		return fmt.Errorf("json: cannot unmarshal %T into Go value of type %T", raw, *l)
	}

	return nil
}

// List of valid notification levels.
const (
	DisabledNotificationLevel NotificationLevelValue = iota
	ParticipatingNotificationLevel
	WatchNotificationLevel
	GlobalNotificationLevel
	MentionNotificationLevel
	CustomNotificationLevel
)

var notificationLevelNames = [...]string{
	"disabled",
	"participating",
	"watch",
	"global",
	"mention",
	"custom",
}

var notificationLevelTypes = map[string]NotificationLevelValue{
	"disabled":      DisabledNotificationLevel,
	"participating": ParticipatingNotificationLevel,
	"watch":         WatchNotificationLevel,
	"global":        GlobalNotificationLevel,
	"mention":       MentionNotificationLevel,
	"custom":        CustomNotificationLevel,
}

// NotificationLevel is a helper routine that allocates a new NotificationLevelValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func NotificationLevel(v NotificationLevelValue) *NotificationLevelValue {
	return Ptr(v)
}

// ProjectCreationLevelValue represents a project creation level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
type ProjectCreationLevelValue string

// List of available project creation levels.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
const (
	NoOneProjectCreation      ProjectCreationLevelValue = "noone"
	MaintainerProjectCreation ProjectCreationLevelValue = "maintainer"
	DeveloperProjectCreation  ProjectCreationLevelValue = "developer"
)

// ProjectCreationLevel is a helper routine that allocates a new ProjectCreationLevelValue
// to store v and returns a pointer to it.
// Please use Ptr instead.
func ProjectCreationLevel(v ProjectCreationLevelValue) *ProjectCreationLevelValue {
	return Ptr(v)
}

// ProjectHookEvent represents a project hook event.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#hook-events
type ProjectHookEvent string

// List of available project hook events.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#hook-events
const (
	ProjectHookEventPush                ProjectHookEvent = "push_events"
	ProjectHookEventTagPush             ProjectHookEvent = "tag_push_events"
	ProjectHookEventIssues              ProjectHookEvent = "issues_events"
	ProjectHookEventConfidentialIssues  ProjectHookEvent = "confidential_issues_events"
	ProjectHookEventNote                ProjectHookEvent = "note_events"
	ProjectHookEventMergeRequests       ProjectHookEvent = "merge_requests_events"
	ProjectHookEventJob                 ProjectHookEvent = "job_events"
	ProjectHookEventPipeline            ProjectHookEvent = "pipeline_events"
	ProjectHookEventWiki                ProjectHookEvent = "wiki_page_events"
	ProjectHookEventReleases            ProjectHookEvent = "releases_events"
	ProjectHookEventEmoji               ProjectHookEvent = "emoji_events"
	ProjectHookEventResourceAccessToken ProjectHookEvent = "resource_access_token_events"
)

// ResourceGroupProcessMode represents a process mode for a resource group
// within a GitLab project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/ci/resource_groups/index.html#process-modes
type ResourceGroupProcessMode string

// List of available resource group process modes.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/ci/resource_groups/index.html#process-modes
const (
	Unordered   ResourceGroupProcessMode = "unordered"
	OldestFirst ResourceGroupProcessMode = "oldest_first"
	NewestFirst ResourceGroupProcessMode = "newest_first"
)

// SharedRunnersSettingValue determines whether shared runners are enabled for a
// group’s subgroups and projects.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#options-for-shared_runners_setting
type SharedRunnersSettingValue string

// List of available shared runner setting levels.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#options-for-shared_runners_setting
const (
	EnabledSharedRunnersSettingValue                  SharedRunnersSettingValue = "enabled"
	DisabledAndOverridableSharedRunnersSettingValue   SharedRunnersSettingValue = "disabled_and_overridable"
	DisabledAndUnoverridableSharedRunnersSettingValue SharedRunnersSettingValue = "disabled_and_unoverridable"

	// Deprecated: DisabledWithOverrideSharedRunnersSettingValue is deprecated
	// in favor of DisabledAndOverridableSharedRunnersSettingValue.
	DisabledWithOverrideSharedRunnersSettingValue SharedRunnersSettingValue = "disabled_with_override"
)

// SharedRunnersSetting is a helper routine that allocates a new SharedRunnersSettingValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func SharedRunnersSetting(v SharedRunnersSettingValue) *SharedRunnersSettingValue {
	return Ptr(v)
}

// SubGroupCreationLevelValue represents a sub group creation level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
type SubGroupCreationLevelValue string

// List of available sub group creation levels.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
const (
	OwnerSubGroupCreationLevelValue      SubGroupCreationLevelValue = "owner"
	MaintainerSubGroupCreationLevelValue SubGroupCreationLevelValue = "maintainer"
)

// SubGroupCreationLevel is a helper routine that allocates a new SubGroupCreationLevelValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func SubGroupCreationLevel(v SubGroupCreationLevelValue) *SubGroupCreationLevelValue {
	return Ptr(v)
}

// SquashOptionValue represents a squash optional level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
type SquashOptionValue string

// List of available squash options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
const (
	SquashOptionNever      SquashOptionValue = "never"
	SquashOptionAlways     SquashOptionValue = "always"
	SquashOptionDefaultOff SquashOptionValue = "default_off"
	SquashOptionDefaultOn  SquashOptionValue = "default_on"
)

// SquashOption is a helper routine that allocates a new SquashOptionValue
// to store s and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func SquashOption(s SquashOptionValue) *SquashOptionValue {
	return Ptr(s)
}

// TasksCompletionStatus represents tasks of the issue/merge request.
type TasksCompletionStatus struct {
	Count          int `json:"count"`
	CompletedCount int `json:"completed_count"`
}

// TodoAction represents the available actions that can be performed on a todo.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/todos.html
type TodoAction string

// The available todo actions.
const (
	TodoAssigned          TodoAction = "assigned"
	TodoMentioned         TodoAction = "mentioned"
	TodoBuildFailed       TodoAction = "build_failed"
	TodoMarked            TodoAction = "marked"
	TodoApprovalRequired  TodoAction = "approval_required"
	TodoDirectlyAddressed TodoAction = "directly_addressed"
)

// TodoTargetType represents the available target that can be linked to a todo.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/todos.html
type TodoTargetType string

const (
	TodoTargetAlertManagement  TodoTargetType = "AlertManagement::Alert"
	TodoTargetDesignManagement TodoTargetType = "DesignManagement::Design"
	TodoTargetIssue            TodoTargetType = "Issue"
	TodoTargetMergeRequest     TodoTargetType = "MergeRequest"
)

// UploadType represents the available upload types.
type UploadType string

// The available upload types.
const (
	UploadAvatar UploadType = "avatar"
	UploadFile   UploadType = "file"
)

// VariableTypeValue represents a variable type within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
type VariableTypeValue string

// List of available variable types.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
const (
	EnvVariableType  VariableTypeValue = "env_var"
	FileVariableType VariableTypeValue = "file"
)

// VariableType is a helper routine that allocates a new VariableTypeValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func VariableType(v VariableTypeValue) *VariableTypeValue {
	return Ptr(v)
}

// VisibilityValue represents a visibility level within GitLab.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
type VisibilityValue string

// List of available visibility levels.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/
const (
	PrivateVisibility  VisibilityValue = "private"
	InternalVisibility VisibilityValue = "internal"
	PublicVisibility   VisibilityValue = "public"
)

// Visibility is a helper routine that allocates a new VisibilityValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func Visibility(v VisibilityValue) *VisibilityValue {
	return Ptr(v)
}

// WikiFormatValue represents the available wiki formats.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/wikis.html
type WikiFormatValue string

// The available wiki formats.
const (
	WikiFormatMarkdown WikiFormatValue = "markdown"
	WikiFormatRDoc     WikiFormatValue = "rdoc"
	WikiFormatASCIIDoc WikiFormatValue = "asciidoc"
	WikiFormatOrg      WikiFormatValue = "org"
)

// WikiFormat is a helper routine that allocates a new WikiFormatValue
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func WikiFormat(v WikiFormatValue) *WikiFormatValue {
	return Ptr(v)
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func Bool(v bool) *bool {
	return Ptr(v)
}

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func Int(v int) *int {
	return Ptr(v)
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func String(v string) *string {
	return Ptr(v)
}

// Time is a helper routine that allocates a new time.Time value
// to store v and returns a pointer to it.
//
// Deprecated: Please use Ptr instead.
func Time(v time.Time) *time.Time {
	return Ptr(v)
}

// BoolValue is a boolean value with advanced json unmarshaling features.
type BoolValue bool

// UnmarshalJSON allows 1, 0, "true", and "false" to be considered as boolean values
// Needed for:
// https://gitlab.com/gitlab-org/gitlab-ce/issues/50122
// https://gitlab.com/gitlab-org/gitlab/-/issues/233941
// https://github.com/gitlabhq/terraform-provider-gitlab/issues/348
func (t *BoolValue) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"1"`:
		*t = true
		return nil
	case `"0"`:
		*t = false
		return nil
	case `"true"`:
		*t = true
		return nil
	case `"false"`:
		*t = false
		return nil
	default:
		var v bool
		err := json.Unmarshal(b, &v)
		*t = BoolValue(v)
		return err
	}
}
