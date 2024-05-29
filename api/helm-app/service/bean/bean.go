package bean

import (
	"fmt"
	"strconv"
)

type AppIdentifier struct {
	ClusterId   int    `json:"clusterId"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releaseName"`
}

// GetUniqueAppNameIdentifier returns unique app name identifier, we store all helm releases in kubelink cache with key
// as what is returned from this func, this is the case where an app across diff namespace or cluster can have same name,
// so to identify then uniquely below implementation would serve as good unique identifier for an external app.
func (r *AppIdentifier) GetUniqueAppNameIdentifier() string {
	return fmt.Sprintf("%s-%s-%s", r.ReleaseName, r.Namespace, strconv.Itoa(r.ClusterId))
}

func (r *AppIdentifier) GetUniqueAppIdentifierForGivenNamespaceAndCluster(namespace, clusterId string) string {
	return fmt.Sprintf("%s-%s-%s", r.ReleaseName, namespace, clusterId)
}
