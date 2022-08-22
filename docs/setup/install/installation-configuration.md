# Installation Configuration

## Configuration

**Configure Secrets**

For `helm` installation this section referes to _**secrets**_ section of `values.yaml`. For `kubectl` based installation it refers to `kind: secret` in _**install/devtron-operator-configs.yaml**_.

Configure the following properties:

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **POSTGRESQL\_PASSWORD** | Using this parameter the auto-generated password for Postgres can be edited as per requirement(Used by Devtron to store the app information) | |
| **WEBHOOK\_TOKEN** | If you want to continue using Jenkins for CI then provide this for authentication of requests should be base64 encoded |  |

**Configure ConfigMaps**

For `helm` installation this section refers to _**configs**_ section of `values.yaml`. For `kubectl` based installation it refers to `kind: ConfigMap` in _**install/devtron-operator-configs.yaml**_.

Configure the following properties:

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **BASE\_URL\_SCHEME** | Either of HTTP or HTTPS \(required\) | HTTP |
| **BASE\_URL** | URL without scheme and trailing slash, this is the domain pointing to the cluster on which the Devtron platform is being installed. For example, if you have directed domain `devtron.example.com` to the cluster and the ingress controller is listening on port `32080` then URL will be `devtron.example.com:32080` \(required\) | `change-me` |
| **DEX\_CONFIG** | dex config if you want to integrate login with SSO \(optional\) for more information check [Argocd documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/) |  |
| **EXTERNAL\_SECRET\_AMAZON\_REGION** | AWS region for the secret manager to pick \(required\) |  |
| **PROMETHEUS\_URL** | URL of Prometheus where all cluster data is stored; if this is wrong, you will not be able to see application metrics like CPU, RAM, HTTP status code, latency, and throughput \(required\) |  |

**Configure Overrides**

For `helm` installation this section refers to _**customOverrides**_ section of `values.yaml`. In this section you can override values of devtron-cm which you want to keep persistent. For example:

You can configure the following properties:

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **CI\_NODE\_LABEL\_SELECTOR** | Labels for a particular nodegroup which you want to use for running CIs | |
| **CI\_NODE\_TAINTS\_KEY** | Key for toleration if nodegroup chosen for CIs have some taints | |
| **CI\_NODE\_TAINTS\_VALUE** | Value for toleration if nodegroup chosen for CIs have some taints |  |

## Storage for Logs and Cache

### AWS SPECIFIC	

While installing Devtron and using the AWS-S3 bucket for storing the logs and caches, the below parameters are to be used in the ConfigMap.

> NOTE: For using the S3 bucket it is important to add the S3 permission policy to the IAM role attached to the nodes of the cluster.

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **DEFAULT\_CACHE\_BUCKET** | AWS bucket to store docker cache, it should be created beforehand \(required\) |  |
| **DEFAULT\_BUILD\_LOGS\_BUCKET** | AWS bucket to store build logs, it should be created beforehand \(required\) |  |
| **DEFAULT\_CACHE\_BUCKET\_REGION** | AWS region of S3 bucket to store cache \(required\) |  |
| **DEFAULT\_CD\_LOGS\_BUCKET\_REGION** | AWS region of S3 bucket to store CD logs \(required\) |  |

### AZURE SPECIFIC

While installing Devtron using Azure Blob Storage for storing logs and caches, the below parameters will be used in the ConfigMap.

