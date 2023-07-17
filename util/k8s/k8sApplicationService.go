package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/version"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	yamlUtil "github.com/devtron-labs/devtron/util/yaml"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/url"
)

const (
	DEFAULT_CLUSTER = "default_cluster"
)

type K8sApplicationService interface {
	ValidatePodLogsRequestQuery(r *http.Request) (*ResourceRequestBean, error)
	ValidateTerminalRequestQuery(r *http.Request) (*terminal.TerminalSessionRequest, *ResourceRequestBean, error)
	DecodeDevtronAppId(applicationId string) (*DevtronAppIdentifier, error)
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
	GetManifestsByBatch(ctx context.Context, request []ResourceRequestBean) ([]BatchResourceResponse, error)
	FilterServiceAndIngress(ctx context.Context, resourceTreeInf map[string]interface{}, validRequests []ResourceRequestBean, appDetail bean.AppDetailContainer, appId string) []ResourceRequestBean
	GetUrlsByBatch(ctx context.Context, resp []BatchResourceResponse) []interface{}
	GetAllApiResources(ctx context.Context, clusterId int, isSuperAdmin bool, userId int32) (*application.GetAllApiResourcesResponse, error)
	GetResourceList(ctx context.Context, token string, request *ResourceRequestBean, validateResourceAccess func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) (*util.ClusterResourceListMap, error)
	ApplyResources(ctx context.Context, token string, request *application.ApplyResourcesRequest, resourceRbacHandler func(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool) ([]*application.ApplyResourcesResponse, error)
	FetchConnectionStatusForCluster(k8sClientSet *kubernetes.Clientset, clusterId int) error
	RotatePods(ctx context.Context, request *RotatePodRequest) (*RotatePodResponse, error)
	CreatePodEphemeralContainers(req *cluster.EphemeralContainerRequest) error
	TerminatePodEphemeralContainer(req cluster.EphemeralContainerRequest) (bool, error)
	GetPodContainersList(clusterId int, namespace, podName string) (*PodContainerList, error)
	GetK8sServerVersion(clusterId int) (*version.Info, error)
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
	terminalSession             terminal.TerminalSessionHandler
	ephemeralContainerService   cluster.EphemeralContainerService
}

type K8sApplicationServiceConfig struct {
	BatchSize        int `env:"BATCH_SIZE" envDefault:"5"`
	TimeOutInSeconds int `env:"TIMEOUT_IN_SECONDS" envDefault:"5"`
}

func NewK8sApplicationServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	pump connector.Pump, k8sClientService application.K8sClientService,
	helmAppService client.HelmAppService, K8sUtil *util.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig,
	K8sResourceHistoryService kubernetesResourceAuditLogs.K8sResourceHistoryService,
	terminalSession terminal.TerminalSessionHandler,
	ephemeralContainerService cluster.EphemeralContainerService) *K8sApplicationServiceImpl {
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
		terminalSession:             terminalSession,
		ephemeralContainerService:   ephemeralContainerService,
	}
}

const (
	// App Type Identifiers
	DevtronAppType = 0 // Identifier for Devtron Apps
	HelmAppType    = 1 // Identifier for Helm Apps

	// Deployment Type Identifiers
	HelmInstalledType = 0 // Identifier for Helm deployment
	ArgoInstalledType = 1 // Identifier for ArgoCD deployment
)

type ResourceRequestBean struct {
	AppId                string                      `json:"appId"`
	AppType              int                         `json:"appType,omitempty"`        // 0: DevtronApp, 1: HelmApp
	DeploymentType       int                         `json:"deploymentType,omitempty"` // 0: DevtronApp, 1: HelmApp
	AppIdentifier        *client.AppIdentifier       `json:"-"`
	K8sRequest           *application.K8sRequestBean `json:"k8sRequest"`
	DevtronAppIdentifier *DevtronAppIdentifier       `json:"-"`         // For Devtron App Resources
	ClusterId            int                         `json:"clusterId"` // clusterId is used when request is for direct cluster (not for helm release)
}

