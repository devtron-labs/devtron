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
	"github.com/argoproj/argo-cd/v2/util/settings"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/go-pg/pg"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	grpc_prometheus.EnableClientHandlingTimeHistogram()
}

type ArgoCDConnectionManager interface {
	GetConnection(token string) *grpc.ClientConn
}
type ArgoCDConnectionManagerImpl struct {
	logger           *zap.SugaredLogger
	settingsManager  *settings.SettingsManager
	moduleRepository moduleRepo.ModuleRepository
	argoCDSettings   *settings.ArgoCDSettings
}

func NewArgoCDConnectionManagerImpl(Logger *zap.SugaredLogger, settingsManager *settings.SettingsManager,
	moduleRepository moduleRepo.ModuleRepository) (*ArgoCDConnectionManagerImpl, error) {
	argoUserServiceImpl := &ArgoCDConnectionManagerImpl{
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
func (impl *ArgoCDConnectionManagerImpl) GetConnection(token string) *grpc.ClientConn {
	conf, err := GetConfig()
	if err != nil {
		impl.logger.Errorw("error on get acd config while creating connection", "err", err)
		return nil
	}
	settings := impl.getArgoCdSettings()
	var option []grpc.DialOption
	option = append(option, grpc.WithTransportCredentials(GetTLS(settings.Certificate)))
	if len(token) > 0 {
		option = append(option, grpc.WithPerRPCCredentials(TokenAuth{token: token}))
	}
	option = append(option, grpc.WithChainUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor, otelgrpc.UnaryClientInterceptor()), grpc.WithChainStreamInterceptor(grpc_prometheus.StreamClientInterceptor, otelgrpc.StreamClientInterceptor()))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", conf.Host, conf.Port), option...)
	if err != nil {
		return nil
	}
	return conn
}

func SettingsManager(cfg *Config) (*settings.SettingsManager, error) {
	clientSet, kubeConfig := getK8sClient()
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, err
	}
	//TODO: remove this hardcoding
	if len(cfg.Namespace) >= 0 {
		namespace = cfg.Namespace
	}
	return settings.NewSettingsManager(context.Background(), clientSet, namespace), nil
}

func getK8sClient() (k8sClient *kubernetes.Clientset, k8sConfig clientcmd.ClientConfig) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	clientSet := kubernetes.NewForConfigOrDie(config)
	return clientSet, kubeConfig
}

func (impl *ArgoCDConnectionManagerImpl) getArgoCdSettings() *settings.ArgoCDSettings {
	settings := impl.argoCDSettings
	if settings == nil {
		module, err := impl.moduleRepository.FindOne(ModuleNameArgoCd)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		if module == nil || module.Status != ModuleStatusInstalled {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		settings, err = impl.settingsManager.GetSettings()
		if err != nil {
			impl.logger.Errorw("error on get acd connection", "err", err)
			return nil
		}
		impl.argoCDSettings = settings
	}
	return impl.argoCDSettings
}
