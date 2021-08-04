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
|ACD_PASSWORD | ArgoCD Password| Auto-Generated| Optional|
|AZURE_ACCOUNT_KEY | Account Key to access Azure objects| ""| Mandatory (If using Azure)|
|GRAFANA_PASSWORD | Password for Graphana| Auto-Generated| Optional|
|POSTGRESQL_PASSWORD | Password for Postgresql| Auto-Generated| Optional|

### ConfigMaps

|Parameter | Description| Default| Necessity|
|-|-|-|-|
|AZURE_ACCOUNT_NAME | Azure Account Name using| ""| Mandatory (If using Azure)|
|AZURE_BLOB_CONTAINER_CI_LOG | Name for CI Log| ci-log-container| Optional|
|AZURE_BLOB_CONTAINER_CI_CACHE | Name for CI Cache| ci-cache-container| Optional|
|BLOB_STORAGE_PROVIDER | Cloud provider name| MINIO| Mandatory (If using any cloud), MINIO/AZURE/S3|
|DEFAULT_BUILD_LOGS_BUCKET | S3 Bucket name for Build Logs| devtron-ci-log| Mandoatory (If using AWS)|
|DEFAULT_CD_LOGS_BUCKET_REGION | Amazon S3 Bucket CD Logs region| us-east-1| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET | S3 Bucket name used for Cache (Do not include s3://)| devtron-ci-cache| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET_REGION | S3 Bucket region for Cache| us-east-1| Mandoatory (If using AWS)|
|EXTERNAL_SECRET_AMAZON_REGION | Region where Devtron is installed| ""| Mandatory (If using AWS)|
|ENABLE_INGRESS | To enable Ingress (True/False)| False| Optional|
|INGRESS_ANNOTATIONS | Annotations for ingress| ""| Optional|
|PROMETHEUS_URL | Exisitng Prometheous URL| ""| Optional|


