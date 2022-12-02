# Install Devtron with CI/CD along with GitOps (Argo CD)


## Before you begin

Install [Helm](https://helm.sh/docs/intro/install/) if you have not installed it.


## Run the following command to add helm repo

```bash
helm repo add devtron https://helm.devtron.ai
```

## Run the following command to install Devtron with CI/CD along with GitOps (Argo CD) integration


```bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set argo-cd.enabled=true
```

**Note**: To install Devtron on clusters with the multi-architecture nodes (ARM and AMD), append the installation command with `--set installer.arch=multi-arch`.

## Configure Blob Storage

If you want to configure blob storage during the installation of Devtron with CI/CD along with GitOPs (Argo CD), please do not skip this step.

Configuring Blob Storage in your Devtron environment allows you to store build logs and cache. In case, if you do not configure the Blob Storage, then:
* You will not be able to access the build and deployment logs after an hour.
* Build time for commit hash takes longer as cache is not available.
* Artifact reports cannot be generated in pre/post build and deployment stages.


Use the following command to install Devtron along with MinIO for storing logs and cache. 

```bash
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true \
--set argo-cd.enabled=true
```

Note: Unlike global cloud providers such as AWS S3 Bucket, Azure Blob Storage and Google Cloud Storage, MinIO can be hosted locally also.


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
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

#### For Devtron version less than v0.6.0

Use username:`admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

## Cleaning Devtron Helm installer

Please make sure that you do not have anything inside namespaces devtroncd, devtron-cd, devtron-ci, and devtron-demo as the below steps will clean everything inside these namespaces.

Run the following command to clean Devtron Helm installer:

```bash
helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo

```

## Next, you can proceed with Configuratios, if require

Refer our [Configurations](installation-configuration.md) documentation for detail.


**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
