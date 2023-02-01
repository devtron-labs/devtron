# Installer Devtron sur Minikube, Microk8s, K3s, Kind
Vous pouvez installer et essayer Devtron sur une machine haut de gamme ou sur une VM Cloud. Si vous l'installez sur un ordinateur portable/PC, il peut commencer à répondre lentement, il est donc recommandé de désinstaller Devtron de votre système avant de l'éteindre.
## Configurations système pour l'installation de Devtron
1. 2 vCPU
1. 4 Go et plus de mémoire libre
1. 20 Go et plus d'espace disque libre
## Avant de commencer
Avant de commencer et d'installer Devtron, vous devez configurer le cluster dans votre serveur et installer les pré-requis :

* Créez un cluster en utilisant [Minikube](https://minikube.sigs.k8s.io/docs/start/) ou [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) ou [K3s](https://rancher.com/docs/k3s/latest/en/installation/).
* Installez [Helm3](https://helm.sh/docs/intro/install/).
* Installez [kubectl](https://kubernetes.io/docs/tasks/tools/).
## Installer Devtron
{% tabs %}

{% tab title=" Minikube/Kind cluster" %}

Pour installer Devtron sur un cluster ``Minikube/kind``, exécutez la commande suivante :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 
~~~

{% endtab %}

{% tab title="k3s Cluster" %}
Pour installer Devtron sur un cluster ``k3s``, exécutez la commande suivante :
~~~ bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort
~~~

{% endtab %}

{% endtabs %}
## Tableau de bord Devtron
Pour accéder au tableau de bord Devtron lorsque vous utilisez ``Minikube`` comme cluster, exécutez la commande suivante :
~~~ bash
minikube service devtron-service --namespace devtroncd
~~~

Pour accéder au tableau de bord Devtron lorsque vous utilisez ``Kind/k3s`` comme cluster, exécutez la commande suivante pour transférer le service devtron vers le port 8000 :
~~~ bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
~~~

**Tableau de bord** : <http://127.0.0.1:8000>.
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
## Installer Devtron sur une VM Cloud (AWS ec2, Azure VM, GCP VM)
Il est recommandé d'utiliser une VM Cloud avec 2vCPU+, 4 Go et plus de mémoire libre, 20 Go et plus de stockage, un type de VM Compute Optimized et l'une des saveurs du système d'exploitation Ubuntu.
### Créer un cluster Microk8s
~~~ bash
sudo snap install microk8s --classic --channel=1.22
sudo usermod -a -G microk8s $USER
sudo chown -f -R $USER ~/.kube
newgrp microk8s
microk8s enable dns storage helm3
echo "alias kubectl='microk8s kubectl '" >> .bashrc
echo "alias helm='microk8s helm3 '" >> .bashrc
source .bashrc
~~~
### Installer devtron
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 
~~~
### Exécutez la commande suivante pour obtenir le numéro de port de devtron-service :
~~~ bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
~~~

Assurez-vous que le port sur lequel fonctionne le devtron-service reste ouvert dans le groupe de sécurité de la VM ou le groupe de sécurité du réseau.

**Note** : Si vous souhaitez désinstaller Devtron ou nettoyer l'installateur Helm Devtron, reportez-vous à la rubrique [Désinstaller Devtron](https://docs.devtron.ai/install/uninstall-devtron).

Si vous avez des questions, veuillez nous en faire part sur notre canal discord. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
