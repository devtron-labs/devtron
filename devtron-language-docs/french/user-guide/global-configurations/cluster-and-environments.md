# Clusters et Environnements
Vous pouvez ajouter vos clusters et environnements Kubernetes existants dans la section `Clusters et Environnements`. Vous devez avoir un accès [super admin](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permissions) pour ajouter un cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)
## Ajouter un cluster :
Pour ajouter un cluster, allez dans la section `Clusters & Environnements` des `Configurations générales`. Cliquez sur **Ajouter cluster**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

Renseignez les champs suivants pour ajouter votre cluster kubernetes :

|Champ|Description|
| :- | :- |
|`Nom`|Entrez le nom de votre cluster.|
|`URL du serveur`|URL du serveur d'un cluster.<br>Note : Nous recommandons d'utiliser une [URL auto-hébergée](#benefits-of-self-hosted-url) plutôt qu'une URL hébergée dans le cloud.</br>|
|`Jeton porteur`|Jeton porteur d'un cluster.|

### Obtenir les identifiants du cluster
> **Prérequis :** `kubectl` et `jq` doivent être installés sur le bastion.

**Note** : Nous recommandons d'utiliser une URL auto-hébergée plutôt qu'une URL hébergée dans le cloud. Référez-vous aux avantages liés à l'[URL auto-hébergée](#benefits-of-self-hosted-url).

Vous pouvez obtenir l'**`URL du serveur`** et le **`Jeton porteur`** en exécutant la commande suivante selon le fournisseur du cluster :

{% tabs %}
{% tab title="k8s Cluster Providers" %}
Si vous utilisez Kubernetes géré par EKS, AKS, GKE, Kops ou Digital Ocean, exécutez la commande suivante pour générer l'URL du serveur et le jeton porteur :
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh \
&& bash kubernetes_export_sa.sh cd-user devtroncd \
https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% tab title="Microk8s Cluster" %}
Si vous utilisez un **`cluster microk8s`**, exécutez la commande suivante pour générer l'URL du serveur et le jeton porteur :
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' \
kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user \
devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% endtabs %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)
### Avantages de l'URL auto-hébergée
* Reprise après sinistre :
  * Il n'est pas possible de modifier l'URL du serveur d'un fournisseur cloud spécifique. Si vous utilisez une URL EKS (e.g.` *****.eu-west-1.elb.amazonaws.com`), ajouter un nouveau cluster et migrer tous les services un par un représente une opération pénible.
  * Mais en cas d'utilisation d'une URL auto-hébergée (par exemple, `clear.example.com`), vous pouvez simplement pointer vers l'URL du serveur du nouveau cluster dans le gestionnaire de DNS et mettre à jour le nouveau jeton de cluster et synchroniser tous les déploiements.
* Migrations de clusters simples :
  * Dans le cas des clusters Kubernetes managés (tels que EKS, AKS, GKE, etc.) qui sont spécifiques à un fournisseur de cloud, la migration de votre cluster d'un fournisseur à un autre se révélera une perte de temps et d'efforts.
  * En revanche, la migration d'une URL auto-hébergée est facile car l'URL est un domaine hébergé unique indépendant du fournisseur de cloud.
### Configurer Prometheus (Activer les métriques des applications)
Si vous voulez voir les métriques sur les applications déployées dans le cluster, Prometheus doit être déployé dans le cluster. Prometheus est un outil puissant pour fournir un aperçu graphique du comportement de votre application.
> **Note :** Assurez-vous d'avoir installé `Monitoring (Grafana)` à partir de `Devtron Stack Manager` pour configurer prometheus.
> Si vous n'installez pas `Monitoring (Grafana)`, l'option de configuration de Prometheus ne sera pas disponible.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

Activez les métriques de l'application pour configurer prometheus et renseignez les champs suivants :

|Champ|Description|
| :- | :- |
|`Endpoint Prometheus`|Indiquez l'URL de votre prometheus.|
|`Type d'authentification`|Prometheus prend en charge deux types d'authentification :<ul><li>**Basic** : Si vous sélectionnez le type d'authentification `basic`, dans ce cas vous devez indiquer le `Nom d'utilisateur` et le `Mot de passe` de prometheus pour l'authentification.</li></ul> <ul><li>**Anonymous** : Si vous sélectionnez le type d'authentification `Anonymous`, dans ce cas vous n'avez pas besoin de fournir le `Nom d'utilisateur` et le `Mot de passe`.<br>Note : Les champs `Nom d'utilisateur` et `Mot de passe` ne seront pas disponibles par défaut.</li></ul>|
|`Clé TLS `et` certificat TLS`|La `Clé TLS` et le `Certificat TLS` sont facultatifs, ces options sont utilisées lorsque vous utilisez une URL personnalisée.|

Maintenant, cliquez sur `Enregistrer le cluster` pour sauvegarder votre cluster sur Devtron.
### Installation de Devtron Agent
Lorsque vous enregistrez les configurations du cluster, votre cluster Kubernetes est mappé avec Devtron. Maintenant, Devtron agent doit être installé sur le cluster ajouté afin que vous puissiez déployer vos applications sur ce cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

Lorsque Devtron agent commence à s'installer, cliquez sur `Détails` pour vérifier l'état de l'installation.

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

Une nouvelle fenêtre s'ouvre et affiche tous les détails concernant Devtron agent.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster\_gc5.jpg)
## Ajouter un environnement
Une fois que vous avez ajouté votre cluster dans `Clusters et Environnements`, vous pouvez ajouter l'environnement en cliquant sur `Ajouter un environnement`.

Une nouvelle fenêtre d'environnement s'ouvre.

|Champ|Description|
| :- | :- |
|`Nom de l'environnement`|Saisissez le nom de votre environnement.|
|`Saisir le namespace`|Entrez un namespace correspondant à votre environnement.<br>**Note** : Si ce namespace n'existe pas déjà dans votre cluster, Devtron le créera. S'il existe déjà, Devtron fera correspondre l'environnement au namespace existant.</br>|
|`Type d'environnement`|Sélectionnez votre type d'environnement : <ul><li>`Production`</li></ul> <ul><li>`Non-production`</li></ul>Note : Devtron affiche les métriques de déploiement (métriques DORA) pour les environnements étiquetés en tant que `Production` uniquement.|

Cliquez sur `Sauvegarder` et votre environnement sera créé.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)
## Mise à jour de l'environnement
* Vous pouvez également mettre à jour un environnement en cliquant sur l'environnement.
* Vous ne pouvez modifier que les options de `Production` et `Non-Production`.
* Vous ne pouvez pas modifier le` Nom de l'environnement `ni` le Nom du namespace`.
* Assurez-vous de cliquer sur **Mettre à jour** pour mettre à jour votre environnement.
