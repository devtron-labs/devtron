package k8s

import (
	"context"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/pkg/cluster"
	bean3 "github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	"github.com/devtron-labs/devtron/util/k8s"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"strconv"
	"strings"
	"sync"
	"time"
)

type K8sCommonService interface {
	GetResource(ctx context.Context, request *ResourceRequestBean) (resp *k8s.ManifestResponse, err error)
	CreateResource(ctx context.Context, request *ResourceRequestBean) (resp *k8s.ManifestResponse, err error)
	UpdateResource(ctx context.Context, request *ResourceRequestBean) (resp *k8s.ManifestResponse, err error)
	DeleteResource(ctx context.Context, request *ResourceRequestBean, userId int32) (resp *k8s.ManifestResponse, err error)
	ListEvents(ctx context.Context, request *ResourceRequestBean) (*k8s.EventsResponse, error)
	GetRestConfigByClusterId(ctx context.Context, clusterId int) (*rest.Config, error, *cluster.ClusterBean)
	GetManifestsByBatch(ctx context.Context, request []ResourceRequestBean) ([]BatchResourceResponse, error)
	FilterServiceAndIngress(ctx context.Context, resourceTreeInf map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean
	GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{}
	RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error)
}
type K8sCommonServiceImpl struct {
	logger                      *zap.SugaredLogger
	K8sUtil                     *k8s.K8sUtil
	helmAppService              client.HelmAppService
	K8sResourceHistoryService   kubernetesResourceAuditLogs.K8sResourceHistoryService
	clusterService              cluster.ClusterService
	K8sApplicationServiceConfig *K8sApplicationServiceConfig
}
type K8sApplicationServiceConfig struct {
	BatchSize        int `env:"BATCH_SIZE" envDefault:"5"`
	TimeOutInSeconds int `env:"TIMEOUT_IN_SECONDS" envDefault:"5"`
}

func NewK8sCommonServiceImpl(Logger *zap.SugaredLogger, k8sUtils *k8s.K8sUtil, helmAppService client.HelmAppService, K8sResourceHistoryService kubernetesResourceAuditLogs.K8sResourceHistoryService, clusterService cluster.ClusterService) *K8sCommonServiceImpl {
	cfg := &K8sApplicationServiceConfig{}
	err := env.Parse(cfg)
	if err != nil {
		Logger.Infow("error occurred while parsing K8sApplicationServiceConfig,so setting batchSize and timeOutInSeconds to default value", "err", err)
	}
	return &K8sCommonServiceImpl{
		logger:                      Logger,
		K8sUtil:                     k8sUtils,
		helmAppService:              helmAppService,
		K8sResourceHistoryService:   K8sResourceHistoryService,
		clusterService:              clusterService,
		K8sApplicationServiceConfig: cfg,
	}
}

