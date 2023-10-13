package sshTunnel

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

// AuthType is the type of authentication to use for SSH.
type AuthType int

const (
	// AuthTypeKey uses the keys from a SSH key file read from the system.
	AuthTypeKey AuthType = iota
	// AuthTypePassword uses a password directly.
	AuthTypePassword
)

func (tunnel *SSHTunnel) getSSHAuthMethods() ([]ssh.AuthMethod, error) {
	authMethods := make([]ssh.AuthMethod, 0, len(tunnel.authTypes))
	for _, authType := range tunnel.authTypes {
		switch authType {
		case AuthTypeKey:
			authMethodForKey, err := tunnel.getSSHAuthMethodForKey()
			if err != nil {
				log.Printf("error in getting SSH auth method for key : %v", err)
				return nil, err
			}
			authMethods = append(authMethods, authMethodForKey)
		case AuthTypePassword:
			if len(tunnel.authPassword) == 0 {
				return nil, fmt.Errorf("error in adding auth method for password, invalid password")
			}
			authMethodForPassword := ssh.Password(tunnel.authPassword)
			authMethods = append(authMethods, authMethodForPassword)
		default:
			return nil, fmt.Errorf("unknown auth type: %d", authType)
		}
	}
	return authMethods, nil
}

func (tunnel *SSHTunnel) getSSHAuthMethodForKey() (ssh.AuthMethod, error) {
	var key ssh.Signer
	var err error
	key, err = ssh.ParsePrivateKey([]byte(tunnel.authKey))
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}
	return ssh.PublicKeys(key), nil
}
