/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"context"
	"encoding/json"
	error2 "errors"
	"flag"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	v13 "k8s.io/api/policy/v1"
	v1beta12 "k8s.io/api/policy/v1beta1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"net/http"
	"net/url"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"

	"github.com/devtron-labs/authenticator/client"
	"go.uber.org/zap"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type K8sUtil struct {
	logger        *zap.SugaredLogger
	runTimeConfig *client.RuntimeConfig
	kubeconfig    *string
}

type ClusterConfig struct {
	ClusterName           string
	Host                  string
	BearerToken           string
	InsecureSkipTLSVerify bool
	KeyData               string
	CertData              string
	CAData                string
}

func NewK8sUtil(logger *zap.SugaredLogger, runTimeConfig *client.RuntimeConfig) *K8sUtil {
	usr, err := user.Current()
	if err != nil {
		return nil
	}
	var kubeconfig *string
	if runTimeConfig.LocalDevMode {
		kubeconfig = flag.String("kubeconfig-authenticator-xyz", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}

	flag.Parse()
	return &K8sUtil{logger: logger, runTimeConfig: runTimeConfig, kubeconfig: kubeconfig}
}

func (impl K8sUtil) GetRestConfigByCluster(configMap *ClusterConfig) (*rest.Config, error) {
	bearerToken := configMap.BearerToken
	var restConfig *rest.Config
	var err error
	if configMap.ClusterName == DefaultCluster && len(bearerToken) == 0 {
		restConfig, err = impl.GetK8sClusterRestConfig()
		if err != nil {
			impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
			return nil, err
		}
	} else {
		restConfig = &rest.Config{Host: configMap.Host, BearerToken: bearerToken, TLSClientConfig: rest.TLSClientConfig{Insecure: configMap.InsecureSkipTLSVerify}}
		if configMap.InsecureSkipTLSVerify == false {
			restConfig.TLSClientConfig.ServerName = restConfig.ServerName
			restConfig.TLSClientConfig.KeyData = []byte(configMap.KeyData)
			restConfig.TLSClientConfig.CertData = []byte(configMap.CertData)
			restConfig.TLSClientConfig.CAData = []byte(configMap.CAData)
		}
	}
	return restConfig, nil
}

func (impl K8sUtil) GetClient(clusterConfig *ClusterConfig) (*v12.CoreV1Client, error) {
	cfg, err := impl.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
		return nil, err
	}
	httpClient, err := OverrideK8sHttpClientWithTracer(cfg)
	if err != nil {
		return nil, err
	}
	v12Client, err := v12.NewForConfigAndClient(cfg, httpClient)
	return v12Client, err
}

func (impl K8sUtil) GetClientSet(clusterConfig *ClusterConfig) (*kubernetes.Clientset, error) {
	cfg, err := impl.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
		return nil, err
	}
	httpClient, err := OverrideK8sHttpClientWithTracer(cfg)
	if err != nil {
		return nil, err
	}
	kubernetesClient, err := kubernetes.NewForConfigAndClient(cfg, httpClient)
	return kubernetesClient, err
}

func (impl K8sUtil) getKubeConfig(devMode client.LocalDevMode) (*rest.Config, error) {
	if devMode {
		restConfig, err := clientcmd.BuildConfigFromFlags("", *impl.kubeconfig)
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	} else {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	}
}

func (impl K8sUtil) GetClientForInCluster() (*v12.CoreV1Client, error) {
	// creates the in-cluster config
	config, err := impl.getKubeConfig(impl.runTimeConfig.LocalDevMode)
	// creates the clientset
	httpClient, err := OverrideK8sHttpClientWithTracer(config)
	if err != nil {
		return nil, err
	}
	clientset, err := v12.NewForConfigAndClient(config, httpClient)
	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
	return clientset, err
}

