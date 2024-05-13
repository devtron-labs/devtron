package types

import (
	"encoding/json"
	"github.com/devtron-labs/scoop/types"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

type EventConfiguration struct {
	Selectors       []Selector        `json:"selectors" validate:"dive,min=1"`
	K8sResources    []*K8sResource    `json:"k8sResources" validate:"required"`
	EventExpression string            `json:"eventExpression"`
	SelectedActions []types.EventType `json:"selectedActions" validate:"required"`
}

func (ec *EventConfiguration) GetEnvsFromSelectors() []string {
	envNames := make([]string, 0)
	for _, selector := range ec.Selectors {
		envNames = append(envNames, selector.Names...)
	}
	return envNames
}

func (ec *EventConfiguration) GetK8sResources() []schema.GroupVersionKind {
	gvks := make([]schema.GroupVersionKind, 0, len(ec.K8sResources))
	for _, gvk := range ec.K8sResources {
		gvks = append(gvks, gvk.GetGVK())
	}
	return gvks
}

type SelectorType string

const EnvironmentSelector SelectorType = "environment"
const AllClusterGroup = "ALL"

type Selector struct {
	Type SelectorType `json:"type" validate:"oneof= environment"`
	// SubGroup is INCLUDED,EXCLUDED,ALL_PROD,ALL_NON_PROD
	SubGroup types.InterestCriteria `json:"subGroup" validate:"oneof= INCLUDED EXCLUDED ALL_PROD ALL_NON_PROD"`
	Names    []string               `json:"names"`
	// GroupName "ALL CLUSTER" or selected env name
	GroupName string `json:"groupName"`
}

func GetNamespaceSelector(selector Selector) types.NamespaceSelector {
	return types.NamespaceSelector{
		InterestGroup: selector.SubGroup,
		Namespaces:    selector.Names,
	}
}

type K8sResource struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (gvk *K8sResource) GetGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

type RuntimeParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Trigger struct {
	Id             int         `json:"-"`
	IdentifierType TriggerType `json:"identifierType" validate:"required,oneof=DEVTRON_JOB"`
	Data           TriggerData `json:"data" validate:"dive"`
	WatcherId      int         `json:"-"`
}
type TriggerType string

const (
	DEVTRON_JOB TriggerType = "DEVTRON_JOB"
)

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
	Id                 int                `json:"-"`
	Name               string             `json:"name" validate:"global-entity-name"`
	Description        string             `json:"description"`
	EventConfiguration EventConfiguration `json:"eventConfiguration" validate:"dive"`
	Triggers           []Trigger          `json:"triggers" validate:"dive"`
}

type WatcherItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// below data should be in an array
	JobPipelineName string    `json:"jobPipelineName"`
	JobPipelineId   int       `json:"jobPipelineId"`
	JobId           int       `json:"jobId"`
	WorkflowId      int       `json:"workflowId"`
	TriggeredAt     time.Time `json:"triggeredAt"`
}
type WatchersResponse struct {
	Size   int           `json:"size"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
	List   []WatcherItem `json:"list"`
}
type InterceptedResponse struct {
	Offset int                    `json:"offset"`
	Size   int                    `json:"size"`
	Total  int                    `json:"total"`
	List   []InterceptedEventsDto `json:"list"`
}

type InterceptedEventsDto struct {
	// Message        string `json:"message"`
	InterceptedEventId int    `json:"interceptedEventId"`
	Action             string `json:"action"`
	InvolvedObjects    string `json:"involvedObjects"`
	Metadata           string `json:"metadata"`

	ClusterName     string `json:"clusterName"`
	ClusterId       int    `json:"clusterId"`
	Namespace       string `json:"namespace"`
	EnvironmentName string `json:"environmentName"`

	WatcherName        string  `json:"watcherName"`
	InterceptedTime    string  `json:"interceptedTime"`
	ExecutionStatus    Status  `json:"executionStatus"`
	TriggerId          int     `json:"triggerId"`
	TriggerExecutionId int     `json:"triggerExecutionId"`
	Trigger            Trigger `json:"trigger"`
	ExecutionMessage   string  `json:"executionMessage"`
}

type Status string

const (
	Failure     Status = "Failure"
	Success     Status = "Success"
	Progressing Status = "Progressing"
	Errored     Status = "Error"
)

func FetchGvksFromK8sResources(resources []*K8sResource) (string, error) {
	gvks, err := json.Marshal(resources)
	if err != nil {
		return "", err
	}
	return string(gvks), nil
}

func GetSelectorJson(selectors []Selector) (string, error) {
	selectorBytes, err := json.Marshal(&selectors)
	return string(selectorBytes), err
}

func GetSelectorsFromJson(selectorsJson string) ([]Selector, error) {
	selectors := make([]Selector, 0)
	err := json.Unmarshal([]byte(selectorsJson), &selectors)
	return selectors, err
}

func GetClusterSelector(clusterName string, selectors []Selector) *Selector {
	for _, selector := range selectors {
		if selector.GroupName == AllClusterGroup || selector.GroupName == clusterName {
			return &selector
		}
	}

	return nil
}

// ComputeImpactedClusterNames
// since watchers are created with atleast one cluster, return empty list if all clusters are impacted
func ComputeImpactedClusterNames(oldWatcherSelectors, newWatcherSelectors []Selector) []string {
	clusterNamesMap := make(map[string]bool)

	// newWatcherSelectors can be null for delete case
	if newWatcherSelectors != nil {
		for _, selector := range newWatcherSelectors {
			if selector.GroupName == AllClusterGroup {
				// return empty list if all clusters are impacted
				return []string{}
			}
			clusterNamesMap[selector.GroupName] = true
		}
	}

	// oldWatcher can be null for create
	if oldWatcherSelectors != nil {
		for _, selector := range oldWatcherSelectors {
			if selector.GroupName == AllClusterGroup {
				// return empty list if all clusters are impacted
				return []string{}
			}
			clusterNamesMap[selector.GroupName] = true
		}
	}

	return maps.Keys(clusterNamesMap)

}
