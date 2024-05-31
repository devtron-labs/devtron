package bean

type RemoteConnectionMethod string

const (
	RemoteConnectionMethodProxy  RemoteConnectionMethod = "PROXY"
	RemoteConnectionMethodSSH    RemoteConnectionMethod = "SSH"
	RemoteConnectionMethodDirect RemoteConnectionMethod = "DIRECT"
)

type ConnectionMethod string

const (
	ConnectionMethod_Proxy ConnectionMethod = "PROXY"
	ConnectionMethod_SSH   ConnectionMethod = "SSH"
)

type ProxyConfig struct {
	ProxyUrl string `json:"proxyUrl,omitempty"`
}

type SSHTunnelConfig struct {
	SSHServerAddress string `json:"sshServerAddress,omitempty"`
	SSHUsername      string `json:"sshUsername,omitempty"`
	SSHPassword      string `json:"sshPassword,omitempty"`
	SSHAuthKey       string `json:"sshAuthKey,omitempty"`
}

type RemoteConnectionConfigBean struct {
	RemoteConnectionConfigId int                    `json:"remoteConnectionConfigId"`
	ConnectionMethod         RemoteConnectionMethod `json:"connectionMethod,omitempty"`
	ProxyConfig              *ProxyConfig           `json:"proxyConfig,omitempty"`
	SSHTunnelConfig          *SSHTunnelConfig       `json:"sshConfig,omitempty"`
}

type RegistryConfig struct {
	RegistryId                string
	RegistryUrl               string
	RegistryUsername          string
	RegistryPassword          string
	RegistryConnectionType    string //secure, insecure, secure-with-cert
	RegistryCertificateString string
	RegistryCAFilePath        string
	IsPublicRegistry          bool
	ConnectionMethod          ConnectionMethod //ssh, proxy
	ProxyConfig               *ProxyConfig
	SSHConfig                 *SSHTunnelConfig
}
