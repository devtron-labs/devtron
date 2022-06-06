package k8s

type ClusterCapacityDetail struct {
	Id              int                   `json:"id"`
	Name            string                `json:"name"`
	NodeCount       int                   `json:"nodeCount,omitempty"`
	NodeErrors      []string              `json:"nodeErrors,omitempty"`
	NodeK8sVersions []string              `json:"nodeK8sVersions,omitempty"`
	Cpu             *ResourceDetailObject `json:"cpu"`
	Memory          *ResourceDetailObject `json:"memory"`
}

type NodeCapacityDetails struct {
	Name       string                `json:"name"`
	Status     string                `json:"status"`
	Roles      []string              `json:"roles"`
	K8sVersion string                `json:"k8SVersion"`
	PodCount   int                   `json:"podCount"`
	TaintCount int                   `json:"taintCount"`
	Cpu        *ResourceDetailObject `json:"cpu"`
	Memory     *ResourceDetailObject `json:"memory"`
	Age        string                `json:"age"`
	Errors     []string              `json:"errors"`
}

type ResourceDetailObject struct {
	ResourceName      string `json:"name,omitempty"`
	Capacity          string `json:"capacity,omitempty"`
	Allocatable       string `json:"allocatable,omitempty"`
	Usage             string `json:"usage,omitempty"`
	Request           string `json:"request,omitempty"`
	Limit             string `json:"limit,omitempty"`
	UsagePercentage   string `json:"usagePercentage,omitempty"`
	RequestPercentage string `json:"requestPercentage,omitempty"`
	LimitPercentage   string `json:"limitPercentage,omitempty"`
}

type ResourceMetric struct {
	ResourceType string `json:"resourceType"`
	Allocatable  string `json:"allocatable"`
	Utilization  string `json:"utilization"`
	Request      string `json:"request"`
	Limit        string `json:"limit"`
}
