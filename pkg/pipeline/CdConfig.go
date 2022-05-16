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

package pipeline

import (
	"flag"
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CdConfig struct {
	Mode                       string   `env:"MODE" envDefault:"DEV"`
	LimitCpu                   string   `env:"CD_LIMIT_CI_CPU" envDefault:"0.5"`
	LimitMem                   string   `env:"CD_LIMIT_CI_MEM" envDefault:"3G"`
	ReqCpu                     string   `env:"CD_REQ_CI_CPU" envDefault:"0.5"`
	ReqMem                     string   `env:"CD_REQ_CI_MEM" envDefault:"3G"`
	TaintKey                   string   `env:"CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	WorkflowServiceAccount     string   `env:"CD_WORKFLOW_SERVICE_ACCOUNT" envDefault:"cd-runner"`
	DefaultBuildLogsKeyPrefix  string   `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" `
	DefaultArtifactKeyPrefix   string   `env:"DEFAULT_CD_ARTIFACT_KEY_LOCATION" `
	TaintValue                 string   `env:"CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	DefaultBuildLogsBucket     string   `env:"DEFAULT_BUILD_LOGS_BUCKET" `
	NodeLabelSelector          []string `env:"CD_NODE_LABEL_SELECTOR"`
	CdArtifactLocationFormat   string   `env:"CD_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	DefaultNamespace           string   `env:"DEFAULT_CD_NAMESPACE"`
	DefaultImage               string   `env:"DEFAULT_CI_IMAGE" `
	DefaultTimeout             int64    `env:"DEFAULT_CD_TIMEOUT" envDefault:"3600"`
	DefaultCdLogsBucketRegion  string   `env:"DEFAULT_CD_LOGS_BUCKET_REGION" `
	WfControllerInstanceID     string   `env:"WF_CONTROLLER_INSTANCE_ID" envDefault:"devtron-runner"`
	OrchestratorHost           string   `env:"ORCH_HOST" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats"`
	OrchestratorToken          string   `env:"ORCH_TOKEN" envDefault:""`
	ClusterConfig              *rest.Config
	NodeLabel                  map[string]string
	CloudProvider              string `env:"BLOB_STORAGE_PROVIDER" envDefault:"S3"`
	AzureAccountName           string `env:"AZURE_ACCOUNT_NAME"`
	AzureBlobContainerCiLog    string `env:"AZURE_BLOB_CONTAINER_CI_LOG"`
	AzureBlobContainerCiCache  string `env:"AZURE_BLOB_CONTAINER_CI_CACHE"`
	MinioEndpoint              string `env:"MINIO_ENDPOINT"`
	MinioAccessKey             string `env:"MINIO_ACCESS_KEY"`
	MinioSecretKey             string `env:"MINIO_SECRET_KEY"`
	AzureAccountKey            string `env:"AZURE_ACCOUNT_KEY"`
	DefaultAddressPoolBaseCidr string `env:"CD_DEFAULT_ADDRESS_POOL_BASE_CIDR"`
	DefaultAddressPoolSize     int    `env:"CD_DEFAULT_ADDRESS_POOL_SIZE"`
}

func GetCdConfig() (*CdConfig, error) {
	cfg := &CdConfig{}
	err := env.Parse(cfg)
	if cfg.Mode == DevMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig_cd := flag.String("kubeconfig_cd", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		cfg.ClusterConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig_cd)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.ClusterConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	cfg.NodeLabel = make(map[string]string)
	for _, l := range cfg.NodeLabelSelector {
		if l == "" {
			continue
		}
		kv := strings.Split(l, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid ci node label selector %s, it must be in form key=value, key2=val2", kv)
		}
		cfg.NodeLabel[kv[0]] = kv[1]
	}

	return cfg, err
}
