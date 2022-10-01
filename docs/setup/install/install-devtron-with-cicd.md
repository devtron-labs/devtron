# Install Devtron with CICD integration

Are you installing Devtron on Minikube, Microk8s, K3s, Kind? See Instructions [here](./Install-devtron-on-Minikube-Microk8s-K3s-Kind.md)

## Before you begin

Install [Helm](https://helm.sh/docs/intro/install/).

## Installing Devtron using Helm

1. Add Devtron repository
2. Install Devtron

Install with one of the following commands:
{% tabs %}
{% tab title="Default" %}

Use the following command to install Devtron without Blob Storage. 

Configuring Blob Storage in your Devtron environment allows you to store build logs and cache.
In case, if you do not configure the Blob Storage, then:

- You will not be able to access the build and deployment logs after an hour.
- Build time for commit hash takes longer as cache is not available.
- Artifact reports cannot be generated in pre/post build and deployment stages.

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}
```

{% endtab %}

{% tab title="MinIO Storage" %}

Use the following command to install Devtron along with MinIO for storing logs and cache.
**Note**: Unlike global cloud providers such as AWS S3 Bucket, Azure Blob Storage and Google Cloud Storage, MinIO can be hosted locally also.

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true
```

{% endtab %}

{% tab title="AWS S3 Bucket" %}
Use the following command to install Devtron along with AWS S3 buckets for storing build logs and cache. Refer to the `AWS specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#aws-specific) page.

*  Install using S3 IAM policy.

>NOTE: Pleasee ensure that S3 permission policy to the IAM role attached to the nodes of the cluster if you are using the below command.

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

*  Install using access-key and secret-key for AWS S3 authentication:

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key>
```

*  Install using S3 compatible storages: 

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key> \
--set configs.BLOB_STORAGE_S3_ENDPOINT=<endpoint>
```

{% endtab %}

{% tab title="Azure Blob Storage" %}
Use the following command to install Devtron along with Azure Blob Storage for storing build logs and cache.
Refer to the `Azure specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#azure-specific) page.

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

{% tab title="Google Cloud Storage" %}
Use the following command to install Devtron along with Google Cloud Storage for storing build logs and cache.
Refer to the `Google Cloud specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#google-cloud-storage-specific) page.


```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER= GCP \
--set secrets.BLOB_STORAGE_GCP_CREDENTIALS_JSON= eyJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsInByb2plY3RfaWQiOiAiPHlvdXItcHJvamVjdC1pZD4iLCJwcml2YXRlX2tleV9pZCI6ICI8eW91ci1wcml2YXRlLWtleS1pZD4iLCJwcml2YXRlX2tleSI6ICI8eW91ci1wcml2YXRlLWtleT4iLCJjbGllbnRfZW1haWwiOiAiPHlvdXItY2xpZW50LWVtYWlsPiIsImNsaWVudF9pZCI6ICI8eW91ci1jbGllbnQtaWQ+IiwiYXV0aF91cmkiOiAiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLCJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vb2F1dGgyL3YxL2NlcnRzIiwiY2xpZW50X3g1MDlfY2VydF91cmwiOiAiPHlvdXItY2xpZW50LWNlcnQtdXJsPiJ9Cg== \
--set configs.DEFAULT_CACHE_BUCKET= cache-bucket
--set configs.DEFAULT_BUILD_LOGS_BUCKET= log-bucket
```

{% endtab %}
{% endtabs %}

> Append the command with `--set installer.release="vX.X.X"` to install a particular version of Devtron. Where `vx.x.x` is the [release tag](https://github.com/devtron-labs/devtron/releases).

For those countries/users where GitHub is blocked, you can use Gitee as the installation source:

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

#### For Devtron version v0.6.0 and higher

Use username:`admin` and for password run command mentioned below.
```bash
$ kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

#### For Devtron version less than v0.6.0

Use username:`admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

## Cleaning Devtron Helm installer

Please make sure that you do not have anything inside namespaces devtroncd, devtron-cd, devtron-ci, and devtron-demo as the below steps will clean everything inside these namespaces:

```bash
helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo
```

## What's next

[Configurations](installation-configuration.md)

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