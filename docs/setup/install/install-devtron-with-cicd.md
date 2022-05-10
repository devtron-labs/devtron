# Install Devtron with CICD integration

## Before you begin

Install [Helm](https://helm.sh/docs/intro/install/).

## Installing Devtron using Helm

1. Add Devtron repository
2. Install Devtron

{% tabs %}
{% tab title="Install with default configurations" %}
This installation will use Minio for storing build logs and cache. 

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}

```

{% endtab %}

{% tab title="Install with AWS S3 Buckets" %}
This installation will use AWS s3 buckets for storing build logs and cache. Refer to the `AWS specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#storage-for-logs-and-cache) page.

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1
```
{% endtab %}

{% tab title="Install with Azure Blob Storage" %}
This installation will use Azure Blob Storage for storing build logs and cache.
Refer to the `Azure specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#storage-for-logs-and-cache) page.

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set secrets.AZURE_ACCOUNT_KEY=xxxxxxxxxx \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```
{% endtab %}
{% endtabs %}


For those countries/users where Github is blocked, you can use Gitee as the installation source:

{% tabs %}
{% tab title="Install with Gitee" %}
```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.source=gitee
```
{% endtab %}
{% endtabs %}

If you are planning to use Devtron for `production deployments`, please refer to our recommended overrides for [Devtron Installation](override-default-devtron-installation-configs.md).

## Check Devtron installation status

The install commands start Devtron-operator, which takes about 20 minutes to spin up all of the Devtron microservices one by one. You can use the following command to check the status of the installation:

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

The command executes with one of the following output messages, indicating the status of the installation:

| Status | Description |
| :--- | :--- |
| `Downloaded` | The installer has downloaded all the manifests, and the installation is in progress. |
| `Applied` | The installer has successfully applied all the manifests, and the installation is complete. |

## Check the installer logs

To check the installer logs, run the following command:

```bash
kubectl logs -f -l app=inception -n devtroncd
```

## Devtron dashboard

Use the following command to get the dashboard URL:

```bash
kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
```

You will get an output similar to the one shown below:

```bash
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

The hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` as mentioned above is the Loadbalancer URL where you can access the Devtron dashboard.

If you don't see any results or receive a message that says "service doesn't exist," it means Devtron is still installing; please check back in 5 minutes.

> Note: You can also do a `CNAME` entry corresponding to your domain/subdomain to point to this Loadbalancer URL to access it at a custom domain.

| Host | Type | Points to |
| :--- | :--- | :--- |
| devtron.yourdomain.com | CNAME | aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com |

### Devtron Admin credentials

For admin login, use the username:`admin`, and run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

## Cleaning Devtron Helm installer

Please make sure that you do not have anything inside namespaces devtroncd, devtron-cd, devtron-ci, and devtron-demo as the below steps will clean everything inside these namespaces:

```bash
helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo
```

### FAQs

<details>
  <summary>1. How will I know when the installation is finished?</summary>
  
  Run the following command to check the status of the installation:
  
  ```bash
  kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
  ```

  The above command will print `Applied` once the installation process is complete. The installation process could take up to 30 minutes. 
</details>

<details>
  <summary>2. How do I track the progress of the installation?</summary>

  Run the following command to check the logs of the Pod:

  ```bash
  pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod
  ```
</details>

<details>
  <summary>3. How can I restart the installation if the Devtron installer logs contain an error?</summary>

  First run the below command to clean up components installed by Devtron installer:

  ```bash
  cd devtron-installation-script/
  kubectl delete -n devtroncd -f yamls/
  kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'
  ```

  Next, [install Devtron](./install-devtron.md)
</details>


Still facing issues, please reach out to us on [Discord](https://discord.gg/jsRG5qx2gp).
