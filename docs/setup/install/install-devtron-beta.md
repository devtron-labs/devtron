# Install Devtron(Beta) with CI/CD

We also release beta versions of devtron every few days before the stable release for people who would like to explore and test beta features before everyone else. If you want to install a fresh devtron from beta release channel, use the chart in our official devtron repository.

This chart is currently not available on the official helm repository therefore you need to download it to install it.

1. Clone Devtron Repositry 
2. Upgrade Helm Dependency
3. Install Devtron

```bash 
$ git clone [https://github.com/devtron-labs/devtron.git](https://github.com/devtron-labs/devtron.git)
$ cd devtron/charts/devtron
$ helm dependency up
$ #modify values in values.yaml
$ helm install devtron . --create-namespace --namespace devtroncd \
--set installer.modules={cicd}

```

{% tab title="Install with AWS S3 Buckets" %}
This installation will use AWS s3 buckets for storing build logs and cache. Refer to the `AWS specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#storage-for-logs-and-cache) page.
```bash
$ git clone [https://github.com/devtron-labs/devtron.git](https://github.com/devtron-labs/devtron.git)
$ cd devtron/charts/devtron
$ helm dependency up
$ #modify values in values.yaml
$ helm install devtron . --create-namespace --namespace devtroncd \
--set installer.modules={cicd}\
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1
```

{% tab title="Install with Azure Blob Storage" %}
This installation will use Azure Blob Storage for storing build logs and cache.
Refer to the `Azure specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#storage-for-logs-and-cache) page.

```bash
$ git clone [https://github.com/devtron-labs/devtron.git](https://github.com/devtron-labs/devtron.git)
$ cd devtron/charts/devtron
$ helm dependency up
$ #modify values in values.yaml
$ helm install devtron .  --create-namespace --namespace devtroncd \
--set installer.modules={cicd}\
--set secrets.AZURE_ACCOUNT_KEY=xxxxxxxxxx \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```

> Note: There is no option to upgrade to beta on stack manager UI as of now and you may always see upgrade available for latest stable version using which you'll be moved to latest stable version available.

## Install Devtron with CI/CD and Additional Integration

To install additional integrations along with CI/CD using the Devtron Helm chart, you can append the configurations for those integrations to the installation command in the same way you did for the cicd integration. Assuming you want to add an integration named `GitOps (Argo CD)`, the installation command would look like this:


```bash
$ git clone [https://github.com/devtron-labs/devtron.git](https://github.com/devtron-labs/devtron.git)
$ cd devtron/charts/devtron
$ helm dependency up
$ #modify values in values.yaml
$ helm install devtron . --create-namespace --namespace devtroncd \
--set installer.modules={cicd}
--set argo-cd.enabled=true
```

| Integration | Configuration |
| :---: | :---: |
| `GitOps (Argo CD)`  |  `--set notifier.enabled=true` |
|  `Notifications` |  `--set notifier.enabled=true` |
| `Clair`  | `--set security.enabled=true` <br/> `--set security.clair.enabled=true`  |
| `Trivy`  |  `--set security.enabled=true` <br/> `--set security.trivy.enabled=true`  |
| `Monitoring (Grafana)` | `--set monitoring.grafana.enabled=true` |


## Install Multi-Architecture Nodes (ARM and AMD)

To install Devtron on clusters with the multi-architecture nodes (ARM and AMD), append the Devtron installation command with `--set installer.arch=multi-arch`.

**Note**: 

* If you want to install Devtron for `production deployments`, please refer to our recommended overrides for [Devtron Installation](override-default-devtron-installation-configs.md).
