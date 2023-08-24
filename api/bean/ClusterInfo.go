package bean

type ClusterInfo struct {
	ClusterId             int    `json:"clusterId"`
	ClusterName           string `json:"clusterName"`
	BearerToken           string `json:"bearerToken"`
	ServerUrl             string `json:"serverUrl"`
	InsecureSkipTLSVerify bool   `json:"insecureSkipTLSVerify"`
	KeyData               string `json:"keyData"`
	CertData              string `json:"certData"`
	CAData                string `json:"CAData"`
	ProxyUrl              string `json:"proxyUrl"`
}
