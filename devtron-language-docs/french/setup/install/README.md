\# Installer Devtron


Devtron s'installe sur un cluster Kubernetes. Une fois le cluster Kubernetes créé, Devtron peut être installé de manière autonome ou avec une intégration CI/CD.

Choisissez l'une des options en fonction de vos besoins :

| Options d'installation | Description | Quand choisir  |

\| --- | --- | --- |

| [Devtron avec CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) | L'installation de Devtron avec l'intégration CI/CD est utilisée pour effectuer le CI/CD, l'analyse de sécurité, GitOps, le débogage et l'observabilité. | Utilisez cette option pour installer Devtron avec l'intégration `Build and Deploy CI/CD` . |

| [Tableau de bord Helm de Devtron](https://docs.devtron.ai/install/install-devtron) | Le tableau de bord Helm de Devtron, qui est une installation autonome, comprend des fonctionnalités pour déployer, observer, gérer et déboguer les applications Helm existantes dans de multiples clusters. Vous pouvez également installer des intégrations à partir de [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations). | Utilisez cette option si vous gérez les applications via Helm et que vous souhaitez utiliser Devtron pour déployer, observer, gérer et déboguer les applications Helm. |

| [Devtron avec CI/CD en parallèle avec GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops) | Avec cette option, vous pouvez installer Devtron avec CI/CD en activant GitOps pendant l'installation. Vous pouvez également installer d'autres intégrations à partir de [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations). |  Utilisez cette option pour installer Devtron avec CI/CD en activant GitOps, qui est la méthode la plus évolutive en termes de contrôle de version, de collaboration, de conformité et d'automatisation d'infrastructure.  |


\*\*Note\*\* : Si vous avez des questions, veuillez nous en faire part sur notre canal discord. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
