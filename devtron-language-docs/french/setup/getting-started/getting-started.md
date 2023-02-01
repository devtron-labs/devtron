# Mise en route
Dans cette section, vous trouverez des informations sur la configuration minimale requise pour l'installation et l'utilisation de **Devtron**.

Devtron s'installe sur un cluster Kubernetes. Après la création du cluster Kubernetes, Devtron peut être installé de manière autonome ou avec une intégration CI/CD :

* [Devtron avec CI/CD](setup/install/install-devtron-with-cicd.md) : l'installation de Devtron avec une intégration CI/CD permet de mettre en œuvre le principe CI/CD, l'analyse de sécurité, le GitOps, le débogage et l'observabilité.
* [Tableau de bord Helm de Devtron](setup/install/install-devtron.md) : une installation autonome, le tableau de bord Helm de Devtron comporte des fonctionnalités permettant de déployer, d'observer, de gérer et de déboguer des applications Helm existantes dans plusieurs clusters. Vous pouvez aussi installer des intégrations à partir de [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=).

Dans cette section, vous allez découvrir les éléments de base qui vous permettront une mise en route rapide de **Devtron**.
Pour commencer, examinons les prérequis nécessaires à l'installation de Devtron.
## Prérequis nécessaires
* Création d'un [cluster Kubernetes, de préférence la version K8s 1.16 ou ultérieure](#create-a-kubernetes-cluster)
* [Installation de Helm](https://helm.sh/docs/intro/install/)
* [Ressources recommandées](#recommended-resources)
### Création d'un cluster Kubernetes
En vue de l'installation de Devtron, vous pouvez créer n'importe quel [cluster Kubernetes](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (de préférence la version K8s 1.16 ou ultérieure).

Pour la création d'un cluster, vous pouvez avoir recours à l'un des fournisseurs de cloud suivants, selon vos besoins :

|Fournisseur de cloud|Description|
| :-: | :-: |
|**AWS EKS**|Création d'un cluster à l'aide d'[AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html). <br>`Remarque` : vous pouvez aussi consulter notre documentation spécifique à l'installation de `Devtron avec CI/CD` sur AWS EKS [ici](setup/install/install-devtron-on-AWS-EKS.md).</br>|
|**Google Kubernetes Engine (GKE)**|Création d'un cluster à l'aide de [GKE](https://cloud.google.com/kubernetes-engine/).|
|**Azure Kubernetes Service (AKS)**|Création d'un cluster à l'aide d'[AKS](https://learn.microsoft.com/en-us/azure/aks/).|
|**k3s - Kubernetes version allégée**|Création d'un cluster à l'aide de [k3s - Kubernetes version allégée](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/).<br>`Remarque` : vous pouvez aussi consulter notre documentation spécifique à l'installation du `tableau de bord Helm de Devtron` sur `Minikube, Microk8s, K3s, Kind` [ici](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).</br>|

### Installation de Helm
Veillez à installer [Helm](https://helm.sh/docs/intro/install/).
### Ressources recommandées
La configuration minimale requise pour l'installation du `tableau de bord Helm de Devtron` et de `Devtron avec CI/CD` en fonction du nombre d'applications que vous souhaitez gérer sur `Devtron` est indiquée ci-dessous :

* Pour la configuration de ressources moindres (gestion d'un maximum de 5 applications sur Devtron) :

|Intégration|CPU|Mémoire|
| :-: | :-: | :-: |
|**Devtron avec CI/CD**|2|6 Go|
|**Tableau de bord Helm de Devtron**|1|1 Go|

* Pour la configuration de ressources moyennes/importantes (gestion de plus de 5 applications sur Devtron) :

|Intégration|CPU|Mémoire|
| :-: | :-: | :-: |
|**Devtron avec CI/CD**|6|13 Go|
|**Tableau de bord Helm de Devtron**|2|3 Go|

> Pour plus d'informations, reportez-vous à la section [Configurations de contournement](setup/install/override-default-devtron-installation-configs.md).

> **Remarque :**

* Avant d'effectuer l'installation de Devtron, veuillez vérifier que les ressources recommandées sont disponibles sur votre cluster Kubernetes.
* Afin de bénéficier d'une homogénéité des performances, il est recommandé, pour l'installation de Devtron, de NE PAS utiliser de machines virtuelles à processeur burstable (série T d'AWS, série B d'Azure et E2/N1 de GCP).
## Installation de Devtron
Vous pouvez installer Devtron de manière autonome (tableau de bord Helm de Devtron) ou avec une intégration CI/CD. Sinon, vous pouvez mettre à niveau Devtron vers la dernière version.

Choisissez l'une de ces options en fonction de vos besoins :

|Options d'installation|Description|
| :-: | :-: |
|[Devtron avec CI/CD](setup/install/install-devtron-with-cicd.md)|L'installation de Devtron avec une intégration CI/CD est utilisée pour la mise en œuvre du principe CI/CD, de l'analyse de sécurité, du GitOps, du débogage et de l'observabilité.|
|[Tableau de bord Helm de Devtron](setup/install/install-devtron.md)|Une installation autonome, le tableau de bord Helm de Devtron comporte des fonctionnalités permettant de déployer, d'observer, de gérer et de déboguer des applications Helm existantes dans plusieurs clusters. Vous pouvez aussi installer des intégrations à partir de [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=).|
|[Devtron avec CI/CD et GitOps (Argo CD)](setup/install/install-devtron-with-cicd-with-gitops.md)|Cette option permet d'installer Devtron avec CI/CD en activant GitOps pendant l'installation. Vous pouvez aussi installer d'autres intégrations à partir de [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=).|
|**Mise à niveau de Devtron vers la dernière version**|La mise à niveau de Devtron peut se faire selon l'une des méthodes suivantes :<ul><li>[Mettre à niveau Devtron à l'aide de Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Mettre à niveau Devtron à partir de l'UI](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li>|

**Remarque** : n'hésitez pas à nous faire part de vos questions sur notre canal Discord. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
