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
	"fmt"
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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
	LocalDevMode                LocalDevMode `env:"RUNTIME_CONFIG_LOCAL_DEV" envDefault:"false"`
	DevtronDefaultNamespaceName string       `env:"DEVTRON_DEFAULT_NAMESPACE" envDefault:"devtroncd"`
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

// TODO use it as generic function across system
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

func (impl *K8sClient) GetRestClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(impl.config)
}

func (impl *K8sClient) GetArgocdConfig() (secret *v1.Secret, cm *v1.ConfigMap, err error) {
	clientSet, err := kubernetes.NewForConfig(impl.config)
	if err != nil {
		return nil, nil, err
	}
	secret, err = clientSet.CoreV1().Secrets(impl.runtimeConfig.DevtronDefaultNamespaceName).Get(context.Background(), ArgocdSecretName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	cm, err = clientSet.CoreV1().ConfigMaps(impl.runtimeConfig.DevtronDefaultNamespaceName).Get(context.Background(), ArgocdConfigMapName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return secret, cm, nil
}

func (impl *K8sClient) GetDevtronConfig() (secret *v1.Secret, err error) {
	dexConfig, err := DexConfigConfigFromEnv()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(impl.config)
	if err != nil {
		return nil, err
	}
	secret, err = clientSet.CoreV1().Secrets(impl.runtimeConfig.DevtronDefaultNamespaceName).Get(context.Background(), dexConfig.DevtronSecretName, v12.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (impl *K8sClient) GetDevtronNamespace() string {
	return impl.runtimeConfig.DevtronDefaultNamespaceName
}

// argocd specific conf
const (
	SettingAdminPasswordHashKey = "admin.password"
	// SettingAdminPasswordMtimeKey designates the key for a root password mtime inside a Kubernetes secret.
	SettingAdminPasswordMtimeKey = "admin.passwordMtime"
	SettingAdminEnabledKey       = "admin.enabled"
	SettingAdminTokensKey        = "admin.tokens"
	SettingServerSignatureKey    = "server.secretkey"
	SettingURLKey                = "url"
	CallbackEndpoint             = "/auth/callback"
	SettingDexConfigKey          = "dex.config"
	DexCallbackEndpoint          = "/api/dex/callback"
	InitialPasswordLength        = 16
	DevtronSecretName            = "devtron-secret"
	DevtronConfigMapName         = "devtron-cm"

	ArgocdConfigMapName        = "argocd-cm"
	ArgocdSecretName           = "argocd-secret"
	ADMIN_PASSWORD             = "ADMIN_PASSWORD"
	SettingAdminAcdPasswordKey = "ACD_PASSWORD"
)

func (impl *K8sClient) GetServerSettings() (*DexConfig, error) {
	cfg := &DexConfig{}
	secret, err := impl.GetDevtronConfig()
	if err != nil {
		return nil, err
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	if settingServerSignatur, ok := secret.Data[SettingServerSignatureKey]; ok {
		cfg.ServerSecret = string(settingServerSignatur)
	}
	if settingURLByte, ok := secret.Data[SettingURLKey]; ok {
		cfg.Url = string(settingURLByte)
	}
	if adminPasswordMtimeBytes, ok := secret.Data[SettingAdminPasswordMtimeKey]; ok {
		if mTime, err := time.Parse(time.RFC3339, string(adminPasswordMtimeBytes)); err == nil {
			cfg.AdminPasswordMtime = mTime
		}
	}
	if dexConfigBytes, ok := secret.Data[SettingDexConfigKey]; ok {
		cfg.DexConfigRaw = string(dexConfigBytes)
	}
	return cfg, nil
}

func (impl *K8sClient) GenerateDexConfigYAML(settings *DexConfig) ([]byte, error) {
	redirectURL, err := settings.RedirectURL()
	if err != nil {
		return nil, fmt.Errorf("failed to infer redirect url from config: %v", err)
	}
	var dexCfg map[string]interface{}
	if len(settings.DexConfigRaw) > 0 {
		err = yaml.Unmarshal([]byte(settings.DexConfigRaw), &dexCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal dex.config from configmap: %v", err)
		}
	}
	if dexCfg == nil {
		dexCfg = make(map[string]interface{})
	}
	issuer, err := settings.GetDexProxyUrl()
	if err != nil {
		return nil, fmt.Errorf("failed to find issuer url: %v", err)
	}
	dexCfg["issuer"] = issuer
	dexCfg["storage"] = map[string]interface{}{
		"type": "memory",
	}
	dexCfg["web"] = map[string]interface{}{
		"http": "0.0.0.0:5556",
	}
	dexCfg["grpc"] = map[string]interface{}{
		"addr": "0.0.0.0:5557",
	}
	dexCfg["telemetry"] = map[string]interface{}{
		"http": "0.0.0.0:5558",
	}
	dexCfg["oauth2"] = map[string]interface{}{
		"skipApprovalScreen": true,
	}

	argoCDStaticClient := map[string]interface{}{
		"id":     settings.DexClientID,
		"name":   "devtron",
		"secret": settings.DexOAuth2ClientSecret(),
		"redirectURIs": []string{
			redirectURL,
		},
	}

	staticClients, ok := dexCfg["staticClients"].([]interface{})
	if ok {
		dexCfg["staticClients"] = append([]interface{}{argoCDStaticClient}, staticClients...)
	} else {
		dexCfg["staticClients"] = []interface{}{argoCDStaticClient}
	}

	dexRedirectURL, err := settings.DexRedirectURL()
	if err != nil {
		return nil, err
	}
	connectors, ok := dexCfg["connectors"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed Dex configuration found")
	}
	for i, connectorIf := range connectors {
		connector, ok := connectorIf.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("malformed Dex configuration found")
		}
		connectorType := connector["type"].(string)
		if !needsRedirectURI(connectorType) {
			continue
		}
		connectorCfg, ok := connector["config"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("malformed Dex configuration found")
		}
		connectorCfg["redirectURI"] = dexRedirectURL
		connector["config"] = connectorCfg
		connectors[i] = connector
	}
	dexCfg["connectors"] = connectors
	return yaml.Marshal(dexCfg)
}

// needsRedirectURI returns whether or not the given connector type needs a redirectURI
// Update this list as necessary, as new connectors are added
// https://github.com/dexidp/dex/tree/master/Documentation/connectors
func needsRedirectURI(connectorType string) bool {
	switch connectorType {
	case "oidc", "saml", "microsoft", "linkedin", "gitlab", "github", "bitbucket-cloud", "openshift":
		return true
	}
	return false
}

func (impl *K8sClient) ConfigUpdateNotify() (chan bool, error) {
	clusterClient, err := kubernetes.NewForConfig(impl.config)
	if err != nil {
		return nil, err
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute, kubeinformers.WithNamespace(impl.runtimeConfig.DevtronDefaultNamespaceName))
	cmInformenr := informerFactory.Core().V1().ConfigMaps()
	secretInformer := informerFactory.Core().V1().Secrets()
	chanConfigUpdate := make(chan bool)
	tryNotify := func() {
		fmt.Println("setting updated")
		chanConfigUpdate <- true
	}
	now := time.Now()
	cmInformenr.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if metaObj, ok := obj.(metav1.Object); ok {
				if metaObj.GetCreationTimestamp().After(now) {
					tryNotify()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMeta, oldOk := oldObj.(metav1.Common)
			newMeta, newOk := newObj.(metav1.Common)
			if oldOk && newOk && oldMeta.GetResourceVersion() != newMeta.GetResourceVersion() {
				tryNotify()
			}
		},
		DeleteFunc: func(obj interface{}) {

		},
	})
	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if metaObj, ok := obj.(metav1.Object); ok {
				if metaObj.GetCreationTimestamp().After(now) {
					tryNotify()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMeta, oldOk := oldObj.(metav1.Common)
			newMeta, newOk := newObj.(metav1.Common)
			if oldOk && newOk && oldMeta.GetResourceVersion() != newMeta.GetResourceVersion() {
				tryNotify()
			}
		},
		DeleteFunc: func(obj interface{}) {

		},
	})
	informerFactory.Start(wait.NeverStop)
	return chanConfigUpdate, nil
}
