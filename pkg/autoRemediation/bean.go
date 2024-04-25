package autoRemediation

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
	IdentifierType string      `json:"identifierType"`
	Data           TriggerData `json:"data"`
}

type TriggerData struct {
	RuntimeParameters      []RuntimeParameter `json:"runtimeParameters"`
	JobId                  int                `json:"jobId"`
	JobName                string             `json:"jobName"`
	PipelineId             int                `json:"pipelineId"`
	PipelineName           string             `json:"pipelineName"`
	ExecutionEnvironment   string             `json:"executionEnvironment"`
	ExecutionEnvironmentId int                `json:"executionEnvironmentId"`
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
	Message            string  `json:"message"`
	MessageType        string  `json:"messageType"`
	Event              string  `json:"event"`
	InvolvedObject     string  `json:"involvedObject"`
	ClusterId          int     `json:"clusterId"`
	ClusterName        string  `json:"clusterName"`
	Namespace          string  `json:"namespace"`
	WatcherName        string  `json:"watcherName"`
	InterceptedTime    string  `json:"interceptedTime"`
	ExecutionStatus    string  `json:"executionStatus"`
	TriggerExecutionId int     `json:"triggerExecutionId"`
	Trigger            Trigger `json:"trigger"`
}