type ResourceInfo struct {
	PodName string `json:"podName"`
}

type DevtronAppIdentifier struct {
	ClusterId int `json:"clusterId"`
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
}

type BatchResourceResponse struct {
	ManifestResponse *application.ManifestResponse
	Err              error
}

type PodContainerList struct {
	Containers          []string
	InitContainers      []string
	EphemeralContainers []string
}

func (impl *K8sApplicationServiceImpl) ValidatePodLogsRequestQuery(r *http.Request) (*ResourceRequestBean, error) {
	v, vars := r.URL.Query(), mux.Vars(r)
	request := &ResourceRequestBean{}
	podName := vars["podName"]
	/*sinceSeconds, err := strconv.Atoi(v.Get("sinceSeconds"))
	if err != nil {
		sinceSeconds = 0
	}*/
	containerName, clusterIdString := v.Get("containerName"), v.Get("clusterId")
	prevContainerLogs := v.Get("previous")
	isPrevLogs, err := strconv.ParseBool(prevContainerLogs)
	if err != nil {
		isPrevLogs = false
	}
	appId := v.Get("appId")
	follow, err := strconv.ParseBool(v.Get("follow"))
	if err != nil {
		follow = false
	}
	tailLines, err := strconv.Atoi(v.Get("tailLines"))
	if err != nil {
		tailLines = 0
	}
	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name:             podName,
			GroupVersionKind: schema.GroupVersionKind{},
		},
		PodLogsRequest: application.PodLogsRequest{
			//SinceTime:     sinceSeconds,
			TailLines:                  tailLines,
			Follow:                     follow,
			ContainerName:              containerName,
			IsPrevContainerLogsEnabled: isPrevLogs,
		},
	}
	request.K8sRequest = k8sRequest
	if appId != "" {
		// Validate App Type
		appType, err := strconv.Atoi(v.Get("appType"))
		if err != nil || !(appType == DevtronAppType || appType == HelmAppType) {
			impl.logger.Errorw("Invalid appType", "err", err, "appType", appType)
			return nil, err
		}
		request.AppType = appType
		// Validate Deployment Type
		deploymentType, err := strconv.Atoi(v.Get("deploymentType"))
		if err != nil || !(deploymentType == HelmInstalledType || deploymentType == ArgoInstalledType) {
			impl.logger.Errorw("Invalid deploymentType", "err", err, "deploymentType", deploymentType)
			return nil, err
		}
		request.DeploymentType = deploymentType
		// Validate App Id
		if request.AppType == HelmAppType {
			// For Helm App resources
			appIdentifier, err := impl.helmAppService.DecodeAppId(appId)
			if err != nil {
				impl.logger.Errorw("error in decoding appId", "err", err, "appId", appId)
				return nil, err
			}
			request.AppIdentifier = appIdentifier
			request.ClusterId = appIdentifier.ClusterId
			request.K8sRequest.ResourceIdentifier.Namespace = appIdentifier.Namespace
		} else if request.AppType == DevtronAppType {
			// For Devtron App resources
			devtronAppIdentifier, err := impl.DecodeDevtronAppId(appId)
			if err != nil {
				impl.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
				return nil, err
			}
			request.DevtronAppIdentifier = devtronAppIdentifier
			request.ClusterId = devtronAppIdentifier.ClusterId
			namespace := v.Get("namespace")
			if namespace == "" {
				err = fmt.Errorf("missing required field namespace")
				impl.logger.Errorw("empty namespace", "err", err, "appId", request.AppId)
				return nil, err
			}
			request.K8sRequest.ResourceIdentifier.Namespace = namespace
		}
	} else if clusterIdString != "" {
		// Validate Cluster Id
		clusterId, err := strconv.Atoi(clusterIdString)
		if err != nil {
			impl.logger.Errorw("invalid cluster id", "clusterId", clusterIdString, "err", err)
			return nil, err
		}
		request.ClusterId = clusterId
		namespace := v.Get("namespace")
		if namespace == "" {
			err = fmt.Errorf("missing required field namespace")
			impl.logger.Errorw("empty namespace", "err", err, "appId", request.AppId)
			return nil, err
		}
		request.K8sRequest.ResourceIdentifier.Namespace = namespace
		request.K8sRequest.ResourceIdentifier.GroupVersionKind = schema.GroupVersionKind{
			Group:   "",
			Kind:    "Pod",
			Version: "v1",
		}
	}
	return request, nil
}

