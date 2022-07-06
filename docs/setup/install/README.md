# Installing Devtron
 
[![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
 
Devtron is installed over a Kubernetes cluster and can be installed standalone or along with CI/CD integration:

* [Devtron with CI/CD](install-devtron-with-cicd.md): Devtron installation with the CI/CD integration is used to perform CI/CD, security scanning, GitOps, debugging, and observability.
* [Devtron](install-devtron.md): The Devtron installation includes functionalities to deploy, observe, manage, and debug existing Helm applications in multiple clusters and deeply integrate with multiple tools using extensions.

## Recommended resources

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

> Refer to the [Override Configurations](./override-default-devtron-installation-configs.md) section for more information.
 
## Before you begin
 
Create a [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s 1.16 or higher) if you haven't done that already!
 
Refer to the [Creating a Production grade EKS cluster using EKSCTL](https://devtron.ai/blog/creating-production-grade-kubernetes-eks-cluster-eksctl/) article to set up a cluster in the production environment.

## Installing Devtron
 
* [Install Devtron with CI/CD integration](install-devtron-with-cicd.md)
* [Install Devtron](install-devtron.md)
* [Upgrade Devtron to latest version](/docs/setup/upgrade/README.md)
 
