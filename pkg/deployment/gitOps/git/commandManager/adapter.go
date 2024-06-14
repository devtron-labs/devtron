package commandManager

func BuildTlsInfoPath(caCertPath, tlsKeyPath, tlsCertPath string) *TlsPathInfo {
	return &TlsPathInfo{
		CaCertPath:  caCertPath,
		TlsKeyPath:  tlsKeyPath,
		TlsCertPath: tlsCertPath,
	}
}