func (impl *K8sApplicationServiceImpl) ValidateTerminalRequestQuery(r *http.Request) (*terminal.TerminalSessionRequest, *ResourceRequestBean, error) {
	request := &terminal.TerminalSessionRequest{}
	v := r.URL.Query()
	vars := mux.Vars(r)
	request.ContainerName = vars["container"]
	request.Namespace = vars["namespace"]
	request.PodName = vars["pod"]
	request.Shell = vars["shell"]
	resourceRequestBean := &ResourceRequestBean{}
	identifier := vars["identifier"]
	if strings.Contains(identifier, "|") {
		// Validate App Type
		appType, err := strconv.Atoi(v.Get("appType"))
		if err != nil || appType < DevtronAppType && appType > HelmAppType {
			impl.logger.Errorw("Invalid appType", "err", err, "appType", appType)
			return nil, nil, err
		}
		request.ApplicationId = identifier
		if appType == HelmAppType {
			appIdentifier, err := impl.helmAppService.DecodeAppId(request.ApplicationId)
			if err != nil {
				impl.logger.Errorw("invalid app id", "err", err, "appId", request.ApplicationId)
				return nil, nil, err
			}
			resourceRequestBean.AppIdentifier = appIdentifier
			resourceRequestBean.ClusterId = appIdentifier.ClusterId
			request.ClusterId = appIdentifier.ClusterId
		} else if appType == DevtronAppType {
			devtronAppIdentifier, err := impl.DecodeDevtronAppId(request.ApplicationId)
			if err != nil {
				impl.logger.Errorw("invalid app id", "err", err, "appId", request.ApplicationId)
				return nil, nil, err
			}
			resourceRequestBean.DevtronAppIdentifier = devtronAppIdentifier
			resourceRequestBean.ClusterId = devtronAppIdentifier.ClusterId
			request.ClusterId = devtronAppIdentifier.ClusterId
		}
	} else {
		// Validate Cluster Id
		clsuterId, err := strconv.Atoi(identifier)
		if err != nil || clsuterId <= 0 {
			impl.logger.Errorw("Invalid cluster id", "err", err, "clusterId", identifier)
			return nil, nil, err
		}
		resourceRequestBean.ClusterId = clsuterId
		request.ClusterId = clsuterId
		k8sRequest := &application.K8sRequestBean{
			ResourceIdentifier: application.ResourceIdentifier{
				Name:      request.PodName,
				Namespace: request.Namespace,
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "",
					Kind:    "Pod",
					Version: "v1",
				},
			},
		}
		resourceRequestBean.K8sRequest = k8sRequest
	}
	return request, resourceRequestBean, nil
}

