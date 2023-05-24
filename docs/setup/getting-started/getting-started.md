# Getting Started
 
This section includes information about the minimum requirements you need to install and use **Devtron**.

Devtron is installed over a Kubernetes cluster. Once you create a Kubernetes cluster, Devtron can be installed standalone or along with CI/CD integration:

* [Devtron with CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd): Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability.
* [Helm Dashboard by Devtron](https://docs.devtron.ai/install/install-devtron): The Helm Dashboard by Devtron, which is a standalone installation, includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations).

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
| **AWS EKS** | Create a cluster using [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html). <br>`Note`: You can also refer our customized documentation for installing  `Devtron with CI/CD` on AWS EKS [here](https://github.com/devtron-labs/devtron/blob/b33a37bb608d07966c8f8b89e4f59287db873c6c/docs/setup/install/install-devtron-on-aws-eks.md).</br>  |
| **Google Kubernetes Engine (GKE)** | Create a cluster using [GKE](https://cloud.google.com/kubernetes-engine/). |
| **Azure Kubernetes Service (AKS)** | Create a cluster using [AKS](https://learn.microsoft.com/en-us/azure/aks/). | 
| **k3s - Lightweight Kubernetes** | Create a cluster using [k3s - Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/).<br>`Note`: You can also refer our customized documentation for installing `Helm Dashboard by Devtron` on `Minikube, Microk8s, K3s, Kind` [here](https://docs.devtron.ai/install/install-devtron-on-minikube-microk8s-k3s-kind).</br> | 



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

> Refer to the [Override Configurations](https://docs.devtron.ai/configurations-overview/override-default-devtron-installation-configs) section for more information.

**Note:**
* Please make sure that the recommended resources are available on your Kubernetes cluster before you proceed with Devtron installation.
* It is NOT recommended to use brustable CPU VMs (T series in AWS, B Series in Azure and E2/N1 in GCP) for Devtron installation to experience consistency in performance.
 

## Installation of Devtron

You can install Devtron standalone (Helm Dashboard by Devtron) or along with CI/CD integration. Or, you can upgrade Devtron to the latest version.

Choose one of the options as per your requirements:

| Installation Options | Description |
| --- | --- |
| [Devtron with CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) | Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability. |
| [Helm Dashboard by Devtron](https://docs.devtron.ai/install/install-devtron) | The Helm Dashboard by Devtron which is a standalone installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters. You can also install integrations from [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations). |
| [Devtron with CI/CD along with GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops) | With this option, you can install Devtron with CI/CD by enabling GitOps during the installation. You can also install other integrations from [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations). |
| **Upgrade Devtron to latest version** | You can upgrade Devtron in one of the following ways:<ul><li>[Upgrade Devtron using Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Upgrade Devtron from UI](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li> |

**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)


