package bean

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
)

type PrometheusAuth struct {
	UserName      string `json:"userName,omitempty"`
	Password      string `json:"password,omitempty"`
	TlsClientCert string `json:"tlsClientCert,omitempty"`
	TlsClientKey  string `json:"tlsClientKey,omitempty"`
	IsAnonymous   bool   `json:"isAnonymous"`
}

type ClusterBean struct {
	Id                      int                        `json:"id" validate:"number"`
	ClusterName             string                     `json:"cluster_name,omitempty" validate:"required"`
	Description             string                     `json:"description"`
	ServerUrl               string                     `json:"server_url,omitempty" validate:"url,required"`
	PrometheusUrl           string                     `json:"prometheus_url,omitempty" validate:"validate-non-empty-url"`
	Active                  bool                       `json:"active"`
	Config                  map[string]string          `json:"config,omitempty"`
	PrometheusAuth          *PrometheusAuth            `json:"prometheusAuth,omitempty"`
	DefaultClusterComponent []*DefaultClusterComponent `json:"defaultClusterComponent"`
	AgentInstallationStage  int                        `json:"agentInstallationStage,notnull"` // -1=external, 0=not triggered, 1=progressing, 2=success, 3=fails
	K8sVersion              string                     `json:"k8sVersion"`
	HasConfigOrUrlChanged   bool                       `json:"-"`
	UserName                string                     `json:"userName,omitempty"`
	InsecureSkipTLSVerify   bool                       `json:"insecureSkipTlsVerify"`
	ErrorInConnecting       string                     `json:"errorInConnecting"`
	IsCdArgoSetup           bool                       `json:"isCdArgoSetup"`
	IsVirtualCluster        bool                       `json:"isVirtualCluster"`
	ClusterUpdated          bool                       `json:"clusterUpdated"`
}

func (bean ClusterBean) GetClusterConfig() *k8s.ClusterConfig {
	host := bean.ServerUrl
	configMap := bean.Config
	bearerToken := configMap[commonBean.BearerToken]
	clusterCfg := &k8s.ClusterConfig{Host: host, BearerToken: bearerToken}
	clusterCfg.InsecureSkipTLSVerify = bean.InsecureSkipTLSVerify
	if bean.InsecureSkipTLSVerify == false {
		clusterCfg.KeyData = configMap[commonBean.TlsKey]
		clusterCfg.CertData = configMap[commonBean.CertData]
		clusterCfg.CAData = configMap[commonBean.CertificateAuthorityData]
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
