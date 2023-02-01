#安装Devtron


Devtron安装在Kubernetes集群上。您创建Kubernetes集群后，Devtron可以独立安装或与CI/CD集成一起安装。

根据您的要求选择以下其中一个选项：

| 安装选项 | 说明 | 何时选择 |

\| --- | --- | --- |

| [带有CI/CD的 Devtron]（https://docs.devtron.ai/install/install-devtron-with-cicd）| Devtron安装与CI/CD集成用于执行CI/CD、安全扫描、GitOps、调试和可观察性。| 使用此选项以通过“构建和部署CI/CD”集成安装Devtron。 |

| [Devtron的Helm Dashboard]（https://docs.devtron.ai/install/install-devtron）| Devtron的 Helm Dashboard是一个独立安装，包括在多个集群中部署、观察、管理和调试现有Helm应用程序的功能。您还可以从 [Devtron Stack Manager]（https://docs.devtron.ai/v/v0.6/usage/integrations）安装集成。 | 如果您通过Helm管理应用程序并且想要使用Devtron部署、观察、管理和调试Helm应用程序，请使用此选项。 |

| [带有CI/CD以及GitOps（Argo CD）的Devtron]（https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops） | 使用此选项，您可以通过在安装期间启用GitOps来使用CI/CD安装Devtron。您也可以从 [Devtron Stack Manager]（https://docs.devtron.ai/v/v0.6/usage/integrations）安装其他集成。 | 使用此选项通过启用GitOps来安装带有CI/CD的Devtron。这是在版本控制、协作、合规性和基础设施自动化方面最具可扩展性的方法。 |


\*\*注意\*\*：如果您有任何疑问，请通过我们的Discord频道告诉我们。 [![加入 Discord] (https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
