/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package sql

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// SSL modes supported for postgres connections, mirroring libpq / AWS RDS semantics.
const (
	SslModeDisable    = "disable"     // no TLS (default, preserves legacy behaviour)
	SslModeRequire    = "require"     // encrypt only, no server certificate verification
	SslModeVerifyCA   = "verify-ca"   // encrypt + verify the server cert chains to a trusted CA
	SslModeVerifyFull = "verify-full" // encrypt + verify CA + server hostname
)

// BuildTLSConfig returns a *tls.Config for the given sslMode, or nil for "disable"/empty
// (in which case go-pg connects in plaintext, i.e. the pre-existing behaviour).
//
// rootCertPath is the path to a PEM CA bundle (for AWS RDS this is the downloadable
// global-bundle.pem) and is required for verify-ca and verify-full.
//
// host is the server host used as the expected certificate hostname (ServerName) for
// verify-full; for RDS this must be the real DB endpoint (e.g. *.rds.amazonaws.com).
func BuildTLSConfig(sslMode, rootCertPath, host string) (*tls.Config, error) {
	switch strings.ToLower(strings.TrimSpace(sslMode)) {
	case "", SslModeDisable:
		return nil, nil
	case SslModeRequire:
		// Encrypt the connection but do not validate the server certificate.
		return &tls.Config{InsecureSkipVerify: true}, nil
	case SslModeVerifyCA, SslModeVerifyFull:
		rootCAs, err := loadRootCAs(rootCertPath)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{RootCAs: rootCAs}
		if strings.EqualFold(strings.TrimSpace(sslMode), SslModeVerifyFull) {
			// verify-full: full verification including the server hostname.
			tlsConfig.ServerName = hostWithoutPort(host)
		} else {
			// verify-ca: validate the certificate chain but not the hostname.
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.VerifyPeerCertificate = verifyChainOnly(rootCAs)
		}
		return tlsConfig, nil
	default:
		return nil, fmt.Errorf("pg: unsupported PG_SSL_MODE %q (supported: disable, require, verify-ca, verify-full)", sslMode)
	}
}

func loadRootCAs(rootCertPath string) (*x509.CertPool, error) {
	if strings.TrimSpace(rootCertPath) == "" {
		return nil, errors.New("pg: PG_SSL_ROOT_CERT is required for verify-ca/verify-full ssl modes")
	}
	pemData, err := os.ReadFile(rootCertPath)
	if err != nil {
		return nil, fmt.Errorf("pg: failed to read ssl root cert %q: %w", rootCertPath, err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("pg: no certificates found in ssl root cert %q", rootCertPath)
	}
	return rootCAs, nil
}

// verifyChainOnly verifies that the presented certificate chains to a trusted CA
// without checking the server hostname (verify-ca semantics). This mirrors what
// go's default verifier does, minus the DNS name check.
func verifyChainOnly(rootCAs *x509.CertPool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		certs := make([]*x509.Certificate, 0, len(rawCerts))
		for _, raw := range rawCerts {
			cert, err := x509.ParseCertificate(raw)
			if err != nil {
				return err
			}
			certs = append(certs, cert)
		}
		if len(certs) == 0 {
			return errors.New("pg: no server certificate presented")
		}
		intermediates := x509.NewCertPool()
		for _, cert := range certs[1:] {
			intermediates.AddCert(cert)
		}
		_, err := certs[0].Verify(x509.VerifyOptions{
			Roots:         rootCAs,
			Intermediates: intermediates,
		})
		return err
	}
}

// hostWithoutPort strips a trailing ":port" if present so the value can be used as
// a TLS ServerName.
func hostWithoutPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
