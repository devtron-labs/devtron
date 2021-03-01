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
	"github.com/caarlos0/env"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"strings"
)

const DevMode = "DEV"
const ProdMode = "PROD"

type CiConfig struct {
	DefaultCacheBucket        string   `env:"DEFAULT_CACHE_BUCKET" envDefault:"ci-caching"`
	DefaultCacheBucketRegion  string   `env:"DEFAULT_CACHE_BUCKET_REGION" envDefault:"us-east-2"`
	CiLogsKeyPrefix           string   `env:"CI_LOGS_KEY_PREFIX" envDefault:"my-artifacts"`
	DefaultImage              string   `env:"DEFAULT_CI_IMAGE" envDefault:"686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47"`
	DefaultNamespace          string   `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci"`
	DefaultTimeout            int64    `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
	Mode                      string   `env:"MODE" envDefault:"DEV"`
	DefaultBuildLogsBucket    string   `env:"DEFAULT_BUILD_LOGS_BUCKET" envDefault:"devtron-pro-ci-logs"`
	LimitCpu                  string   `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	LimitMem                  string   `env:"LIMIT_CI_MEM" envDefault:"3G"`
	ReqCpu                    string   `env:"REQ_CI_CPU" envDefault:"0.5"`
	ReqMem                    string   `env:"REQ_CI_MEM" envDefault:"3G"`
	TaintKey                  string   `env:"CI_NODE_TAINTS_KEY" envDefault:""`
	TaintValue                string   `env:"CI_NODE_TAINTS_VALUE" envDefault:""`
	NodeLabelSelector         []string `env:"CI_NODE_LABEL_SELECTOR"`
	CacheLimit                int64    `env:"CACHE_LIMIT" envDefault:"5000000000"` // TODO: Add to default db config also
	DefaultBuildLogsKeyPrefix string   `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" envDefault:"arsenal-v1"`
	DefaultArtifactKeyPrefix  string   `env:"DEFAULT_ARTIFACT_KEY_LOCATION" envDefault:"arsenal-v1/ci-artifacts"`
	WorkflowServiceAccount    string   `env:"WORKFLOW_SERVICE_ACCOUNT" envDefault:"ci-runner"`
	ExternalCiApiSecret       string   `env:"EXTERNAL_CI_API_SECRET" envDefault:"devtroncd-secret"`
	ExternalCiWebhookUrl      string   `env:"EXTERNAL_CI_WEB_HOOK_URL" envDefault:""`
	ExternalCiPayload         string   `env:"EXTERNAL_CI_PAYLOAD" envDefault:"{\"ciProjectDetails\":[{\"gitRepository\":\"https://github.com/srj92/getting-started-nodejs.git\",\"checkoutPath\":\"./abc\",\"commitHash\":\"239077135f8cdeeccb7857e2851348f558cb53d3\",\"commitTime\":\"2019-10-31T20:55:21+05:30\",\"branch\":\"master\",\"message\":\"Update README.md\",\"author\":\"Suraj Gupta \"}],\"dockerImage\":\"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2\",\"digest\":\"test1\",\"dataSource\":\"ext\",\"materialType\":\"git\"}"`
	CiArtifactLocationFormat  string   `env:"CI_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	ImageScannerEndpoint      string   `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-new-demo-devtroncd-service.devtroncd:80"`
	CloudProvider             string   `env:"BLOB_STORAGE_PROVIDER" envDefault:"S3"`
	AzureAccountName          string   `env:"AZURE_ACCOUNT_NAME"`
	AzureBlobContainerCiLog   string   `env:"AZURE_BLOB_CONTAINER_CI_LOG"`
	AzureBlobContainerCiCache string   `env:"AZURE_BLOB_CONTAINER_CI_CACHE"`
	MinioEndpoint             string   `env:"MINIO_ENDPOINT"`
	MinioAccessKey            string   `env:"MINIO_ACCESS_KEY"`
	MinioSecretKey            string   `env:"MINIO_SECRET_KEY"`

	AzureAccountKey string `env:"AZURE_ACCOUNT_KEY"`
	ClusterConfig   *rest.Config
	NodeLabel       map[string]string
}

const ExternalCiWebhookPath = "orchestrator/webhook/ext-ci"

func GetCiConfig() (*CiConfig, error) {
	cfg := &CiConfig{}
	err := env.Parse(cfg)

	if cfg.Mode == DevMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig := flag.String("kubeconfig", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		cfg.ClusterConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
		if len(l) == 0 {
			continue
		}
		kv := strings.Split(l, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid ci node label selector %s, it must be in form key=value, key2=val2", kv)
		}
		cfg.NodeLabel[kv[0]] = kv[1]
	}
	//validation for supported cloudproviders
	if cfg.CloudProvider != BLOB_STORAGE_S3 && cfg.CloudProvider != BLOB_STORAGE_AZURE && cfg.CloudProvider != BLOB_STORAGE_MINIO {
		return nil, fmt.Errorf("unsupported cloudprovider: %s", cfg.CloudProvider)
	}
	return cfg, err
}