func (impl *K8sCommonServiceImpl) GetResource(ctx context.Context, request *ResourceRequestBean) (*k8s.ManifestResponse, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resourceIdentifier := request.K8sRequest.ResourceIdentifier
	resp, err := impl.K8sUtil.GetResource(ctx, resourceIdentifier.Namespace, resourceIdentifier.Name, resourceIdentifier.GroupVersionKind, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) CreateResource(ctx context.Context, request *ResourceRequestBean) (*k8s.ManifestResponse, error) {
	resourceIdentifier := &openapi.ResourceIdentifier{
		Name:      &request.K8sRequest.ResourceIdentifier.Name,
		Namespace: &request.K8sRequest.ResourceIdentifier.Namespace,
		Group:     &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Group,
		Version:   &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Version,
		Kind:      &request.K8sRequest.ResourceIdentifier.GroupVersionKind.Kind,
	}
	manifestRes, err := impl.helmAppService.GetDesiredManifest(ctx, request.AppIdentifier, resourceIdentifier)
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
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.K8sUtil.CreateResources(ctx, restConfig, *manifest, request.K8sRequest.ResourceIdentifier.GroupVersionKind, request.K8sRequest.ResourceIdentifier.Namespace)
	if err != nil {
		impl.logger.Errorw("error in creating resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) UpdateResource(ctx context.Context, request *ResourceRequestBean) (*k8s.ManifestResponse, error) {
	//getting rest config by clusterId
	clusterId := request.ClusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resourceIdentifier := request.K8sRequest.ResourceIdentifier
	resp, err := impl.K8sUtil.UpdateResource(ctx, restConfig, resourceIdentifier.GroupVersionKind, resourceIdentifier.Namespace, request.K8sRequest.Patch)
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "clusterId", clusterId)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) DeleteResource(ctx context.Context, request *ResourceRequestBean, userId int32) (*k8s.ManifestResponse, error) {
	//getting rest config by clusterId
	clusterId := request.ClusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resourceIdentifier := request.K8sRequest.ResourceIdentifier
	resp, err := impl.K8sUtil.DeleteResource(ctx, restConfig, resourceIdentifier.GroupVersionKind, resourceIdentifier.Namespace, resourceIdentifier.Name, request.K8sRequest.ForceDelete)
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "clusterId", clusterId)
		return nil, err
	}
	if request.AppIdentifier != nil {
		saveAuditLogsErr := impl.K8sResourceHistoryService.SaveHelmAppsResourceHistory(request.AppIdentifier, request.K8sRequest, userId, bean3.Delete)
		if saveAuditLogsErr != nil {
			impl.logger.Errorw("error in saving audit logs for delete resource request", "err", err)
		}
	}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) ListEvents(ctx context.Context, request *ResourceRequestBean) (*k8s.EventsResponse, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resourceIdentifier := request.K8sRequest.ResourceIdentifier
	list, err := impl.K8sUtil.ListEvents(restConfig, resourceIdentifier.Namespace, resourceIdentifier.GroupVersionKind, ctx, resourceIdentifier.Name)
	if err != nil {
		impl.logger.Errorw("error in listing events", "err", err, "clusterId", clusterId)
		return nil, err
	}
	return &k8s.EventsResponse{list}, nil

}