func (impl K8sUtil) GetK8sClient() (*v12.CoreV1Client, error) {
	var config *rest.Config
	var err error
	if impl.runTimeConfig.LocalDevMode {
		config, err = clientcmd.BuildConfigFromFlags("", *impl.kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		impl.logger.Errorw("error fetching cluster config", "error", err)
		return nil, err
	}
	httpClient, err := OverrideK8sHttpClientWithTracer(config)
	if err != nil {
		return nil, err
	}
	v12Client, err := v12.NewForConfigAndClient(config, httpClient)
	if err != nil {
		impl.logger.Errorw("error creating k8s client", "error", err)
		return nil, err
	}
	return v12Client, err
}

func (impl K8sUtil) GetK8sDiscoveryClient(clusterConfig *ClusterConfig) (*discovery.DiscoveryClient, error) {
	cfg, err := impl.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config for default cluster", "err", err)
		return nil, err
	}
	httpClient, err := OverrideK8sHttpClientWithTracer(cfg)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfigAndClient(cfg, httpClient)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return nil, err
	}
	return discoveryClient, err
}

func (impl K8sUtil) GetK8sDiscoveryClientInCluster() (*discovery.DiscoveryClient, error) {
	var config *rest.Config
	var err error
	if impl.runTimeConfig.LocalDevMode {
		config, err = clientcmd.BuildConfigFromFlags("", *impl.kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
	httpClient, err := OverrideK8sHttpClientWithTracer(config)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfigAndClient(config, httpClient)
	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
	return discoveryClient, err
}

func (impl K8sUtil) CreateNsIfNotExists(namespace string, clusterConfig *ClusterConfig) (err error) {
	v12Client, err := impl.GetClient(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return err
	}
	exists, err := impl.checkIfNsExists(namespace, v12Client)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return err
	}
	if exists {
		return nil
	}
	impl.logger.Infow("ns not exists creating", "ns", namespace)
	_, err = impl.createNs(namespace, v12Client)
	return err
}

func (impl K8sUtil) checkIfNsExists(namespace string, client *v12.CoreV1Client) (exists bool, err error) {
	ns, err := client.Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	//ns, err := impl.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	impl.logger.Debugw("ns fetch", "name", namespace, "res", ns)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}

}

func (impl K8sUtil) createNs(namespace string, client *v12.CoreV1Client) (ns *v1.Namespace, err error) {
	nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	ns, err = client.Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		return ns, nil
	}
}

func (impl K8sUtil) deleteNs(namespace string, client *v12.CoreV1Client) error {
	err := client.Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	return err
}

func (impl K8sUtil) GetConfigMap(namespace string, name string, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) CreateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) UpdateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) PatchConfigMap(namespace string, clusterConfig *ClusterConfig, name string, data map[string]interface{}) (*v1.ConfigMap, error) {
	k8sClient, err := impl.GetClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	cm, err := k8sClient.ConfigMaps(namespace).Patch(context.Background(), name, types.PatchType(types.MergePatchType), b, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
	return cm, nil
}

func (impl K8sUtil) PatchConfigMapJsonType(namespace string, clusterConfig *ClusterConfig, name string, data interface{}, path string) (*v1.ConfigMap, error) {
	v12Client, err := impl.GetClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	var patches []*JsonPatchType
	patch := &JsonPatchType{
		Op:    "replace",
		Path:  path,
		Value: data,
	}
	patches = append(patches, patch)
	b, err := json.Marshal(patches)
	if err != nil {
		panic(err)
	}

	cm, err := v12Client.ConfigMaps(namespace).Patch(context.Background(), name, types.PatchType(types.JSONPatchType), b, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
	return cm, nil
}

type JsonPatchType struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func (impl K8sUtil) GetSecret(namespace string, name string, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret, err := client.Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) CreateSecret(namespace string, data map[string][]byte, secretName string, secretType v1.SecretType, client *v12.CoreV1Client, labels map[string]string, stringData map[string]string) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
	}
	if labels != nil && len(labels) > 0 {
		secret.ObjectMeta.Labels = labels
	}
	if stringData != nil && len(stringData) > 0 {
		secret.StringData = stringData
	}
	if data != nil && len(data) > 0 {
		secret.Data = data
	}
	if len(secretType) > 0 {
		secret.Type = secretType
	}
	secret, err := client.Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) UpdateSecret(namespace string, secret *v1.Secret, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret, err := client.Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) DeleteSecret(namespace string, name string, client *v12.CoreV1Client) error {
	err := client.Secrets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (impl K8sUtil) DeleteJob(namespace string, name string, clusterConfig *ClusterConfig) error {
	clientSet, err := impl.GetClientSet(clusterConfig)
	if err != nil {
		impl.logger.Errorw("clientSet err, DeleteJob", "err", err)
		return err
	}
	jobs := clientSet.BatchV1().Jobs(namespace)

	job, err := jobs.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("get job err, DeleteJob", "err", err)
		return nil
	}

	if job != nil {
		err := jobs.Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("delete err, DeleteJob", "err", err)
			return err
		}
	}

	return nil
}

