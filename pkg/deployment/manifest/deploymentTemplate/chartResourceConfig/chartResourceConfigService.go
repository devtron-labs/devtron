package chartResourceConfig

import "github.com/devtron-labs/devtron/pkg/resourceQualifiers"

func GetStopTemplate(scope *resourceQualifiers.Scope) (string, error) {
	stopTemplate := `{"replicaCount":0,"autoscaling":{"MinReplicas":0,"MaxReplicas":0 ,"enabled": false},"kedaAutoscaling":{"minReplicaCount":0,"maxReplicaCount":0 ,"enabled": false}}`
	return stopTemplate, nil
}
