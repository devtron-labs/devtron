# Getting Started
 

Before effectively using Devtron, it is necessary to understand the underlying technology that the Devtron platform is built on. 
It is very IMPORTANT to know the prerequisite requirements so that you can onboard Devtron smoothly as possible according to your needs. You can install Devtron after fulfilling the prerequisite requirements.

Devtron is installed over a Kubernetes cluster and can be installed standalone or along with CI/CD integration:

* [Devtron with CI/CD](setup/install/install-devtron-with-cicd.md): Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability.
* [Helm Dashboard by Devtron](setup/install/install-devtron.md): The Helm Dashboard by Devtron which is a standalone installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from `Devtron Stack Manager`.

In this section, we will cover the basic details on how you can quickly get started with **Devtron**.
First, lets see what are the prerequisite requirements before you install Devtron.

## Pre-requisite Requirements
* Create a [Kubernetes cluster, preferably K8s version 1.16 or higher](#create-a-kubernetes-cluster)
* [Helm Installation](https://helm.sh/docs/intro/install/)
* [Recommended Resources](#recommended-resources)


### Create a Kubernetes Cluster
 
You can create any [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s version 1.16 or higher) for trying Devtron out in a local environment.

You can create a cluster using one of the following cloud providers as per your requirements:

| Cloud Provider | Description |
| --- | --- |
| **AWS EKS** | Create a cluster using [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html). <br>`Note`: If you want to refer to our customized documentation for installing `Devtron with CI/CD` on AWS EKS, please refer [here](setup/install/install-devtron-on-AWS-EKS.md).</br>  |
| **Google Kubernetes Engine (GKE)** | Create a cluster using [GKE](https://cloud.google.com/kubernetes-engine/). |
| **Azure Kubernetes Service (AKS)** | Create a cluster using [AKS](https://learn.microsoft.com/en-us/azure/aks/). | 
| **k3s - Lightweight Kubernetes** | Create a cluster using [k3s - Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/).<br>`Note`: If you want to install `Helm Dashboard by Devtron` on `Minikube, Microk8s, K3s, Kind`, please refer our customized documentation [here](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).</br> | 



### Install Helm

Make sure to install [helm]((https://helm.sh/docs/intro/install/))

Helm is a package manager for Kubernetes that makes it possible to download charts, which are pre-packaged collections of all the necessary versioned, pre-configured resources required to deploy a container. Helm charts are most useful when first setting up a Kubernetes cluster to deploy simple applications.


### Recommended Resources

When you specify a Pod, you can optionally specify how much of each resource a container needs. The most common resources to specify are CPU and memory (RAM); CPU and memory are collectively referred to as compute resources, or resources. Compute resources are measurable quantities that can be requested, allocated, and consumed.

The minimum requirements for installing `Helm Dashboard by Devtron` and `Devtron with CI/CD` as per the number of applications you want to manage on `Devtron` are provided below:

* For configuring small resources (to manage not more than 5 apps on Devtron):

| Integration | CPU | Memory |
| --- | :---: | :---: |
| **Devtron with CI/CD** | 2 | 6 GB |
| **Helm Dashboard by Devtron** | 1 | 1 GB |

* For configuring medium/larger resources (to manage more than 5 apps on Devtron):

| Integration | CPU | Memory |
| --- | :---: | :---: |
| **Devtron with CI/CD** | 6 | 13 GB |
| **Helm Dashboard by Devtron** | 2 | 3 GB |

> Refer to the [Override Configurations](setup/install/override-default-devtron-installation-configs.md) section for more information.

>**Note:** It is NOT recommended to use brustable CPU VMs (T series in AWS, B Series in Azure and E2/N1 in GCP) for Devtron installation.
 


## Installation of Devtron

As it is mentioned before, you can install Devtron standalone (Helm Dashboard by Devtron) or along with CI/CD integration.

| Installation Options | Description |
| --- | --- |
| [Devtron with CI/CD](setup/install/install-devtron-with-cicd.md) | Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability. |
| [Helm Dashboard by Devtron](setup/install/install-devtron.md) | The Helm Dashboard by Devtron which is a standalone installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from `Devtron Stack Manager`. |
| **Upgrade Devtron to latest version** | You can upgrade Devtron in one of the following ways:<ul><li>[Upgrade Devtron using Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Upgrade Devtron from UI](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li> |

**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)


