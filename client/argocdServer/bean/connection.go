/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package bean

import (
	"github.com/caarlos0/env/v6"
	"k8s.io/client-go/rest"
)

type ArgoGRPCConfig struct {
	ConnectionConfig *Config
	AuthConfig       *AcdAuthConfig
}

type ArgoK8sConfig struct {
	RestConfig       *rest.Config
	AcdNamespace     string
	AcdConfigMapName string
}

type AcdAuthConfig struct {
	ClusterId                 int
	DevtronSecretName         string
	DevtronDexSecretNamespace string
	UserName                  string
	Password                  string
}

type Config struct {
	Host      string `env:"CD_HOST" envDefault:"localhost" description:"Host for the devtron stack"`
	Port      string `env:"CD_PORT" envDefault:"8000" description:"Port for pre/post-cd" `
	Namespace string `env:"CD_NAMESPACE" envDefault:"devtroncd"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
