# Devtron On Graviton
In cloud computing, optimizing performance, efficiency, and cost-effectiveness is an endless pursuit. As technology evolves, new opportunities arise to achieve these goals. One such advancement is the introduction of AWS Graviton instances, which are rapidly gaining prominence as a game-changer in cloud architecture. 

AWS Graviton instances are a family of Arm-based processors developed by Amazon Web Services (AWS). These processors are designed to deliver high performance while maintaining energy efficiency.

We are thrilled to announce that Devtron seamlessly supports Graviton instances, and it's fantastic to note the substantial benefits we've experienced in terms of resource utilization. With 5% less memory utilization and 2% less CPU utilization compared to AMD instances, underscore the advantages of leveraging Graviton architecture. This not only translates into cost savings but also contributes to a more environmentally sustainable cloud infrastructure.

## Installation

To install Devtron on graviton-cluster, refer this [link](../setup/install/install-devtron-with-cicd.md#install-multi-architecture-nodes-arm-and-amd)

## Inferences

### 1. Reduced Build Time

The utilization of Graviton machines for building Graviton architecture has led to reduction in build time by approximately 30% and less CPU/Memory utilization within Devtron.

**AMD Build**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/amd-build.png)

![Resource Utilization on AMD](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/amd/build-amd.png)

**ARM Build**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/arm-build.png)

![Resource Utilization on ARM](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resources/graviton/arm/build-arm.png)

<hr />

### 2. Similar Performance

Experience Devtron's equivalent performance to that of other architectures, all while exhibiting slightly lower resource utilization.

<hr />


### 3. Less Resource Utilization

Notably, Graviton instances exhibit slightly lower resource utilization compared to AMD Nodes, users can take the opportunity for cost savings in cloud operations.

We have attached some snapshots of the resource utilization for the critical micro-services on Devtron having the GitOps option enabled and more than 115 applications deployed. For an accurate performance comparison, We have used a single-node cluster for both architectures (AMD and ARM).

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




