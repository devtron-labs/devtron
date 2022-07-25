/*
 * Copyright (c) 2021 Devtron Labs
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

package client

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"time"
)

type LocalDevMode bool

type K8sClient struct {
	runtimeConfig *RuntimeConfig
	config        *rest.Config
}

type RuntimeConfig struct {
	LocalDevMode LocalDevMode `env:"RUNTIME_CONFIG_LOCAL_DEV" envDefault:"false"`
}

func GetRuntimeConfig() (*RuntimeConfig, error) {
	cfg := &RuntimeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewK8sClient(runtimeConfig *RuntimeConfig) (*K8sClient, error) {
	config, err := getKubeConfig(runtimeConfig.LocalDevMode)
	if err != nil {
		return nil, err
	}
	return &K8sClient{
		runtimeConfig: runtimeConfig,
		config:        config,
	}, nil
}

//TODO use it as generic function across system
func getKubeConfig(devMode LocalDevMode) (*rest.Config, error) {
	if devMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig := flag.String("kubeconfig-authenticator", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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

func (impl *K8sClient) GetArgoConfig() (secret *v1.Secret, cm *v1.ConfigMap, err error) {
	clientSet, err := kubernetes.NewForConfig(impl.config)
	if err != nil {
		return nil, nil, err
	}
	secret, err = clientSet.CoreV1().Secrets(ArgocdNamespaceName).Get(context.Background(), ArgoCDSecretName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	cm, err = clientSet.CoreV1().ConfigMaps(ArgocdNamespaceName).Get(context.Background(), ArgoCDConfigMapName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return secret, cm, nil
}

// argocd specific conf
const (
	SettingAdminPasswordHashKey = "admin.password"
	// SettingAdminPasswordMtimeKey designates the key for a root password mtime inside a Kubernetes secret.
	SettingAdminPasswordMtimeKey = "admin.passwordMtime"
	SettingAdminEnabledKey       = "admin.enabled"
	SettingAdminTokensKey        = "admin.tokens"

	SettingServerSignatureKey = "server.secretkey"
	settingURLKey             = "url"
	ArgoCDConfigMapName       = "argocd-cm"
	ArgoCDSecretName          = "argocd-secret"
	ArgocdNamespaceName       = "devtroncd"
)

func (impl *K8sClient) GetServerSettings() (*DexConfig, error) {
	cfg := &DexConfig{}
	secret, cm, err := impl.GetArgoConfig()
	if err != nil {
		return nil, err
	}
	if settingServerSignatur, ok := secret.Data[SettingServerSignatureKey]; ok {
		cfg.ServerSecret = string(settingServerSignatur)
	}
	if settingURL, ok := cm.Data[settingURLKey]; ok {
		cfg.Url = settingURL
	}
	if adminPasswordMtimeBytes, ok := secret.Data[SettingAdminPasswordMtimeKey]; ok {
		if mTime, err := time.Parse(time.RFC3339, string(adminPasswordMtimeBytes)); err == nil {
			cfg.AdminPasswordMtime = mTime
		}
	}
	return cfg, nil
}
