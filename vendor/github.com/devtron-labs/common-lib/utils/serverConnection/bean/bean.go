package bean

type ServerConnectionMethod string

const (
	ServerConnectionMethodProxy ServerConnectionMethod = "PROXY"
	ServerConnectionMethodSSH   ServerConnectionMethod = "SSH"
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
	SSHServerAddress string `json:"SSHServerAddress,omitempty"`
	SSHUsername      string `json:"SSHUsername,omitempty"`
	SSHPassword      string `json:"SSHPassword,omitempty"`
	SSHAuthKey       string `json:"SSHAuthKey,omitempty"`
}

type ServerConnectionConfigBean struct {
	ServerConnectionConfigId int                    `json:"serverConnectionConfigId"`
	ConnectionMethod         ServerConnectionMethod `json:"connectionMethod,omitempty"`
	ProxyConfig              *ProxyConfig           `json:"proxyConfig,omitempty"`
	SSHTunnelConfig          *SSHTunnelConfig       `json:"SSHConfig,omitempty"`
}

type RegistryConfig struct {
	RegistryId       string
	RegistryUrl      string
	RegistryUsername string
	RegistryPassword string
	ConnectionMethod ConnectionMethod
	ProxyConfig      ProxyConfig
	SSHConfig        SSHTunnelConfig
}
