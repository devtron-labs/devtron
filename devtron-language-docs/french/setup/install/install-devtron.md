# Installer Devtron
Dans cette section, nous décrivons comment vous pouvez installer le tableau de bord Helm de Devtron sans aucune intégration. Les intégrations peuvent être ajoutées ultérieurement en utilisant [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations).

Si vous souhaitez installer Devtron sur Minikube, Microk8s, K3s, Kind, reportez-vous à cette [section](./Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).
## Avant de commencer
Installez [Helm](https://helm.sh/docs/intro/install/) si vous ne l'avez pas installé.
## Ajoutez Helm Repo
~~~ bash
helm repo add devtron https://helm.devtron.ai
~~~
## Installer le tableau de bord Helm de Devtron
**Note** : Cette commande d'installation n'installera pas l'intégration CI/CD. Pour CI/CD, reportez-vous à la section [Installer Devtron avec CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd).

Exécutez la commande suivante pour installer le tableau de bord Helm de Devtron :
~~~ bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd
~~~
## Installer des nœuds multi-architecture (ARM et AMD)
Pour installer Devtron sur des clusters avec des nœuds multi-architecture (ARM et AMD), ajoutez à la commande d'installation de Devtron `--set installer.arch=multi-arch`.

[//]: # (Si vous envisagez d'utiliser Hyperion pour des `déploiements de production`, veuillez vous référer à nos recommandations de remplacement pour l'[Installation Devtron]&#40;override-default-devtron-installation-configs.md&#41;.)

[//]: # (## Installation status)

[//]: # ()
[//]: # (Exécutez la commande suivante)

[//]: # ()
[//]: # (```bash)

[//]: # (kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}')

[//]: # (```)
## Tableau de bord Devtron
Exécutez la commande suivante pour obtenir l'URL du tableau de bord :
~~~ text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
~~~

Vous obtiendrez le résultat suivant :
~~~ text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

Le nom d'hôte `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` comme mentionné ci-dessus est l'URL du Loadbalancer où vous pouvez accéder au tableau de bord Devtron.
> Vous pouvez également faire une entrée CNAME correspondant à votre domaine/sous-domaine pour pointer vers cette URL du Loadbalancer afin d'y accéder à un domaine personnalisé.

|> Hôte|> Type|> Pointe vers|
| :- | :- | :- |
|> devtron.yourdomain.com|> CNAME|> aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|
>
## Identifiants Admin Devtron
### Pour la version Devtron v0.6.0 et ultérieure
**Nom d'utilisateur** : `admin` <br>
**Mot de passe** : Exécutez la commande suivante pour obtenir le mot de passe admin :
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
~~~
### Pour les versions de Devtron antérieures à la v0.6.0
**Nom d'utilisateur** : `admin` <br>
**Mot de passe** : Exécutez la commande suivante pour obtenir le mot de passe admin :
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
~~~

**Note** : Si vous souhaitez désinstaller Devtron ou nettoyer l'installateur Helm Devtron, reportez-vous à la rubrique [Désinstaller Devtron](https://docs.devtron.ai/install/uninstall-devtron).
## Mettre à jour
Pour utiliser les capacités CI/CD avec Devtron, vous pouvez installer [Devtron avec CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) ou [Devtron avec CI/CD en parallèle avec GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops).
