package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	yamlUtil "github.com/devtron-labs/devtron/util/yaml"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"io"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_CLUSTER = "default_cluster"
)

type K8sApplicationService interface {
	GetResource(ctx context.Context, request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	CreateResource(ctx context.Context, request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	UpdateResource(ctx context.Context, request *ResourceRequestBean) (resp *application.ManifestResponse, err error)
	DeleteResource(ctx context.Context, request *ResourceRequestBean, userId int32) (resp *application.ManifestResponse, err error)
	ListEvents(ctx context.Context, request *ResourceRequestBean) (*application.EventsResponse, error)
	GetPodLogs(ctx context.Context, request *ResourceRequestBean) (io.ReadCloser, error)
	ValidateResourceRequest(ctx context.Context, appIdentifier *client.AppIdentifier, request *application.K8sRequestBean) (bool, error)
	ValidateClusterResourceRequest(ctx context.Context, clusterResourceRequest *ResourceRequestBean,
		rbacCallback func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool) (bool, error)
	ValidateClusterResourceBean(ctx context.Context, clusterId int, manifest unstructured.Unstructured, gvk schema.GroupVersionKind, rbacCallback func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool) bool
	GetResourceInfo(ctx context.Context) (*ResourceInfo, error)
	GetRestConfigByClusterId(ctx context.Context, clusterId int) (*rest.Config, error)
	GetRestConfigByCluster(ctx context.Context, cluster *cluster.ClusterBean) (*rest.Config, error)
	GetManifestsByBatch(ctx context.Context, request []ResourceRequestBean) ([]BatchResourceResponse, error)
	FilterServiceAndIngress(ctx context.Context, resourceTreeInf map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean
	GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{}
	GetAllApiResources(ctx context.Context, clusterId int, isSuperAdmin bool, userId int32) (*application.GetAllApiResourcesResponse, error)
	GetResourceList(ctx context.Context, token string, request *ResourceRequestBean, validateResourceAccess func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) (*util.ClusterResourceListMap, error)
	ApplyResources(ctx context.Context, token string, request *application.ApplyResourcesRequest, resourceRbacHandler func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) ([]*application.ApplyResourcesResponse, error)
	FetchConnectionStatusForCluster(k8sClientSet *kubernetes.Clientset, clusterId int) error
	RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error)
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
	K8sResourceHistoryService   kubernetesResourceAuditLogs.K8sResourceHistoryService
}

type K8sApplicationServiceConfig struct {
	BatchSize        int `env:"BATCH_SIZE" envDefault:"5"`
	TimeOutInSeconds int `env:"TIMEOUT_IN_SECONDS" envDefault:"5"`
}

func NewK8sApplicationServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	pump connector.Pump, k8sClientService application.K8sClientService,
	helmAppService client.HelmAppService, K8sUtil *util.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig,
	K8sResourceHistoryService kubernetesResourceAuditLogs.K8sResourceHistoryService) *K8sApplicationServiceImpl {
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
		K8sResourceHistoryService:   K8sResourceHistoryService,
	}
}

