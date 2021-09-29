# Install Devtron using Helm2

To install Helm2, please check [Installing Helm2](https://v2.helm.sh/docs/install//) Make sure you have [Installed tiller using helm init](https://v2.helm.sh/docs/install/#installing-tiller)

{% tabs %}
{% tab title="Install with default configurations" %}
This installation will use Minio for storing build logs and cache. 
```bash
kubectl create namespace devtroncd
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/crds/crd-devtron.yaml
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd --set installer.source=gitee
```
{% endtab %}

{% tab title="Install with AWS S3 Buckets" %}
This installation will use AWS s3 buckets for storing build logs and cache

```bash
kubectl create namespace devtroncd
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/crds/crd-devtron.yaml
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1
```
{% endtab %}

{% tab title="Install with Azure Blob Storage" %}
This installation will use Azure Blob Storage for storing build logs and cache

```bash
kubectl create namespace devtroncd
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/crds/crd-devtron.yaml
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```
{% endtab %}
{% endtabs %}

For those countries/users where Github is blocked , you can use Gitee as the installation source.

{% tabs %}
{% tab title="Install with Gitee" %}
```bash
kubectl create namespace devtroncd

kubectl apply -f https://gitee.com/devtron-labs/devtron/raw/main/manifests/crds/crd-devtron.yaml

helm install devtron devtron/devtron-operator --namespace devtroncd --set installer.source=gitee
```
{% endtab %}
{% endtabs %}

If you are planning to use Devtron for `production deployments`, please refer to our recommended overrides for [Devtron Installation](override-default-devtron-installation-configs.md).

## Installation status

Run following command

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

The install commands initiates Devtron-operator which spins up all the Devtron micro-services one by one in about 20 mins. You can use the above command to check the status of the installation if the installation is still in progress, it will print `Downloaded`. When the installation is complete, it prints `Applied`.

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

### Cleaning Installer Helm2

```bash
helm delete devtron --purge
#Deleting CRDs manually
kubectl delete -f https://raw.githubusercontent.com/devtron-labs/devtron-installation-script/main/charts/devtron/crds/crd-devtron.yaml
```
