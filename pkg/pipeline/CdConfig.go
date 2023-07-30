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

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

type CdConfig struct {
	//Mode                             string   `env:"MODE" envDefault:"DEV"`
	CdLimitCpu                  string   `env:"CD_LIMIT_CI_CPU" envDefault:"0.5"`
	CdLimitMem                  string   `env:"CD_LIMIT_CI_MEM" envDefault:"3G"`
	CdReqCpu                    string   `env:"CD_REQ_CI_CPU" envDefault:"0.5"`
	CdReqMem                    string   `env:"CD_REQ_CI_MEM" envDefault:"3G"`
	CdTaintKey                  string   `env:"CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	CdWorkflowServiceAccount    string   `env:"CD_WORKFLOW_SERVICE_ACCOUNT" envDefault:"cd-runner"`
	CdDefaultBuildLogsKeyPrefix string   `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" `
	CdDefaultArtifactKeyPrefix  string   `env:"DEFAULT_CD_ARTIFACT_KEY_LOCATION" `
	CdTaintValue                string   `env:"CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	CdDefaultBuildLogsBucket    string   `env:"DEFAULT_BUILD_LOGS_BUCKET" `
	CdNodeLabelSelector         []string `env:"CD_NODE_LABEL_SELECTOR"`
	CdArtifactLocationFormat    string   `env:"CD_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	CdDefaultNamespace          string   `env:"DEFAULT_CD_NAMESPACE"`
	CdDefaultImage              string   `env:"DEFAULT_CI_IMAGE"`
	CdDefaultTimeout            int64    `env:"DEFAULT_CD_TIMEOUT" envDefault:"3600"`
	CdDefaultCdLogsBucketRegion string   `env:"DEFAULT_CD_LOGS_BUCKET_REGION" `
	WfControllerInstanceID      string   `env:"WF_CONTROLLER_INSTANCE_ID" envDefault:"devtron-runner"`
	//OrchestratorHost                 string   `env:"ORCH_HOST" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats"`
	//OrchestratorToken                string   `env:"ORCH_TOKEN" envDefault:""`
	//ClusterConfig                    *rest.Config
	//NodeLabel                        map[string]string
	//CloudProvider                    blob_storage.BlobStorageType        `env:"BLOB_STORAGE_PROVIDER" envDefault:"S3"`
	//BlobStorageEnabled               bool                                `env:"BLOB_STORAGE_ENABLED" envDefault:"false"`
	//BlobStorageS3AccessKey           string                              `env:"BLOB_STORAGE_S3_ACCESS_KEY"`
	//BlobStorageS3SecretKey           string                              `env:"BLOB_STORAGE_S3_SECRET_KEY"`
	//BlobStorageS3Endpoint            string                              `env:"BLOB_STORAGE_S3_ENDPOINT"`
	//BlobStorageS3EndpointInsecure    bool                                `env:"BLOB_STORAGE_S3_ENDPOINT_INSECURE" envDefault:"false"`
	//BlobStorageS3BucketVersioned     bool                                `env:"BLOB_STORAGE_S3_BUCKET_VERSIONED" envDefault:"true"`
	//BlobStorageGcpCredentialJson     string                              `env:"BLOB_STORAGE_GCP_CREDENTIALS_JSON"`
	//AzureAccountName                 string                              `env:"AZURE_ACCOUNT_NAME"`
	//AzureGatewayUrl                  string                              `env:"AZURE_GATEWAY_URL" envDefault:"http://devtron-minio.devtroncd:9000"`
	//AzureGatewayConnectionInsecure   bool                                `env:"AZURE_GATEWAY_CONNECTION_INSECURE" envDefault:"true"`
	//AzureBlobContainerCiLog          string                              `env:"AZURE_BLOB_CONTAINER_CI_LOG"`
	//AzureBlobContainerCiCache        string                              `env:"AZURE_BLOB_CONTAINER_CI_CACHE"`
	//AzureAccountKey                  string                              `env:"AZURE_ACCOUNT_KEY"`
	//BuildLogTTLValue                 int                                 `env:"BUILD_LOG_TTL_VALUE_IN_SECS" envDefault:"3600"`
	CdDefaultAddressPoolBaseCidr     string `env:"CD_DEFAULT_ADDRESS_POOL_BASE_CIDR"`
	CdDefaultAddressPoolSize         int    `env:"CD_DEFAULT_ADDRESS_POOL_SIZE"`
	ExposeCDMetrics                  bool   `env:"EXPOSE_CD_METRICS" envDefault:"false"`
	UseBlobStorageConfigInCdWorkflow bool   `env:"USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW" envDefault:"true"`
	//BaseLogLocationPath              string                              `env:"BASE_LOG_LOCATION_PATH" envDefault:"/home/devtron/"`
	CdWorkflowExecutorType pipelineConfig.WorkflowExecutorType `env:"CD_WORKFLOW_EXECUTOR_TYPE" envDefault:"AWF"`
	//InAppLoggingEnabled              bool                                `env:"IN_APP_LOGGING_ENABLED" envDefault:"false"`
	*CiCdConfig
}

//func GetCdConfig() (*CiCdConfig, error) {
//	ciCdConfig := &CiCdConfig{}
//	err := env.Parse(ciCdConfig)
//	cfg := &CiCdConfig{CiCdConfig: ciCdConfig}
//	err = env.Parse(cfg)
//	if cfg.Mode == DevMode {
//		usr, err := user.Current()
//		if err != nil {
//			return nil, err
//		}
//		kubeconfig_cd := flag.String("kubeconfig_cd", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
//		flag.Parse()
//		cfg.ClusterConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig_cd)
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		cfg.ClusterConfig, err = rest.InClusterConfig()
//		if err != nil {
//			return nil, err
//		}
//	}
//	cfg.NodeLabel = make(map[string]string)
//	for _, l := range cfg.CdNodeLabelSelector {
//		if l == "" {
//			continue
//		}
//		kv := strings.Split(l, "=")
//		if len(kv) != 2 {
//			return nil, fmt.Errorf("invalid ci node label selector %s, it must be in form key=value, key2=val2", kv)
//		}
//		cfg.NodeLabel[kv[0]] = kv[1]
//	}
//
//	return cfg, err
//}
