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
|DEFAULT_BUILD_LOGS_BUCKET | S3 Bucket name used for storing Build Logs| devtron-ci-log| Mandoatory (If using AWS)|
|DEFAULT_CD_LOGS_BUCKET_REGION | Region of S3 Bucket where CD Logs are being stored| us-east-1| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET | S3 Bucket name used for storing CACHE (Do not include s3://)| devtron-ci-cache| Mandatory (If using AWS)|
|DEFAULT_CACHE_BUCKET_REGION | S3 Bucket region where Cache is being stored| us-east-1| Mandoatory (If using AWS)|
|EXTERNAL_SECRET_AMAZON_REGION | Region where the cluster is setup for Devtron installation| ""| Mandatory (If using AWS)|
|ENABLE_INGRESS | To enable Ingress (True/False)| False| Optional|
|INGRESS_ANNOTATIONS | Annotations for ingress| ""| Optional|
|PROMETHEUS_URL | Existing Prometheus URL if it is installed| ""| Optional|


