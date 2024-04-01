package bean

type RemoteConnectionMethod string

const (
	RemoteConnectionMethodProxy RemoteConnectionMethod = "PROXY"
	RemoteConnectionMethodSSH   RemoteConnectionMethod = "SSH"
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