func (impl K8sUtil) GetK8sConfigAndClients(clusterConfig *ClusterConfig) (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	restConfig, err := impl.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, nil, nil, err
	}
	k8sHttpClient, err := OverrideK8sHttpClientWithTracer(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s http client set by rest config", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, nil, nil, err
	}
	k8sClientSet, err := kubernetes.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, nil, nil, err
	}
	return restConfig, k8sHttpClient, k8sClientSet, nil
}

func (impl K8sUtil) DiscoveryClientGetLiveZCall(cluster *ClusterConfig) ([]byte, error) {
	_, _, k8sClientSet, err := impl.GetK8sConfigAndClients(cluster)
	if err != nil {
		impl.logger.Errorw("errir in getting clients and configs", "err", err, "clusterName", cluster.ClusterName)
		return nil, err
	}
	//using livez path as healthz path is deprecated
	response, err := impl.GetLiveZCall(LiveZ, k8sClientSet)
	return response, err

}
func (impl K8sUtil) GetLiveZCall(path string, k8sClientSet *kubernetes.Clientset) ([]byte, error) {
	response, err := k8sClientSet.Discovery().RESTClient().Get().AbsPath(path).DoRaw(context.Background())
	return response, err
}

func (impl K8sUtil) CreateJob(namespace string, name string, clusterConfig *ClusterConfig, job *batchV1.Job) error {
	clientSet, err := impl.GetClientSet(clusterConfig)
	if err != nil {
		impl.logger.Errorw("clientSet err, CreateJob", "err", err)
	}
	time.Sleep(5 * time.Second)

	jobs := clientSet.BatchV1().Jobs(namespace)
	_, err = jobs.Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		impl.logger.Errorw("get job err, CreateJob", "err", err)
		time.Sleep(5 * time.Second)
		_, err = jobs.Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return error2.New("job deletion takes more time than expected, please try after sometime")
		}
	}

	_, err = jobs.Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		impl.logger.Errorw("create err, CreateJob", "err", err)
		return err
	}
	return nil
}

// DeletePod delete pods with label job-name

const Running = "Running"

func (impl K8sUtil) DeletePodByLabel(namespace string, labels string, clusterConfig *ClusterConfig) error {
	clientSet, err := impl.GetClientSet(clusterConfig)
	if err != nil {
		impl.logger.Errorw("clientSet err, DeletePod", "err", err)
		return err
	}

	time.Sleep(2 * time.Second)

	pods := clientSet.CoreV1().Pods(namespace)
	podList, err := pods.List(context.Background(), metav1.ListOptions{LabelSelector: labels})
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("get pod err, DeletePod", "err", err)
		return nil
	}

	for _, pod := range (*podList).Items {
		if pod.Status.Phase != Running {
			podName := pod.ObjectMeta.Name
			err := pods.Delete(context.Background(), podName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				impl.logger.Errorw("delete err, DeletePod", "err", err)
				return err
			}
		}
	}
	return nil
}

