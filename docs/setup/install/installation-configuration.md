# Installation Configuration

The following tables contains parameters and its details for Secrets and ConfigMaps which are configured during installation of Devtron. While installing Devtron using kubectl the following parameters can be tweaked in [devtron-operator-configs.yaml](https://github.com/devtron-labs/devtron/blob/main/manifests/install/devtron-operator-configs.yaml) file. If the installation is being proceeded using helm, the values can be tweaked in [values.yaml](https://github.com/devtron-labs/charts/blob/main/charts/devtron/values.yaml) file.

To override the default values while installing using helm, we can use `--set` flag. For example, if we want to update POSTGRESQL_PASSWORD and BLOB_STORAGE_PROVIDER, we can simply use the install command as -

```bash
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
```

### Secrets

|Parameter | Description| Default| Necessity|
|-|-|-|-|
|ACD_PASSWORD | ArgoCD Password for CD Workflow| Auto-Generated| Optional|
|AZURE_ACCOUNT_KEY | Account key to access Azure objects such as BLOB_CONTAINER_CI_LOG or CI_CACHE| ""| Mandatory (If using Azure)|
|GRAFANA_PASSWORD | Password for Graphana to display graphs| Auto-Generated| Optional|
|POSTGRESQL_PASSWORD | Password for your Postgresql database which will be used to access database| Auto-Generated| Optional|

### ConfigMaps

|Parameter | Description| Default| Necessity|
|-|-|-|-|
|AZURE_ACCOUNT_NAME | Azure Account Name which you will use| ""| Mandatory (If using Azure)|
|AZURE_BLOB_CONTAINER_CI_LOG | Name of container created for storing CI_LOG| ci-log-container| Optional|
|AZURE_BLOB_CONTAINER_CI_CACHE | Name of container created for storing CI_CACHE| ci-cache-container| Optional|
|BLOB_STORAGE_PROVIDER | Cloud provider name which you will use| MINIO| Mandatory (If using any cloud other than MINIO), MINIO/AZURE/S3|
|DEFAULT_BUILD_LOGS_BUCKET | S3 Bucket name used for storing Build Logs| devtron-ci-log| Mandatory (If using AWS)|
|DEFAULT_CD_LOGS_BUCKET_REGION | Region of S3 Bucket where CD Logs are being stored| us-east-1| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET | S3 Bucket name used for storing CACHE (Do not include s3://)| devtron-ci-cache| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET_REGION | S3 Bucket region where Cache is being stored| us-east-1| Mandatory (If using AWS)|
|EXTERNAL_SECRET_AMAZON_REGION | Region where the cluster is setup for Devtron installation| ""| Mandatory (If using AWS)|
|ENABLE_INGRESS | To enable Ingress (True/False)| False| Optional|
|INGRESS_ANNOTATIONS | Annotations for ingress| ""| Optional|
|PROMETHEUS_URL | Existing Prometheus URL if it is installed| ""| Optional|
|CI_NODE_LABEL_SELECTOR | Label of CI worker node| "" | Optional| 
|CI_NODE_TAINTS_KEY | Taint key name of CI worker node | "" | Optional|
|CI_NODE_TAINTS_VALUE |Value of taint key of CI node | "" | Optional|
|CI_DEFAULT_ADDRESS_POOL_BASE_CIDR | CIDR ranges used to allocate subnets in each IP address pool for CI | "" | Optional|
|CI_DEFAULT_ADDRESS_POOL_SIZE | The subnet size to allocate from the base pool for CI | "" | Optional|
|CD_NODE_LABEL_SELECTOR | Label of CD node | kubernetes.io/os=linux| Optional|
|CD_NODE_TAINTS_KEY| Taint key name of CD node| dedicated | Optional|
|CD_NODE_TAINTS_VALUE| Value of taint key of CD node| ci | Optional|
|CD_LIMIT_CI_CPU|CPU limit for pre and post CD Pod |0.5| Optional|
|CD_LIMIT_CI_MEM| Memory limit for pre and post CD Pod|3G|Optional|
|CD_REQ_CI_CPU|CPU request for CI Pod|0.5|Optional|
|CD_REQ_CI_MEM|Memory request for CI Pod |1G|Optional|
|CD_DEFAULT_ADDRESS_POOL_BASE_CIDR | CIDR ranges used to allocate subnets in each IP address pool for CD | "" | Optional|
|CD_DEFAULT_ADDRESS_POOL_SIZE | The subnet size to allocate from the base pool for CD | "" | Optional|
|GITOPS_REPO_PREFIX | Prefix for Gitops repository | devtron |Optional|


### Dashboard Configurations

```bash
RECOMMEND_SECURITY_SCANNING=false
FORCE_SECURITY_SCANNING=false
HIDE_DISCORD=false
```

|Parameter | Description|
|-|-|
|RECOMMEND_SECURITY_SCANNING | If True `security scanning` is `enabled` by default for a new build pipeline. Users can however turn it off in the new or existing pipelines.|
|FORCE_SECURITY_SCANNING | If set True, `security scanning` is forcefully `enabled` by default for a new build pipeline. User can not turn it off for new as well as for existing build pipelines. Old pipelines that have security scanning disabled will remain unchanged and image scanning should be enabled manually for them.|
|HIDE_DISCORD | Hides discord chat bot from dashboard.|



