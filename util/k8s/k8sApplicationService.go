package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

const DEFAULT_CLUSTER = "default_cluster"

type K8sApplicationService interface {
	GetResource(request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	CreateResource(request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	UpdateResource(request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	DeleteResource(request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	ListEvents(request *ResourceRequestBean) (*application.EventsResponse, error)
	GetPodLogs(request *ResourceRequestBean) (io.ReadCloser, error)
	ValidateResourceRequest(appIdentifier *client.AppIdentifier, request *application.K8sRequestBean) (bool, error)
	GetResourceInfo() (*ResourceInfo, error)
	GetRestConfigByClusterId(clusterId int) (*rest.Config, error)
}
type K8sApplicationServiceImpl struct {
	logger           *zap.SugaredLogger
	clusterService   cluster.ClusterService
	pump             connector.Pump
	k8sClientService application.K8sClientService
	helmAppService   client.HelmAppService
	K8sUtil          *util.K8sUtil
	aCDAuthConfig    *util3.ACDAuthConfig
}

func NewK8sApplicationServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	pump connector.Pump, k8sClientService application.K8sClientService,
	helmAppService client.HelmAppService, K8sUtil *util.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig) *K8sApplicationServiceImpl {
	return &K8sApplicationServiceImpl{
		logger:           Logger,
		clusterService:   clusterService,
		pump:             pump,
		k8sClientService: k8sClientService,
		helmAppService:   helmAppService,
		K8sUtil:          K8sUtil,
		aCDAuthConfig:    aCDAuthConfig,
	}
}

type ResourceRequestBean struct {
	AppId         string                      `json:"appId"`
	AppIdentifier *client.AppIdentifier       `json:"-"`
	K8sRequest    *application.K8sRequestBean `json:"k8sRequest"`
}

type ResourceInfo struct {
	PodName string `json:"podName"`
}

func (impl *K8sApplicationServiceImpl) GetResource(request *ResourceRequestBean) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.GetResource(restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) CreateResource(request *ResourceRequestBean) (*application.ManifestResponse, error) {
	resourceIdentifier := &openapi.ResourceIdentifier{
		Name:      &request.K8sRequest.ResourceIdentifier.Name,
		Namespace: &request.K8sRequest.ResourceIdentifier.Namespace,
		Group:     &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Group,
		Version:   &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Version,
		Kind:      &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Kind,
	}
	manifestRes, err := impl.helmAppService.GetDesiredManifest(context.Background(), request.AppIdentifier, resourceIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting desired manifest for validation", "err", err)
		return nil, err
	}
	manifest, manifestOk := manifestRes.GetManifestOk()
	if manifestOk == false || len(*manifest) == 0 {
		impl.logger.Debugw("invalid request, desired manifest not found", "err", err)
		return nil, fmt.Errorf("no manifest found for this request")
	}

	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.CreateResource(restConfig, request.K8sRequest, *manifest)
	if err != nil {
		impl.logger.Errorw("error in creating resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) UpdateResource(request *ResourceRequestBean) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.UpdateResource(restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) DeleteResource(request *ResourceRequestBean) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.DeleteResource(restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) ListEvents(request *ResourceRequestBean) (*application.EventsResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.ListEvents(restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) GetPodLogs(request *ResourceRequestBean) (io.ReadCloser, error) {
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.GetPodLogs(restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) GetRestConfigByClusterId(clusterId int) (*rest.Config, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId")
		return nil, err
	}
	//configMap := cluster.Config
	//bearerToken := configMap["bearer_token"]
	cluster.ServerUrl = "https://api.demo.devtron.info"
	bearerToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6Inp4OVhUNU9IdWZXelY3WWhRc0Z3Q21zeDk5eTdSMFpkYXN1bzFLZk1ZNEkifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZXZ0cm9uY2QiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiY2QtdXNlci10b2tlbi13NG5xNCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJjZC11c2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMWFmMTRjNGMtOGYzOC00ZmYxLTgyOTEtMDRlYWI0OTNjZWM3Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRldnRyb25jZDpjZC11c2VyIn0.iHXJOGv4Lw-4GzHTsvqrrCvYBTNCqm9ssYo0kWFMgf5WAWDLVbj0__9vvrhcfLbAmvCE1HahCt1-FQJMZA8NqHv1gJi95j5QgLCRFyvTARgmQRfLELHYIOdaEPN2IRM4G6J1ApTWINoDK26vJbi8Tuisg5aVKwVw4kCPLPc1IMVsnwRE719hlmLqZ5P4pIFjZivrN-jqkE9GqsmZ9ssDJrWPrsYp688zmhoib9jC07XqrT2qqa1CJAgoZ8fP94Dg5deifO-o32wLHkubgboiScONeIcFEubnUwwT5RwduOiEu4OiREYSpW7NMCA3zG-pRugwhqS-38DafU1QsqYkLBQrB8EQbPn3hoYx6n7iCWGYAY7SX_DZNOBd1sVy0f3PXR4yMwo1NOalMO_S3HWqv5TNG-adWKJyVcIq6dABmbrnYGw8z2ctu11HFbUxYUNmHB1aR1T5FdNZIjD0cvu49Xi5pyGYXHMO6TjJIjvhd8K6gJFbjoa6RjET3fN8CiZDG7sizOQZbr2tlVHQP55VTOu4YEfmSjsV8q5ueJYDA7UkNN3jyxP-8a_n-qtXIx4-AeE2VLhJEU8iwbS6288248uRYAUs1KvvNYq5d2C1u0H7ApjjaMN-rRGRxfAa_akex8i6_H-9F3CgtNNhvsBE9QQeNbGNm1qGpRQUKYjt3S4"
	var restConfig *rest.Config
	if cluster.ClusterName == DEFAULT_CLUSTER && len(bearerToken) == 0 {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
			return nil, err
		}
	} else {
		restConfig = &rest.Config{Host: cluster.ServerUrl, BearerToken: bearerToken, TLSClientConfig: rest.TLSClientConfig{Insecure: true}}
	}
	return restConfig, nil
}

func (impl *K8sApplicationServiceImpl) ValidateResourceRequest(appIdentifier *client.AppIdentifier, request *application.K8sRequestBean) (bool, error) {
	app, err := impl.helmAppService.GetApplicationDetail(context.Background(), appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting app detail", "err", err, "appDetails", appIdentifier)
		return false, err
	}
	valid := false
	for _, node := range app.ResourceTreeResponse.Nodes {
		nodeDetails := application.ResourceIdentifier{
			Name:      node.Name,
			Namespace: node.Namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   node.Group,
				Version: node.Version,
				Kind:    node.Kind,
			},
		}
		if nodeDetails == request.ResourceIdentifier {
			valid = true
			break
		}
	}
	if !valid {
		for _, pod := range app.ResourceTreeResponse.PodMetadata {
			if pod.Name == request.ResourceIdentifier.Name {
				for _, container := range pod.Containers {
					if container == request.PodLogsRequest.ContainerName {
						valid = true
						break
					}
				}
			}
		}
	}
	return valid, nil
}

func (impl *K8sApplicationServiceImpl) GetResourceInfo() (*ResourceInfo, error) {
	pod, err := impl.K8sUtil.GetResourceInfoByLabelSelector(impl.aCDAuthConfig.ACDConfigMapNamespace, "app=inception")
	if err != nil {
		impl.logger.Errorw("error on getting resource from k8s, unable to fetch installer pod", "err", err)
		return nil, err
	}
	response := &ResourceInfo{PodName: pod.Name}
	return response, nil
}
