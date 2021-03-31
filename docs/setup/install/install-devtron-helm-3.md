# Install Devtron using Helm3

To install Helm3, please check [Installing Helm3](https://helm.sh/docs/intro/install/)

{% tabs %}
{% tab title="Install with default configurations" %}
This installation will use Minio for storing build logs and cache. Please make sure to edit the POSTGRESQL_PASSWORD value.

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me
```
{% endtab %}

{% tab title="Install with AWS S3 Buckets" %}
This installation will use AWS s3 buckets for storing build logs and cache

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=AWS \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1
```
{% endtab %}

{% tab title="Install with Azure Blob Storage" %}
This installation will use AWS s3 buckets for storing build logs and cache

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```
{% endtab %}
{% endtabs %}

## Installation status

Run following command

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

Once installation process is complete, above command will print `Applied` It may take around 30 minutes for installation to complete.

## Access Devtron dashboard

If you did not provide a **BASE\_URL** during install or have used the default installation, Devtron creates a loadbalancer for you on its own. Use the following command to get the dashboard url.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

You will get result something like below

```text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

The hostname mentioned here \( aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com \) is the Loadbalancer URL where you can access the Devtron dashboard.

**PS:** You can also do a CNAME entry corresponding to your domain/subdomain to point to this Loadbalancer URL to access it at a custom domain.

| Host | Type | Points to |
| ---: | :--- | :--- |
| devtron.yourdomain.com | CNAME | aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com |

### Devtron Admin credentials

For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

### Cleaning Devtron Installer Helm3

```
helm delete devtron --namespace devtroncd
```
