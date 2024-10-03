package helper

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"strconv"
	"strings"
)

func DecodeExternalArgoAppId(appId string) (*bean.ArgoAppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 3 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &bean.ArgoAppIdentifier{
		ClusterId: clusterId,
		Namespace: component[1],
		AppName:   component[2],
	}, nil
}

func ConvertClusterBeanToGrpcConfig(cluster repository.Cluster) *gRPC.ClusterConfig {
	config := &gRPC.ClusterConfig{
		ApiServerUrl:          cluster.ServerUrl,
		Token:                 cluster.Config[commonBean.BearerToken],
		ClusterId:             int32(cluster.Id),
		ClusterName:           cluster.ClusterName,
		InsecureSkipTLSVerify: cluster.InsecureSkipTlsVerify,
	}
	if cluster.InsecureSkipTlsVerify == false {
		config.KeyData = cluster.Config[commonBean.TlsKey]
		config.CertData = cluster.Config[commonBean.CertData]
		config.CaData = cluster.Config[commonBean.CertificateAuthorityData]
	}
	return config

}
