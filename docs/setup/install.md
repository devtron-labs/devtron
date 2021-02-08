[![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)

# Devtron Installation

## Prerequisites

You will need to be ready with following prerequisites before Devtron installation
 - A Kubernetes cluster (preferably K8s 1.16 or above) created on AWS (EKS or KOPS) or on AZURE (AKS or KOPS). Check [Creating a Production grade EKS cluster using EKSCTL](https://devtron.ai/blog/creating-production-grade-kubernetes-eks-cluster-eksctl/)
 - An Nginx ingress controller pre-configured within the cluster either exposed as LoadBalancer or NodePort.
 - 2 S3 buckets/ AZURE Blob storage containers for ci-caching and ci-logs and their access permissions added to the cluster role. You will need to put these in configs [here](#storage-for-logs-and-cache)


## Introduction

This installer bootstraps deployment of all required components for installation of [Devtron Platform](https://github.com/devtron-labs) on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager or kubectl cli.

It packages third party components like:

 - [Grafana](https://github.com/grafana/grafana) for displaying application metrics
 - [Argocd](https://github.com/argoproj/argo-cd/) for gitops
 - [Argo workflows](https://github.com/argoproj/argo) for CI
 - [Clair](https://github.com/quay/clair) & [Guard](https://github.com/guard/guard) for image scanning
 - [Kubernetes External Secrets](https://github.com/godaddy/kubernetes-external-secrets) for integrating with external secret management stores like [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/) or [HashiCorp Vault](https://www.vaultproject.io/)
 - [Nats](https://github.com/nats-io) for event streaming
 - [Postgres](https://github.com/postgres/postgres) as datastore
 - Fork of [Argo Rollout](https://github.com/argoproj/argo-rollouts)

## How to use it



### Install with Helm

## Helm 3

```bash
git clone https://github.com/devtron-labs/devtron-installation-script.git
cd devtron-installation-script/charts
```
Copy and edit the `devtron/values.yaml` to configure your Devtron installation.
```bash
cp devtron/values.yaml devtron/install-values.yaml
vim devtron/install-values.yaml
```
For more details about configuration see the [helm chart configuration](#configuration)
Once your configurations are ready, continue with following steps
```bash
#Create devtroncd namespace
kubectl create ns devtroncd
helm install devtron devtron/ --namespace devtroncd -f devtron/install-values.yaml
```

## Helm 2

```bash
git clone https://github.com/devtron-labs/devtron-installation-script.git
cd devtron-installation-script/charts
#Create CRDs manually when using Helm2
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron-installation-script/main/charts/devtron/crds/crd-devtron.yaml
```
Copy and edit the `devtron/values.yaml` to configure your Devtron installation.
```bash
cp devtron/values.yaml devtron/install-values.yaml
vim devtron/install-values.yaml
```
For more details about configuration see the [helm chart configuration](#configuration)
Once your configurations are ready, continue with following steps
```bash
#Create devtroncd namespace
kubectl create ns devtroncd
helm install devtron/ --name devtron --namespace devtroncd -f devtron/install-values.yaml
```



### Install with kubectl

If you don't want to install helm and just want to use `kubectl` to install `devtron platform`, then please follow the steps mentioned below:

```bash
git clone https://github.com/devtron-labs/devtron-installation-script.git
cd devtron-installation-script/
#Use a preferred editor to edit the values in install/devtron-operator-configs.yaml
vim install/devtron-operator-configs.yaml
```
Edit the `install/devtron-operator-configs.yaml` to configure your Devtron installation. For more details about it, see [configuration](#configuration)
Once your configurations are ready, continue with following steps
```bash
kubectl create ns devtroncd
kubectl -n devtroncd apply -f charts/devtron/crds
# wait for crd to install
kubectl apply -n devtroncd -f charts/devtron/templates/install.yaml
#edit install/devtron-operator-configs.yaml and input the
kubectl apply -n devtroncd -f install/devtron-operator-configs.yaml
kubectl apply -n devtroncd -f charts/devtron/templates/devtron-installer.yaml
```

### Installation status
Run following command
   ```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

Once installation process is complete, above command will print `Applied`
It may take around 30 minutes for installation to complete.

### Access devtron dashboard

#### Dashboard URL
Devtron dashboard in now available at the `BASE_URL/dashboard`, where `BASE_URL` is same as
provided in `values.yaml` in case of installation via helm chart
OR
provided in `install/devtron-operator-configs.yaml` in case of installation via kubectl.

Run following command to get dashboard
```bash
scheme=`kubectl -n devtroncd get cm devtron-operator-cm -o jsonpath='{.data.BASE_URL_SCHEME}'` && url=`kubectl -n devtroncd get cm devtron-operator-cm -o jsonpath='{.data.BASE_URL}'` && echo "$scheme://$url/dashboard"
```
**Please Note:** URL should be pointing to the cluster on which you have installed the platform. For example if you have directed domain `devtron.example.com` to the cluster and ingress controller is listening on port `32080` then url will be `devtron.example.com:32080`

#### Login credentials
For login use username:`admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```
### Access grafana

#### Grafana URL
grafana is available at `BASE_URL/grafana`

#### Login credentials
To log into grafana use username: `admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-grafana-cred-secret -o jsonpath='{.data.admin-password}' | base64 -d
```

### Configuration

#### Configure Secrets
For `helm` installation this section referes to ***secrets.env*** section of `values.yaml`.
For `kubectl` based installation it refers to `kind: secret` in ***install/devtron-operator-configs.yaml***.

Following properties should be configured

| Parameter | Description | Default |
|----------:|:------------|:--------|
| **POSTGRESQL_PASSWORD** | password for postgres database, should be base64 encoded (required) | change-me |
| **GIT_TOKEN** | git token for the gitops work flow, please note this is not for source code of repo and this token should have full access to create, delete, update repository, should be base64 encoded (required) |  |
| **WEBHOOK_TOKEN** | If you want to continue using jenkins for CI then please provide this for authentication of requests, should be base64 encoded  |  |

#### Configure ConfigMaps
For `helm` installation this section referes to ***configs*** section of `values.yaml`.
For `kubectl` based installation it refers to `kind: ConfigMap` in ***install/devtron-operator-configs.yaml***.

Following properties should be configured

| Parameter | Description | Default |
|----------:|:------------|:--------|
| **BASE_URL_SCHEME** | either of http or https (required) | http |
| **BASE_URL** | url without scheme and trailing slash, this is the domain pointing to the cluster on which devtron platform is being installed. For example if you have directed domain `devtron.example.com` to the cluster and ingress controller is listening on port `32080` then url will be `devtron.example.com:32080` (required) | `change-me` |
| **DEX_CONFIG** | dex config if you want to integrate login with SSO (optional) for more information check [Argocd documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/) |
| **GIT_PROVIDER** | git provider for storing config files for gitops, currently only GITHUB and GITLAB are supported (required) | `GITHUB` | |
| **GITLAB_NAMESPACE_NAME** | if GIT_PROVIDER is GITLAB, this is mandatory and should be already created | |
| **GIT_USERNAME** | git username for the GIT_PROVIDER  (required) | |
| **GITHUB_ORGANIZATION** | if GIT_PROVIDER is GITHUB, this is mandatory and should be already created | |
| **EXTERNAL_SECRET_AMAZON_REGION** | AWS region for secret manager to pick (required) |  |
| **PROMETHEUS_URL** | url of prometheus where all cluster data is stored, if this is wrong, you will not be able to see application metrics like cpu, ram, http status code, latency and throughput (required) |  |
| **GIT_HOST** | if GIT_PROVIDER is GITLAB, this is required only when user want to use self hosted gitlab. provide valid git host URL | |

### Storage for Logs and Cache

AWS SPECIFIC
| Parameter | Description | Default |
|----------:|:------------|:--------|
| **DEFAULT_CD_LOGS_BUCKET_REGION** | AWS region of bucket to store CD logs, this should be created before hand (required) | |
| **DEFAULT_CACHE_BUCKET** | AWS bucket to store docker cache, this should be created before hand (required) |  |
| **DEFAULT_CACHE_BUCKET_REGION** | AWS region of cache bucket defined in previous step (required) | |
| **DEFAULT_BUILD_LOGS_BUCKET** | AWS bucket to store build logs, this should be created before hand (required) | |


AZURE SPECIFIC
| Parameter | Description | Default |
|----------:|:------------|:--------|
|  **AZURE_ACCOUNT_NAME** | AZURE Blob storage account name
|  **AZURE_BLOB_CONTAINER_CI_LOG** | AZURE Blob storage container for storing ci-logs
|  **AZURE_BLOB_CONTAINER_CI_CACHE** | AZURE Blob storage container for storing ci cache


To convert string to base64 use

```bash
echo -n "string" | base64 -d
```
**Please Note:**
1) Ensure that the **cluster has read and write access** to the S3 buckets/Azure Blob storage container mentioned in DEFAULT_CACHE_BUCKET, DEFAULT_BUILD_LOGS_BUCKET or AZURE_BLOB_CONTAINER_CI_LOG, AZURE_BLOB_CONTAINER_CI_CACHE
2) Ensure that cluster has **read access** to AWS secrets backends (SSM & secrets manager)


### Cleanup

Run following commands to delete all the components installed by devtron
```bash
cd devtron-installation-script/
kubectl delete -n devtroncd -f yamls/
kubectl delete -n devtroncd -f charts/devtron/templates/devtron-installer.yaml
kubectl delete -n devtroncd -f charts/devtron/templates/install.yaml
kubectl delete -n devtroncd -f charts/devtron/crds
kubectl delete ns devtroncd
```
#### Cleaning Installer Helm3
```bash
cd devtron-installation-script/
helm delete devtron --namespace devtroncd
```

#### Cleaning Installer Helm2
```bash
cd devtron-installation-script/
helm delete devtron --purge
#Deleting CRDs manually
kubectl delete -f https://raw.githubusercontent.com/devtron-labs/devtron-installation-script/main/charts/devtron/crds/crd-devtron.yaml
```


### Trouble shooting steps

 **1**. How do I know when installation is complete?

Run following command
   ```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

Once installation process is complete, above command will print `Applied`
It may take around 30 mins for installation to complete

**2**. How do I track progress of installation?

Run following command to check logs of the pod
 ```bash
pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod
```
**3**. devtron installer logs have error and I want to restart installation.

Run following command to clean up components installed by devtron installer
 ```bash
cd devtron-installation-script/
kubectl delete -n devtroncd -f yamls/
kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'
```

 In case you are still facing issues please feel free to reach out to us on [discord](https://discord.gg/jsRG5qx2gp)
