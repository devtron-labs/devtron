# Getting Started
 

Before effectively using Devtron, it is necessary to understand the underlying technology that the Devtron platform is built on. 
It is very IMPORTANT to know the prerequisite requirements so that you can onboard Devtron smoothly as possible according to your needs. 

In this section, we will cover the basic details on how you can quickly get started with **Devtron**.
 First, lets see what are the prerequisite requirements before you install Devtron.

## Pre-requisite Requirements
* [Recommended Resources](#recommended-resources)
*  [Kubernetes Cluster, preferably K8s version 1.16 or higher](#create-a-kubernetes-cluster)
* [Helm Installation](https://helm.sh/docs/intro/install/)

You can install Devtron after fulfilling the prerequisite requirements.

Devtron is installed over a Kubernetes cluster and can be installed standalone or along with CI/CD integration:

* [Devtron with CI/CD](setup/install/install-devtron-with-cicd.md): Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability.
* [Devtron](setup/install/install-devtron.md): The Devtron standalone installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters and can also integrate with multiple tools using extensions.

 

### Recommended Resources

The minimum requirements for Devtron and Devtron with CI/CD integration in production and non-production environments include:

* Non-production

| Integration | CPU | Memory |
| --- | :---: | :---: |
| **Devtron with CI/CD** | 2 | 6 GB |
| **Devtron** | 1 | 1 GB |

* Production (assumption based on 5 clusters)

| Integration | CPU | Memory |
| --- | :---: | :---: |
| **Devtron with CI/CD** | 6 | 13 GB |
| **Devtron** | 2 | 3 GB |

> Refer to the [Override Configurations](setup/install/override-default-devtron-installation-configs.md) section for more information.

>**Note:** It is NOT recommended to use brustable CPU VMs (T series in AWS, B Series in Azure and E2/N1 in GCP) for Devtron installation.
 
### Create a Kubernetes Cluster
 
You can create any [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s version 1.16 or higher) for trying Devtron out in a local development environment.

If you want to set up a cluster in the production environment in `AWS EKS`, we have the customized documentation for installing Devtron with CI/CD in [AWS EKS](setup/install/install-devtron-on-AWS-EKS.md).

Or, if you want to install **Devtron** on `Minikube, Microk8s, K3s, Kind`, we have the customized documentation available, refer [Minikube, Microk8s, K3s, Kind](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).
 
If you want to set up a cluster in the production environment, refer to the [Creating a Production grade EKS cluster using EKSCTL](https://devtron.ai/blog/creating-production-grade-kubernetes-eks-cluster-eksctl/) article.

### Install Helm

Make sure to install [helm]((https://helm.sh/docs/intro/install/))

Helm is a package manager for Kubernetes that makes it possible to download charts, which are pre-packaged collections of all the necessary versioned, pre-configured resources required to deploy a container. Helm charts are most useful when first setting up a Kubernetes cluster to deploy simple applications.


## Installation of Devtron

| Installation Options | Description |
| --- | --- |
| **Devtron with CI/CD** | `Install Devtron with CI/CD` on hosted Kubernetes with one of the options:<ul><li>[AWS EKS](install-devtron-on-AWS-EKS.md)</ul></li><ul><li>[Install Devtron with CI/CD](setup/install/install-devtron-with-cicd.md) on your preferred cluster</ul></li>  |
| **Devtron** | `Install Devtron without CI/CD` with one of the options:<ul><li>[Install Devtron](setup/install/install-devtron.md) on your preferred cluster</ul></li><ul><li>[Install Devtron on Minikube, Microk8s, K3s, Kind](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md)</ul></li>|
| **Upgrade Devtron to latest version** | You can upgrade Devtron in one of the following ways:<ul><li>[Upgrade Devtron using Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Upgrade Devtron from UI](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li> |

**Note**: If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)


