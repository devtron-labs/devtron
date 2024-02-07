package k8s

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	bean3 "github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	"github.com/devtron-labs/devtron/util"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strconv"
	"sync"
	"time"
)

type K8sCommonService interface {
	GetResource(ctx context.Context, request *ResourceRequestBean) (resp *ResourceGetResponse, err error)
	UpdateResource(ctx context.Context, request *ResourceRequestBean) (resp *k8s2.ManifestResponse, err error)
	DeleteResource(ctx context.Context, request *ResourceRequestBean) (*k8s2.ManifestResponse, error)
	ListEvents(ctx context.Context, request *ResourceRequestBean) (*k8s2.EventsResponse, error)
	GetRestConfigByClusterId(ctx context.Context, clusterId int) (*rest.Config, error, *cluster.ClusterBean)
	GetManifestsByBatch(ctx context.Context, request []ResourceRequestBean) ([]BatchResourceResponse, error)
	FilterK8sResources(ctx context.Context, resourceTreeInf map[string]interface{}, appDetail bean.AppDetailContainer, appId string, kindsToBeFiltered []string) []ResourceRequestBean
	GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{}
	RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error)
	GetCoreClientByClusterId(clusterId int) (*kubernetes.Clientset, *v1.CoreV1Client, error)
	GetK8sServerVersion(clusterId int) (*version.Info, error)
	PortNumberExtraction(resp []BatchResourceResponse, resourceTree map[string]interface{}) map[string]interface{}
}
type K8sCommonServiceImpl struct {
	logger                      *zap.SugaredLogger
	K8sUtil                     *k8s.K8sUtilExtended
	helmAppService              service.HelmAppService
	K8sResourceHistoryService   kubernetesResourceAuditLogs.K8sResourceHistoryService
	clusterService              cluster.ClusterService
	K8sApplicationServiceConfig *K8sApplicationServiceConfig
}

type K8sApplicationServiceConfig struct {
	BatchSize        int `env:"BATCH_SIZE" envDefault:"5"`
	TimeOutInSeconds int `env:"TIMEOUT_IN_SECONDS" envDefault:"5"`
}

func NewK8sCommonServiceImpl(Logger *zap.SugaredLogger, k8sUtils *k8s.K8sUtilExtended, helmAppService service.HelmAppService,
	K8sResourceHistoryService kubernetesResourceAuditLogs.K8sResourceHistoryService, clusterService cluster.ClusterService) *K8sCommonServiceImpl {
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

func (impl *K8sCommonServiceImpl) GetResource(ctx context.Context, request *ResourceRequestBean) (*ResourceGetResponse, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err, _ := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resourceIdentifier := request.K8sRequest.ResourceIdentifier
	resp, err := impl.K8sUtil.K8sService.GetResource(ctx, resourceIdentifier.Namespace, resourceIdentifier.Name, resourceIdentifier.GroupVersionKind, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "resource", resourceIdentifier.Name)
		return nil, err
	}
	response := &ResourceGetResponse{
		ManifestResponse: resp,
	}
	return response, nil
}

func (impl *K8sCommonServiceImpl) UpdateResource(ctx context.Context, request *ResourceRequestBean) (*k8s2.ManifestResponse, error) {
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
		statusError, ok := err.(*errors.StatusError)
		if ok {
			err = &util2.ApiError{Code: "400", HttpStatusCode: int(statusError.ErrStatus.Code), UserMessage: statusError.Error()}
		}
		return nil, err
	}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) DeleteResource(ctx context.Context, request *ResourceRequestBean) (*k8s2.ManifestResponse, error) {
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
	//if request.AppIdentifier != nil {
	//	saveAuditLogsErr := impl.K8sResourceHistoryService.SaveHelmAppsResourceHistory(request.AppIdentifier, request.K8sRequest, userId, bean3.Delete)
	//	if saveAuditLogsErr != nil {
	//		impl.logger.Errorw("error in saving audit logs for delete resource request", "err", err)
	//	}
	//}
	return resp, nil
}

func (impl *K8sCommonServiceImpl) ListEvents(ctx context.Context, request *ResourceRequestBean) (*k8s2.EventsResponse, error) {
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
	return &k8s2.EventsResponse{list}, nil

}

