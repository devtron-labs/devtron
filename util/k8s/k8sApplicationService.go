package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/pkg/cluster"
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
}
type K8sApplicationServiceImpl struct {
	logger           *zap.SugaredLogger
	clusterService   cluster.ClusterService
	pump             connector.Pump
	k8sClientService application.K8sClientService
	helmAppService   client.HelmAppService
}

func NewK8sApplicationServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	pump connector.Pump, k8sClientService application.K8sClientService,
	helmAppService client.HelmAppService) *K8sApplicationServiceImpl {
	return &K8sApplicationServiceImpl{
		logger:           Logger,
		clusterService:   clusterService,
		pump:             pump,
		k8sClientService: k8sClientService,
		helmAppService:   helmAppService,
	}
}

type ResourceRequestBean struct {
	AppId         string                      `json:"appId"`
	AppIdentifier *client.AppIdentifier       `json:"-"`
	K8sRequest    *application.K8sRequestBean `json:"k8sRequest"`
}

func (impl *K8sApplicationServiceImpl) GetResource(request *ResourceRequestBean) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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
	restConfig, err := impl.getRestConfigByClusterId(request.AppIdentifier.ClusterId)
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

func (impl *K8sApplicationServiceImpl) getRestConfigByClusterId(clusterId int) (*rest.Config, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId")
		return nil, err
	}
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
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