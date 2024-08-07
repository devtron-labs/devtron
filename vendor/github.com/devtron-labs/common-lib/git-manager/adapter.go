package git_manager

func BuildTlsData(tlsKey, tlsCert, caCert string, tlsVerificationEnabled bool) *TLSData {
	return &TLSData{
		TLSKey:                 tlsKey,
		TLSCertificate:         tlsCert,
		CACert:                 caCert,
		TlsVerificationEnabled: tlsVerificationEnabled,
	}
}
