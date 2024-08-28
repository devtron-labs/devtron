/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package informer

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"
)

func NewGlobalMapClusterNamespace() sync.Map {
	var globalMapClusterNamespace sync.Map
	return globalMapClusterNamespace
}

type K8sInformerFactoryImpl struct {
	logger                    *zap.SugaredLogger
	globalMapClusterNamespace sync.Map // {"cluster1":{"ns1":true","ns2":true"}}
	informerStopper           map[string]chan struct{}
	k8sUtil                   *k8s.K8sServiceImpl
}

type K8sInformerFactory interface {
	GetLatestNamespaceListGroupByCLuster() map[string]map[string]bool
	BuildInformer(clusterInfo []*bean.ClusterInfo)
	CleanNamespaceInformer(clusterName string)
}

func NewK8sInformerFactoryImpl(logger *zap.SugaredLogger, globalMapClusterNamespace sync.Map, k8sUtil *k8s.K8sServiceImpl) *K8sInformerFactoryImpl {
	informerFactory := &K8sInformerFactoryImpl{
		logger:                    logger,
		globalMapClusterNamespace: globalMapClusterNamespace,
		k8sUtil:                   k8sUtil,
	}
	informerFactory.informerStopper = make(map[string]chan struct{})
	return informerFactory
}

func (impl *K8sInformerFactoryImpl) GetLatestNamespaceListGroupByCLuster() map[string]map[string]bool {
	copiedClusterNamespaces := make(map[string]map[string]bool)
	impl.globalMapClusterNamespace.Range(func(key, value interface{}) bool {
		clusterName := key.(string)
		allNamespaces := value.(*sync.Map)
		namespaceMap := make(map[string]bool)
		allNamespaces.Range(func(nsKey, nsValue interface{}) bool {
			namespaceMap[nsKey.(string)] = nsValue.(bool)
			return true
		})
		copiedClusterNamespaces[clusterName] = namespaceMap
		return true
	})
	return copiedClusterNamespaces
}

func (impl *K8sInformerFactoryImpl) BuildInformer(clusterInfo []*bean.ClusterInfo) {
	for _, info := range clusterInfo {
		clusterConfig := &k8s.ClusterConfig{
			ClusterName:           info.ClusterName,
			BearerToken:           info.BearerToken,
			Host:                  info.ServerUrl,
			InsecureSkipTLSVerify: info.InsecureSkipTLSVerify,
			KeyData:               info.KeyData,
			CertData:              info.CertData,
			CAData:                info.CAData,
		}
		impl.buildInformerAndNamespaceList(info.ClusterName, clusterConfig)
	}
	return
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, clusterConfig *k8s.ClusterConfig) sync.Map {
	allNamespaces := sync.Map{}
	impl.globalMapClusterNamespace.Store(clusterName, &allNamespaces)
	_, _, clusterClient, err := impl.k8sUtil.GetK8sConfigAndClients(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s clientset", "err", err, "clusterName", clusterConfig.ClusterName)
		return impl.globalMapClusterNamespace
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute)
	nsInformer := informerFactory.Core().V1().Namespaces()
	stopper := make(chan struct{})
	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if mobject, ok := obj.(metav1.Object); ok {
				value, _ := impl.globalMapClusterNamespace.Load(clusterName)
				allNamespaces := value.(*sync.Map)
				allNamespaces.Store(mobject.GetName(), true)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if object, ok := obj.(metav1.Object); ok {
				value, _ := impl.globalMapClusterNamespace.Load(clusterName)
				allNamespaces := value.(*sync.Map)
				allNamespaces.Delete(object.GetName())
			}
		},
	})
	informerFactory.Start(stopper)
	impl.informerStopper[clusterName] = stopper
	return impl.globalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) CleanNamespaceInformer(clusterName string) {
	stopper := impl.informerStopper[clusterName]
	if stopper != nil {
		close(stopper)
		delete(impl.informerStopper, clusterName)
	}
	return
}
