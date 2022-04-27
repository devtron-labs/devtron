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

package module

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type ModuleCacheService interface {
}

type ModuleCacheServiceImpl struct {
	logger           *zap.SugaredLogger
	mutex            sync.Mutex
	K8sUtil          *util.K8sUtil
	moduleEnvConfig  *ModuleEnvConfig
	serverEnvConfig  *serverEnvConfig.ServerEnvConfig
	moduleDataStore  *ModuleDataStore
	serverDataStore  *serverDataStore.ServerDataStore
	moduleRepository ModuleRepository
}

func NewModuleCacheServiceImpl(logger *zap.SugaredLogger, K8sUtil *util.K8sUtil, moduleEnvConfig *ModuleEnvConfig, serverEnvConfig *serverEnvConfig.ServerEnvConfig, moduleDataStore *ModuleDataStore,
	serverDataStore *serverDataStore.ServerDataStore, moduleRepository ModuleRepository) *ModuleCacheServiceImpl {
	impl := &ModuleCacheServiceImpl{
		logger:           logger,
		K8sUtil:          K8sUtil,
		moduleEnvConfig:  moduleEnvConfig,
		serverEnvConfig:  serverEnvConfig,
		moduleDataStore:  moduleDataStore,
		serverDataStore:  serverDataStore,
		moduleRepository: moduleRepository,
	}

	// build informer to listen on installer object
	go impl.buildInformerToListenOnInstallerObject()

	return impl
}

func (impl *ModuleCacheServiceImpl) buildInformerToListenOnInstallerObject() {
	impl.logger.Debug("building informer cache to listen on installer object")
	clusterConfig, err := impl.K8sUtil.GetK8sClusterRestConfig()
	if err != nil {
		log.Fatalln("not able to get k8s cluster rest config.", "error", err)
	}

	clusterClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln("not able to get config from rest config.", "error", err)
	}

	installerResource := schema.GroupVersionResource{
		Group:    impl.serverEnvConfig.InstallerCrdObjectGroupName,
		Version:  impl.serverEnvConfig.InstallerCrdObjectVersion,
		Resource: impl.serverEnvConfig.InstallerCrdObjectResource,
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		clusterClient, time.Minute, impl.serverEnvConfig.InstallerCrdNamespace, nil)
	informer := factory.ForResource(installerResource).Informer()
	
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.handleInstallerObjectChange(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			impl.handleInstallerObjectChange(newObj)
		},
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go informer.Run(ctx.Done())
	<-ctx.Done()
}

func (impl *ModuleCacheServiceImpl) handleInstallerObjectChange(obj interface{}) {
	u := obj.(*unstructured.Unstructured)
	val, _, _ := unstructured.NestedString(u.Object, "status", "sync", "status")
	impl.serverDataStore.InstallerCrdObjectStatus = val

	// if installer status is applied and server is of full mode and status is notInstalled then check ciCd module from DB
	// if status is installing or unknown then save as installed in DB
	// if status is in installing/unknown state, then mark it as installed and update in-memory
	// if status is installed, then update in-memory
	if !impl.moduleDataStore.IsCiCdModuleInstalled && val == serverBean.InstallerCrdObjectStatusApplied && util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_FULL {
		ciCdModule, err := impl.moduleRepository.FindOne(ModuleCiCdName)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching ciCd module", "err", err)
			return
		}
		if ciCdModule.Status == ModuleStatusInstalling || ciCdModule.Status == ModuleStatusUnknown {
			ciCdModule.Status = ModuleStatusInstalled
			ciCdModule.UpdatedOn = time.Now()
			err = impl.moduleRepository.Update(ciCdModule)
			if err != nil {
				impl.logger.Errorw("error in updating module status to installed", "name", ciCdModule.Name, "err", err)
			} else {
				impl.moduleDataStore.IsCiCdModuleInstalled = true
			}
		} else if ciCdModule.Status == ModuleStatusInstalled {
			impl.moduleDataStore.IsCiCdModuleInstalled = true
		}
	}
}
