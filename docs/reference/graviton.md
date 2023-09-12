# Devtron On Graviton

## Installation

To install Devtron on graviton-cluster, refer this [link](https://docs.devtron.ai/install/install-devtron-with-cicd#install-multi-architecture-nodes-arm-and-amd)

## Inferences

### 1. Reduced Build Time

You can infer from the below snapshots that the build time on Devtron reduced to approximately 2 minutes when using Graviton machine for ARM build.

**AMD Build**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/amd-build.png)

![Resource Utilization on AMD](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/build-amd.png)

**ARM Build**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/arm-build.png)

![Resource Utilization on ARM](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/build-arm.png)

<hr />

### 2. Similar Performance

Performance is same as that of other architectures. There is no significant difference.

<hr />


### 3. Less Resource Utilization

Slightly less resource utilization than the AMD Node which can definitely save costs on cloud.

We have attached few snapshots depicting the resource utilization of critical micro-services on Devtron, having the GitOps option enabled and more than 115 applications deployed. To get the actual comparison, we have used a single node cluster for both the architecture (AMD and ARM).

#### 1. orchestrator

**AMD-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/devtron-amd.png)

**ARM-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/devtron-arm.png)


#### 2. argocd-server

**AMD-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/amd-argo-server.png)

**ARM-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/argocd-server-arm.png)

#### 3. argocd-application-controller

**AMD-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/amd-app-controller.png)

**ARM-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/app-controller-arm.png)

#### 4. argocd-repo-server

**AMD-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/amd-repo-server.png)

**ARM-Based**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/repo-server-arm.png)




