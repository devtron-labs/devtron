package connection

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	commonk8s "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	config2 "github.com/devtron-labs/devtron/client/argocdServer/config"
	"github.com/devtron-labs/devtron/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type externalHealthServer struct {
	healthpb.UnimplementedHealthServer
}

func (externalHealthServer) Check(context.Context, *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

type externalConfigGetter struct {
	config2.ArgoCDConfigGetter
	config *bean.ArgoGRPCConfig
}

func (g externalConfigGetter) GetGRPCConfig() (*bean.ArgoGRPCConfig, error) {
	return g.config, nil
}

func externalManager(config *bean.Config) (*ArgoCDConnectionManagerImpl, error) {
	return NewArgoCDConnectionManagerImpl(zap.NewNop().Sugar(), nil, nil,
		&util.EnvironmentVariables{}, nil, nil, nil, nil,
		&commonk8s.RuntimeConfig{LocalDevMode: true},
		externalConfigGetter{config: &bean.ArgoGRPCConfig{ConnectionConfig: config}}, nil)
}

func externalCertificate(t *testing.T, ip string) (tls.Certificate, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Minute),
		NotAfter: time.Now().Add(time.Hour), IPAddresses: []net.IP{net.ParseIP(ip)},
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	certificate, err := tls.X509KeyPair(certificatePEM, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
	require.NoError(t, err)
	return certificate, certificatePEM
}

func TestExternalArgoTLSAndToken(t *testing.T) {
	for _, test := range []struct {
		name, certificateIP, trust, token string
		accept                            bool
	}{
		{"trusted CA", "127.0.0.1", "server", "fixture-first", true},
		{"untrusted CA", "127.0.0.1", "other", "fixture-first", false},
		{"system roots reject private CA", "127.0.0.1", "system", "fixture-first", false},
		{"hostname mismatch", "127.0.0.2", "server", "fixture-first", false},
		{"missing token", "127.0.0.1", "server", "missing", false},
		{"empty token", "127.0.0.1", "server", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			certificate, ca := externalCertificate(t, test.certificateIP)
			var calls atomic.Int64
			var expectedToken atomic.Value
			expectedToken.Store("fixture-first")
			server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				calls.Add(1)
				incoming, _ := metadata.FromIncomingContext(ctx)
				if values := incoming.Get("token"); len(values) != 1 || values[0] != expectedToken.Load().(string) {
					return nil, status.Error(codes.Unauthenticated, "invalid fixture token")
				}
				return handler(ctx, request)
			}))
			healthpb.RegisterHealthServer(server, externalHealthServer{})
			httpServer := httptest.NewUnstartedServer(server)
			httpServer.EnableHTTP2 = true
			httpServer.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}}
			httpServer.StartTLS()
			t.Cleanup(httpServer.Close)
			t.Cleanup(server.Stop)
			host, port, err := net.SplitHostPort(strings.TrimPrefix(httpServer.URL, "https://"))
			require.NoError(t, err)
			config := &bean.Config{Host: host, Port: port, TokenFile: filepath.Join(t.TempDir(), "token")}
			if test.token != "missing" {
				require.NoError(t, os.WriteFile(config.TokenFile, []byte(test.token+"\n"), 0600))
			}
			if test.trust != "system" {
				if test.trust == "other" {
					_, ca = externalCertificate(t, "127.0.0.1")
				}
				config.CAFile = filepath.Join(t.TempDir(), "ca.crt")
				require.NoError(t, os.WriteFile(config.CAFile, ca, 0600))
			}
			manager, err := externalManager(config)
			require.NoError(t, err)
			grpcConfig := &bean.ArgoGRPCConfig{ConnectionConfig: config}
			// Nil bootstrap dependencies deliberately catch attempts to access Argo settings, accounts or Kubernetes.
			require.Empty(t, manager.ValidateGitOpsAndGetOrUpdateArgoCdUserDetail(grpcConfig))
			require.Empty(t, manager.GetOrUpdateArgoCdUserDetail(grpcConfig))
			connection := manager.GetGrpcClientConnection(grpcConfig)
			require.NotNil(t, connection)
			t.Cleanup(func() { _ = connection.Close() })
			check := func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_, err := healthpb.NewHealthClient(connection).Check(ctx, &healthpb.HealthCheckRequest{})
				return err
			}
			if !test.accept {
				require.Error(t, check())
				require.Zero(t, calls.Load())
				return
			}
			require.NoError(t, check())
			expectedToken.Store("fixture-rotated")
			require.NoError(t, os.WriteFile(config.TokenFile, []byte("fixture-rotated\n"), 0600))
			require.NoError(t, check(), "token rotation must work on the existing connection")
			require.EqualValues(t, 2, calls.Load())
		})
	}
}

func TestExternalArgoInvalidCAFailsInitialization(t *testing.T) {
	for _, missing := range []bool{false, true} {
		file := filepath.Join(t.TempDir(), "ca.crt")
		if !missing {
			require.NoError(t, os.WriteFile(file, []byte("invalid PEM"), 0600))
		}
		manager, err := externalManager(&bean.Config{TokenFile: "mounted-token", CAFile: file})
		require.Error(t, err)
		require.Nil(t, manager)
	}
}

func TestExternalTokenRequiresTLS(t *testing.T) {
	require.True(t, TokenAuth{tokenFile: "mounted-token"}.RequireTransportSecurity())
	require.False(t, TokenAuth{token: "legacy-token"}.RequireTransportSecurity())
	metadata, err := (TokenAuth{token: "legacy-token"}).GetRequestMetadata(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string]string{"token": "legacy-token"}, metadata)
}
