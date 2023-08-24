# Install Devtron with CICD

In this section, we describe the steps in detail on how you can install Devtron with CI/CD integration.


## Before you begin

Install [Helm](https://helm.sh/docs/intro/install/), if you have not installed it.


## Install Devtron with CI/CD

Run the following command to install the latest version of Devtron along with the CI/CD module:

```bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}
```

**Note**: If you want to configure Blob Storage during the installation, refer [configure blob storage duing installation](#configure-blob-storage-duing-installation).


## Install AWS EBS CSI Driver, if require

If you are using EKS version 1.23 or above, you must install [aws-ebs-csi-driver](https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html).


Run the following command to install aws ebs csi driver using Helm:

```bash
helm repo add aws-ebs-csi-driver \
https://kubernetes-sigs.github.io/aws-ebs-csi-driver \
helm repo update \
helm upgrade --install aws-ebs-csi-driver \
--namespace kube-system aws-ebs-csi-driver/aws-ebs-csi-driver
```

## Install Multi-Architecture Nodes (ARM and AMD)

To install Devtron on clusters with the multi-architecture nodes (ARM and AMD), append the Devtron installation command with `--set installer.arch=multi-arch`.

**Note**: 

* If you want to install Devtron for `production deployments`, please refer to our recommended overrides for [Devtron Installation](override-default-devtron-installation-configs.md).


## Configure Blob Storage duing Installation

Configuring Blob Storage in your Devtron environment allows you to store build logs and cache.
In case, if you do not configure the Blob Storage, then:

- You will not be able to access the build and deployment logs after an hour.
- Build time for commit hash takes longer as cache is not available.
- Artifact reports cannot be generated in pre/post build and deployment stages.

Choose one of the options to configure blob storage:

{% tabs %}

{% tab title="MinIO Storage" %}

Run the following command to install Devtron along with MinIO for storing logs and cache.

```bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true
```
**Note**: Unlike global cloud providers such as AWS S3 Bucket, Azure Blob Storage and Google Cloud Storage, MinIO can be hosted locally also.

{% endtab %}

{% tab title="AWS S3 Bucket" %}

Refer to the `AWS specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#aws-specific) page.

Run the following command to install Devtron along with AWS S3 buckets for storing build logs and cache:

*  Install using S3 IAM policy.

>Note: Please ensure that S3 permission policy to the IAM role attached to the nodes of the cluster if you are using below command.

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
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

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
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

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
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

Refer to the `Azure specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#azure-specific) page.

Run the following command to install Devtron along with Azure Blob Storage for storing build logs and cache:

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set secrets.AZURE_ACCOUNT_KEY=xxxxxxxxxx \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```

{% endtab %}

{% tab title="Google Cloud Storage" %}

Refer to the `Google Cloud specific` parameters on the [Storage for Logs and Cache](./installation-configuration.md#google-cloud-storage-specific) page.

Run the following command to install Devtron along with Google Cloud Storage for storing build logs and cache:

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=GCP \
--set secrets.BLOB_STORAGE_GCP_CREDENTIALS_JSON=eyJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsInByb2plY3RfaWQiOiAiPHlvdXItcHJvamVjdC1pZD4iLCJwcml2YXRlX2tleV9pZCI6ICI8eW91ci1wcml2YXRlLWtleS1pZD4iLCJwcml2YXRlX2tleSI6ICI8eW91ci1wcml2YXRlLWtleT4iLCJjbGllbnRfZW1haWwiOiAiPHlvdXItY2xpZW50LWVtYWlsPiIsImNsaWVudF9pZCI6ICI8eW91ci1jbGllbnQtaWQ+IiwiYXV0aF91cmkiOiAiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLCJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vb2F1dGgyL3YxL2NlcnRzIiwiY2xpZW50X3g1MDlfY2VydF91cmwiOiAiPHlvdXItY2xpZW50LWNlcnQtdXJsPiJ9Cg== \
--set configs.DEFAULT_CACHE_BUCKET=cache-bucket \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=log-bucket
```

{% endtab %}
{% endtabs %}


## Check Status of Devtron Installation

**Note**: The installation takes about 15 to 20 minutes to spin up all of the Devtron microservices one by one.

 Run the following command to check the status of the installation:

```bash
kubectl -n devtroncd get installers installer-devtron \
-o jsonpath='{.status.sync.status}'
```

The command executes with one of the following output messages, indicating the status of the installation:

| Status | Description |
| :--- | :--- |
| `Downloaded` | The installer has downloaded all the manifests, and the installation is in progress. |
| `Applied` | The installer has successfully applied all the manifests, and the installation is completed. |


## Check the installer logs

Run the following command to check the installer logs:

```bash
kubectl logs -f -l app=inception -n devtroncd
```

## Devtron dashboard

Run the following command to get the Devtron dashboard URL:

```bash
kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
```

You will get an output similar to the example shown below:

```bash
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

Use the hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` (Loadbalancer URL) to access the Devtron dashboard.

**Note**: If you do not get a hostname or receive a message that says "service doesn't exist," it means Devtron is still installing. 
Please wait until the installation is completed.

**Note**: You can also use a `CNAME` entry corresponding to your domain/subdomain to point to the Loadbalancer URL to access at a customized domain.

| Host | Type | Points to |
| :--- | :--- | :--- |
| devtron.yourdomain.com | CNAME | aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com |


## Devtron Admin credentials

When you install Devtron for the first time, it creates a default admin user and password (with unrestricted access to Devtron). You can use that credentials to log in as an administrator. 

After the initial login, we recommend you set up any SSO service like Google, GitHub, etc., and then add other users (including yourself). Subsequently, all the users can use the same SSO (let's say, GitHub) to log in to Devtron's dashboard.

The sections below will help you understand the process of getting the administrator password.

### For Devtron version v0.6.0 and higher

**Username**: `admin` <br>
**Password**: Run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

<details>
<summary>For Devtron version less than v0.6.0</summary>

**Username**: `admin` <br>
**Password**: Run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```
</details>


* If you want to uninstall Devtron or clean Devtron helm installer, refer our [uninstall Devtron](https://docs.devtron.ai/install/uninstall-devtron).

* Related to installaltion, please also refer [FAQ](https://docs.devtron.ai/install/faq-on-installation) section also.



**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)

