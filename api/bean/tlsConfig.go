package bean

type TLSConfig struct {
	CaData      string `json:"caData"`
	TLSCertData string `json:"tlsCertData"`
	TLSKeyData  string `json:"tlsKeyData"`
}