// DeleteAndCreateJob Deletes and recreates if job exists else creates the job
func (impl K8sUtil) DeleteAndCreateJob(content []byte, namespace string, clusterConfig *ClusterConfig) error {
	// Job object from content
	var job batchV1.Job
	err := yaml.Unmarshal(content, &job)
	if err != nil {
		impl.logger.Errorw("Unmarshal err, CreateJobSafely", "err", err)
		return err
	}

	// delete job if exists
	err = impl.DeleteJob(namespace, job.Name, clusterConfig)
	if err != nil {
		impl.logger.Errorw("DeleteJobIfExists err, CreateJobSafely", "err", err)
		return err
	}

	labels := "job-name=" + job.Name
	err = impl.DeletePodByLabel(namespace, labels, clusterConfig)
	if err != nil {
		impl.logger.Errorw("DeleteJobIfExists err, CreateJobSafely", "err", err)
		return err
	}
	// create job
	err = impl.CreateJob(namespace, job.Name, clusterConfig, &job)
	if err != nil {
		impl.logger.Errorw("CreateJob err, CreateJobSafely", "err", err)
		return err
	}

	return nil
}

func (impl K8sUtil) ListNamespaces(client *v12.CoreV1Client) (*v1.NamespaceList, error) {
	nsList, err := client.Namespaces().List(context.Background(), metav1.ListOptions{})
	if errors.IsNotFound(err) {
		return nsList, nil
	} else if err != nil {
		return nsList, err
	} else {
		return nsList, nil
	}
}

func (impl K8sUtil) GetClientByToken(serverUrl string, token map[string]string) (*v12.CoreV1Client, error) {
	bearerToken := token[BearerToken]
	clusterCfg := &ClusterConfig{Host: serverUrl, BearerToken: bearerToken}
	v12Client, err := impl.GetClient(clusterCfg)
	if err != nil {
		impl.logger.Errorw("error in k8s client", "error", err)
		return nil, err
	}
	return v12Client, nil
}

func (impl K8sUtil) GetResourceInfoByLabelSelector(ctx context.Context, namespace string, labelSelector string) (*v1.Pod, error) {
	inClusterClient, err := impl.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("cluster config error", "err", err)
		return nil, err
	}
	pods, err := inClusterClient.Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	} else if len(pods.Items) > 1 {
		err = &util.ApiError{Code: "406", HttpStatusCode: 200, UserMessage: "found more than one pod for label selector"}
		return nil, err
	} else if len(pods.Items) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no pod found for label selector"}
		return nil, err
	} else {
		return &pods.Items[0], nil
	}
}

func (impl K8sUtil) GetK8sClusterRestConfig() (*rest.Config, error) {
	impl.logger.Debug("getting k8s rest config")
	if impl.runTimeConfig.LocalDevMode {
		restConfig, err := clientcmd.BuildConfigFromFlags("", *impl.kubeconfig)
		if err != nil {
			impl.logger.Errorw("Error while building kubernetes cluster rest config", "error", err)
			return nil, err
		}
		return restConfig, nil
	} else {
		clusterConfig, err := rest.InClusterConfig()
		if err != nil {
			impl.logger.Errorw("error in fetch default cluster config", "err", err)
			return nil, err
		}
		return clusterConfig, nil
	}
}

func (impl K8sUtil) GetPodByName(namespace string, name string, client *v12.CoreV1Client) (*v1.Pod, error) {
	pod, err := client.Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error in fetch pod name", "err", err)
		return nil, err
	} else {
		return pod, nil
	}
}

