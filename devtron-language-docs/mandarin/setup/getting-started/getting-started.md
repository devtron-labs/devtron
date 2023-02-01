# 入门
本节包含关于安装和使用 **Devtron** 所需的最低要求的信息。

Devtron 安装在 Kubernetes 群集上。创建 Kubernetes 群集后，Devtron 可以独立安装，也可以与 CI/CD 集成一起安装：

* [带 CI/CD 的 Devtron](setup/install/install-devtron-with-cicd.md)：带 CI/CD 集成的 Devtron 安装用于执行 CI/CD、安全扫描、GitOps、调试和可观测性。
* [Devtron Helm 仪表板](setup/install/install-devtron.md)：Devtron Helm 仪表板是一个独立安装，包括在多个群集中部署、观测、管理和调试现有 Helm 应用程序的功能。您也可以从 [Devtron 堆栈管理器](https://docs.devtron.ai/v/v0.6/usage/integrations?q=)安装集成。

在本节中，我们将介绍如何快速开始使用 **Devtron** 的基本细节。
首先，让我们看看在安装 Devtron 之前有哪些必备要求。
## 必备要求
* 创建 [Kubernetes 群集，最好是 K8s v1.16 或更高版本](#create-a-kubernetes-cluster)
* [Helm 安装](https://helm.sh/docs/intro/install/)
* [建议的资源](#recommended-resources)
### 创建 Kubernetes 群集
您可以创建任何 [Kubernetes 群集](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/)（最好是 K8s v1.16 或更高版本）来安装 Devtron。

您可以根据需要使用以下云提供程序之一创建群集：

|云提供程序|说明|
| :-: | :-: |
|**AWS EKS**|使用 [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html) 创建群集。<br>`注意`：[在这里](setup/install/install-devtron-on-AWS-EKS.md)，您还可以参考我们的定制文档，在 AWS EKS 上安装`带 CI/CD 的 Devtron`。</br>|
|**谷歌 Kubernetes 引擎 (GKE)**|使用 [GKE](https://cloud.google.com/kubernetes-engine/) 创建群集。|
|**Azure Kubernetes Service (AKS)**|使用 [AKS](https://learn.microsoft.com/en-us/azure/aks/) 创建群集。|
|**k3s - Lightweight Kubernetes**|使用 [k3s - Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/) 创建群集。<br>`注意`：[在这里](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md)，您也可以参考我们的定制文档，在 `Minikube、Microk8s、K3s 和 Kind` 上安装 `Devtron Helm 仪表板`。</br>|

### 安装 Helm
确保已安装 [Helm](https://helm.sh/docs/intro/install/)。
### 建议的资源
根据您希望在 `Devtron` 上管理的应用程序数量，安装 `Devtron Helm 仪表板`和`带 CI/CD 的 Devtron` 的最低要求如下：

* 对于配置小型资源（在 Devtron 上管理不超过 5 个应用程序）：

|集成|CPU|内存|
| :-: | :-: | :-: |
|**带 CI/CD 的 Devtron**|2|6 GB|
|**Devtron Helm 仪表板**|1|1 GB|

* 对于配置中型/大型资源（在 Devtron 上管理 5 个以上的应用程序）：

|集成|CPU|内存|
| :-: | :-: | :-: |
|**带 CI/CD 的 Devtron**|6|13 GB|
|**Devtron Helm 仪表板**|2|3 GB|

> 有关更多信息，请参考[替代配置](setup/install/override-default-devtron-installation-configs.md)部分。

> **注意：**

* 在继续安装 Devtron 之前，请确保您的 Kubernetes 群集上有建议的可用资源。
* 不建议使用可突增性能 CPU 的 VMs（AWS 中的 T 系列、Azure 中的 B 系列和 GCP 中的 E2/N1）来安装 Devtron，以体验性能的一致性。
## 安装 Devtron
您可以单独安装 Devtron（Devtron Helm 仪表板）或与 CI/CD 集成一起安装。或者，您可以将 Devtron 升级到最新版本。

根据您的要求选择其中一个选项：

|安装选项|说明|
| :-: | :-: |
|[带 CI/CD 的 Devtron](setup/install/install-devtron-with-cicd.md)|带 CI/CD 集成的 Devtron 安装用于执行 CI/CD、安全扫描、GitOps、调试和可观测性。|
|[Devtron Helm 仪表板](setup/install/install-devtron.md)|Devtron Helm 仪表板是一个独立安装，包括在多个群集中部署、观测、管理和调试现有 Helm 应用程序的功能。您也可以从 [Devtron 堆栈管理器](https://docs.devtron.ai/v/v0.6/usage/integrations?q=)安装集成。|
|[带 CI/CD 的 Devtron 与 GitOps 一起 (Argo CD)](setup/install/install-devtron-with-cicd-with-gitops.md)|使用此选项，您可以通过在安装过程中启用 GitOps 来安装带 CI/CD 的 Devtron。您也可以从 [Devtron 堆栈管理器](https://docs.devtron.ai/v/v0.6/usage/integrations?q=)安装其他集成。|
|**将 Devtron 升级到最新版本**|您可以通过以下方法之一升级 Devtron：<ul><li>[使用 Helm 升级 Devtron](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[从 UI 升级 Devtron](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li>|

**注意**：如果您有任何问题，请通过我们的 Discord 频道告诉我们。![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
