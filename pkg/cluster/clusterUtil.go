package cluster

import "fmt"

const (
	CmName = "cluster-event"
)

func ParseCmNameForK8sInformerOnClusterEvent(clusterId int) string {
	return fmt.Sprintf("%s-%d", CmName, clusterId)
}
