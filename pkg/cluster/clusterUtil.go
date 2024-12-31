package cluster

import "fmt"

const (
	SECRET_NAME = "cluster-event"
)

func ParseSecretNameForKubelinkInformer(clusterId int) string {
	return fmt.Sprintf("%s-%d", SECRET_NAME, clusterId)
}
