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

package argocdServer

import (
	"context"
	"fmt"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"

	"github.com/argoproj/argo-cd/v2/util/settings"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func init() {
	grpc_prometheus.EnableClientHandlingTimeHistogram()
}

type ArgoCdConnection interface {
	GetConnection(token string) *grpc.ClientConn
}
type ArgoCdConnectionImpl struct {
	logger           *zap.SugaredLogger
	settingsManager  *settings.SettingsManager
	moduleRepository moduleRepo.ModuleRepository
	argoCDSettings   *settings.ArgoCDSettings
}

func NewArgoCdConnectionImpl(Logger *zap.SugaredLogger, settingsManager *settings.SettingsManager,
	moduleRepository moduleRepo.ModuleRepository) (*ArgoCdConnectionImpl, error) {
	argoUserServiceImpl := &ArgoCdConnectionImpl{
		logger:           Logger,
		settingsManager:  settingsManager,
		moduleRepository: moduleRepository,
		argoCDSettings:   nil,
	}
	return argoUserServiceImpl, nil
}

const (
	ModuleNameArgoCd      string = "argo-cd"
	ModuleStatusInstalled string = "installed"
)

// GetConnection - this function will call only for acd connection
func (impl *ArgoCdConnectionImpl) GetConnection(token string) *grpc.ClientConn {
	conf, err := GetConfig()
	if err != nil {
		impl.logger.Errorw("error on get acd connection", "err", err)
		log.Fatal(err)
	}
	settings := impl.argoCDSettings
	if settings == nil {
		module, err := impl.moduleRepository.FindOne(ModuleNameArgoCd)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error on get acd connection", "err", err)
			//log.Fatal(err)
			return nil
		}
		if module == nil || module.Status != ModuleStatusInstalled {
			impl.logger.Errorw("error on get acd connection", "err", err)
			//log.Fatal(err)
			return nil
		}
		settings, err = impl.settingsManager.GetSettings()
		if err != nil {
			impl.logger.Errorw("error on get acd connection", "err", err)
			log.Fatal(err)
		}
		impl.argoCDSettings = settings
	}
	var option []grpc.DialOption
	option = append(option, grpc.WithTransportCredentials(GetTLS(settings.Certificate)))
	if len(token) > 0 {
		option = append(option, grpc.WithPerRPCCredentials(TokenAuth{token: token}))
	}
	option = append(option, grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor), grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", conf.Host, conf.Port), option...)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func SettingsManager(cfg *Config) (*settings.SettingsManager, error) {
	clientset, kubeconfig := GetK8sclient()
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		return nil, err
	}
	//TODO: remove this hardcoding
	if len(cfg.Namespace) >= 0 {
		namespace = cfg.Namespace
	}
	return settings.NewSettingsManager(context.Background(), clientset, namespace), nil
}

func GetK8sclient() (k8sClient *kubernetes.Clientset, k8sConfig clientcmd.ClientConfig) {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)
	return clientset, kubeconfig
}
