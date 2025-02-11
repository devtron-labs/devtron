package common

import (
	fmt "fmt"
)

// ValidateTLSConfig validates a TLS configuration.
func ValidateTLSConfig(tlsConfig *TLSConfig) error {
	if tlsConfig == nil {
		return nil
	}

	if tlsConfig.InsecureSkipVerify {
		return nil
	}

	var caCertSet, clientCertSet, clientKeySet bool

	if tlsConfig.CACertSecret != nil {
		caCertSet = true
	}

	if tlsConfig.ClientCertSecret != nil {
		clientCertSet = true
	}

	if tlsConfig.ClientKeySecret != nil {
		clientKeySet = true
	}

	if !caCertSet && !clientCertSet && !clientKeySet {
		return fmt.Errorf("invalid tls config, please configure either caCertSecret, or clientCertSecret and clientKeySecret, or both")
	}

	if (clientCertSet || clientKeySet) && (!clientCertSet || !clientKeySet) {
		return fmt.Errorf("invalid tls config, both clientCertSecret and clientKeySecret need to be configured")
	}
	return nil
}

func ValidateBasicAuth(auth *BasicAuth) error {
	if auth == nil {
		return nil
	}
	if auth.Username == nil {
		return fmt.Errorf("username missing")
	}
	if auth.Password == nil {
		return fmt.Errorf("password missing")
	}
	return nil
}

func ValidateSASLConfig(saslConfig *SASLConfig) error {
	if saslConfig == nil {
		return nil
	}

	switch saslConfig.Mechanism {
	case "", "PLAIN", "OAUTHBEARER", "SCRAM-SHA-256", "SCRAM-SHA-512", "GSSAPI":
	default:
		return fmt.Errorf("invalid sasl config. Possible values for SASL Mechanism are `OAUTHBEARER`, `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512` and `GSSAPI`")
	}

	// user and password must both be set
	if saslConfig.UserSecret == nil || saslConfig.PasswordSecret == nil {
		return fmt.Errorf("invalid sasl config, both userSecret and passwordSecret must be defined")
	}

	return nil
}
