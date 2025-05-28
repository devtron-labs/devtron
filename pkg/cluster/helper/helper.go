package helper

import (
	"fmt"
	informerBean "github.com/devtron-labs/common-lib/informer"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"time"
)

func CreateClusterModifyEventData(clusterId int, action string) (map[string]string, map[string]string) {
	data := make(map[string]string)
	data[informerBean.CmFieldClusterId] = fmt.Sprintf("%v", clusterId)
	data[informerBean.CmFieldAction] = action
	data[clusterBean.CmFieldUpdatedOn] = time.Now().String()

	labels := make(map[string]string)
	labels[informerBean.ClusterModifyEventSecretTypeKey] = informerBean.ClusterModifyEventCmLabelValue

	return data, labels
}
