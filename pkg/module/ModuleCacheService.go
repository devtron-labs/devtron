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
	client "github.com/devtron-labs/devtron/api/helm-app"
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
	serverDataStore  *serverDataStore.ServerDataStore
	moduleRepository ModuleRepository
	helmAppService   client.HelmAppService
}

func NewModuleCacheServiceImpl(logger *zap.SugaredLogger, K8sUtil *util.K8sUtil, moduleEnvConfig *ModuleEnvConfig, serverEnvConfig *serverEnvConfig.ServerEnvConfig,
	serverDataStore *serverDataStore.ServerDataStore, moduleRepository ModuleRepository, helmAppService client.HelmAppService) *ModuleCacheServiceImpl {
	impl := &ModuleCacheServiceImpl{
		logger:           logger,
		K8sUtil:          K8sUtil,
		moduleEnvConfig:  moduleEnvConfig,
		serverEnvConfig:  serverEnvConfig,
		serverDataStore:  serverDataStore,
		moduleRepository: moduleRepository,
		helmAppService:   helmAppService,
	}

	// DB migration - if server mode is not base stack and data in modules table is empty, then insert entries in DB
	if !util2.IsBaseStack() {
		exists, err := impl.moduleRepository.ModuleExists()
		if err != nil {
			log.Fatalln("Error while checking if any module exists in database.", "error", err)
		}
		if !exists {
			for _, supportedModuleName := range SupportedModuleNamesList {
				module := &Module{
					Name:      supportedModuleName,
					Version:   serverDataStore.CurrentVersion,
					Status:    ModuleStatusInstalled,
					UpdatedOn: time.Now(),
				}
				err = impl.moduleRepository.Save(module)
				if err != nil {
					log.Fatalln("Error while saving module.", "error", err)
				}
			}
		}
	}

	// if devtron user type is OSS_HELM then only installer object and modules installation is useful
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeOssHelm {
		// for base mode, installer crd won't come in picture
		// for non-base mode, need to update modules to installed in db in found as installing
		if !util2.IsBaseStack() {
			// handle module status (logic for installed status)
			impl.updateModuleStatusToInstalled()
		}

		// listen in installer object to save status in-memory
		// build informer to listen on installer object
		go impl.buildInformerToListenOnInstallerObject()
	}

	return impl
}

func (impl *ModuleCacheServiceImpl) updateModuleStatusToInstalled() {
	impl.logger.Debug("updating module status to installed")
	modules, err := impl.moduleRepository.FindAll()
	if err != nil {
		log.Fatalln("not able to get all the module from DB.", "error", err)
	}

	for _, module := range modules {
		if module.Status != ModuleStatusInstalling {
			continue
		}
		appIdentifier := client.AppIdentifier{
			ClusterId:   1,
			Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
			ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
		}
		appDetail, err := impl.helmAppService.GetApplicationDetail(context.Background(), &appIdentifier)
		if err != nil {
			log.Fatalln("Error occurred while fetching helm application detail to check if module is installed", "moduleName", module.Name, "err", err)
		}
		if appDetail.ApplicationStatus == serverBean.AppHealthStatusHealthy {
			impl.updateModuleToInstalled(module)
		}
	}
}

func (impl *ModuleCacheServiceImpl) updateModuleToInstalled(module Module) {
	module.Status = ModuleStatusInstalled
	module.UpdatedOn = time.Now()
	err := impl.moduleRepository.Update(&module)
	if err != nil {
		log.Fatalln("error in updating module status to installed", "name", module.Name, "err", err)
	}
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
		DeleteFunc: func(obj interface{}) {
			impl.serverDataStore.InstallerCrdObjectStatus = ""
			impl.serverDataStore.InstallerCrdObjectExists = false
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
	impl.serverDataStore.InstallerCrdObjectExists = true
}
