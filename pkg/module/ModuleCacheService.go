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
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	"github.com/devtron-labs/devtron/pkg/team"
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
	moduleRepository moduleRepo.ModuleRepository
	teamService      team.TeamService
}

func NewModuleCacheServiceImpl(logger *zap.SugaredLogger, K8sUtil *util.K8sUtil, moduleEnvConfig *ModuleEnvConfig, serverEnvConfig *serverEnvConfig.ServerEnvConfig,
	serverDataStore *serverDataStore.ServerDataStore, moduleRepository moduleRepo.ModuleRepository, teamService team.TeamService) *ModuleCacheServiceImpl {
	impl := &ModuleCacheServiceImpl{
		logger:           logger,
		K8sUtil:          K8sUtil,
		moduleEnvConfig:  moduleEnvConfig,
		serverEnvConfig:  serverEnvConfig,
		serverDataStore:  serverDataStore,
		moduleRepository: moduleRepository,
		teamService:      teamService,
	}

	// DB migration - if server mode is not base stack and data in modules table is empty, then insert entries in DB
	if !util2.IsBaseStack() {
		exists, err := impl.moduleRepository.ModuleExists()
		if err != nil {
			log.Fatalln("Error while checking if any module exists in database.", "error", err)
		}
		if !exists {
			// insert cicd module entry
			impl.updateModuleToInstalled(ModuleNameCicd)

			// if old installation (i.e. project was created more than 1 hour ago then insert rest entries)
			teamId := 1
			team, err := teamService.FetchOne(teamId)
			if err != nil {
				log.Fatalln("Error while getting team.", "teamId", teamId, "err", err)
			}

			// insert first release components if this was old release and user installed full mode at that time
			if time.Now().After(team.CreatedOn.Add(1 * time.Hour)) {
				for _, supportedModuleName := range SupportedModuleNamesListFirstReleaseExcludingCicd {
					impl.updateModuleToInstalled(supportedModuleName)
				}
			}
		}
	}

	// if devtron user type is OSS_HELM then only installer object and modules installation is useful
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeOssHelm {
		// listen in installer object to save status in-memory
		// build informer to listen on installer object
		go impl.buildInformerToListenOnInstallerObject()
	}

	return impl
}

func (impl *ModuleCacheServiceImpl) updateModuleToInstalled(moduleName string) {
	module := &moduleRepo.Module{
		Name:      moduleName,
		Version:   impl.serverDataStore.CurrentVersion,
		Status:    ModuleStatusInstalled,
		UpdatedOn: time.Now(),
	}
	err := impl.moduleRepository.Save(module)
	if err != nil {
		log.Fatalln("Error while saving module.", "moduleName", moduleName, "error", err)
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
