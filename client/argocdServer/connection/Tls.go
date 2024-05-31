/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package connection

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"google.golang.org/grpc/credentials"
	"os"
)

func GetTLS(cert *tls.Certificate) credentials.TransportCredentials {
	//These certificates are to be read from secret create by argocd
	//cert, err := tls.X509KeyPair([]byte(TLSCert), []byte(TLSKey))
	//if err != nil {
	//}
	certPool := x509.NewCertPool()
	pemCertBytes, _ := EncodeX509KeyPair(*cert)
	ok := certPool.AppendCertsFromPEM(pemCertBytes)
	if !ok {
		panic("bad certs")
	}
	/* #nosec */
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}
	dCreds := credentials.NewTLS(tlsConfig)
	return dCreds
}

// EncodeX509KeyPair encodes a TLS Certificate into its pem encoded format for storage
func EncodeX509KeyPair(cert tls.Certificate) ([]byte, []byte) {

	certpem := []byte{}
	for _, certtmp := range cert.Certificate {
		certpem = append(certpem, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certtmp})...)
	}
	keypem := pem.EncodeToMemory(pemBlockForKey(cert.PrivateKey))
	return certpem, keypem
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