func (impl K8sUtil) BuildK8sObjectListTableData(manifest *unstructured.UnstructuredList, namespaced bool, gvk schema.GroupVersionKind, validateResourceAccess func(namespace string, group string, kind string, resourceName string) bool) (*ClusterResourceListMap, error) {
	clusterResourceListMap := &ClusterResourceListMap{}
	// build headers
	var headers []string
	columnIndexes := make(map[int]string)
	kind := gvk.Kind
	if kind == "Event" {
		headers, columnIndexes = impl.getEventKindHeader()
	} else {
		columnDefinitionsUncast := manifest.Object[K8sClusterResourceColumnDefinitionKey]
		if columnDefinitionsUncast != nil {
			columnDefinitions := columnDefinitionsUncast.([]interface{})
			for index, cd := range columnDefinitions {
				if cd == nil {
					continue
				}
				columnMap := cd.(map[string]interface{})
				columnNameUncast := columnMap[K8sClusterResourceNameKey]
				if columnNameUncast == nil {
					continue
				}
				priorityUncast := columnMap[K8sClusterResourcePriorityKey]
				if priorityUncast == nil {
					continue
				}
				columnName := columnNameUncast.(string)
				columnName = strings.ToLower(columnName)
				priority := priorityUncast.(int64)
				if namespaced && index == 1 {
					headers = append(headers, K8sClusterResourceNamespaceKey)
				}
				if priority == 0 || (manifest.GetKind() == "Event" && columnName == "source") {
					columnIndexes[index] = columnName
					headers = append(headers, columnName)
				}
			}
		}
	}

	// build rows
	rowsMapping := make([]map[string]interface{}, 0)
	rowsDataUncast := manifest.Object[K8sClusterResourceRowsKey]
	var namespace string
	var allowed bool
	if rowsDataUncast != nil {
		rows := rowsDataUncast.([]interface{})
		for _, row := range rows {
			namespace = ""
			allowed = true
			rowIndex := make(map[string]interface{})
			rowMap := row.(map[string]interface{})
			cellsUncast := rowMap[K8sClusterResourceCellKey]
			if cellsUncast == nil {
				continue
			}
			rowCells := cellsUncast.([]interface{})
			for index, columnName := range columnIndexes {
				cellValUncast := rowCells[index]
				var cell interface{}
				if cellValUncast == nil {
					cell = ""
				} else {
					cell = cellValUncast.(interface{})
				}
				rowIndex[columnName] = cell
			}

			cellObjUncast := rowMap[K8sClusterResourceObjectKey]
			var cellObj map[string]interface{}
			if cellObjUncast != nil {
				cellObj = cellObjUncast.(map[string]interface{})
				if cellObj != nil && cellObj[K8sClusterResourceMetadataKey] != nil {
					metadata := cellObj[K8sClusterResourceMetadataKey].(map[string]interface{})
					if metadata[K8sClusterResourceNamespaceKey] != nil {
						namespace = metadata[K8sClusterResourceNamespaceKey].(string)
						if namespaced {
							rowIndex[K8sClusterResourceNamespaceKey] = namespace
						}
					}
				}
			}
			allowed = impl.ValidateResource(cellObj, gvk, validateResourceAccess)
			if allowed {
				rowsMapping = append(rowsMapping, rowIndex)
			}
		}
	}

	clusterResourceListMap.Headers = headers
	clusterResourceListMap.Data = rowsMapping
	impl.logger.Debugw("resource listing response", "clusterResourceListMap", clusterResourceListMap)
	return clusterResourceListMap, nil
}

func (impl K8sUtil) ValidateResource(resourceObj map[string]interface{}, gvk schema.GroupVersionKind, validateCallback func(namespace string, group string, kind string, resourceName string) bool) bool {
	resKind := gvk.Kind
	groupName := gvk.Group
	metadata := resourceObj[K8sClusterResourceMetadataKey]
	if metadata == nil {
		return false
	}
	metadataMap := metadata.(map[string]interface{})
	var namespace, resourceName string
	var ownerReferences []interface{}
	if metadataMap[K8sClusterResourceNamespaceKey] != nil {
		namespace = metadataMap[K8sClusterResourceNamespaceKey].(string)
	}
	if metadataMap[K8sClusterResourceMetadataNameKey] != nil {
		resourceName = metadataMap[K8sClusterResourceMetadataNameKey].(string)
	}
	if metadataMap[K8sClusterResourceOwnerReferenceKey] != nil {
		ownerReferences = metadataMap[K8sClusterResourceOwnerReferenceKey].([]interface{})
	}
	if len(ownerReferences) > 0 {
		for _, ownerRef := range ownerReferences {
			allowed := impl.validateForResource(namespace, ownerRef, validateCallback)
			if allowed {
				return allowed
			}
		}
	}
	// check current RBAC in case not matched with above one
	return validateCallback(namespace, groupName, resKind, resourceName)
}

