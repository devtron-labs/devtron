package cluster

import "fmt"

const (
	SecretName = "cluster-event"
)

func ParseSecretNameForKubelinkInformer(clusterId int) string {
	return fmt.Sprintf("%s-%d", SecretName, clusterId)
}