func (impl *K8sCommonServiceImpl) FilterServiceAndIngress(ctx context.Context, resourceTree map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean {
	noOfNodes := len(resourceTree["nodes"].([]interface{}))
	resourceNodeItemss := resourceTree["nodes"].([]interface{})
	for i := 0; i < noOfNodes; i++ {
		resourceItem := resourceNodeItemss[i].(map[string]interface{})
		var kind, name, namespace string
		kind = impl.extractResourceValue(resourceItem, "kind")
		name = impl.extractResourceValue(resourceItem, "name")
		namespace = impl.extractResourceValue(resourceItem, "namespace")

		if appId == "" {
			appId = strconv.Itoa(appDetail.ClusterId) + "|" + namespace + "|" + (appDetail.AppName + "-" + appDetail.EnvironmentName)
		}
		if strings.Compare(kind, Service) == 0 || strings.Compare(kind, Ingress) == 0 {
			group := impl.extractResourceValue(resourceItem, Group)
			version := impl.extractResourceValue(resourceItem, Version)
			req := ResourceRequestBean{
				AppId:     appId,
				ClusterId: appDetail.ClusterId,
				AppIdentifier: &client.AppIdentifier{
					ClusterId: appDetail.ClusterId,
				},
				K8sRequest: &k8s.K8sRequestBean{
					ResourceIdentifier: k8s.ResourceIdentifier{
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

func (impl *K8sCommonServiceImpl) GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{} {
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

func (impl *K8sCommonServiceImpl) GetManifestsByBatch(ctx context.Context, requests []ResourceRequestBean) ([]BatchResourceResponse, error) {
	ch := make(chan []BatchResourceResponse)
	var res []BatchResourceResponse
	ctx, cancel := context.WithTimeout(ctx, time.Duration(impl.K8sApplicationServiceConfig.TimeOutInSeconds)*time.Second)
	defer cancel()
	go func() {
		ans := impl.getManifestsByBatch(ctx, requests)
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

func (impl *K8sCommonServiceImpl) GetRestConfigByClusterId(ctx context.Context, clusterId int) (*rest.Config, error, *cluster.ClusterBean) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "K8sApplicationService.GetRestConfigByClusterId")
	defer span.End()
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId", clusterId)
		return nil, err, nil
	}
	clusterConfig := cluster.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
	if err != nil {
		impl.logger.Errorw("Error in getting rest config", "err", err, "clusterId", clusterId)
		return restConfig, err, nil
	}
	return restConfig, nil, cluster
}

func (impl *K8sCommonServiceImpl) RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error) {

	clusterId := request.ClusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "clusterId", clusterId, "err", err)
		return nil, err
	}
	response := &RotatePodResponse{}
	var resourceResponses []*bean3.RotatePodResourceResponse
	var containsError bool
	for _, resourceIdentifier := range request.Resources {
		resourceResponse := &bean3.RotatePodResourceResponse{
			ResourceIdentifier: resourceIdentifier,
		}
		groupVersionKind := resourceIdentifier.GroupVersionKind
		name := resourceIdentifier.Name
		namespace := resourceIdentifier.Namespace
		resourceKind := groupVersionKind.Kind
		// validate one of deployment, statefulset, daemonSet, Rollout
		if resourceKind != kube.DeploymentKind && resourceKind != kube.StatefulSetKind && resourceKind != kube.DaemonSetKind && resourceKind != k8s.K8sClusterResourceRolloutKind {
			impl.logger.Errorf("restarting not supported for kind %s name %s", resourceKind, resourceIdentifier.Name)
			containsError = true
			resourceResponse.ErrorResponse = k8s.RestartingNotSupported
		} else {
			activitySnapshot := time.Now().Format(time.RFC3339)
			data := fmt.Sprintf(`{"metadata": {"annotations": {"devtron.ai/restartedAt": "%s"}},"spec": {"template": {"metadata": {"annotations": {"devtron.ai/activity": "%s"}}}}}`, activitySnapshot, activitySnapshot)
			var patchType types.PatchType
			if resourceKind != k8s.K8sClusterResourceRolloutKind {
				patchType = types.StrategicMergePatchType
			} else {
				// rollout does not support strategic merge type
				patchType = types.MergePatchType
			}
			_, err = impl.K8sUtil.PatchResourceRequest(ctx, restConfig, patchType, data, name, namespace, groupVersionKind)
			if err != nil {
				containsError = true
				resourceResponse.ErrorResponse = err.Error()
			}
		}
		resourceResponses = append(resourceResponses, resourceResponse)
	}

	response.Responses = resourceResponses
	response.ContainsError = containsError
	return response, nil
}

func (impl *K8sCommonServiceImpl) getManifestsByBatch(ctx context.Context, requests []ResourceRequestBean) []BatchResourceResponse {
	//total batch length
	batchSize := impl.K8sApplicationServiceConfig.BatchSize
	if requests == nil {
		impl.logger.Errorw("Empty requests for getManifestsInBatch")
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
				resp.ManifestResponse, resp.Err = impl.GetResource(ctx, &requests[i+j])
				res[i+j] = resp
				wg.Done()
			}(j)
		}
		wg.Wait()
		i += batchSize
	}
	return res
}

func (impl *K8sCommonServiceImpl) extractResourceValue(resourceItem map[string]interface{}, resourceName string) string {
	if _, ok := resourceItem[resourceName]; ok && resourceItem[resourceName] != nil {
		return resourceItem[resourceName].(string)
	}
	return ""
}

func (impl *K8sCommonServiceImpl) getUrls(manifest *k8s.ManifestResponse) bean3.Response {
	var res bean3.Response
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