func (impl K8sUtil) validateForResource(namespace string, resourceRef interface{}, validateCallback func(namespace string, group string, kind string, resourceName string) bool) bool {
	resourceReference := resourceRef.(map[string]interface{})
	resKind := resourceReference[K8sClusterResourceKindKey].(string)
	apiVersion := resourceReference[K8sClusterResourceApiVersionKey].(string)
	groupName := ""
	if strings.Contains(apiVersion, "/") {
		groupName = apiVersion[:strings.LastIndex(apiVersion, "/")] // extracting group from this apiVersion
	}
	resName := ""
	if resourceReference["name"] != "" {
		resName = resourceReference["name"].(string)
		switch resKind {
		case kube.ReplicaSetKind:
			// check deployment first, then RO and then RS
			if strings.Contains(resName, "-") {
				deploymentName := resName[:strings.LastIndex(resName, "-")]
				allowed := validateCallback(namespace, groupName, kube.DeploymentKind, deploymentName)
				if allowed {
					return true
				}
				allowed = validateCallback(namespace, K8sClusterResourceRolloutGroup, K8sClusterResourceRolloutKind, deploymentName)
				if allowed {
					return true
				}
			}
			allowed := validateCallback(namespace, groupName, resKind, resName)
			if allowed {
				return true
			}
		case kube.JobKind:
			// check CronJob first, then Job
			if strings.Contains(resName, "-") {
				cronJobName := resName[:strings.LastIndex(resName, "-")]
				allowed := validateCallback(namespace, groupName, K8sClusterResourceCronJobKind, cronJobName)
				if allowed {
					return true
				}
			}
			allowed := validateCallback(namespace, groupName, resKind, resName)
			if allowed {
				return true
			}
		case kube.DeploymentKind, K8sClusterResourceCronJobKind, kube.StatefulSetKind, kube.DaemonSetKind, K8sClusterResourceRolloutKind, K8sClusterResourceReplicationControllerKind:
			allowed := validateCallback(namespace, groupName, resKind, resName)
			if allowed {
				return true
			}
		}
	}
	return false
}

func (impl K8sUtil) getEventKindHeader() ([]string, map[int]string) {
	headers := []string{"type", "message", "namespace", "involved object", "source", "count", "age", "last seen"}
	columnIndexes := make(map[int]string)
	columnIndexes[0] = "last seen"
	columnIndexes[1] = "type"
	columnIndexes[2] = "namespace"
	columnIndexes[3] = "involved object"
	columnIndexes[5] = "source"
	columnIndexes[6] = "message"
	columnIndexes[7] = "age"
	columnIndexes[8] = "count"
	return headers, columnIndexes
}

func OverrideK8sHttpClientWithTracer(restConfig *rest.Config) (*http.Client, error) {
	httpClientFor, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		fmt.Println("error occurred while overriding k8s client", "reason", err)
		return nil, err
	}
	httpClientFor.Transport = otelhttp.NewTransport(httpClientFor.Transport)
	return httpClientFor, nil
}
func (impl K8sUtil) GetKubeVersion() (*version.Info, error) {
	discoveryClient, err := impl.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("eexception caught in getting discoveryClient", "err", err)
		return nil, err
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	return k8sServerVersion, err
}

func (impl K8sUtil) GetConfigAndClientsInCluster() (*v12.CoreV1Client, error) {
	restConfig := &rest.Config{}
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		impl.logger.Error("Error in creating config for default cluster", "err", err)
		return nil, err
	}
	httpClientFor, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		impl.logger.Error("error occurred while overriding k8s client", "reason", err)
		return nil, err
	}
	k8sClient, err := v12.NewForConfigAndClient(restConfig, httpClientFor)
	if err != nil {
		impl.logger.Error("error creating k8s client", "error", err)
		return nil, err
	}
	return k8sClient, err
}
func (impl K8sUtil) GetNodesList(ctx context.Context, k8sClientSet *kubernetes.Clientset) (*v1.NodeList, error) {
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	return nodeList, err
}
func (impl K8sUtil) GetNodeByName(ctx context.Context, k8sClientSet *kubernetes.Clientset, name string) (*v1.Node, error) {
	node, err := k8sClientSet.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	return node, err
}