func (impl *K8sApplicationServiceImpl) DecodeDevtronAppId(applicationId string) (*DevtronAppIdentifier, error) {
	component := strings.Split(applicationId, "|")
	if len(component) != 3 {
		return nil, fmt.Errorf("malformed app id %s", applicationId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	appId, err := strconv.Atoi(component[1])
	if err != nil {
		return nil, err
	}
	envId, err := strconv.Atoi(component[2])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 || appId <= 0 || envId <= 0 {
		return nil, fmt.Errorf("invalid app identifier")
	}
	return &DevtronAppIdentifier{
		ClusterId: clusterId,
		AppId:     appId,
		EnvId:     envId,
	}, nil
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
	clusterConfig := cluster.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
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
	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
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
	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
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
	k8sServerVersion, err := impl.GetK8sServerVersion(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting k8s server version", "clusterId", clusterId, "err", err)
		//return nil, err
	} else {
		resourceList.ServerVersion = k8sServerVersion.String()
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
	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
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
	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(&clusterConfig)
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
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			err = fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*errors2.StatusError); ok {
			if statusError != nil {
				errReason := statusError.ErrStatus.Reason
				var errMsg string
				if errReason == metav1.StatusReasonUnauthorized {
					errMsg = "token seems invalid or does not have sufficient permissions"
				} else {
					errMsg = statusError.ErrStatus.Message
				}
				err = fmt.Errorf("%s : %s", errReason, errMsg)
			} else {
				err = fmt.Errorf("Validation failed : %v", err)
			}
		} else {
			err = fmt.Errorf("Validation failed : %v", err)
		}
	} else if err == nil && string(response) != "ok" {
		err = fmt.Errorf("Validation failed with response : %s", string(response))
	}
	return err
}

func (impl *K8sApplicationServiceImpl) CreatePodEphemeralContainers(req *cluster.EphemeralContainerRequest) error {

	clientSet, v1Client, err := impl.getCoreClientByClusterId(req.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting coreV1 client by clusterId", "clusterId", req.ClusterId, "err", err)
		return err
	}
	compatible, err := impl.K8sUtil.K8sServerVersionCheckForEphemeralContainers(clientSet)
	if err != nil {
		impl.logger.Errorw("error in checking kubernetes server version compatability for ephemeral containers", "clusterId", req.ClusterId, "err", err)
		return err
	}
	if !compatible {
		return errors.New("This feature is supported on and above Kubernetes v1.23 only.")
	}
	pod, err := impl.K8sUtil.GetPodByName(req.Namespace, req.PodName, v1Client)
	if err != nil {
		impl.logger.Errorw("error in getting pod", "clusterId", req.ClusterId, "namespace", req.Namespace, "podName", req.PodName, "err", err)
		return err
	}

	podJS, err := json.Marshal(pod)
	if err != nil {
		impl.logger.Errorw("error occurred in unMarshaling pod object", "podObject", pod, "err", err)
		return fmt.Errorf("error creating JSON for pod: %v", err)
	}
	debugPod, debugContainer, err := impl.generateDebugContainer(pod, *req)
	if err != nil {
		impl.logger.Errorw("error in generateDebugContainer", "request", req, "err", err)
		return err
	}

	debugJS, err := json.Marshal(debugPod)
	if err != nil {
		impl.logger.Errorw("error occurred in unMarshaling debugPod object", "debugPod", debugPod, "err", err)
		return fmt.Errorf("error creating JSON for pod: %v", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(podJS, debugJS, pod)
	if err != nil {
		impl.logger.Errorw("error occurred in CreateTwoWayMergePatch", "podJS", podJS, "debugJS", debugJS, "pod", pod, "err", err)
		return fmt.Errorf("error creating patch to add debug container: %v", err)
	}

	_, err = v1Client.Pods(req.Namespace).Patch(context.Background(), pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		if serr, ok := err.(*errors2.StatusError); ok && serr.Status().Reason == metav1.StatusReasonNotFound && serr.ErrStatus.Details.Name == "" {
			impl.logger.Errorw("error occurred while creating ephemeral containers", "err", err, "reason", "ephemeral containers are disabled for this cluster")
			return fmt.Errorf("ephemeral containers are disabled for this cluster (error from kubernetes server: %q)", err)
		}
		if runtime.IsNotRegisteredError(err) {
			patch, err := json.Marshal([]map[string]interface{}{{
				"op":    "add",
				"path":  "/ephemeralContainers/-",
				"value": debugContainer,
			}})
			if err != nil {
				impl.logger.Errorw("error occured while trying to create epehemral containers with legacy API", "err", err)
				return fmt.Errorf("error creating JSON 6902 patch for old /ephemeralcontainers API: %s", err)
			}
			//try with legacy API
			result := v1Client.RESTClient().Patch(types.JSONPatchType).
				Namespace(pod.Namespace).
				Resource("pods").
				Name(pod.Name).
				SubResource("ephemeralcontainers").
				Body(patch).
				Do(context.Background())
			return result.Error()
		}
		return err
	}

	if err == nil {
		req.AdvancedData = &cluster.EphemeralContainerAdvancedData{
			Manifest: string(debugJS),
		}
		req.BasicData = &cluster.EphemeralContainerBasicData{
			ContainerName:       debugContainer.Name,
			TargetContainerName: debugContainer.TargetContainerName,
			Image:               debugContainer.Image,
		}
		err = impl.ephemeralContainerService.AuditEphemeralContainerAction(*req, repository.ActionCreate)
		if err != nil {
			impl.logger.Errorw("error in saving ephemeral container data", "err", err)
			return err
		}
	}

	impl.logger.Errorw("error in creating ephemeral containers ", "err", err, "clusterId", req.ClusterId, "namespace", req.Namespace, "podName", req.PodName, "ephemeralContainerSpec", debugContainer)
	return err
}

func (impl *K8sApplicationServiceImpl) generateDebugContainer(pod *corev1.Pod, req cluster.EphemeralContainerRequest) (*corev1.Pod, *corev1.EphemeralContainer, error) {
	copied := pod.DeepCopy()
	ephemeralContainer := &corev1.EphemeralContainer{}
	if req.AdvancedData != nil {
		err := json.Unmarshal([]byte(req.AdvancedData.Manifest), ephemeralContainer)
		if err != nil {
			impl.logger.Errorw("error occurred i unMarshaling advanced ephemeral data", "err", err, "advancedData", req.AdvancedData.Manifest)
			return copied, ephemeralContainer, err
		}
		if ephemeralContainer.TargetContainerName == "" || ephemeralContainer.Name == "" || ephemeralContainer.Image == "" {
			return copied, ephemeralContainer, errors.New("containerName,targetContainerName and image cannot be empty")
		}
	} else {
		ephemeralContainer = &corev1.EphemeralContainer{
			EphemeralContainerCommon: corev1.EphemeralContainerCommon{
				Name:                     req.BasicData.ContainerName,
				Env:                      nil,
				Image:                    req.BasicData.Image,
				ImagePullPolicy:          corev1.PullIfNotPresent,
				Stdin:                    true,
				TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				TTY:                      true,
			},
			TargetContainerName: req.BasicData.TargetContainerName,
		}
	}
	ephemeralContainer.Name = ephemeralContainer.Name + "-" + util2.Generate(5)
	scriptCreateCommand := fmt.Sprintf("echo 'while true; do sleep 600; done;' > %s-devtron.sh", ephemeralContainer.Name)
	scriptRunCommand := fmt.Sprintf("sh %s-devtron.sh", ephemeralContainer.Name)
	ephemeralContainer.Command = []string{"sh", "-c", scriptCreateCommand + " && " + scriptRunCommand}
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ephemeralContainer)
	ephemeralContainer = &copied.Spec.EphemeralContainers[len(copied.Spec.EphemeralContainers)-1]
	return copied, ephemeralContainer, nil

}

func (impl *K8sApplicationServiceImpl) TerminatePodEphemeralContainer(req cluster.EphemeralContainerRequest) (bool, error) {
	terminalReq := &terminal.TerminalSessionRequest{
		PodName:       req.PodName,
		ClusterId:     req.ClusterId,
		Namespace:     req.Namespace,
		ContainerName: req.BasicData.ContainerName,
	}
	
	containerKillCommand := fmt.Sprintf("kill -16 $(ps aux | awk '/%s-devtron/ {print $2; exit}')", terminalReq.ContainerName)
	cmds := []string{"sh", "-c", containerKillCommand}
	_, errBuf, err := impl.terminalSession.RunCmdInRemotePod(terminalReq, cmds)
	if err != nil {
		impl.logger.Errorw("failed to execute commands ", "err", err, "commands", cmds, "podName", req.PodName, "namespace", req.Namespace)
		return false, err
	}
	errBufString := errBuf.String()
	if errBufString != "" {
		impl.logger.Errorw("error response on executing commands ", "err", errBufString, "commands", cmds, "podName", req.Namespace, "namespace", req.Namespace)
		return false, err
	}

	if err == nil {

		err = impl.ephemeralContainerService.AuditEphemeralContainerAction(req, repository.ActionTerminate)
		if err != nil {
			impl.logger.Errorw("error in saving ephemeral container data", "err", err)
			return true, err
		}

	}

	return true, nil
}

func (impl *K8sApplicationServiceImpl) getCoreClientByClusterId(clusterId int) (*kubernetes.Clientset, *v1.CoreV1Client, error) {
	clusterBean, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error occurred in finding clusterBean by Id", "clusterId", clusterId, "err", err)
		return nil, nil, err
	}

	clusterConfig := clusterBean.GetClusterConfig()
	v1Client, err := impl.K8sUtil.GetClient(&clusterConfig)
	if err != nil {
		//not logging clusterConfig as it contains sensitive data
		impl.logger.Errorw("error occurred in getting v1Client with cluster config", "err", err, "clusterId", clusterId)
		return nil, nil, err
	}
	clientSet, err := impl.K8sUtil.GetClientSet(&clusterConfig)
	if err != nil {
		//not logging clusterConfig as it contains sensitive data
		impl.logger.Errorw("error occurred in getting clientSet with cluster config", "err", err, "clusterId", clusterId)
		return nil, v1Client, err
	}
	return clientSet, v1Client, nil
}

func (impl *K8sApplicationServiceImpl) GetPodContainersList(clusterId int, namespace, podName string) (*PodContainerList, error) {
	_, v1Client, err := impl.getCoreClientByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting coreV1 client by clusterId", "clusterId", clusterId, "err", err)
		return nil, err
	}
	pod, err := impl.K8sUtil.GetPodByName(namespace, podName, v1Client)
	if err != nil {
		impl.logger.Errorw("error in getting pod", "clusterId", clusterId, "namespace", namespace, "podName", podName, "err", err)
		return nil, err
	}
	ephemeralContainerStatusMap := make(map[string]bool)
	for _, c := range pod.Status.EphemeralContainerStatuses {
		//c.state contains three states running,waiting and terminated
		// at any point of time only one state will be there
		if c.State.Running != nil {
			ephemeralContainerStatusMap[c.Name] = true
		}
	}
	containers := make([]string, len(pod.Spec.Containers))
	initContainers := make([]string, len(pod.Spec.InitContainers))
	ephemeralContainers := make([]string, 0, len(pod.Spec.EphemeralContainers))

	for i, c := range pod.Spec.Containers {
		containers[i] = c.Name
	}

	for _, ec := range pod.Spec.EphemeralContainers {
		if _, ok := ephemeralContainerStatusMap[ec.Name]; ok {
			ephemeralContainers = append(ephemeralContainers, ec.Name)
		}
	}

	for i, ic := range pod.Spec.InitContainers {
		initContainers[i] = ic.Name
	}

	return &PodContainerList{
		Containers:          containers,
		EphemeralContainers: ephemeralContainers,
		InitContainers:      initContainers,
	}, nil
}

func (impl *K8sApplicationServiceImpl) GetK8sServerVersion(clusterId int) (*version.Info, error) {
	clientSet, _, err := impl.getCoreClientByClusterId(clusterId)
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
