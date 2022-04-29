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
	"encoding/json"
	error2 "errors"
	"flag"
	"os/user"
	"path/filepath"
	"time"

	"github.com/devtron-labs/authenticator/client"
	"github.com/ghodss/yaml"
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
)

type K8sUtil struct {
	logger        *zap.SugaredLogger
	runTimeConfig *client.RuntimeConfig
	kubeconfig    *string
}

type ClusterConfig struct {
	Host        string
	BearerToken string
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

func (impl K8sUtil) GetClient(clusterConfig *ClusterConfig) (*v12.CoreV1Client, error) {
	cfg := &rest.Config{}
	cfg.Host = clusterConfig.Host
	cfg.BearerToken = clusterConfig.BearerToken
	cfg.Insecure = true
	client, err := v12.NewForConfig(cfg)
	return client, err
}

func (impl K8sUtil) GetClientSet(clusterConfig *ClusterConfig) (*kubernetes.Clientset, error) {
	cfg := &rest.Config{}
	cfg.Host = clusterConfig.Host
	cfg.BearerToken = clusterConfig.BearerToken
	cfg.Insecure = true
	client, err := kubernetes.NewForConfig(cfg)
	return client, err
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
	clientset, err := v12.NewForConfig(config)
	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
	return clientset, err
}

func (impl K8sUtil) GetK8sDiscoveryClient(clusterConfig *ClusterConfig) (*discovery.DiscoveryClient, error) {
	cfg := &rest.Config{}
	cfg.Host = clusterConfig.Host
	cfg.BearerToken = clusterConfig.BearerToken
	cfg.Insecure = true
	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return nil, err
	}
	return client, err
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
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
	return client, err
}

func (impl K8sUtil) CreateNsIfNotExists(namespace string, clusterConfig *ClusterConfig) (err error) {
	client, err := impl.GetClient(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return err
	}
	exists, err := impl.checkIfNsExists(namespace, client)
	if err != nil {
		impl.logger.Errorw("error", "error", err, "clusterConfig", clusterConfig)
		return err
	}
	if exists {
		return nil
	}
	impl.logger.Infow("ns not exists creating", "ns", namespace)
	_, err = impl.createNs(namespace, client)
	return err
}

func (impl K8sUtil) checkIfNsExists(namespace string, client *v12.CoreV1Client) (exists bool, err error) {
	ns, err := client.Namespaces().Get(namespace, metav1.GetOptions{})
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
	ns, err = client.Namespaces().Create(nsSpec)
	if err != nil {
		return nil, err
	} else {
		return ns, nil
	}
}

func (impl K8sUtil) deleteNs(namespace string, client *v12.CoreV1Client) error {
	err := client.Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	return err
}

func (impl K8sUtil) GetConfigMap(namespace string, name string, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) CreateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Create(cm)
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) UpdateConfigMap(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Update(cm)
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) PatchConfigMap(namespace string, clusterConfig *ClusterConfig, name string, data map[string]interface{}) (*v1.ConfigMap, error) {
	client, err := impl.GetClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	cm, err := client.ConfigMaps(namespace).Patch(name, types.PatchType(types.MergePatchType), b)
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
	return cm, nil
}

func (impl K8sUtil) PatchConfigMapJsonType(namespace string, clusterConfig *ClusterConfig, name string, data interface{}, path string) (*v1.ConfigMap, error) {
	client, err := impl.GetClient(clusterConfig)
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

	cm, err := client.ConfigMaps(namespace).Patch(name, types.PatchType(types.JSONPatchType), b)
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
	secret, err := client.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) CreateSecret(namespace string, data map[string][]byte, secretName string, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: data,
	}
	secret, err := client.Secrets(namespace).Create(secret)
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) UpdateSecret(namespace string, secret *v1.Secret, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret, err := client.Secrets(namespace).Update(secret)
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) DeleteJob(namespace string, name string, clusterConfig *ClusterConfig) error {
	clientSet, err := impl.GetClientSet(clusterConfig)
	if err != nil {
		impl.logger.Errorw("clientSet err, DeleteJob", "err", err)
		return err
	}
	jobs := clientSet.BatchV1().Jobs(namespace)

	job, err := jobs.Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("get job err, DeleteJob", "err", err)
		return nil
	}

	if job != nil {
		err := jobs.Delete(name, &metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("delete err, DeleteJob", "err", err)
			return err
		}
	}

	return nil
}

func (impl K8sUtil) CreateJob(namespace string, name string, clusterConfig *ClusterConfig, job *batchV1.Job) error {
	clientSet, err := impl.GetClientSet(clusterConfig)
	if err != nil {
		impl.logger.Errorw("clientSet err, CreateJob", "err", err)
	}
	time.Sleep(5 * time.Second)

	jobs := clientSet.BatchV1().Jobs(namespace)
	_, err = jobs.Get(name, metav1.GetOptions{})
	if err == nil {
		impl.logger.Errorw("get job err, CreateJob", "err", err)
		time.Sleep(5 * time.Second)
		_, err = jobs.Get(name, metav1.GetOptions{})
		if err == nil {
			return error2.New("job deletion takes more time than expected, please try after sometime")
		}
	}

	_, err = jobs.Create(job)
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
	podList, err := pods.List(metav1.ListOptions{LabelSelector: labels})
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("get pod err, DeletePod", "err", err)
		return nil
	}

	for _, pod := range (*podList).Items {
		if pod.Status.Phase != Running {
			podName := pod.ObjectMeta.Name
			err := pods.Delete(podName, &metav1.DeleteOptions{})
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
	nsList, err := client.Namespaces().List(metav1.ListOptions{})
	if errors.IsNotFound(err) {
		return nsList, nil
	} else if err != nil {
		return nsList, err
	} else {
		return nsList, nil
	}
}

func (impl K8sUtil) GetClientByToken(serverUrl string, token map[string]string) (*v12.CoreV1Client, error) {
	bearerToken := token["bearer_token"]
	clusterCfg := &ClusterConfig{Host: serverUrl, BearerToken: bearerToken}
	client, err := impl.GetClient(clusterCfg)
	if err != nil {
		impl.logger.Errorw("error in k8s client", "error", err)
		return nil, err
	}
	return client, nil
}

func (impl K8sUtil) GetResourceInfoByLabelSelector(namespace string, labelSelector string) (*v1.Pod, error) {
	client, err := impl.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("cluster config error", "err", err)
		return nil, err
	}
	pods, err := client.Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	} else if len(pods.Items) > 1 {
		err = &ApiError{Code: "406", HttpStatusCode: 200, UserMessage: "found more than one pod for label selector"}
		return nil, err
	} else if len(pods.Items) == 0 {
		err = &ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no pod found for label selector"}
		return nil, err
	} else {
		return &pods.Items[0], nil
	}
}
