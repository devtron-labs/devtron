package autoRemediation

import "github.com/devtron-labs/devtron/pkg/autoRemediation/repository"

type EventConfiguration struct {
	Selectors       []Selector    `json:"selectors"`
	K8sResources    []K8sResource `json:"k8sResources"`
	EventExpression string        `json:"eventExpression"`
}

type Selector struct {
	Type      string   `json:"type"`
	Names     []string `json:"names"`
	GroupName string   `json:"groupName"`
}

type K8sResource struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type RuntimeParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Trigger struct {
	Id             int                    `json:"-"`
	IdentifierType repository.TriggerType `json:"identifierType"`
	Data           TriggerData            `json:"data"`
}

type TriggerData struct {
	RuntimeParameters      []RuntimeParameter `json:"runtimeParameters"`
	JobId                  int                `json:"jobId"`
	JobName                string             `json:"jobName"`
	PipelineId             int                `json:"pipelineId"`
	PipelineName           string             `json:"pipelineName"`
	ExecutionEnvironment   string             `json:"executionEnvironment"`
	ExecutionEnvironmentId int                `json:"executionEnvironmentId"`
	WorkflowId             int                `json:"workflowId"`
}

type WatcherDto struct {
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	EventConfiguration EventConfiguration `json:"eventConfiguration"`
	Triggers           []Trigger          `json:"triggers"`
}

type WatcherItem struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	JobPipelineName string `json:"jobPipelineName"`
	JobPipelineId   int    `json:"jobPipelineId"`
	WorkflowId      int    `json:"workflowId"`
}
type WatchersResponse struct {
	Size   int           `json:"size"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
	List   []WatcherItem `json:"list"`
}
type InterceptedEventsDto struct {
	Message        string `json:"message"`
	MessageType    string `json:"messageType"`
	Event          string `json:"event"`
	InvolvedObject string `json:"involvedObject"`

	ClusterName     string `json:"clusterName"`
	ClusterId       int    `json:"clusterId"`
	Namespace       string `json:"namespace"`
	EnvironmentName string `json:"environmentName"`

	WatcherName     string            `json:"watcherName"`
	InterceptedTime string            `json:"interceptedTime"`
	ExecutionStatus repository.Status `json:"executionStatus"`
	TriggerId       int               `json:"triggerId"`
	TriggerExecutionId int     `json:"triggerExecutionId"`
	Trigger         Trigger           `json:"trigger"`
}