func (impl *K8sCommonServiceImpl) FilterK8sResources(ctx context.Context, resourceTree map[string]interface{}, appDetail bean.AppDetailContainer, appId string, kindsToBeFiltered []string) []ResourceRequestBean {
	validRequests := make([]ResourceRequestBean, 0)
	kindsToBeFilteredMap := util.ConvertStringSliceToMap(kindsToBeFiltered)
	resourceTreeNodes, ok := resourceTree["nodes"]
	if !ok {
		return validRequests
	}
	noOfNodes := len(resourceTreeNodes.([]interface{}))
	resourceNodeItemss := resourceTreeNodes.([]interface{})
	for i := 0; i < noOfNodes; i++ {
		resourceItem := resourceNodeItemss[i].(map[string]interface{})
		var kind, name, namespace string
		kind = impl.extractResourceValue(resourceItem, "kind")
		name = impl.extractResourceValue(resourceItem, "name")
		namespace = impl.extractResourceValue(resourceItem, "namespace")

		if appId == "" {
			appId = strconv.Itoa(appDetail.ClusterId) + "|" + namespace + "|" + (appDetail.AppName + "-" + appDetail.EnvironmentName)
		}
		if kindsToBeFilteredMap[kind] {
			group := impl.extractResourceValue(resourceItem, k8sCommonBean.Group)
			version := impl.extractResourceValue(resourceItem, k8sCommonBean.Version)
			req := ResourceRequestBean{
				AppId:     appId,
				ClusterId: appDetail.ClusterId,
				AppIdentifier: &service.AppIdentifier{
					ClusterId: appDetail.ClusterId,
				},
				K8sRequest: &k8s2.K8sRequestBean{
					ResourceIdentifier: k8s2.ResourceIdentifier{
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
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
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
		if resourceKind != k8sCommonBean.DeploymentKind && resourceKind != k8sCommonBean.StatefulSetKind &&
			resourceKind != k8sCommonBean.DaemonSetKind && resourceKind != k8sCommonBean.K8sClusterResourceRolloutKind {
			impl.logger.Errorf("restarting not supported for kind %s name %s", resourceKind, resourceIdentifier.Name)
			containsError = true
			resourceResponse.ErrorResponse = k8s2.RestartingNotSupported
		} else {
			activitySnapshot := time.Now().Format(time.RFC3339)
			data := fmt.Sprintf(`{"metadata": {"annotations": {"devtron.ai/restartedAt": "%s"}},"spec": {"template": {"metadata": {"annotations": {"devtron.ai/activity": "%s"}}}}}`, activitySnapshot, activitySnapshot)
			var patchType types.PatchType
			if resourceKind != k8sCommonBean.K8sClusterResourceRolloutKind {
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
				response, err := impl.GetResource(ctx, &requests[i+j])
				if response != nil {
					resp.ManifestResponse = response.ManifestResponse
				}
				resp.Err = err
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

func (impl *K8sCommonServiceImpl) getUrls(manifest *k8s2.ManifestResponse) bean3.Response {
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

func (impl *K8sCommonServiceImpl) GetK8sServerVersion(clusterId int) (*version.Info, error) {
	clientSet, _, err := impl.GetCoreClientByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting coreV1 client by clusterId", "clusterId", clusterId, "err", err)
		return nil, err
	}
	k8sVersion, err := impl.K8sUtil.GetK8sServerVersion(clientSet)
	if err != nil {
		impl.logger.Errorw("error in getting k8s server version", "clusterId", clusterId, "err", err)
		return nil, err
	}
	return k8sVersion, err
}
func (impl *K8sCommonServiceImpl) GetCoreClientByClusterId(clusterId int) (*kubernetes.Clientset, *v1.CoreV1Client, error) {
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error occurred in finding clusterBean by Id", "clusterId", clusterId, "err", err)
		return nil, nil, err
	}

	clusterConfig := clusterBean.GetClusterConfig()
	v1Client, err := impl.K8sUtil.GetCoreV1Client(clusterConfig)
	if err != nil {
		//not logging clusterConfig as it contains sensitive data
		impl.logger.Errorw("error occurred in getting v1Client with cluster config", "err", err, "clusterId", clusterId)
		return nil, nil, err
	}
	_, _, clientSet, err := impl.K8sUtil.GetK8sConfigAndClients(clusterConfig)
	if err != nil {
		//not logging clusterConfig as it contains sensitive data
		impl.logger.Errorw("error occurred in getting clientSet with cluster config", "err", err, "clusterId", clusterId)
		return nil, v1Client, err
	}
	return clientSet, v1Client, nil
}

func (impl K8sCommonServiceImpl) PortNumberExtraction(resp []BatchResourceResponse, resourceTree map[string]interface{}) map[string]interface{} {
	servicePortMapping := make(map[string]interface{})
	endpointPortMapping := make(map[string]interface{})
	endpointSlicePortMapping := make(map[string]interface{})

	for _, portHolder := range resp {
		if portHolder.ManifestResponse == nil {
			continue
		}
		kind, ok := portHolder.ManifestResponse.Manifest.Object[k8sCommonBean.Kind]
		if !ok {
			impl.logger.Warnw("kind not found in resource tree, unable to extract port no")
			continue
		}
		metadataResp, ok := portHolder.ManifestResponse.Manifest.Object[k8sCommonBean.K8sClusterResourceMetadataKey]
		if !ok {
			impl.logger.Warnw("metadata not found in resource tree, unable to extract port no")
			continue
		}
		metadata, ok := metadataResp.(map[string]interface{})
		if !ok {
			impl.logger.Warnw("metadata not found in resource tree, unable to extract port no")
			continue
		}
		serviceNameResp, ok := metadata[k8sCommonBean.K8sClusterResourceMetadataNameKey]
		if !ok {
			impl.logger.Warnw("service name not found in resource tree, unable to extract port no")
			continue
		}
		serviceName, ok := serviceNameResp.(string)
		if !ok {
			impl.logger.Warnw("service name not found in resource tree, unable to extract port no")
			continue
		}
		if kind == k8sCommonBean.ServiceKind {
			specField, ok := portHolder.ManifestResponse.Manifest.Object[k8sCommonBean.Spec]
			if !ok {
				impl.logger.Warnw("spec not found in resource tree, unable to extract port no")
				continue
			}
			spec, ok := specField.(map[string]interface{})
			if !ok {
				impl.logger.Warnw("spec not found in resource tree, unable to extract port no")
				continue
			}

			if spec != nil {
				ports, ok := spec[k8sCommonBean.Ports]
				if !ok {
					impl.logger.Warnw("ports not found in resource tree, unable to extract port no")
					continue
				}
				portList, ok := ports.([]interface{})
				if !ok {
					impl.logger.Warnw("portList not found in resource tree, unable to extract port no")
					continue
				}
				servicePorts := make([]int64, 0)
				for _, portItem := range portList {
					portItems, ok := portItem.(map[string]interface{})
					if !ok {
						impl.logger.Warnw("portItems not found in resource tree, unable to extract port no")
						continue
					}
					if portItems != nil {
						portNumbers, ok := portItems[k8sCommonBean.Port]
						if !ok {
							impl.logger.Warnw("ports number found in resource tree, unable to extract port no")
							continue
						}
						portNumber, ok := portNumbers.(int64)
						if !ok {
							impl.logger.Warnw("portNumber(int64) not found in resource tree, unable to extract port no")
							continue
						}
						if portNumber != 0 {
							servicePorts = append(servicePorts, portNumber)
						}
					}
				}
				servicePortMapping[serviceName] = servicePorts
			} else {
				impl.logger.Warnw("spec doest not contain data", "spec", spec)
				continue
			}
		}
		if kind == k8sCommonBean.EndpointsKind {
			subsetsField, ok := portHolder.ManifestResponse.Manifest.Object[k8sCommonBean.Subsets]
			if !ok {
				impl.logger.Warnw("spec not found in resource tree, unable to extract port no")
				continue
			}
			if subsetsField != nil {
				subsets, ok := subsetsField.([]interface{})
				if !ok {
					impl.logger.Warnw("subsets not found in resource tree, unable to extract port no")
					continue
				}
				for _, subset := range subsets {
					subsetObj, ok := subset.(map[string]interface{})
					if !ok {
						impl.logger.Warnw("subsetObj not found in resource tree, unable to extract port no")
						continue
					}
					if subsetObj != nil {
						ports, ok := subsetObj[k8sCommonBean.Ports]
						if !ok {
							impl.logger.Warnw("ports not found in resource tree endpoints, unable to extract port no")
							continue
						}
						portsIfs, ok := ports.([]interface{})
						if !ok {
							impl.logger.Warnw("portsIfs not found in resource tree, unable to extract port no")
							continue
						}
						endpointPorts := make([]int64, 0)
						for _, portsIf := range portsIfs {
							portsIfObj, ok := portsIf.(map[string]interface{})
							if !ok {
								impl.logger.Warnw("portsIfObj not found in resource tree, unable to extract port no")
								continue
							}
							if portsIfObj != nil {
								port, ok := portsIfObj[k8sCommonBean.Port].(int64)
								if !ok {
									impl.logger.Warnw("port not found in resource tree, unable to extract port no")
									continue
								}
								endpointPorts = append(endpointPorts, port)
							}
						}
						endpointPortMapping[serviceName] = endpointPorts
					}
				}
			}
		}
		if kind == k8sCommonBean.EndPointsSlice {
			portsField, ok := portHolder.ManifestResponse.Manifest.Object[k8sCommonBean.Ports]
			if !ok {
				impl.logger.Warnw("ports not found in resource tree endpoint, unable to extract port no")
				continue
			}
			if portsField != nil {
				endPointsSlicePorts, ok := portsField.([]interface{})
				if !ok {
					impl.logger.Warnw("endPointsSlicePorts not found in resource tree endpoint, unable to extract port no")
					continue
				}
				endpointSlicePorts := make([]int64, 0)
				for _, val := range endPointsSlicePorts {
					portNumbers, ok := val.(map[string]interface{})[k8sCommonBean.Port]
					if !ok {
						impl.logger.Warnw("endPointsSlicePorts not found in resource tree endpoint, unable to extract port no")
						continue
					}
					portNumber, ok := portNumbers.(int64)
					if !ok {
						impl.logger.Warnw("portNumber(int64) not found in resource tree endpoint, unable to extract port no")
						continue
					}
					if portNumber != 0 {
						endpointSlicePorts = append(endpointSlicePorts, portNumber)
					}
				}
				endpointSlicePortMapping[serviceName] = endpointSlicePorts
			}
		}
	}
	if val, ok := resourceTree[k8sCommonBean.Nodes]; ok {
		resourceTreeVal, ok := val.([]interface{})
		if !ok {
			impl.logger.Warnw("resourceTreeVal not found in resourceTree, unable to extract port no")
			return resourceTree
		}
		for _, val := range resourceTreeVal {
			value, ok := val.(map[string]interface{})
			if !ok {
				impl.logger.Warnw("value not found in resourceTreeVal, unable to extract port no")
				continue
			}
			serviceNameRes, ok := value[k8sCommonBean.K8sClusterResourceMetadataNameKey]
			if !ok {
				impl.logger.Warnw("service name not found in resourceTreeVal, unable to extract port no")
				continue
			}
			serviceName, ok := serviceNameRes.(string)
			if !ok {
				impl.logger.Warnw("service name not found in resourceTreeVal, unable to extract port no")
				continue
			}
			for key, _type := range value {
				if key == k8sCommonBean.Kind && _type == k8sCommonBean.EndpointsKind {
					if port, ok := endpointPortMapping[serviceName]; ok {
						value[k8sCommonBean.Port] = port
					}
				}
				if key == k8sCommonBean.Kind && _type == k8sCommonBean.ServiceKind {
					if port, ok := servicePortMapping[serviceName]; ok {
						value[k8sCommonBean.Port] = port
					}
				}
				if key == k8sCommonBean.Kind && _type == k8sCommonBean.EndPointsSlice {
					if port, ok := endpointSlicePortMapping[serviceName]; ok {
						value[k8sCommonBean.Port] = port
					}
				}
			}
		}
	}
	return resourceTree
}
