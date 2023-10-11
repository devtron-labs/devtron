package sshTunnel

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib-private/sshTunnel/bean"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// SSHTunnel represents a SSH tunnel
type SSHTunnel struct {
	mutex        *sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	started      bool
	username     string
	authTypes    []AuthType
	authKey      string
	authPassword string
	server       *Endpoint
	local        *Endpoint
	remote       *Endpoint
	Timeout      time.Duration
	Active       int
	sshClient    *ssh.Client
	sshConfig    *ssh.ClientConfig
}

type Endpoint struct {
	host string
	port int
}

func NewTCPEndpoint(host string, port int) *Endpoint {
	return &Endpoint{
		host: host,
		port: port,
	}
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

// NewSSHTunnel creates a new SSH tunnel to the specified server redirecting a port on local localhost to a port on remote host.
// By default, the SSH connection is made to port 22 as root
func NewSSHTunnel(username, password, authKey, server, remoteAddress string,
	remotePort, localPort int, timeout time.Duration) *SSHTunnel {
	tunnel := &SSHTunnel{
		mutex:   &sync.Mutex{},
		server:  NewTCPEndpoint(server, bean.SSHPort),
		local:   NewTCPEndpoint(bean.LocalHostAddress, localPort),
		remote:  NewTCPEndpoint(remoteAddress, remotePort),
		Timeout: timeout,
	}
	tunnel.SetUser(username)
	tunnel.SetAuthData(password, authKey)
	return tunnel
}

// SetUser updates the username used to make the SSH connection
func (tunnel *SSHTunnel) SetUser(user string) {
	if len(user) == 0 {
		user = bean.RootUser
	}
	tunnel.username = user
}

func (tunnel *SSHTunnel) SetAuthData(password, authKey string) {
	if len(password) != 0 {
		tunnel.SetAuthTypePassword(password)
	}
	if len(authKey) != 0 {
		tunnel.SetAuthKey(authKey)
	}
}

// SetAuthKey changes the authentication to key-based and uses the specified key data.
func (tunnel *SSHTunnel) SetAuthKey(key string) {
	tunnel.authTypes = append(tunnel.authTypes, AuthTypeKey)
	tunnel.authKey = key
}

// SetAuthTypePassword changes the authentication to password-based and uses the specified password.
func (tunnel *SSHTunnel) SetAuthTypePassword(password string) {
	tunnel.authTypes = append(tunnel.authTypes, AuthTypePassword)
	tunnel.authPassword = password
}

// Start starts the SSH tunnel. It can be stopped by calling `Stop` or cancelling its context.
// This call will block until the tunnel is stopped either calling those methods or by an error.
func (tunnel *SSHTunnel) Start(ctx context.Context) error {
	tunnel.mutex.Lock()
	if tunnel.started {
		tunnel.mutex.Unlock()
		return fmt.Errorf("tunnel is already started")
	} else {
		tunnel.started = true
		tunnel.ctx, tunnel.cancel = context.WithCancel(ctx)
		tunnel.mutex.Unlock()
	}

	log.Printf("SSH tunnel is starting : %v \n", tunnel)
	config, err := tunnel.InitSSHConfig()
	if err != nil {
		return tunnel.stop(fmt.Errorf("ssh config failed: %w", err))
	}
	tunnel.sshConfig = config

	listenConfig := net.ListenConfig{}
	localListener, err := listenConfig.Listen(tunnel.ctx, bean.EndpointTypeTCP, tunnel.local.String())
	if err != nil {
		return tunnel.stop(fmt.Errorf("local listen %s on %s failed: %w", bean.EndpointTypeTCP, tunnel.local.String(), err))
	}

	errChan := make(chan error)
	go func() {
		errChan <- tunnel.listen(localListener)
	}()

	return tunnel.stop(<-errChan)
}

// Stop closes all connections and makes Start exit gracefully.
func (tunnel *SSHTunnel) Stop() {
	tunnel.mutex.Lock()
	defer tunnel.mutex.Unlock()
	if tunnel.started {
		tunnel.cancel()
	}
}

func (tunnel *SSHTunnel) InitSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: tunnel.username,
		//we are ignoring hostKeyCallback currently, mentioned here for easy discoverability in future
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: tunnel.Timeout,
	}
	authMethods, err := tunnel.getSSHAuthMethods()
	if err != nil {
		return nil, err
	}
	config.Auth = authMethods
	return config, nil
}

func (tunnel *SSHTunnel) stop(err error) error {
	tunnel.mutex.Lock()
	tunnel.started = false
	tunnel.mutex.Unlock()
	return err
}

func (tunnel *SSHTunnel) listen(localListener net.Listener) error {
	errGroup, groupCtx := errgroup.WithContext(tunnel.ctx)

	errGroup.Go(func() error {
		for {
			localConn, err := localListener.Accept()
			if err != nil {
				return fmt.Errorf("local accept %s on %s failed: %w", bean.EndpointTypeTCP, tunnel.local.String(), err)
			}

			errGroup.Go(func() error {
				return tunnel.handle(localConn)
			})
		}
	})

	<-groupCtx.Done()

	localListener.Close()

	err := errGroup.Wait()

	select {
	case <-tunnel.ctx.Done():
	default:
		return err
	}

	return nil
}

func (tunnel *SSHTunnel) handle(localConn net.Conn) error {
	err := tunnel.AddConnection()
	if err != nil {
		return err
	}

	tunnel.forward(localConn)
	tunnel.CloseConnection()

	return nil
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	from := localConn.RemoteAddr().String()
	remoteConn, err := tunnel.sshClient.Dial(bean.EndpointTypeTCP, tunnel.remote.String())
	if err != nil {
		localConn.Close()
		return
	}

	connStr := fmt.Sprintf("%s -(%s)> %s -(ssh)> %s -(%s)> %s", from, bean.EndpointTypeTCP, tunnel.local.String(),
		tunnel.server.String(), bean.EndpointTypeTCP, tunnel.remote.String())

	connCtx, connCancel := context.WithCancel(tunnel.ctx)
	errGroup := &errgroup.Group{}

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(remoteConn, localConn)
		if err != nil {
			return fmt.Errorf("failed copying bytes from remote to local: %w", err)
		}

		return nil
	})

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(localConn, remoteConn)
		if err != nil {
			return fmt.Errorf("failed copying bytes from local to remote: %w", err)
		}

		return nil
	})

	<-connCtx.Done()

	localConn.Close()
	remoteConn.Close()

	err = errGroup.Wait()

	select {
	case <-tunnel.ctx.Done():
	default:
		if err != nil {
			log.Printf("error encountered in waiting for errGroup closing : %v", err)
		}
	}
	log.Printf("connection closed : %s", connStr)
}

func (tunnel *SSHTunnel) AddConnection() error {
	tunnel.mutex.Lock()
	defer tunnel.mutex.Unlock()

	if tunnel.Active == 0 {
		sshClient, err := ssh.Dial(bean.EndpointTypeTCP, tunnel.server.String(), tunnel.sshConfig)
		if err != nil {
			return fmt.Errorf("ssh dial %s to %s failed: %w", bean.EndpointTypeTCP, tunnel.server.String(), err)
		}
		tunnel.sshClient = sshClient
	}

	tunnel.Active += 1

	return nil
}

func (tunnel *SSHTunnel) CloseConnection() {
	tunnel.mutex.Lock()
	defer tunnel.mutex.Unlock()

	tunnel.Active -= 1

	if tunnel.Active == 0 {
		tunnel.sshClient.Close()
		tunnel.sshClient = nil
	}
}
