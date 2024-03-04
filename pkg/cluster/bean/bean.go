package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/common-lib/utils/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
)

type ClusterBean struct {
	Id                      int                              `json:"id" validate:"number"`
	ClusterName             string                           `json:"cluster_name,omitempty" validate:"required"`
	Description             string                           `json:"description"`
	ServerUrl               string                           `json:"server_url,omitempty" validate:"url,required"`
	PrometheusUrl           string                           `json:"prometheus_url,omitempty" validate:"validate-non-empty-url"`
	Active                  bool                             `json:"active"`
	ProxyUrl                string                           `json:"proxyUrl,omitempty"`
	Config                  map[string]string                `json:"config,omitempty"`
	PrometheusAuth          *PrometheusAuth                  `json:"prometheusAuth,omitempty"`
	DefaultClusterComponent []*DefaultClusterComponent       `json:"defaultClusterComponent"`
	AgentInstallationStage  int                              `json:"agentInstallationStage,notnull"` // -1=external, 0=not triggered, 1=progressing, 2=success, 3=fails
	K8sVersion              string                           `json:"k8sVersion"`
	HasConfigOrUrlChanged   bool                             `json:"-"`
	UserName                string                           `json:"userName,omitempty"`
	InsecureSkipTLSVerify   bool                             `json:"insecureSkipTlsVerify"`
	ErrorInConnecting       string                           `json:"errorInConnecting"`
	IsCdArgoSetup           bool                             `json:"isCdArgoSetup"`
	IsVirtualCluster        bool                             `json:"isVirtualCluster"`
	isClusterNameEmpty      bool                             `json:"-"`
	ClusterUpdated          bool                             `json:"clusterUpdated"`
	ToConnectWithSSHTunnel  bool                             `json:"toConnectWithSSHTunnel,omitempty"`
	SSHTunnelConfig         *SSHTunnelConfig                 `json:"sshTunnelConfig,omitempty"`
	ServerConnectionConfig  *bean.ServerConnectionConfigBean `json:"-"`
}

type VirtualClusterBean struct {
	Id               int    `json:"id,omitempty" validate:"number"`
	ClusterName      string `json:"clusterName,omitempty" validate:"required"`
	Active           bool   `json:"active"`
	IsVirtualCluster bool   `json:"isVirtualCluster" default:"true"`
}

type PrometheusAuth struct {
	UserName      string `json:"userName,omitempty"`
	Password      string `json:"password,omitempty"`
	TlsClientCert string `json:"tlsClientCert,omitempty"`
	TlsClientKey  string `json:"tlsClientKey,omitempty"`
	IsAnonymous   bool   `json:"isAnonymous"`
}

type SSHTunnelConfig struct {
	User             string `json:"user"`
	Password         string `json:"password"`
	AuthKey          string `json:"authKey"`
	SSHServerAddress string `json:"sshServerAddress"`
}

func (bean ClusterBean) GetClusterConfig() *k8s.ClusterConfig {
	//bean = *adapter.ConvertClusterBeanToNewClusterBean(&bean)
	configMap := bean.Config
	bearerToken := configMap[k8s.BearerToken]
	clusterCfg := &k8s.ClusterConfig{
		ClusterId:             bean.Id,
		ClusterName:           bean.ClusterName,
		Host:                  bean.ServerUrl,
		BearerToken:           bearerToken,
		InsecureSkipTLSVerify: bean.InsecureSkipTLSVerify,
	}
	if bean.InsecureSkipTLSVerify == false {
		clusterCfg.KeyData = configMap[k8s.TlsKey]
		clusterCfg.CertData = configMap[k8s.CertData]
		clusterCfg.CAData = configMap[k8s.CertificateAuthorityData]
	}
	if bean.ServerConnectionConfig != nil {
		clusterCfg.ServerConnectionConfig = &bean2.ServerConnectionConfigBean{
			ServerConnectionConfigId: bean.ServerConnectionConfig.ServerConnectionConfigId,
			ConnectionMethod:         bean2.ServerConnectionMethod(bean.ServerConnectionConfig.ConnectionMethod),
			ProxyConfig:              (*bean2.ProxyConfig)(bean.ServerConnectionConfig.ProxyConfig),
			SSHTunnelConfig:          (*bean2.SSHTunnelConfig)(bean.ServerConnectionConfig.SSHTunnelConfig),
		}
	}
	return clusterCfg
}

type UserInfo struct {
	UserName          string            `json:"userName,omitempty"`
	Config            map[string]string `json:"config,omitempty"`
	ErrorInConnecting string            `json:"errorInConnecting"`
}

type ValidateClusterBean struct {
	UserInfos map[string]*UserInfo `json:"userInfos,omitempty""`
	*ClusterBean
}

type UserClusterBeanMapping struct {
	Mapping map[string]*ClusterBean `json:"mapping"`
}

type Kubeconfig struct {
	Config string `json:"config"`
}

type DefaultClusterComponent struct {
	ComponentName  string `json:"name"`
	AppId          int    `json:"appId"`
	InstalledAppId int    `json:"installedAppId,omitempty"`
	EnvId          int    `json:"envId"`
	EnvName        string `json:"envName"`
	Status         string `json:"status"`
}
