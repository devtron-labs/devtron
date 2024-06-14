/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

type RemoteConnectionMethod string

const (
	RemoteConnectionMethodProxy  RemoteConnectionMethod = "PROXY"
	RemoteConnectionMethodSSH    RemoteConnectionMethod = "SSH"
	RemoteConnectionMethodDirect RemoteConnectionMethod = "DIRECT"
)

type ConnectionMethod string

const (
	ConnectionMethod_Proxy  ConnectionMethod = "PROXY"
	ConnectionMethod_SSH    ConnectionMethod = "SSH"
	ConnectionMethod_DIRECT ConnectionMethod = "DIRECT"
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
