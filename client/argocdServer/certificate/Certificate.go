package certificate

import (
	"context"
	"errors"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type Client interface {
	ListCertificates(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error)
	CreateCertificate(ctx context.Context, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error)
	DeleteCertificate(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error)
}

type ServiceClientImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewServiceClientImpl(
	logger *zap.SugaredLogger,
	argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (c *ServiceClientImpl) getService(ctx context.Context) (certificate.CertificateServiceClient, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	//defer conn.Close()
	return certificate.NewCertificateServiceClient(conn), nil
}

func (c *ServiceClientImpl) ListCertificates(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := c.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListCertificates(ctx, query)
}

func (c *ServiceClientImpl) CreateCertificate(ctx context.Context, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := c.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.CreateCertificate(ctx, query)
}

func (c *ServiceClientImpl) DeleteCertificate(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := c.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.DeleteCertificate(ctx, query, opts...)
}
