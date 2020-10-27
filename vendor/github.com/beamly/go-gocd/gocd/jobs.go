package gocd

import (
	"context"
	"encoding/json"
)

const (
	// JobStateTransitionPassed "Passed"
	JobStateTransitionPassed = "Passed"
	// JobStateTransitionScheduled "Scheduled"
	JobStateTransitionScheduled = "Scheduled"
)

// JobsService describes actions which can be performed on jobs
type JobsService service

// Job describes a job which can be performed in GoCD
// codebeat:disable[TOO_MANY_IVARS]
type Job struct {
	AgentUUID            string                 `json:"agent_uuid,omitempty"`
	Name                 string                 `json:"name"`
	JobStateTransitions  []*JobStateTransition  `json:"job_state_transitions,omitempty"`
	ScheduledDate        int                    `json:"scheduled_date,omitempty"`
	OriginalJobID        string                 `json:"original_job_id,omitempty"`
	PipelineCounter      int                    `json:"pipeline_counter,omitempty"`
	Rerun                bool                   `json:"rerun,omitempty"`
	PipelineName         string                 `json:"pipeline_name,omitempty"`
	Result               string                 `json:"result,omitempty"`
	State                string                 `json:"state,omitempty"`
	ID                   int                    `json:"id,omitempty"`
	StageCounter         string                 `json:"stage_counter,omitempty"`
	StageName            string                 `json:"stage_name,omitempty"`
	RunInstanceCount     int                    `json:"run_instance_count,omitempty"`
	Timeout              TimeoutField           `json:"timeout,omitempty"`
	EnvironmentVariables []*EnvironmentVariable `json:"environment_variables,omitempty"`
	Properties           []*JobProperty         `json:"properties,omitempty"`
	Resources            []string               `json:"resources,omitempty"`
	Tasks                []*Task                `json:"tasks,omitempty"`
	Tabs                 []*Tab                 `json:"tabs,omitempty"`
	Artifacts            []*Artifact            `json:"artifacts,omitempty"`
	ElasticProfileID     string                 `json:"elastic_profile_id,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// Artifact describes the result of a job
type Artifact struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Tab description in a gocd job
type Tab struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// JobProperty describes the property for a job
type JobProperty struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	XPath  string `json:"xpath"`
}

// EnvironmentVariable describes an environment variable key/pair.
type EnvironmentVariable struct {
	Name           string `json:"name"`
	Value          string `json:"value,omitempty"`
	EncryptedValue string `json:"encrypted_value,omitempty"`
	Secure         bool   `json:"secure"`
}

type unencryptedEnvironmentVariable struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secure bool   `json:"secure"`
}

type encryptedEnvironmentVariable struct {
	Name           string `json:"name"`
	EncryptedValue string `json:"encrypted_value"`
	Secure         bool   `json:"secure"`
}

// MarshalJSON is a custom JSON Marhsaller for EnvironmentVariables to handle empty values
func (v *EnvironmentVariable) MarshalJSON() ([]byte, error) {
	if v.EncryptedValue != "" {
		return json.Marshal(&encryptedEnvironmentVariable{
			Name:           v.Name,
			Secure:         v.Secure,
			EncryptedValue: v.EncryptedValue,
		})
	}

	return json.Marshal(&unencryptedEnvironmentVariable{
		Name:   v.Name,
		Secure: v.Secure,
		Value:  v.Value,
	})
}

// PluginConfigurationKVPair describes a key/value pair of plugin configurations.
type PluginConfigurationKVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Task Describes a Task object in the GoCD api.
type Task struct {
	Type       string         `json:"type"`
	Attributes TaskAttributes `json:"attributes"`
}

// TaskAttributes describes all the properties for a Task.
// codebeat:disable[TOO_MANY_IVARS]
type TaskAttributes struct {
	RunIf               []string                    `json:"run_if,omitempty"`
	Command             string                      `json:"command,omitempty"`
	WorkingDirectory    string                      `json:"working_directory,omitempty"`
	Arguments           []string                    `json:"arguments,omitempty"`
	BuildFile           string                      `json:"build_file,omitempty"`
	Target              string                      `json:"target,omitempty"`
	NantPath            string                      `json:"nant_path,omitempty"`
	Pipeline            string                      `json:"pipeline,omitempty"`
	Stage               string                      `json:"stage,omitempty"`
	Job                 string                      `json:"job,omitempty"`
	Source              string                      `json:"source,omitempty"`
	IsSourceAFile       bool                        `json:"is_source_a_file,omitempty"`
	Destination         string                      `json:"destination,omitempty"`
	PluginConfiguration *TaskPluginConfiguration    `json:"plugin_configuration,omitempty"`
	Configuration       []PluginConfigurationKVPair `json:"configuration,omitempty"`
	ArtifactOrigin      string                      `json:"artifact_origin,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// TaskPluginConfiguration is for specifying options for pluggable task
type TaskPluginConfiguration struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// JobStateTransition describes a State Transition object in a GoCD api response
type JobStateTransition struct {
	StateChangeTime int    `json:"state_change_time,omitempty"`
	ID              int    `json:"id,omitempty"`
	State           string `json:"state,omitempty"`
}

// JobRunHistoryResponse describes the api response from
type JobRunHistoryResponse struct {
	Jobs       []*Job              `json:"jobs,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// JobScheduleResponse contains a collection of jobs
type JobScheduleResponse struct {
	Jobs []*JobSchedule `xml:"job"`
}

// JobSchedule describes the event causes for a job
type JobSchedule struct {
	Name                 string               `xml:"name,attr"`
	ID                   string               `xml:"id,attr"`
	Link                 JobScheduleLink      `xml:"link"`
	BuildLocator         string               `xml:"buildLocator"`
	Resources            []string             `xml:"resources>resource"`
	EnvironmentVariables *[]JobScheduleEnvVar `xml:"environmentVariables,omitempty>variable"`
}

// JobScheduleEnvVar describes the environmnet variables for a job schedule
type JobScheduleEnvVar struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

// JobScheduleLink describes the HAL links for a job schedule
type JobScheduleLink struct {
	Rel  string `xml:"rel,attr"`
	HRef string `xml:"href,attr"`
}

// TimeoutField helps manage the marshalling of the timoeut field which can be both "never" and an integer
type TimeoutField int

// ListScheduled lists Pipeline groups
func (js *JobsService) ListScheduled(ctx context.Context) (jobs []*JobSchedule, resp *APIResponse, err error) {
	j := &JobScheduleResponse{}
	_, resp, err = js.client.getAction(ctx, &APIClientRequest{
		Path:         "jobs/scheduled.xml",
		ResponseBody: j,
		ResponseType: responseTypeXML,
	})
	jobs = j.Jobs

	return
}