> NOTE: For using the storage containers it is mandatory to enable versioning on the storage account. [Refer this guide](https://docs.microsoft.com/en-us/azure/storage/blobs/versioning-enable?tabs=portal#enable-blob-versioning) to enable the same.

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **AZURE\_ACCOUNT\_NAME** | Account name for AZURE Blob Storage |  |
| **AZURE\_BLOB\_CONTAINER\_CI\_LOG** | AZURE Blob storage container for storing ci-logs after running the CI pipeline |  |
| **AZURE\_BLOB\_CONTAINER\_CI\_CACHE** | AZURE Blob storage container for storing ci cache after running the CI pipeline |  |

To convert string to base64 use the following command:

```bash
echo -n "string" | base64 -d
```

> **Note**:
> 1. Ensure that the **cluster has read and write access** to the S3 buckets/Azure Blob storage container mentioned in DEFAULT\_CACHE\_BUCKET, DEFAULT\_BUILD\_LOGS\_BUCKET or AZURE\_BLOB\_CONTAINER\_CI\_LOG, or AZURE\_BLOB\_CONTAINER\_CI\_CACHE.
> 2. Ensure that the cluster has **read access** to AWS secrets backends \(SSM & secrets manager\).

---

The following tables contain parameters and their details for Secrets and ConfigMaps that are configured during the installation of Devtron. 
While installing Devtron using `kubectl` the following parameters can be tweaked in [devtron-operator-configs.yaml](https://github.com/devtron-labs/devtron/blob/main/manifests/install/devtron-operator-configs.yaml) file. If the installation is proceeded using `helm3`, the values can be tweaked in [values.yaml](https://github.com/devtron-labs/charts/blob/main/charts/devtron/values.yaml) file.

We can use the `--set` flag to override the default values when installing with Helm. For example, to update POSTGRESQL_PASSWORD and BLOB_STORAGE_PROVIDER, use the install command as:

```bash
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
```

## Secrets

|Parameter | Description| Default| Necessity|
|-|-|-|-|
|ACD_PASSWORD | ArgoCD Password for CD Workflow| Auto-Generated| Optional|
|AZURE_ACCOUNT_KEY | Account key to access Azure objects such as BLOB_CONTAINER_CI_LOG or CI_CACHE| ""| Mandatory (If using Azure)|
|GRAFANA_PASSWORD | Password for Grafana to display graphs| Auto-Generated| Optional|
|POSTGRESQL_PASSWORD | Password for your Postgresql database that will be used to access the database| Auto-Generated| Optional|

## ConfigMaps

|Parameter | Description| Default| Necessity|
|-|-|-|-|
|AZURE_ACCOUNT_NAME | Azure account name which you will use| ""| Mandatory (If using Azure)|
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
|CI_NODE_TAINTS_VALUE | Value of taint key of CI node | "" | Optional|
|CI_DEFAULT_ADDRESS_POOL_BASE_CIDR | CIDR ranges used to allocate subnets in each IP address pool for CI | "" | Optional|
|CI_DEFAULT_ADDRESS_POOL_SIZE | The subnet size to allocate from the base pool for CI | "" | Optional|
|CD_NODE_LABEL_SELECTOR | Label of CD node | kubernetes.io/os=linux| Optional|
|CD_NODE_TAINTS_KEY| Taint key name of CD node| dedicated | Optional|
|CD_NODE_TAINTS_VALUE| Value of taint key of CD node| ci | Optional|
|CD_LIMIT_CI_CPU| CPU limit for pre and post CD Pod |0.5| Optional|
|CD_LIMIT_CI_MEM| Memory limit for pre and post CD Pod|3G|Optional|
|CD_REQ_CI_CPU| CPU request for CI Pod|0.5|Optional|
|CD_REQ_CI_MEM|Memory request for CI Pod |1G|Optional|
|CD_DEFAULT_ADDRESS_POOL_BASE_CIDR | CIDR ranges used to allocate subnets in each IP address pool for CD | "" | Optional|
|CD_DEFAULT_ADDRESS_POOL_SIZE | The subnet size to allocate from the base pool for CD | "" | Optional|
|GITOPS_REPO_PREFIX | Prefix for Gitops repository | devtron |Optional|

## Dashboard Configurations

```bash
RECOMMEND_SECURITY_SCANNING=false
FORCE_SECURITY_SCANNING=false
HIDE_DISCORD=false
```

|Parameter | Description|
|-|-|
|RECOMMEND_SECURITY_SCANNING | If True, `security scanning` is `enabled` by default for a new build pipeline. Users can however turn it off in the new or existing pipelines.|
|FORCE_SECURITY_SCANNING | If set to True, `security scanning` is forcefully `enabled` by default for a new build pipeline. Users can not turn it off for new as well as for existing build pipelines. Old pipelines that have security scanning disabled will remain unchanged and image scanning should be enabled manually for them.|
|HIDE_DISCORD | Hides discord chatbot from the dashboard.|