type ResourceRequestBean struct {
	AppId         string                      `json:"appId"`
	AppIdentifier *client.AppIdentifier       `json:"-"`
	K8sRequest    *application.K8sRequestBean `json:"k8sRequest"`
	ClusterId     int                         `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
}

type ResourceInfo struct {
	PodName string `json:"podName"`
}

type BatchResourceResponse struct {
	ManifestResponse *application.ManifestResponse
	Err              error
}

func (impl *K8sApplicationServiceImpl) FilterServiceAndIngress(ctx context.Context, resourceTree map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean {
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
		if strings.Compare(kind, "Service") == 0 || strings.Compare(kind, "Ingress") == 0 {
			group := impl.extractResourceValue(resourceItem, "group")
			version := impl.extractResourceValue(resourceItem, "version")
			req := ResourceRequestBean{
				AppId:     appId,
				ClusterId: appDetail.ClusterId,
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

func (impl *K8sApplicationServiceImpl) extractResourceValue(resourceItem map[string]interface{}, resourceName string) string {
	if _, ok := resourceItem[resourceName]; ok && resourceItem[resourceName] != nil {
		return resourceItem[resourceName].(string)
	}
	return ""
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

func (impl *K8sApplicationServiceImpl) GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{} {
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

func (impl *K8sApplicationServiceImpl) getManifestsByBatch(ctx context.Context, requests []ResourceRequestBean) []BatchResourceResponse {
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

func (impl *K8sApplicationServiceImpl) GetResource(ctx context.Context, request *ResourceRequestBean) (*application.ManifestResponse, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.GetResource(ctx, restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) CreateResource(ctx context.Context, request *ResourceRequestBean) (*application.ManifestResponse, error) {
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
	restConfig, err := impl.GetRestConfigByClusterId(ctx, request.AppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.CreateResource(ctx, restConfig, request.K8sRequest, *manifest)
	if err != nil {
		impl.logger.Errorw("error in creating resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) UpdateResource(ctx context.Context, request *ResourceRequestBean) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	clusterId := request.ClusterId
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.UpdateResource(ctx, restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in updating resource", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) DeleteResource(ctx context.Context, request *ResourceRequestBean, userId int32) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	clusterId := request.ClusterId
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.DeleteResource(ctx, restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting resource", "err", err, "request", request)
		return nil, err
	}
	if request.AppIdentifier != nil {
		saveAuditLogsErr := impl.K8sResourceHistoryService.SaveHelmAppsResourceHistory(request.AppIdentifier, request.K8sRequest, userId, "delete")
		if saveAuditLogsErr != nil {
			impl.logger.Errorw("error in saving audit logs for delete resource request", "err", err)
		}
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) ListEvents(ctx context.Context, request *ResourceRequestBean) (*application.EventsResponse, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.AppIdentifier.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.ListEvents(ctx, restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) GetPodLogs(ctx context.Context, request *ResourceRequestBean) (io.ReadCloser, error) {
	clusterId := request.ClusterId
	//getting rest config by clusterId
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resp, err := impl.k8sClientService.GetPodLogs(ctx, restConfig, request.K8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting events list", "err", err, "request", request)
		return nil, err
	}
	return resp, nil
}

func (impl *K8sApplicationServiceImpl) GetRestConfigByClusterId(ctx context.Context, clusterId int) (*rest.Config, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "K8sApplicationService.GetRestConfigByClusterId")
	defer span.End()
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId")
		return nil, err
	}
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	var restConfig *rest.Config
	if cluster.ClusterName == DEFAULT_CLUSTER && len(bearerToken) == 0 {
		restConfig, err = impl.K8sUtil.GetK8sClusterRestConfig()
		if err != nil {
			impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
			return nil, err
		}
	} else {
		restConfig = &rest.Config{Host: cluster.ServerUrl, BearerToken: bearerToken, TLSClientConfig: rest.TLSClientConfig{Insecure: true}}
	}
	return restConfig, nil
}

func (impl *K8sApplicationServiceImpl) GetRestConfigByCluster(ctx context.Context, cluster *cluster.ClusterBean) (*rest.Config, error) {
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	var restConfig *rest.Config
	var err error
	if cluster.ClusterName == DEFAULT_CLUSTER && len(bearerToken) == 0 {
		restConfig, err = impl.K8sUtil.GetK8sClusterRestConfig()
		if err != nil {
			impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
			return nil, err
		}
	} else {
		restConfig = &rest.Config{Host: cluster.ServerUrl, BearerToken: bearerToken, TLSClientConfig: rest.TLSClientConfig{Insecure: true}}
	}
	return restConfig, nil
}

func (impl *K8sApplicationServiceImpl) ValidateClusterResourceRequest(ctx context.Context, clusterResourceRequest *ResourceRequestBean,
	rbacCallback func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool) (bool, error) {
	clusterId := clusterResourceRequest.ClusterId
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster Id", "clusterId", clusterId, "err", err)
		return false, err
	}
	clusterName := clusterBean.ClusterName
	restConfig, err := impl.GetRestConfigByCluster(ctx, clusterBean)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "clusterId", clusterId, "err", err)
		return false, err
	}
	k8sRequest := clusterResourceRequest.K8sRequest
	respManifest, err := impl.k8sClientService.GetResource(ctx, restConfig, k8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err, "request", clusterResourceRequest)
		return false, err
	}
	return impl.validateResourceManifest(clusterName, respManifest.Manifest, k8sRequest.ResourceIdentifier.GroupVersionKind, rbacCallback), nil
}

func (impl *K8sApplicationServiceImpl) validateResourceManifest(clusterName string, resourceManifest unstructured.Unstructured, gvk schema.GroupVersionKind, rbacCallback func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool) bool {
	validateCallback := func(namespace, group, kind, resourceName string) bool {
		resourceIdentifier := application.ResourceIdentifier{
			Name:      resourceName,
			Namespace: namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group: group,
				Kind:  kind,
			},
		}
		return rbacCallback(clusterName, resourceIdentifier)
	}
	return impl.K8sUtil.ValidateResource(resourceManifest.Object, gvk, validateCallback)
}

func (impl *K8sApplicationServiceImpl) ValidateClusterResourceBean(ctx context.Context, clusterId int, manifest unstructured.Unstructured, gvk schema.GroupVersionKind, rbacCallback func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool) bool {
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster Id", "clusterId", clusterId, "err", err)
		return false
	}
	return impl.validateResourceManifest(clusterBean.ClusterName, manifest, gvk, rbacCallback)
}

func (impl *K8sApplicationServiceImpl) ValidateResourceRequest(ctx context.Context, appIdentifier *client.AppIdentifier, request *application.K8sRequestBean) (bool, error) {
	app, err := impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
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
	return impl.validateContainerNameIfReqd(valid, request, app), nil
}

func (impl *K8sApplicationServiceImpl) validateContainerNameIfReqd(valid bool, request *application.K8sRequestBean, app *client.AppDetail) bool {
	if !valid {
		requestContainerName := request.PodLogsRequest.ContainerName
		podName := request.ResourceIdentifier.Name
		for _, pod := range app.ResourceTreeResponse.PodMetadata {
			if pod.Name == podName {

				//finding the container name in main Containers
				for _, container := range pod.Containers {
					if container == requestContainerName {
						return true
					}
				}

				//finding the container name in init containers
				for _, initContainer := range pod.InitContainers {
					if initContainer == requestContainerName {
						return true
					}
				}

				//finding the container name in ephemeral containers
				for _, ephemeralContainer := range pod.EphemeralContainers {
					if ephemeralContainer == requestContainerName {
						return true
					}
				}

			}
		}
	}
	return valid
}

func (impl *K8sApplicationServiceImpl) GetResourceInfo(ctx context.Context) (*ResourceInfo, error) {
	pod, err := impl.K8sUtil.GetResourceInfoByLabelSelector(ctx, impl.aCDAuthConfig.ACDConfigMapNamespace, "app=inception")
	if err != nil {
		impl.logger.Errorw("error on getting resource from k8s, unable to fetch installer pod", "err", err)
		return nil, err
	}
	response := &ResourceInfo{PodName: pod.Name}
	return response, nil
}

func (impl *K8sApplicationServiceImpl) GetAllApiResources(ctx context.Context, clusterId int, isSuperAdmin bool, userId int32) (*application.GetAllApiResourcesResponse, error) {
	impl.logger.Infow("getting all api-resources", "clusterId", clusterId)
	restConfig, err := impl.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster rest config", "clusterId", clusterId, "err", err)
		return nil, err
	}
	allApiResources, err := impl.k8sClientService.GetApiResources(restConfig, LIST_VERB)
	if err != nil {
		return nil, err
	}

	// FILTER STARTS
	// 1) remove ""/v1 event kind if event kind exist in events.k8s.io/v1 and ""/v1
	k8sEventIndex := -1
	v1EventIndex := -1
	for index, apiResource := range allApiResources {
		gvk := apiResource.Gvk
		if gvk.Kind == EVENT_K8S_KIND && gvk.Version == "v1" {
			if gvk.Group == "" {
				v1EventIndex = index
			} else if gvk.Group == "events.k8s.io" {
				k8sEventIndex = index
			}
		}
	}
	if k8sEventIndex > -1 && v1EventIndex > -1 {
		allApiResources = append(allApiResources[:v1EventIndex], allApiResources[v1EventIndex+1:]...)
	}
	// FILTER ENDS

	// RBAC FILER STARTS
	allowedAll := isSuperAdmin
	filteredApiResources := make([]*application.K8sApiResource, 0)
	if !isSuperAdmin {
		clusterBean, err := impl.clusterService.FindById(clusterId)
		if err != nil {
			impl.logger.Errorw("failed to find cluster for id", "err", err, "clusterId", clusterId)
			return nil, err
		}
		roles, err := impl.clusterService.FetchRolesFromGroup(userId)
		if err != nil {
			impl.logger.Errorw("error on fetching user roles for cluster list", "err", err)
			return nil, err
		}

		allowedGroupKinds := make(map[string]bool) // group||kind
		for _, role := range roles {
			if clusterBean.ClusterName != role.Cluster {
				continue
			}
			kind := role.Kind
			if role.Group == "" && kind == "" {
				allowedAll = true
				break
			}
			groupName := role.Group
			if groupName == "" {
				groupName = "*"
			} else if groupName == casbin.ClusterEmptyGroupPlaceholder {
				groupName = ""
			}
			allowedGroupKinds[groupName+"||"+kind] = true
			// add children for this kind
			children, found := util.KindVsChildrenGvk[kind]
			if found {
				// if rollout kind other than argo, then neglect only
				if kind != util.K8sClusterResourceRolloutKind || groupName == util.K8sClusterResourceRolloutGroup {
					for _, child := range children {
						allowedGroupKinds[child.Group+"||"+child.Kind] = true
					}
				}
			}
		}

		if !allowedAll {
			for _, apiResource := range allApiResources {
				gvk := apiResource.Gvk
				_, found := allowedGroupKinds[gvk.Group+"||"+gvk.Kind]
				if found {
					filteredApiResources = append(filteredApiResources, apiResource)
				} else {
					_, found = allowedGroupKinds["*"+"||"+gvk.Kind]
					if found {
						filteredApiResources = append(filteredApiResources, apiResource)
					}
				}
			}
		}
	}
	response := &application.GetAllApiResourcesResponse{
		AllowedAll: allowedAll,
	}
	if allowedAll {
		response.ApiResources = allApiResources
	} else {
		response.ApiResources = filteredApiResources
	}
	// RBAC FILER ENDS

	return response, nil
}

func (impl *K8sApplicationServiceImpl) GetResourceList(ctx context.Context, token string, request *ResourceRequestBean, validateResourceAccess func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) (*util.ClusterResourceListMap, error) {
	resourceList := &util.ClusterResourceListMap{}
	clusterId := request.ClusterId
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by cluster Id", "err", err, "clusterId", clusterId)
		return resourceList, err
	}
	restConfig, err := impl.GetRestConfigByCluster(ctx, clusterBean)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", request.ClusterId)
		return resourceList, err
	}
	k8sRequest := request.K8sRequest
	//store the copy of requested resource identifier
	resourceIdentifierCloned := k8sRequest.ResourceIdentifier
	resp, namespaced, err := impl.k8sClientService.GetResourceList(ctx, restConfig, k8sRequest)
	if err != nil {
		impl.logger.Errorw("error in getting resource list", "err", err, "request", request)
		return resourceList, err
	}
	checkForResourceCallback := func(namespace, group, kind, resourceName string) bool {
		resourceIdentifier := resourceIdentifierCloned
		resourceIdentifier.Name = resourceName
		resourceIdentifier.Namespace = namespace
		if group != "" && kind != "" {
			resourceIdentifier.GroupVersionKind = schema.GroupVersionKind{Group: group, Kind: kind}
		}
		k8sRequest.ResourceIdentifier = resourceIdentifier
		return validateResourceAccess(token, clusterBean.ClusterName, *request, casbin.ActionGet)
	}
	resourceList, err = impl.K8sUtil.BuildK8sObjectListTableData(&resp.Resources, namespaced, request.K8sRequest.ResourceIdentifier.GroupVersionKind, checkForResourceCallback)
	if err != nil {
		impl.logger.Errorw("error on parsing for k8s resource", "err", err)
		return resourceList, err
	}
	return resourceList, nil
}

type RotatePodRequest struct {
	ClusterId int                              `json:"clusterId"`
	Resources []application.ResourceIdentifier `json:"resources"`
}

type RotatePodResponse struct {
	Responses     []*RotatePodResourceResponse `json:"responses"`
	ContainsError bool                         `json:"containsError"`
}

type RotatePodResourceResponse struct {
	application.ResourceIdentifier
	ErrorResponse string `json:"errorResponse"`
}

func (impl *K8sApplicationServiceImpl) RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error) {

	clusterId := request.ClusterId
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster Id", "clusterId", clusterId, "err", err)
		return nil, err
	}
	restConfig, err := impl.GetRestConfigByCluster(ctx, clusterBean)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "clusterId", clusterId, "err", err)
		return nil, err
	}
	response := &RotatePodResponse{}
	var resourceResponses []*RotatePodResourceResponse
	var containsError bool
	for _, resourceIdentifier := range request.Resources {
		resourceResponse := &RotatePodResourceResponse{
			ResourceIdentifier: resourceIdentifier,
		}
		groupVersionKind := resourceIdentifier.GroupVersionKind
		resourceKind := groupVersionKind.Kind
		// validate one of deployment, statefulset, daemonSet, Rollout
		if resourceKind != kube.DeploymentKind && resourceKind != kube.StatefulSetKind && resourceKind != kube.DaemonSetKind && resourceKind != util.K8sClusterResourceRolloutKind {
			impl.logger.Errorf("restarting not supported for kind %s name %s", resourceKind, resourceIdentifier.Name)
			containsError = true
			resourceResponse.ErrorResponse = util.RestartingNotSupported
		} else {
			activitySnapshot := time.Now().Format(time.RFC3339)
			data := fmt.Sprintf(`{"metadata": {"annotations": {"devtron.ai/restartedAt": "%s"}},"spec": {"template": {"metadata": {"annotations": {"devtron.ai/activity": "%s"}}}}}`, activitySnapshot, activitySnapshot)
			var patchType types.PatchType
			if resourceKind != util.K8sClusterResourceRolloutKind {
				patchType = types.StrategicMergePatchType
			} else {
				// rollout does not support strategic merge type
				patchType = types.MergePatchType
			}
			k8sRequest := &application.K8sRequestBean{ResourceIdentifier: resourceIdentifier}
			_, err = impl.k8sClientService.PatchResource(ctx, restConfig, patchType, k8sRequest, data)
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

func (impl *K8sApplicationServiceImpl) ApplyResources(ctx context.Context, token string, request *application.ApplyResourcesRequest, validateResourceAccess func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) ([]*application.ApplyResourcesResponse, error) {
	manifests, err := yamlUtil.SplitYAMLs([]byte(request.Manifest))
	if err != nil {
		impl.logger.Errorw("error in splitting yaml in manifest", "err", err)
		return nil, err
	}

	//getting rest config by clusterId
	clusterId := request.ClusterId
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster Id", "clusterId", clusterId, "err", err)
		return nil, err
	}
	restConfig, err := impl.GetRestConfigByCluster(ctx, clusterBean)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "clusterId", clusterId, "err", err)
		return nil, err
	}

	var response []*application.ApplyResourcesResponse
	for _, manifest := range manifests {
		var namespace string
		manifestNamespace := manifest.GetNamespace()
		if len(manifestNamespace) > 0 {
			namespace = manifestNamespace
		} else {
			namespace = DEFAULT_NAMESPACE
		}
		manifestRes := &application.ApplyResourcesResponse{
			Name: manifest.GetName(),
			Kind: manifest.GetKind(),
		}
		resourceRequestBean := ResourceRequestBean{
			ClusterId: clusterId,
			K8sRequest: &application.K8sRequestBean{
				ResourceIdentifier: application.ResourceIdentifier{
					Name:             manifest.GetName(),
					Namespace:        namespace,
					GroupVersionKind: manifest.GroupVersionKind(),
				},
			},
		}
		actionAllowed := validateResourceAccess(token, clusterBean.ClusterName, resourceRequestBean, casbin.ActionUpdate)
		if actionAllowed {
			resourceExists, err := impl.applyResourceFromManifest(ctx, manifest, restConfig, namespace)
			manifestRes.IsUpdate = resourceExists
			if err != nil {
				manifestRes.Error = err.Error()
			}
		} else {
			manifestRes.Error = "permission-denied"
		}
		response = append(response, manifestRes)
	}

	return response, nil
}

func (impl *K8sApplicationServiceImpl) applyResourceFromManifest(ctx context.Context, manifest unstructured.Unstructured, restConfig *rest.Config, namespace string) (bool, error) {
	var isUpdateResource bool
	k8sRequestBean := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name:             manifest.GetName(),
			Namespace:        namespace,
			GroupVersionKind: manifest.GroupVersionKind(),
		},
	}
	jsonStrByteErr, err := json.Marshal(manifest.UnstructuredContent())
	if err != nil {
		impl.logger.Errorw("error in marshalling json", "err", err)
		return isUpdateResource, err
	}
	jsonStr := string(jsonStrByteErr)
	_, err = impl.k8sClientService.GetResource(ctx, restConfig, k8sRequestBean)
	if err != nil {
		statusError, ok := err.(*errors2.StatusError)
		if !ok || statusError == nil || statusError.ErrStatus.Reason != metav1.StatusReasonNotFound {
			impl.logger.Errorw("error in getting resource", "err", err)
			return isUpdateResource, err
		}
		// case of resource not found
		_, err = impl.k8sClientService.CreateResource(ctx, restConfig, k8sRequestBean, jsonStr)
		if err != nil {
			impl.logger.Errorw("error in creating resource", "err", err)
			return isUpdateResource, err
		}
	} else {
		// case of resource update
		isUpdateResource = true
		_, err = impl.k8sClientService.ApplyResource(ctx, restConfig, k8sRequestBean, jsonStr)
		if err != nil {
			impl.logger.Errorw("error in updating resource", "err", err)
			return isUpdateResource, err
		}
	}

	return isUpdateResource, nil
}

func (impl *K8sApplicationServiceImpl) FetchConnectionStatusForCluster(k8sClientSet *kubernetes.Clientset, clusterId int) error {
	//using livez path as healthz path is deprecated
	path := "/livez"
	response, err := k8sClientSet.Discovery().RESTClient().Get().AbsPath(path).DoRaw(context.Background())
	log.Println("received response for cluster livez status", "response", string(response), "err", err, "clusterId", clusterId)
	if err == nil && string(response) != "ok" {
		err = fmt.Errorf("ErrorNotOk : response != 'ok' : %s", string(response))
	}
	return err
}