func (impl K8sUtil) GetServerVersionFromDiscoveryClient(k8sClientSet *kubernetes.Clientset) (*version.Info, error) {
	serverVersion, err := k8sClientSet.DiscoveryClient.ServerVersion()
	return serverVersion, err
}
func (impl K8sUtil) GetPodsListForAllNamespaces(ctx context.Context, k8sClientSet *kubernetes.Clientset) (*v1.PodList, error) {
	podList, err := k8sClientSet.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{})
	return podList, err
}
func (impl K8sUtil) GetNmList(ctx context.Context, metricsClientSet *metrics.Clientset) (*v1beta1.NodeMetricsList, error) {
	nmList, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	return nmList, err
}
func (impl K8sUtil) GetNmByName(ctx context.Context, metricsClientSet *metrics.Clientset, name string) (*v1beta1.NodeMetrics, error) {
	nodeMetrics, err := metricsClientSet.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
	return nodeMetrics, err
}
func (impl K8sUtil) GetMetricsClientSet(restConfig *rest.Config, k8sHttpClient *http.Client) (*metrics.Clientset, error) {
	metricsClientSet, err := metrics.NewForConfigAndClient(restConfig, k8sHttpClient)
	return metricsClientSet, err
}

// DeletePod will delete the given pod, or return an error if it couldn't
func DeletePod(pod v1.Pod, k8sClientSet *kubernetes.Clientset, deleteOptions metav1.DeleteOptions) error {
	return k8sClientSet.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, deleteOptions)
}

// EvictPod will evict the given pod, or return an error if it couldn't
func EvictPod(pod v1.Pod, k8sClientSet *kubernetes.Clientset, evictionGroupVersion schema.GroupVersion, deleteOptions metav1.DeleteOptions) error {
	switch evictionGroupVersion {
	case v13.SchemeGroupVersion:
		// send policy/v1 if the server supports it
		eviction := &v13.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &deleteOptions,
		}
		return k8sClientSet.PolicyV1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)

	default:
		// otherwise, fall back to policy/v1beta1, supported by all servers that support the eviction subresource
		eviction := &v1beta12.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &deleteOptions,
		}
		return k8sClientSet.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
	}
}

// CheckEvictionSupport uses Discovery API to find out if the server support
// eviction subresource If support, it will return its groupVersion; Otherwise,
// it will return an empty GroupVersion
func CheckEvictionSupport(clientset kubernetes.Interface) (schema.GroupVersion, error) {
	discoveryClient := clientset.Discovery()

	// version info available in subresources since v1.8.0 in https://github.com/kubernetes/kubernetes/pull/49971
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return schema.GroupVersion{}, err
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == EvictionSubresource && resource.Kind == EvictionKind && len(resource.Group) > 0 && len(resource.Version) > 0 {
			return schema.GroupVersion{Group: resource.Group, Version: resource.Version}, nil
		}
	}
	return schema.GroupVersion{}, nil
}

func UpdateNodeUnschedulableProperty(desiredUnschedulable bool, node *v1.Node, k8sClientSet *kubernetes.Clientset) (*v1.Node, error) {
	node.Spec.Unschedulable = desiredUnschedulable
	node, err := k8sClientSet.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	return node, err
}
func (impl K8sUtil) CreateK8sClientSet(restConfig *rest.Config) (*kubernetes.Clientset, error) {
	k8sHttpClient, err := OverrideK8sHttpClientWithTracer(restConfig)
	if err != nil {
		impl.logger.Errorw("service err, OverrideK8sHttpClientWithTracer", "err", err, "restConfig", restConfig)
		return nil, err
	}
	k8sClientSet, err := kubernetes.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
		return nil, err
	}
	return k8sClientSet, err
}

func (impl K8sUtil) FetchConnectionStatusForCluster(k8sClientSet *kubernetes.Clientset, clusterId int) error {
	//using livez path as healthz path is deprecated
	path := LiveZ
	response, err := k8sClientSet.Discovery().RESTClient().Get().AbsPath(path).DoRaw(context.Background())
	log.Println("received response for cluster livez status", "response", string(response), "err", err, "clusterId", clusterId)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			err = fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*errors.StatusError); ok {
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
