package k8s

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
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
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_CLUSTER = "default_cluster"
)

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
	GetRestConfigByCluster(cluster *cluster.ClusterBean) (*rest.Config, error)
	GetManifestsByBatch(ctx context.Context, request []ResourceRequestBean) ([]BatchResourceResponse, error)
	FilterServiceAndIngress(resourceTreeInf map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean
	GetUrlsByBatch(resp []BatchResourceResponse) []interface{}
}
type K8sApplicationServiceImpl struct {
	logger                      *zap.SugaredLogger
	clusterService              cluster.ClusterService
	pump                        connector.Pump
	k8sClientService            application.K8sClientService
	helmAppService              client.HelmAppService
	K8sUtil                     *util.K8sUtil
	aCDAuthConfig               *util3.ACDAuthConfig
	K8sApplicationServiceConfig *K8sApplicationServiceConfig
}

type K8sApplicationServiceConfig struct {
	BatchSize        int `env:"BATCH_SIZE" envDefault:"5"`
	TimeOutInSeconds int `env:"TIMEOUT_IN_SECONDS" envDefault:"5"`
}

func NewK8sApplicationServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	pump connector.Pump, k8sClientService application.K8sClientService,
	helmAppService client.HelmAppService, K8sUtil *util.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig) *K8sApplicationServiceImpl {
	cfg := &K8sApplicationServiceConfig{}
	err := env.Parse(cfg)
	if err != nil {
		Logger.Infow("error occurred while parsing K8sApplicationServiceConfig,so setting batchSize and timeOutInSeconds to default value", "err", err)
	}
	return &K8sApplicationServiceImpl{
		logger:                      Logger,
		clusterService:              clusterService,
		pump:                        pump,
		k8sClientService:            k8sClientService,
		helmAppService:              helmAppService,
		K8sUtil:                     K8sUtil,
		aCDAuthConfig:               aCDAuthConfig,
		K8sApplicationServiceConfig: cfg,
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

type BatchResourceResponse struct {
	ManifestResponse *application.ManifestResponse
	Err              error
}

func (impl *K8sApplicationServiceImpl) FilterServiceAndIngress(resourceTree map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean {
	noOfNodes := len(resourceTree["nodes"].([]interface{}))
	resourceNodeItemss := resourceTree["nodes"].([]interface{})
	for i := 0; i < noOfNodes; i++ {
		resourceItem := resourceNodeItemss[i].(map[string]interface{})
		var kind, name, namespace string
		if _, ok := resourceItem["kind"]; ok && resourceItem["kind"] != nil {
			kind = resourceItem["kind"].(string)
		}
		if _, ok := resourceItem["name"]; ok && resourceItem["name"] != nil {
			name = resourceItem["name"].(string)
		}
		if _, ok := resourceItem["namespace"]; ok && resourceItem["namespace"] != nil {
			namespace = resourceItem["namespace"].(string)
		}

		if appId == "" {
			appId = strconv.Itoa(appDetail.ClusterId) + "|" + namespace + "|" + (appDetail.AppName + "-" + appDetail.EnvironmentName)
		}
		if strings.Compare(kind, "Service") == 0 || strings.Compare(kind, "Ingress") == 0 {
			group := ""
			version := ""
			if _, ok := resourceItem["version"]; ok {
				version = resourceItem["version"].(string)
			}
			if _, ok := resourceItem["group"]; ok {
				group = resourceItem["group"].(string)
			}
			req := ResourceRequestBean{
				AppId: appId,
				AppIdentifier: &client.AppIdentifier{
					ClusterId: appDetail.ClusterId,
				},
				K8sRequest: &application.K8sRequestBean{
					ResourceIdentifier: application.ResourceIdentifier{
						Name:      name,
						Namespace: namespace,
						GroupVersionKind: schema.GroupVersionKind{
							Version: version,
							Kind:    kind,
							Group:   group,
						},
					},
				},
			}

			validRequests = append(validRequests, req)
		}
	}
	return validRequests
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

func (impl *K8sApplicationServiceImpl) GetUrlsByBatch(resp []BatchResourceResponse) []interface{} {
	result := make([]interface{}, 0)
	for _, res := range resp {
		err := res.Err
		if err != nil {
			continue
		}
		urlRes := impl.getUrls(res.ManifestResponse)
		result = append(result, urlRes)
	}
	return result
}

func (impl *K8sApplicationServiceImpl) getUrls(manifest *application.ManifestResponse) Response {
	var res Response
	kind := manifest.Manifest.Object["kind"]
	if _, ok := manifest.Manifest.Object["metadata"]; ok {
		metadata := manifest.Manifest.Object["metadata"].(map[string]interface{})
		if metadata != nil {
			name := metadata["name"]
			if name != nil {
				res.Name = name.(string)
			}
		}
	}

	if kind != nil {
		res.Kind = kind.(string)
	}
	res.PointsTo = ""
	urls := make([]string, 0)
	if res.Kind == "Ingress" {
		if manifest.Manifest.Object["spec"] != nil {
			spec := manifest.Manifest.Object["spec"].(map[string]interface{})
			if spec["rules"] != nil {
				rules := spec["rules"].([]interface{})
				for _, rule := range rules {
					ruleMap := rule.(map[string]interface{})
					url := ""
					if ruleMap["host"] != nil {
						url = ruleMap["host"].(string)
					}
					var httpPaths []interface{}
					if ruleMap["http"] != nil && ruleMap["http"].(map[string]interface{})["paths"] != nil {
						httpPaths = ruleMap["http"].(map[string]interface{})["paths"].([]interface{})
					} else {
						continue
					}
					for _, httpPath := range httpPaths {
						path := httpPath.(map[string]interface{})["path"]
						if path != nil {
							url = url + path.(string)
						}
						urls = append(urls, url)
					}
				}
			}
		}
	}

	if manifest.Manifest.Object["status"] != nil {
		status := manifest.Manifest.Object["status"].(map[string]interface{})
		if status["loadBalancer"] != nil {
			loadBalancer := status["loadBalancer"].(map[string]interface{})
			if loadBalancer["ingress"] != nil {
				ingressArray := loadBalancer["ingress"].([]interface{})
				if len(ingressArray) > 0 {
					if hostname, ok := ingressArray[0].(map[string]interface{})["hostname"]; ok {
						res.PointsTo = hostname.(string)
					} else if ip, ok := ingressArray[0].(map[string]interface{})["ip"]; ok {
						res.PointsTo = ip.(string)
					}
				}
			}
		}
	}
	res.Urls = urls
	return res
}

func (impl *K8sApplicationServiceImpl) GetManifestsByBatch(ctx context.Context, requests []ResourceRequestBean) ([]BatchResourceResponse, error) {
	ch := make(chan []BatchResourceResponse)
	var res []BatchResourceResponse
	ctx, cancel := context.WithTimeout(ctx, time.Duration(impl.K8sApplicationServiceConfig.TimeOutInSeconds)*time.Second)
	defer cancel()
	go func() {
		ans := impl.getManifestsByBatch(requests)
		ch <- ans
	}()
	select {
	case ans := <-ch:
		res = ans
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	impl.logger.Info("successfully fetched the requested manifests")
	return res, nil
}

func (impl *K8sApplicationServiceImpl) getManifestsByBatch(requests []ResourceRequestBean) []BatchResourceResponse {
	//total batch length
	batchSize := impl.K8sApplicationServiceConfig.BatchSize
	if requests == nil {
		impl.logger.Error("Empty requests for getManifestsInBatch")
	}
	requestsLength := len(requests)
	//final batch responses
	res := make([]BatchResourceResponse, requestsLength)
	for i := 0; i < requestsLength; {
		//requests left to process
		remainingBatch := requestsLength - i
		if remainingBatch < batchSize {
			batchSize = remainingBatch
		}
		var wg sync.WaitGroup
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			go func(j int) {
				resp := BatchResourceResponse{}
				resp.ManifestResponse, resp.Err = impl.GetResource(&requests[i+j])
				res[i+j] = resp
				wg.Done()
			}(j)
		}
		wg.Wait()
		i += batchSize
	}
	return res
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

func (impl *K8sApplicationServiceImpl) GetRestConfigByCluster(cluster *cluster.ClusterBean) (*rest.Config, error) {
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	var restConfig *rest.Config
	var err error
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
