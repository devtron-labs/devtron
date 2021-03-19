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
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type K8sUtil struct {
	logger *zap.SugaredLogger
}

type ClusterConfig struct {
	Host        string
	BearerToken string
}

func NewK8sUtil(logger *zap.SugaredLogger) *K8sUtil {
	return &K8sUtil{logger: logger}
}

func (impl K8sUtil) GetClient(clusterConfig *ClusterConfig) (*v12.CoreV1Client, error) {
	cfg := &rest.Config{}
	cfg.Host = clusterConfig.Host
	cfg.BearerToken = clusterConfig.BearerToken
	cfg.Insecure = true
	client, err := v12.NewForConfig(cfg)

	return client, err
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

func (impl K8sUtil) CreateNsIfNotExists(namespace string, clusterConfig *ClusterConfig) (err error) {
	client, err := impl.GetClient(clusterConfig)
	if err != nil {
		return err
	}
	exists, err := impl.checkIfNsExists(namespace, client)
	if err != nil {
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

func (impl K8sUtil) GetConfigMap(namespace string, name string, clusterConfig *ClusterConfig) (*v1.ConfigMap, error) {
	client, err := impl.GetClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	cm, err := client.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) GetConfigMapFast(namespace string, name string, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
	cm, err := client.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) UpdateConfigMap(namespace string, cm *v1.ConfigMap, clusterConfig *ClusterConfig) (*v1.ConfigMap, error) {
	client, err := impl.GetClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	cm, err = client.ConfigMaps(namespace).Update(cm)
	if err != nil {
		return nil, err
	} else {
		return cm, nil
	}
}

func (impl K8sUtil) UpdateConfigMapFast(namespace string, cm *v1.ConfigMap, client *v12.CoreV1Client) (*v1.ConfigMap, error) {
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

func (impl K8sUtil) GetSecretFast(namespace string, name string, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret, err := client.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func (impl K8sUtil) CreateSecretFast(namespace string, data map[string][]byte, secretName string, client *v12.CoreV1Client) (*v1.Secret, error) {
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

func (impl K8sUtil) UpdateSecretFast(namespace string, secret *v1.Secret, client *v12.CoreV1Client) (*v1.Secret, error) {
	secret, err := client.Secrets(namespace).Update(secret)
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}
