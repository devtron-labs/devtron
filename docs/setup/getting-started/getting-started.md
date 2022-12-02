# Getting Started
 

Before effectively using Devtron, it is very IMPORTANT to know the prerequisite requirements so that you can onboard Devtron as smoothly as possible. Please make sure to meet all the prerequisite requirements before you proceed with Devtron installation.

Devtron is installed over a Kubernetes cluster. Once you create a Kubernetes cluster, Devtron can be installed standalone or along with CI/CD integration:

* [Devtron with CI/CD](setup/install/install-devtron-with-cicd.md): Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability.
* [Helm Dashboard by Devtron](setup/install/install-devtron.md): The Helm Dashboard by Devtron, which is a standalone installation, includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=).

In this section, we will cover the basic details on how you can quickly get started with **Devtron**.
First, lets see what are the prerequisite requirements before you install Devtron.

## Pre-requisite Requirements
* Create a [Kubernetes cluster, preferably K8s version 1.16 or higher](#create-a-kubernetes-cluster)
* [Helm Installation](https://helm.sh/docs/intro/install/)
* [Recommended Resources](#recommended-resources)


### Create a Kubernetes Cluster
 
You can create any [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s version 1.16 or higher) for installing Devtron.

You can create a cluster using one of the following cloud providers as per your requirements:

| Cloud Provider | Description |
| --- | --- |
| **AWS EKS** | Create a cluster using [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html). <br>`Note`: You can also refer our customized documentation for installing  `Devtron with CI/CD` on AWS EKS [here](setup/install/install-devtron-on-AWS-EKS.md).</br>  |
| **Google Kubernetes Engine (GKE)** | Create a cluster using [GKE](https://cloud.google.com/kubernetes-engine/). |
| **Azure Kubernetes Service (AKS)** | Create a cluster using [AKS](https://learn.microsoft.com/en-us/azure/aks/). | 
| **k3s - Lightweight Kubernetes** | Create a cluster using [k3s - Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/).<br>`Note`: You can also refer our customized documentation for installing `Helm Dashboard by Devtron` on `Minikube, Microk8s, K3s, Kind` [here](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).</br> | 

**Note**: 
* We recommend to create a cluster on [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html), [GKE](https://cloud.google.com/kubernetes-engine/) or [AKS](https://learn.microsoft.com/en-us/azure/aks/) for installing Devtron on production environment. 
* For development environment, we recommend [k3s - Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/).


### Install Helm

Make sure to install [helm](https://helm.sh/docs/intro/install/).



### Recommended Resources

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

>**Note:** 
* Please make sure that the recommended resources are available on your Kubernetes cluster before you proceed with Devtron installation.
* It is NOT recommended to use brustable CPU VMs (T series in AWS, B Series in Azure and E2/N1 in GCP) for Devtron installation.
 


## Installation of Devtron

As it is mentioned before, you can install Devtron standalone (Helm Dashboard by Devtron) or along with CI/CD integration. 
Or, you can upgrade Devtron to the latest version.

Choose one of the options as per your requirements:

| Installation Options | Description |
| --- | --- |
| [Devtron with CI/CD](setup/install/install-devtron-with-cicd.md) | Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability. |
| [Helm Dashboard by Devtron](setup/install/install-devtron.md) | The Helm Dashboard by Devtron which is a standalone installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=). |
| [Devtron with CI/CD along with GitOps](setup/install/install-devtron-with-cicd.md) | Devtron uses GitOps to automate the process of provisioning infrastructure. GitOps configuration files generate the same infrastructure environment every time it’s deployed, just as application source code generates the same application binaries every time it’s built. |
| **Upgrade Devtron to latest version** | You can upgrade Devtron in one of the following ways:<ul><li>[Upgrade Devtron using Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Upgrade Devtron from UI](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li> |

**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)


