package registry

type ConnectionMethod string

const (
	ConnectionMethod_Proxy ConnectionMethod = "PROXY"
	ConnectionMethod_SSH   ConnectionMethod = "SSH"
)

type ProxyConfig struct {
	ProxyUrl string
}

type SSHTunnelConfig struct {
	SSHServerAddress string
	SSHUsername      string
	SSHPassword      string
	SSHAuthKey       string
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
