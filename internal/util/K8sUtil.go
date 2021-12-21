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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	error2 "errors"
	"fmt"
	"github.com/ghodss/yaml"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type K8sUtilimpl interface {
}

type K8sUtil struct {
	logger          *zap.SugaredLogger
	ChartWorkingDir ChartWorkingDir
}

type ClusterConfig struct {
	Host        string
	BearerToken string
}

func NewK8sUtil(logger *zap.SugaredLogger, ChartWorkingDir ChartWorkingDir) *K8sUtil {
	return &K8sUtil{logger: logger, ChartWorkingDir: ChartWorkingDir}
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

func (impl K8sUtil) GetClientForInCluster() (*v12.CoreV1Client, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		impl.logger.Errorw("error", "error", err)
		return nil, err
	}
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
	config, err := rest.InClusterConfig()
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

const watcherRestart = "restart event watcher"

func (impl K8sUtil) WatchConfigMap(namespace string, dir string, client kubernetes.Interface) (string, string, []byte, error) {
	api := client.CoreV1().ConfigMaps(namespace)
	configMaps, err := api.List(metav1.ListOptions{})
	if err != nil {
		return "", "", nil, err
	}

	resourceVersion := configMaps.ListMeta.ResourceVersion
	timeout := int64(1800)
	watcher, err := api.Watch(metav1.ListOptions{ResourceVersion: resourceVersion, TimeoutSeconds: &timeout})
	if err != nil {
		return "", "", nil, err
	}

	ch := watcher.ResultChan()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				// the channel got closed, so we need to restart
				return "", "", nil, error2.New(watcherRestart)
			}
			configMaps, ok := event.Object.(*coreV1.ConfigMap)
			if !ok {
				return "", "", nil, nil
			}
			annotations, ok := configMaps.Annotations["charts.devtron.ai/data"]
			if !ok {
				continue
			}

			if annotations == "mount" {
				for filename, binaryData := range configMaps.BinaryData {
					binaryDataReader := bytes.NewReader(binaryData)
					chartDir := filepath.Join(string(impl.ChartWorkingDir), dir)
					err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
					if err != nil {
						impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
						return "", "", nil, err
					}

					err = impl.ExtractTarGz(binaryDataReader, chartDir)
					if err != nil {
						impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
						return "", "", nil, err
					}
					configmapDirectoryName := strings.Split(filename, ".")

					readFile, err := ioutil.ReadFile(filepath.Join(string(impl.ChartWorkingDir), configmapDirectoryName[0], "Chart.Yaml"))
					if err != nil {
						return "", "", nil, err
					}

					chartContent, err := chartutil.UnmarshalChartfile(readFile)
					if err != nil {
						return "", "", nil, err
					}
					configMapName := chartContent.Name
					configMapVersion := chartContent.Version

					return configMapName, configMapVersion, binaryData, nil
				}

			}

		case <-time.After(30 * time.Minute):
			return "", "", nil, error2.New(watcherRestart)
		}
	}

	return "", "", nil, err
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

func (impl K8sUtil) ExtractTarGz(gzipStream io.Reader, chartDir string) error{
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(filepath.Join(chartDir, header.Name)); os.IsNotExist(err) {
				if err := os.Mkdir(filepath.Join(chartDir, header.Name), 0755); err != nil {
					return err
				}
			} else {
				break
			}

		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(chartDir, header.Name))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()

		default:
			return err

		}

	}
}
