package adapter

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
)

func ConvertClusterBeanToClusterConfig(clusterBean *bean.ClusterBean) *gRPC.ClusterConfig {
	config := &gRPC.ClusterConfig{
		ApiServerUrl:          clusterBean.ServerUrl,
		Token:                 clusterBean.Config[commonBean.BearerToken],
		ClusterId:             int32(clusterBean.Id),
		ClusterName:           clusterBean.ClusterName,
		InsecureSkipTLSVerify: clusterBean.InsecureSkipTLSVerify,
	}

	if clusterBean.InsecureSkipTLSVerify == false {
		config.KeyData = clusterBean.Config[commonBean.TlsKey]
		config.CertData = clusterBean.Config[commonBean.CertData]
		config.CaData = clusterBean.Config[commonBean.CertificateAuthorityData]
	}
	return config
}
