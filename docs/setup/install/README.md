# Installing Devtron
 
[![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
 
Devtron is installed over a Kubernetes cluster and can be installed in 2 modes:

* [Devtron](install-devtron.md): The base installation of Devtron includes basic Helm charts. Devtron base install helps to deploy, observe, manage, and debug existing Helm applications in all the clusters.
* [Devtron with CI/CD](install-devtron-with-cicd.md): This mode includes the base install and the CI/CD module. The CI/CD module installation helps to perform CI/CD, security scanning, GitOps, Access control, debugging, and observability.

## Recommended resources

The minimum requirements for Devtron basic and Devtron with CI/CD installation in production and non-production environments include:

* Non-production

| Install mode | CPU | Memory |
| --- | :---: | :---: |
| **Devtron Base** | 1 | 1 GB |
| **Devtron + CI/CD** | 2 | 6 GB |

* Production (assumption based on 5 clusters)

| Install mode | CPU | Memory |
| --- | :---: | :---: |
| **Devtron Base** | 2 | 3 GB |
| **Devtron + CI/CD** | 6 | 13 GB |

> Refer to the [Override Configurations](./override-default-devtron-installation-configs.md) section for more information.
 
## Before you begin
 
Create a [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s 1.16 or higher) if you haven't done that already!
 
Refer to the [Creating a Production grade EKS cluster using EKSCTL](https://devtron.ai/blog/creating-production-grade-kubernetes-eks-cluster-eksctl/) article to set up a cluster in the production environment.

## Installing Devtron
 
* [Install Devtron](install-devtron.md)
* [Install Devtron with CI/CD module](install-devtron-with-cicd.md)
* [Upgrade Devtron to Latest Version](#upgrade-devtron)
 
## Upgrade Devtron
 
Run the following command to upgrade Devtron to the latest version:
 
```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]'
```
